package auth

import (
	"crypto/tls"
	"fmt"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/go-ldap/ldap/v3"
)

type LDAPAuth struct {
	cfg config.LDAPConfig
}

func NewLDAPAuth(cfg config.LDAPConfig) *LDAPAuth {
	return &LDAPAuth{cfg: cfg}
}

type LDAPUser struct {
	Username    string
	Email       string
	DisplayName string
	Groups      []string
}

func (l *LDAPAuth) Authenticate(username, password string) (*LDAPUser, error) {
	addr := fmt.Sprintf("%s:%d", l.cfg.Host, l.cfg.Port)

	var conn *ldap.Conn
	var err error

	if l.cfg.UseTLS {
		conn, err = ldap.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: false})
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}
	if err != nil {
		return nil, fmt.Errorf("ldap connect failed: %w", err)
	}
	defer conn.Close()

	// Bind with service account to search
	if err := conn.Bind(l.cfg.BindDN, l.cfg.BindPassword); err != nil {
		return nil, fmt.Errorf("ldap bind failed: %w", err)
	}

	// Search for user
	userFilter := fmt.Sprintf(l.cfg.UserFilter, ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		l.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
		userFilter,
		[]string{l.cfg.Attributes.Username, l.cfg.Attributes.Email, l.cfg.Attributes.Name, "memberOf"},
		nil,
	)

	result, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("ldap search failed: %w", err)
	}
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found in LDAP")
	}

	entry := result.Entries[0]

	// Bind as user to verify password
	if err := conn.Bind(entry.DN, password); err != nil {
		return nil, fmt.Errorf("ldap authentication failed: %w", err)
	}

	user := &LDAPUser{
		Username:    entry.GetAttributeValue(l.cfg.Attributes.Username),
		Email:       entry.GetAttributeValue(l.cfg.Attributes.Email),
		DisplayName: entry.GetAttributeValue(l.cfg.Attributes.Name),
		Groups:      entry.GetAttributeValues("memberOf"),
	}

	return user, nil
}
