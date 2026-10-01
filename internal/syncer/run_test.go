package syncer

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

type fakeInventory struct {
	devices []Device
	failIDs map[string]bool
	updates []Device
	err     error
}

func (f *fakeInventory) Devices(context.Context) ([]Device, error) { return f.devices, f.err }
func (f *fakeInventory) Update(_ context.Context, device Device, user User) error {
	if f.failIDs[device.ID] {
		return errors.New("synthetic update failure")
	}
	device.User = user
	f.updates = append(f.updates, device)
	return nil
}

type fakeDirectory map[string]User

func (f fakeDirectory) Lookup(_ context.Context, username string) (User, error) {
	user, ok := f[username]
	if !ok {
		return User{}, errors.New("synthetic missing account")
	}
	return user, nil
}

func TestRunReconcilesEveryOwnedField(t *testing.T) {
	want := User{Username: "student", RealName: "Synthetic Student", Email: "student@example.test", BuildingID: "1", DepartmentID: "6"}
	devices := []Device{{ID: "same", Kind: "macos", User: want}}
	for i, user := range []User{
		{Username: "Student", RealName: want.RealName, Email: want.Email, BuildingID: "1", DepartmentID: "6"},
		{Username: "student", RealName: "Old Name", Email: want.Email, BuildingID: "1", DepartmentID: "6"},
		{Username: "student", RealName: want.RealName, Email: "old@example.test", BuildingID: "1", DepartmentID: "6"},
		{Username: "student", RealName: want.RealName, Email: want.Email, BuildingID: "2", DepartmentID: "6"},
		{Username: "student", RealName: want.RealName, Email: want.Email, BuildingID: "1", DepartmentID: "5"},
	} {
		devices = append(devices, Device{ID: string(rune('a' + i)), Kind: "ios", User: user})
	}
	inventory := &fakeInventory{devices: devices}
	summary, err := Run(t.Context(), inventory, fakeDirectory{"student": want, "Student": want}, false, slog.New(slog.DiscardHandler))
	if err != nil || summary.Unchanged != 1 || summary.Updated != 5 || len(inventory.updates) != 5 {
		t.Fatalf("Run() = %+v, %v; updates %d", summary, err, len(inventory.updates))
	}
	for _, update := range inventory.updates {
		if update.User != want {
			t.Errorf("updated user = %+v, want %+v", update.User, want)
		}
	}
}

func TestRunDryRunNeverWrites(t *testing.T) {
	inventory := &fakeInventory{devices: []Device{{ID: "1", Kind: "macos", User: User{Username: "student", DepartmentID: "5"}}, {ID: "2", Kind: "ios"}}}
	summary, err := Run(t.Context(), inventory, fakeDirectory{"student": {Username: "student", DepartmentID: "6"}}, true, slog.New(slog.DiscardHandler))
	if err != nil || summary.WouldUpdate != 1 || summary.Unassigned != 1 || summary.Updated != 0 || len(inventory.updates) != 0 {
		t.Fatalf("Run() = %+v, %v; updates %d", summary, err, len(inventory.updates))
	}
}

func TestRunAggregatesFailuresAndContinues(t *testing.T) {
	inventory := &fakeInventory{devices: []Device{
		{ID: "1", User: User{Username: "missing"}}, {ID: "2", User: User{Username: "bad"}}, {ID: "3", User: User{Username: "good"}},
	}, failIDs: map[string]bool{"2": true}}
	var diagnostics bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&diagnostics, nil))
	summary, err := Run(t.Context(), inventory, fakeDirectory{"bad": {Username: "bad", DepartmentID: "6"}, "good": {Username: "good", DepartmentID: "6"}}, false, logger)
	if err == nil || summary.LookupFailed != 1 || summary.UpdateFailed != 1 || summary.Updated != 1 || len(inventory.updates) != 1 || inventory.updates[0].ID != "3" {
		t.Fatalf("Run() = %+v, %v; updates %+v", summary, err, inventory.updates)
	}
	for _, field := range []string{`"device_id":"1"`, `"device_id":"2"`, `"operation":"directory_lookup"`, `"operation":"jamf_update"`} {
		if !strings.Contains(diagnostics.String(), field) {
			t.Errorf("missing failure diagnostic %s", field)
		}
	}
}

func TestRunStopsOnInventoryFailureOrCancellation(t *testing.T) {
	inventory := &fakeInventory{err: errors.New("synthetic inventory failure")}
	logger := slog.New(slog.DiscardHandler)
	if _, err := Run(t.Context(), inventory, fakeDirectory{}, false, logger); err == nil {
		t.Fatal("inventory failure was ignored")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	inventory = &fakeInventory{devices: []Device{{ID: "1", User: User{Username: "student"}}}}
	if _, err := Run(ctx, inventory, fakeDirectory{}, false, logger); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() cancellation = %v", err)
	}
}
