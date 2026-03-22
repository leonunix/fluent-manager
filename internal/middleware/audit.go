package middleware

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only log mutating requests
		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			return
		}

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(uint)
		uname, _ := username.(string)

		log := models.AuditLog{
			UserID:   uid,
			Username: uname,
			Action:   method,
			Resource: c.FullPath(),
			Detail:   c.Request.URL.String(),
			IP:       c.ClientIP(),
		}
		models.DB.Create(&log)
	}
}
