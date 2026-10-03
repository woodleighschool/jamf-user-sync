package syncer

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
)

func TestUniqueUserDistinguishesMissingAndAmbiguousAccounts(t *testing.T) {
	entry := ldap.NewEntry("cn=student,dc=example,dc=test", map[string][]string{"sAMAccountName": {"student"}, "userAccountControl": {"512"}})
	for _, tc := range []struct {
		name    string
		entries []*ldap.Entry
		missing bool
		failed  bool
	}{
		{name: "missing", missing: true, failed: true},
		{name: "unique", entries: []*ldap.Entry{entry}},
		{name: "ambiguous", entries: []*ldap.Entry{entry, entry}, failed: true},
		{name: "malformed", entries: []*ldap.Entry{ldap.NewEntry("cn=broken", nil)}, failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user, err := uniqueUser(tc.entries)
			if (err != nil) != tc.failed || errors.Is(err, ErrUserNotFound) != tc.missing {
				t.Fatalf("user=%+v err=%v", user, err)
			}
			if !tc.failed && user.Username != "student" {
				t.Fatalf("username=%q", user.Username)
			}
		})
	}
}
