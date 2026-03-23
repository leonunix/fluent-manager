package handlers

import (
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	JWT  *auth.JWTService
	LDAP *auth.LDAPAuth
	SAML *auth.SAMLAuth
	Svc  services.AuthService
}

type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	AuthSource string `json:"auth_source"`
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
		if h.LDAP == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "LDAP authentication not configured"})
			return
		}
		ldapUser, err := h.LDAP.Authenticate(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "LDAP authentication failed"})
			return
		}
		user, err := h.Svc.FindOrCreateLDAPUser(ldapUser.Username, ldapUser.Email, ldapUser.DisplayName)
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
