package handlers

import (
	"net/http"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	Svc       services.SetupService
	JWT       *auth.JWTService
	CfgPath   string
	RestartCh chan struct{}
}

func (h *SetupHandler) GetStatus(c *gin.Context) {
	initialized, err := h.Svc.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": initialized})
}

func (h *SetupHandler) TestDB(c *gin.Context) {
	var req services.TestDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Svc.TestDBConnection(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "connection successful"})
}

func (h *SetupHandler) Initialize(c *gin.Context) {
	initialized, err := h.Svc.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if initialized {
		c.JSON(http.StatusForbidden, gin.H{"error": "system is already initialized"})
		return
	}

	var req services.SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine if we need target DB initialization (non-default DB)
	needsRestart := (req.DBDriver != "" && req.DBDriver != "sqlite") ||
		(req.DBDriver == "sqlite" && req.DBPath != "" && req.DBPath != "fluent_manager.db")
	currentJWTSecret := h.JWT.CurrentSecret()

	if needsRestart {
		// Initialize on the target database and save config.yaml
		u, err := h.Svc.InitializeTarget(req, h.CfgPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.syncJWTSecretFromConfig()

		token, err := h.JWT.GenerateToken(u.ID, u.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "initialized but failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "system initialized successfully",
			"restart": true,
			"token":   token,
			"user":    userResponse(u),
		})

		// Trigger server restart after response is sent
		go func() {
			h.RestartCh <- struct{}{}
		}()
		return
	}

	// Simple init: keep current SQLite database
	u, err := h.Svc.Initialize(req, h.CfgPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.syncJWTSecretFromConfig() != currentJWTSecret {
		needsRestart = true
	}

	token, err := h.JWT.GenerateToken(u.ID, u.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "initialized but failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "system initialized successfully",
		"restart": needsRestart,
		"token":   token,
		"user":    userResponse(u),
	})

	if needsRestart {
		go func() {
			h.RestartCh <- struct{}{}
		}()
	}
}

func userResponse(u *models.User) gin.H {
	return gin.H{
		"id":           u.ID,
		"username":     u.Username,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"auth_source":  u.AuthSource,
		"roles":        u.Roles,
	}
}

func (h *SetupHandler) syncJWTSecretFromConfig() string {
	if h == nil || h.JWT == nil {
		return ""
	}

	cfg, err := config.Load(h.CfgPath)
	if err != nil {
		return ""
	}

	secret := strings.TrimSpace(cfg.Auth.JWTSecret)
	if secret == "" {
		return ""
	}

	if secret != h.JWT.CurrentSecret() {
		h.JWT.UpdateSecret(secret)
	}
	return secret
}
