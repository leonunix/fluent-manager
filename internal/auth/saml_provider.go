package auth

import (
	"net/http"
	"sync"

	"github.com/fluent-manager/fluent-manager/internal/config"
)

// SAMLProvider holds a thread-safe, hot-swappable SAMLAuth instance.
// Router registers it once at startup; the underlying SP can be replaced at runtime
// when settings are changed from the frontend.
type SAMLProvider struct {
	mu   sync.RWMutex
	auth *SAMLAuth
}

func NewSAMLProvider(initial *SAMLAuth) *SAMLProvider {
	return &SAMLProvider{auth: initial}
}

// Get returns the current SAMLAuth (may be nil or have nil SP).
func (p *SAMLProvider) Get() *SAMLAuth {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.auth
}

// Reload builds a new SAMLAuth from the given config and swaps it in.
func (p *SAMLProvider) Reload(cfg config.SAMLConfig) error {
	if !cfg.Enabled {
		p.mu.Lock()
		p.auth = &SAMLAuth{cfg: cfg}
		p.mu.Unlock()
		return nil
	}
	newAuth, err := NewSAMLAuth(cfg)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.auth = newAuth
	p.mu.Unlock()
	return nil
}

// ServeHTTP delegates to the current SP middleware, if configured.
// This allows the provider to be registered as an http.Handler at startup.
func (p *SAMLProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	auth := p.auth
	p.mu.RUnlock()
	if auth == nil || auth.SP == nil {
		http.Error(w, "SAML not configured", http.StatusNotFound)
		return
	}
	auth.SP.ServeHTTP(w, r)
}
