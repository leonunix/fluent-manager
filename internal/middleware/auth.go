package middleware

import (
	"net/http"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		claims, err := jwtSvc.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// RequirePermission checks if the current user has the specified permission.
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		var user models.User
		if err := models.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Resource == resource && perm.Action == action {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

// AgentAuth authenticates agent nodes via API key.
func AgentAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Agent-Key")
		if key == "" || key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent key"})
			c.Abort()
			return
		}
		c.Next()
	}
}
