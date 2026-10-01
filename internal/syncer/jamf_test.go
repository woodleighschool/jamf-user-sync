package syncer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/woodleighschool/jamf-user-sync/internal/config"
)

func tokenResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"synthetic-token","expires_in":3600}`))
}

func testJamf(t *testing.T, handler http.HandlerFunc) *Jamf {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	jamf, err := NewJamf(config.Config{JamfHost: server.URL, JamfClientID: "synthetic-client", JamfSecret: "synthetic-secret", RequestTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := jamf.Close(); err != nil {
			t.Error(err)
		}
	})
	return jamf
}

func TestJamfSDKEnumeratesAllManagedComputerAndIOSPages(t *testing.T) {
	pages := map[string][]int{}
	jamf := testJamf(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/oauth/token" {
			if r.Method != http.MethodPost {
				t.Errorf("token method = %s", r.Method)
			}
			tokenResponse(w)
			return
		}
		query := r.URL.Query()
		if query.Get("section") != "GENERAL,USER_AND_LOCATION" {
			t.Errorf("section = %q", query.Get("section"))
		}
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil {
			t.Error(err)
		}
		pageSize, err := strconv.Atoi(query.Get("page-size"))
		if err != nil || pageSize != 200 {
			t.Errorf("SDK page size = %d, %v", pageSize, err)
		}
		pages[r.URL.Path] = append(pages[r.URL.Path], page)
		var total int
		switch r.URL.Path {
		case "/api/v4/computers-inventory":
			total = 201
			if query.Get("filter") != "general.remoteManagement.managed==true" || query.Get("sort") != "id:asc" {
				t.Error("computer managed filter or stable sort missing")
			}
		case "/api/v2/mobile-devices/detail":
			total = 202
			if query.Get("filter") != "managed==true" || query.Get("sort") != "mobileDeviceId:asc" {
				t.Error("mobile managed filter or stable sort missing")
			}
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		results := make([]map[string]any, 0)
		for i := page * pageSize; i < min((page+1)*pageSize, total); i++ {
			if total == 201 {
				results = append(results, map[string]any{"id": strconv.Itoa(i), "general": map[string]any{"remoteManagement": map[string]bool{"managed": true}},
					"userAndLocation": map[string]string{"username": "student", "realname": "Synthetic Student", "email": "student@example.test", "buildingId": "1", "departmentId": "5"}})
			} else {
				kind := "ios"
				if i == 201 {
					kind = "appleTv"
				}
				results = append(results, map[string]any{"mobileDeviceId": strconv.Itoa(i), "deviceType": kind, "general": map[string]bool{"managed": true},
					"userAndLocation": map[string]string{"username": "student", "realName": "Synthetic Student", "emailAddress": "student@example.test", "buildingId": "2", "departmentId": "6"}})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"totalCount": total, "results": results})
	})
	devices, err := jamf.Devices(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 402 || devices[200].ID != "200" || devices[401].ID != "200" {
		t.Fatalf("inventory did not include all pages: %d devices", len(devices))
	}
	for path, got := range pages {
		if len(got) != 2 || got[0] != 0 || got[1] != 1 {
			t.Errorf("%s pages = %v, want [0 1]", path, got)
		}
	}
	if devices[0].User.RealName != "Synthetic Student" || devices[201].User.Email != "student@example.test" {
		t.Error("provider field names were not mapped")
	}
}

func TestJamfPatchOwnsOnlyUserFieldsAndClearsEmptyValues(t *testing.T) {
	for _, kind := range []string{"macos", "ios"} {
		t.Run(kind, func(t *testing.T) {
			patches := 0
			jamf := testJamf(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/oauth/token" {
					tokenResponse(w)
					return
				}
				patches++
				expectedPath, section, name, email := "/api/v4/computers-inventory-detail/12", "userAndLocation", "realname", "email"
				if kind == "ios" {
					expectedPath, section, name, email = "/api/v2/mobile-devices/12", "location", "realName", "emailAddress"
				}
				if r.Method != http.MethodPatch || r.URL.Path != expectedPath || r.Header.Get("Authorization") != "Bearer synthetic-token" {
					t.Errorf("unexpected patch request: %s %s", r.Method, r.URL.Path)
				}
				var body map[string]map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				fields := body[section]
				if len(body) != 1 || len(fields) != 5 || fields["username"] != "student" || fields["buildingId"] != "-1" || fields["departmentId"] != "26" {
					t.Errorf("patch escaped ownership: %+v", body)
				}
				for _, field := range []string{name, email} {
					if value, present := fields[field]; !present || value != "" {
						t.Errorf("%s clearing omitted", field)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{}`)
			})
			if err := jamf.Update(t.Context(), Device{ID: "12", Kind: kind}, User{Username: "student", BuildingID: "-1", DepartmentID: "26"}); err != nil {
				t.Fatal(err)
			}
			if patches != 1 {
				t.Errorf("patches = %d", patches)
			}
		})
	}
}
