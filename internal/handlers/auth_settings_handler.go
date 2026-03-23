package handlers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
)

type AuthSettingsHandler struct {
	Svc          services.AuthSettingsService
	SAMLProvider *auth.SAMLProvider
}

const maskedSecretValue = "********"

// dtoToSAMLConfig converts a SAMLSettingsDTO to a SAMLConfig, routing
// cert/key data to PEM fields or file path fields based on content.
func dtoToSAMLConfig(dto services.SAMLSettingsDTO) config.SAMLConfig {
	cfg := config.SAMLConfig{
		Enabled:        dto.Enabled,
		IDPMetadata:    dto.IDPMetadata,
		EntityID:       dto.EntityID,
		ACSURL:         dto.ACSURL,
		GroupAttribute: dto.GroupAttribute,
	}
	if strings.HasPrefix(strings.TrimSpace(dto.CertData), "-----BEGIN") {
		cfg.CertPEM = dto.CertData
	} else {
		cfg.CertFile = dto.CertData
	}
	if strings.HasPrefix(strings.TrimSpace(dto.KeyData), "-----BEGIN") {
		cfg.KeyPEM = dto.KeyData
	} else {
		cfg.KeyFile = dto.KeyData
	}
	return cfg
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	return maskedSecretValue
}

func preserveMaskedSecret(incoming, existing string) string {
	if incoming == maskedSecretValue {
		return existing
	}
	return incoming
}

// GetEnabledMethods returns which auth methods (ldap, saml) are enabled.
// This is a public endpoint used by the login page.
// For SAML, checks that the runtime provider actually has a working SP,
// not just the DB flag, so the login page won't show a dead button.
func (h *AuthSettingsHandler) GetEnabledMethods(c *gin.Context) {
	methods := []string{"local"}
	if ldap, err := h.Svc.GetLDAPSettings(); err == nil && ldap.Enabled {
		methods = append(methods, "ldap")
	}
	if saml, err := h.Svc.GetSAMLSettings(); err == nil && saml.Enabled {
		// Only advertise SAML if the runtime provider is actually loaded
		if h.SAMLProvider != nil {
			if sa := h.SAMLProvider.Get(); sa != nil && sa.SP != nil {
				methods = append(methods, "saml")
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"methods": methods})
}

func (h *AuthSettingsHandler) GetLDAPSettings(c *gin.Context) {
	settings, err := h.Svc.GetLDAPSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Mask bind password in response
	if settings.BindPassword != "" {
		settings.BindPassword = maskedSecretValue
	}
	c.JSON(http.StatusOK, settings)
}

func (h *AuthSettingsHandler) UpdateLDAPSettings(c *gin.Context) {
	var dto services.LDAPSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If password is masked, keep the existing one
	if dto.BindPassword == maskedSecretValue {
		existing, _ := h.Svc.GetLDAPSettings()
		if existing != nil {
			dto.BindPassword = existing.BindPassword
		}
	}

	if err := h.Svc.UpdateLDAPSettings(dto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "LDAP settings updated"})
}

func (h *AuthSettingsHandler) TestLDAPConnection(c *gin.Context) {
	var dto services.LDAPSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If password is masked, use existing
	if dto.BindPassword == maskedSecretValue {
		existing, _ := h.Svc.GetLDAPSettings()
		if existing != nil {
			dto.BindPassword = existing.BindPassword
		}
	}

	addr := fmt.Sprintf("%s:%d", dto.Host, dto.Port)
	var conn *ldap.Conn
	var err error

	if dto.UseTLS {
		conn, err = ldap.DialTLS("tcp", addr, &tls.Config{InsecureSkipVerify: false})
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": fmt.Sprintf("connection failed: %v", err)})
		return
	}
	defer conn.Close()

	if err := conn.Bind(dto.BindDN, dto.BindPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": fmt.Sprintf("bind failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "LDAP connection successful"})
}

func (h *AuthSettingsHandler) GetSAMLSettings(c *gin.Context) {
	settings, err := h.Svc.GetSAMLSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *AuthSettingsHandler) UpdateSAMLSettings(c *gin.Context) {
	var dto services.SAMLSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Svc.UpdateSAMLSettings(dto); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hot-reload the SAML middleware with the new settings
	if h.SAMLProvider != nil {
		cfg := dtoToSAMLConfig(dto)
		if err := h.SAMLProvider.Reload(cfg); err != nil {
			log.Printf("WARNING: SAML reload failed: %v", err)
			c.JSON(http.StatusOK, gin.H{"message": "SAML settings saved but reload failed: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "SAML settings updated"})
}

func (h *AuthSettingsHandler) ListGroupMappings(c *gin.Context) {
	source := c.DefaultQuery("source", "ldap")
	mappings, err := h.Svc.ListGroupMappings(source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mappings})
}

type SetGroupMappingsRequest struct {
	Source   string                       `json:"source" binding:"required"`
	Mappings []services.GroupMappingInput `json:"mappings"`
}

func (h *AuthSettingsHandler) SetGroupMappings(c *gin.Context) {
	var req SetGroupMappingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Svc.SetGroupMappings(req.Source, req.Mappings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group mappings updated"})
}
