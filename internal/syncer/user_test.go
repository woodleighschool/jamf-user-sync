package syncer

import (
	"testing"

	"github.com/go-ldap/ldap/v3"
)

func TestDirectoryUserMapsAccountFields(t *testing.T) {
	for _, tt := range []struct {
		name, campus, department, control, buildingID, departmentID string
	}{
		{"student advances year", "Senior Campus", "Y8", "512", "1", "6"},
		{"junior campus", "Penbank", "Y0", "512", "2", "18"},
		{"staff", "Minimbah", "Staff", "512", "3", "11"},
		{"disabled overrides attributes", "Minimbah", "Y12", "514", "4", "26"},
		{"disabled low bitmask", "Penbank", "Y2", "2", "4", "26"},
		{"other account flag is enabled", "Senior Campus", "IT", "66048", "1", "21"},
		{"unknown attributes", "Other", "Other", "512", "-1", "-1"},
		{"missing optional attributes", "", "", "512", "-1", "-1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			entry := ldap.NewEntry("CN=synthetic", map[string][]string{
				"SAMACCOUNTNAME": {"Student.One"}, "NAME": {"Synthetic Student"}, "mail": {"student@example.test"},
				"CAMPUS": {tt.campus}, "department": {tt.department}, "userAccountControl": {tt.control},
			})
			got, err := directoryUser(entry)
			want := User{Username: "student.one", RealName: "Synthetic Student", Email: "student@example.test", BuildingID: tt.buildingID, DepartmentID: tt.departmentID}
			if err != nil || got != want {
				t.Fatalf("directoryUser() = %+v, %v; want %+v", got, err, want)
			}
		})
	}
}

func TestDirectoryUserRejectsMissingOrMalformedAccountControl(t *testing.T) {
	for _, value := range []string{"", "invalid", "-2", "4294967296"} {
		entry := ldap.NewEntry("CN=synthetic", map[string][]string{"sAMAccountName": {"synthetic"}, "userAccountControl": {value}})
		if _, err := directoryUser(entry); err == nil {
			t.Fatalf("control %q was accepted", value)
		}
	}
}
