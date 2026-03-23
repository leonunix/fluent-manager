package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/fluent-manager/fluent-manager/internal/config"
)

type SAMLAuth struct {
	SP  *samlsp.Middleware
	cfg config.SAMLConfig
}

type SAMLUser struct {
	Username    string
	Email       string
	DisplayName string
	Groups      []string
}

func NewSAMLAuth(cfg config.SAMLConfig) (*SAMLAuth, error) {
	if !cfg.Enabled {
		return &SAMLAuth{cfg: cfg}, nil
	}

	keyPair, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load SAML certificate: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse SAML certificate: %w", err)
	}

	idpMetadataURL, err := url.Parse(cfg.IDPMetadata)
	if err != nil {
		return nil, fmt.Errorf("invalid IDP metadata URL: %w", err)
	}

	rootURL, err := url.Parse(cfg.EntityID)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID URL: %w", err)
	}

	// If ACS URL is explicitly set, derive the root URL from it so that
	// the samlsp middleware computes ACS/metadata/SLO paths correctly
	// (e.g. when behind a reverse proxy with a different external hostname).
	if cfg.ACSURL != "" {
		if acsURL, parseErr := url.Parse(cfg.ACSURL); parseErr == nil {
			acsURL.Path = strings.TrimSuffix(acsURL.Path, "/saml/acs")
			acsURL.Path = strings.TrimSuffix(acsURL.Path, "/")
			rootURL = acsURL
		}
	}

	idpMetadata, err := samlsp.FetchMetadata(
		context.Background(),
		http.DefaultClient,
		*idpMetadataURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IDP metadata: %w", err)
	}

	sp, err := samlsp.New(samlsp.Options{
		URL:               *rootURL,
		EntityID:          cfg.EntityID,
		Key:               keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:       keyPair.Leaf,
		IDPMetadata:       idpMetadata,
		AllowIDPInitiated: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML SP: %w", err)
	}

	return &SAMLAuth{SP: sp, cfg: cfg}, nil
}

// GetUserFromAttributes extracts user info from samlsp session attributes.
// This is the preferred method when using the samlsp middleware session flow.
func (s *SAMLAuth) GetUserFromAttributes(attrs samlsp.Attributes, groupAttribute string) *SAMLUser {
	if groupAttribute == "" {
		groupAttribute = "memberOf"
	}
	user := &SAMLUser{}

	attrMap := map[string]func(string){
		"uid":       func(v string) { user.Username = v },
		"username":  func(v string) { user.Username = v },
		"urn:oid:0.9.2342.19200300.100.1.1": func(v string) { user.Username = v },
		"email":     func(v string) { user.Email = v },
		"mail":      func(v string) { user.Email = v },
		"urn:oid:0.9.2342.19200300.100.1.3": func(v string) { user.Email = v },
		"displayName": func(v string) { user.DisplayName = v },
		"cn":         func(v string) { user.DisplayName = v },
		"urn:oid:2.16.840.1.113730.3.1.241": func(v string) { user.DisplayName = v },
	}

	for name, setter := range attrMap {
		if vals, ok := attrs[name]; ok && len(vals) > 0 && vals[0] != "" {
			setter(vals[0])
		}
	}

	// Extract groups from well-known attribute names plus the configured one
	groupNames := []string{groupAttribute, "memberOf", "groups", "http://schemas.xmlsoap.org/claims/Group"}
	seen := map[string]bool{}
	for _, gn := range groupNames {
		if vals, ok := attrs[gn]; ok {
			for _, v := range vals {
				if v != "" && !seen[v] {
					user.Groups = append(user.Groups, v)
					seen[v] = true
				}
			}
		}
	}

	return user
}

// GetUserFromAssertion extracts user info from a SAML assertion.
// groupAttribute specifies the assertion attribute name for groups (e.g. "memberOf", "groups").
func (s *SAMLAuth) GetUserFromAssertion(assertion *saml.Assertion, groupAttribute string) *SAMLUser {
	if groupAttribute == "" {
		groupAttribute = "memberOf"
	}
	user := &SAMLUser{}
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			val := ""
			if len(attr.Values) > 0 {
				val = attr.Values[0].Value
			}
			switch attr.Name {
			case "uid", "username", "urn:oid:0.9.2342.19200300.100.1.1":
				user.Username = val
			case "email", "mail", "urn:oid:0.9.2342.19200300.100.1.3":
				user.Email = val
			case "displayName", "cn", "urn:oid:2.16.840.1.113730.3.1.241":
				user.DisplayName = val
			}
			// Extract groups
			if attr.Name == groupAttribute || attr.Name == "memberOf" || attr.Name == "groups" ||
				attr.Name == "http://schemas.xmlsoap.org/claims/Group" {
				for _, v := range attr.Values {
					if v.Value != "" {
						user.Groups = append(user.Groups, v.Value)
					}
				}
			}
		}
	}
	return user
}
