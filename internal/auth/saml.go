package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"

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

func (s *SAMLAuth) GetUserFromAssertion(assertion *saml.Assertion) *SAMLUser {
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
		}
	}
	return user
}
