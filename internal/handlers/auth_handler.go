package handlers

import (
	"net/http"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	JWT  *auth.JWTService
	LDAP *auth.LDAPAuth
	SAML *auth.SAMLAuth
}

type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	AuthSource string `json:"auth_source"` // local, ldap
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

	var user models.User

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
		// Find or create local user from LDAP
		result := models.DB.Where("username = ? AND auth_source = ?", ldapUser.Username, "ldap").First(&user)
		if result.RowsAffected == 0 {
			user = models.User{
				Username:    ldapUser.Username,
				Email:       ldapUser.Email,
				DisplayName: ldapUser.DisplayName,
				AuthSource:  "ldap",
				IsActive:    true,
			}
			models.DB.Create(&user)
			// Assign default viewer role
			var viewerRole models.Role
			models.DB.Where("name = ?", "viewer").First(&viewerRole)
			models.DB.Model(&user).Association("Roles").Append(&viewerRole)
		}

	default: // local
		result := models.DB.Where("username = ? AND auth_source = ?", req.Username, "local").First(&user)
		if result.RowsAffected == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}

	token, err := h.JWT.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	now := time.Now()
	models.DB.Model(&user).Update("last_login_at", &now)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"auth_source":  user.AuthSource,
		},
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user models.User
	if err := models.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
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

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.AuthSource != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password change only available for local accounts"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	models.DB.Model(&user).Update("password_hash", string(hash))
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
