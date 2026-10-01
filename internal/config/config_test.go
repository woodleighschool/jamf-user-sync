package config

import (
	"strings"
	"testing"
)

func environment(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"JAMF_HOST": "jamf.example.test", "JAMF_CLIENT_ID": "synthetic-client", "JAMF_CLIENT_SECRET": "synthetic-secret",
		"LDAP_HOST": "dc.example.test", "LDAP_USERNAME": `EXAMPLE\reader`, "LDAP_CREDENTIALS": "cGFzc3dvcmQ=",
		"LDAP_BASE_DN": "dc=woodleighschool,dc=net", "RUN_TIMEOUT": "15m", "REQUEST_TIMEOUT": "30s", "DRY_RUN": "false",
		"LOG_LEVEL": "info",
	} {
		t.Setenv(key, value)
	}
}

func TestLoadPreservesConnectionNamesAndPasswordBytes(t *testing.T) {
	environment(t)
	t.Setenv("LDAP_CREDENTIALS", "cGFzc3dvcmQK")
	t.Setenv("LOG_LEVEL", "DEBUG")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JamfHost != "https://jamf.example.test" || cfg.LDAPHost != "ldap://dc.example.test" || cfg.LDAPPassword != "password\n" || cfg.LDAPCredentials != "" {
		t.Fatal("hostname normalization or password decoding changed")
	}
	if cfg.LogLevel.String() != "DEBUG" {
		t.Fatal("LOG_LEVEL was not preserved")
	}
}

func TestLoadRejectsInvalidConfigurationWithoutLeakingValues(t *testing.T) {
	for _, tt := range []struct{ name, key, value string }{
		{"base64", "LDAP_CREDENTIALS", "invalid-secret"},
		{"empty password", "LDAP_CREDENTIALS", ""},
		{"unbounded run", "RUN_TIMEOUT", "0"},
		{"unbounded request", "REQUEST_TIMEOUT", "-1s"},
		{"plain HTTP", "JAMF_HOST", "http://jamf.example.test"},
		{"embedded credentials", "LDAP_HOST", "ldap://user:secret@dc.example.test"},
		{"invalid DN", "LDAP_BASE_DN", "invalid"},
		{"empty NTLM user", "LDAP_USERNAME", `EXAMPLE\`},
		{"invalid log level", "LOG_LEVEL", "not-a-log-level"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			environment(t)
			t.Setenv(tt.key, tt.value)
			_, err := Load()
			if err == nil {
				t.Fatal("invalid config accepted")
			}
			if tt.value != "" && strings.Contains(err.Error(), tt.value) {
				t.Fatalf("error contains rejected environment value: %s", tt.key)
			}
		})
	}
}
