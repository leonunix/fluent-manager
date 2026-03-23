package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/cache"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

// SAML one-time authorization code infrastructure.
// Uses Redis when available (multi-instance safe), falls back to in-process sync.Map.

const samlCodeTTL = 60 * time.Second
const samlCodePrefix = "saml_code:"

var localCodeStore sync.Map // fallback for single-instance

type samlCodeEntry struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func generateSAMLCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func storeSAMLCode(code, token string) {
	entry := samlCodeEntry{Token: token, Expires: time.Now().Add(samlCodeTTL)}
	if cache.Enabled() {
		cache.SetWithTTL(samlCodePrefix+code, entry, samlCodeTTL)
	} else {
		localCodeStore.Store(code, entry)
	}
}

func consumeSAMLCode(code string) (string, bool) {
	if cache.Enabled() {
		var entry samlCodeEntry
		if !cache.GetAndDelete(samlCodePrefix+code, &entry) {
			return "", false
		}
		if time.Now().After(entry.Expires) {
			return "", false
		}
		return entry.Token, true
	}
	// Fallback: in-process store
	val, ok := localCodeStore.LoadAndDelete(code)
	if !ok {
		return "", false
	}
	entry := val.(samlCodeEntry)
	if time.Now().After(entry.Expires) {
		return "", false
	}
	return entry.Token, true
}

type AuthHandler struct {
	JWT             *auth.JWTService
	LDAP            *auth.LDAPAuth     // from config.yaml (startup fallback)
	SAMLProvider    *auth.SAMLProvider  // hot-swappable SAML provider
	Svc             services.AuthService
	AuthSettingsSvc services.AuthSettingsService
}

type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	AuthSource string `json:"auth_source"`
}

// getLDAPAuth returns an LDAPAuth built from DB settings.
// Falls back to config.yaml instance only when no DB settings exist at all.
// If DB settings exist but enabled=false, returns nil (LDAP is explicitly disabled).
func (h *AuthHandler) getLDAPAuth() *auth.LDAPAuth {
	if h.AuthSettingsSvc == nil {
		return h.LDAP
	}
	dto, err := h.AuthSettingsSvc.GetLDAPSettings()
	if err != nil {
		return h.LDAP
	}
	// DB settings exist — respect the enabled flag
	if !dto.Enabled {
		return nil
	}
	if dto.Host == "" {
		return h.LDAP
	}
	cfg := config.LDAPConfig{
		Enabled:      dto.Enabled,
		Host:         dto.Host,
		Port:         dto.Port,
		UseTLS:       dto.UseTLS,
		BindDN:       dto.BindDN,
		BindPassword: dto.BindPassword,
		BaseDN:       dto.BaseDN,
		UserFilter:   dto.UserFilter,
		GroupFilter:  dto.GroupFilter,
	}
	cfg.Attributes.Username = dto.Attributes.Username
	cfg.Attributes.Email = dto.Attributes.Email
	cfg.Attributes.Name = dto.Attributes.Name
	return auth.NewLDAPAuth(cfg)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AuthSource == "" {
		req.AuthSource = "local"
	}

	var userID uint
	var username, email, displayName, authSource string

	switch req.AuthSource {
	case "ldap":
		ldapAuth := h.getLDAPAuth()
		if ldapAuth == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "LDAP authentication not configured"})
			return
		}
		ldapUser, err := ldapAuth.Authenticate(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "LDAP authentication failed"})
			return
		}
		user, err := h.Svc.FindOrCreateExternalUser(ldapUser.Username, ldapUser.Email, ldapUser.DisplayName, "ldap", ldapUser.Groups)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userID = user.ID
		username = user.Username
		email = user.Email
		displayName = user.DisplayName
		authSource = user.AuthSource

		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
			return
		}

	case "saml":
		// SAML login via password is not supported — SAML uses browser redirect flow.
		// This branch handles the case where the frontend posts auth_source=saml after
		// the SAML callback has already set a session. For now, return an error guiding
		// the user to use the SAML redirect endpoint.
		c.JSON(http.StatusBadRequest, gin.H{"error": "SAML login uses browser redirect via /saml/login, not password-based login"})
		return

	default:
		user, err := h.Svc.LocalLogin(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
			return
		}

		userID = user.ID
		username = user.Username
		email = user.Email
		displayName = user.DisplayName
		authSource = user.AuthSource
	}

	token, err := h.JWT.GenerateToken(userID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.Svc.UpdateLastLogin(userID)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":           userID,
			"username":     username,
			"email":        email,
			"display_name": displayName,
			"auth_source":  authSource,
		},
	})
}

// SAMLCallback handles the post-authentication SAML callback.
// The samlsp middleware has already validated the assertion and established a session.
// This endpoint extracts attributes from that session, creates/syncs the user, and
// redirects to the frontend with a JWT token.
func (h *AuthHandler) SAMLCallback(c *gin.Context) {
	samlAuth := h.SAMLProvider.Get()
	if samlAuth == nil || samlAuth.SP == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SAML not configured"})
		return
	}

	session, err := samlAuth.SP.Session.GetSession(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid SAML session"})
		return
	}

	sa, ok := session.(samlsp.SessionWithAttributes)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected SAML session type"})
		return
	}

	// Determine group attribute from DB settings
	groupAttribute := "memberOf"
	if h.AuthSettingsSvc != nil {
		if dto, err := h.AuthSettingsSvc.GetSAMLSettings(); err == nil && dto.GroupAttribute != "" {
			groupAttribute = dto.GroupAttribute
		}
	}

	samlUser := samlAuth.GetUserFromAttributes(sa.GetAttributes(), groupAttribute)
	if samlUser.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SAML assertion missing username"})
		return
	}

	user, err := h.Svc.FindOrCreateExternalUser(samlUser.Username, samlUser.Email, samlUser.DisplayName, "saml", samlUser.Groups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	token, errTok := h.JWT.GenerateToken(user.ID, user.Username)
	if errTok != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.Svc.UpdateLastLogin(user.ID)

	// Store token as a one-time code to avoid exposing JWT in URL
	code := generateSAMLCode()
	storeSAMLCode(code, token)

	c.Redirect(http.StatusFound, "/?saml_code="+code)
}

// ExchangeSAMLCode exchanges a one-time SAML authorization code for a JWT token.
func (h *AuthHandler) ExchangeSAMLCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, ok := consumeSAMLCode(req.Code)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := h.Svc.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Svc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
