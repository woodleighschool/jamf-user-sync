package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/caarlos0/env/v11"
	"github.com/go-ldap/ldap/v3"
)

type Config struct {
	JamfHost        string        `env:"JAMF_HOST,required,notEmpty"`
	JamfClientID    string        `env:"JAMF_CLIENT_ID,required,notEmpty"`
	JamfSecret      string        `env:"JAMF_CLIENT_SECRET,required,notEmpty"`
	LDAPHost        string        `env:"LDAP_HOST,required,notEmpty"`
	LDAPUsername    string        `env:"LDAP_USERNAME,required,notEmpty"`
	LDAPCredentials string        `env:"LDAP_CREDENTIALS,required,notEmpty"`
	LDAPBaseDN      string        `env:"LDAP_BASE_DN" envDefault:"dc=woodleighschool,dc=net"`
	DryRun          bool          `env:"DRY_RUN" envDefault:"false"`
	RunTimeout      time.Duration `env:"RUN_TIMEOUT" envDefault:"15m"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	LogLevel        slog.Level    `env:"LOG_LEVEL" envDefault:"info"`
	LDAPPassword    string
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		// Parser errors can contain environment values, including credentials.
		return Config{}, errors.New("config: required environment is missing or invalid")
	}
	if cfg.RunTimeout <= 0 || cfg.RequestTimeout <= 0 {
		return Config{}, errors.New("config: RUN_TIMEOUT and REQUEST_TIMEOUT must be positive")
	}
	var err error
	cfg.JamfHost, err = serverURL(cfg.JamfHost, "https", "https")
	if err != nil {
		return Config{}, fmt.Errorf("config: JAMF_HOST: %w", err)
	}
	cfg.LDAPHost, err = serverURL(cfg.LDAPHost, "ldap", "ldap", "ldaps")
	if err != nil {
		return Config{}, fmt.Errorf("config: LDAP_HOST: %w", err)
	}
	if _, err := ldap.ParseDN(cfg.LDAPBaseDN); err != nil || strings.TrimSpace(cfg.LDAPBaseDN) == "" {
		return Config{}, errors.New("config: LDAP_BASE_DN must be a distinguished name")
	}
	password, err := base64.StdEncoding.DecodeString(cfg.LDAPCredentials)
	if err != nil || len(password) == 0 || !utf8.Valid(password) {
		return Config{}, errors.New("config: LDAP_CREDENTIALS must encode a nonempty UTF-8 password in base64")
	}
	cfg.LDAPPassword = string(password)
	cfg.LDAPCredentials = ""
	if domain, user, ntlm := strings.Cut(cfg.LDAPUsername, `\`); ntlm && (domain == "" || user == "" || strings.Contains(user, `\`)) {
		return Config{}, errors.New(`config: LDAP_USERNAME must contain DOMAIN\user or a simple-bind principal`)
	}
	return cfg, nil
}

func serverURL(value, defaultScheme string, schemes ...string) (string, error) {
	if !strings.Contains(value, "://") {
		value = defaultScheme + "://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", errors.New("must be a server hostname or URL without credentials, path, query or fragment")
	}
	if slices.Contains(schemes, parsed.Scheme) {
		return strings.TrimRight(value, "/"), nil
	}
	return "", errors.New("unsupported URL scheme")
}
