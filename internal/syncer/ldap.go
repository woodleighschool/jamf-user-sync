package syncer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/woodleighschool/jamf-user-sync/internal/config"
)

// ErrUserNotFound means the assigned Jamf username has no directory account.
var ErrUserNotFound = errors.New("directory: account not found")

type LDAP struct {
	conn       *ldap.Conn
	baseDN     string
	timeout    time.Duration
	stopCancel func() bool
}

func NewLDAP(ctx context.Context, cfg config.Config) (*LDAP, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}
	deadline, _ := ctx.Deadline()
	conn, err := ldap.DialURL(cfg.LDAPHost, ldap.DialWithDialer(&net.Dialer{Timeout: cfg.RequestTimeout, Deadline: deadline}))
	if err != nil {
		return nil, errors.New("directory: connection failed")
	}
	conn.SetTimeout(cfg.RequestTimeout)
	directory := &LDAP{conn: conn, baseDN: cfg.LDAPBaseDN, timeout: cfg.RequestTimeout}
	directory.stopCancel = context.AfterFunc(ctx, func() { _ = conn.Close() })
	if domain, username, ntlm := strings.Cut(cfg.LDAPUsername, `\`); ntlm {
		err = conn.NTLMBind(domain, username, cfg.LDAPPassword)
	} else {
		err = conn.Bind(cfg.LDAPUsername, cfg.LDAPPassword)
	}
	if err != nil {
		directory.Close()
		if ctx.Err() != nil {
			return nil, fmt.Errorf("directory: %w", ctx.Err())
		}
		return nil, errors.New("directory: bind failed")
	}
	return directory, nil
}

func (d *LDAP) Close() {
	d.stopCancel()
	_ = d.conn.Close()
}

func (d *LDAP) Lookup(ctx context.Context, username string) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	request := ldap.NewSearchRequest(d.baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 2, 0, false,
		"(&(objectClass=user)(sAMAccountName="+ldap.EscapeFilter(username)+"))",
		[]string{"name", "mail", "sAMAccountName", "userAccountControl", "department", "Campus"}, nil)
	response := d.conn.SearchAsync(ctx, request, 2)
	var entries []*ldap.Entry
	for response.Next() {
		if entry := response.Entry(); entry != nil {
			entries = append(entries, entry)
		}
	}
	if err := ctx.Err(); err != nil {
		return User{}, fmt.Errorf("directory search: %w", err)
	}
	if response.Err() != nil {
		return User{}, errors.New("directory: account search failed")
	}
	return uniqueUser(entries)
}

func uniqueUser(entries []*ldap.Entry) (User, error) {
	if len(entries) == 0 {
		return User{}, ErrUserNotFound
	}
	if len(entries) != 1 {
		return User{}, errors.New("directory: account search must return exactly one user")
	}
	return directoryUser(entries[0])
}
