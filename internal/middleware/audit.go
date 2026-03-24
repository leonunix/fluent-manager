package middleware

import (
	"strconv"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

const (
	auditDetailKey       = "audit_detail"
	auditResourceTypeKey = "audit_resource_type"
	auditResourceIDKey   = "audit_resource_id"
)

func SetAuditDetail(c *gin.Context, detail string) {
	if c == nil {
		return
	}
	c.Set(auditDetailKey, detail)
}

func SetAuditResource(c *gin.Context, resourceType string, resourceID uint) {
	if c == nil {
		return
	}
	if resourceType != "" {
		c.Set(auditResourceTypeKey, resourceType)
	}
	if resourceID != 0 {
		c.Set(auditResourceIDKey, resourceID)
	}
}

// resolveAuditResource extracts a resource type and ID from the route path and params.
// For example, /api/v1/nodes/:id -> ("node", <id>).
func resolveAuditResource(c *gin.Context) (string, uint) {
	if overrideType, ok := c.Get(auditResourceTypeKey); ok {
		if resourceType, ok := overrideType.(string); ok && resourceType != "" {
			if overrideID, ok := c.Get(auditResourceIDKey); ok {
				if resourceID, ok := overrideID.(uint); ok {
					return resourceType, resourceID
				}
			}
			return resourceType, 0
		}
	}

	path := c.FullPath()
	// Map route prefixes to resource types
	resourceMap := map[string]string{
		"/api/v1/nodes":             "node",
		"/api/v1/clusters":          "cluster",
		"/api/v1/regions":           "region",
		"/api/v1/datacenters":       "datacenter",
		"/api/v1/configs":           "config",
		"/api/v1/deploys":           "deploy",
		"/api/v1/bootstrap/hosts":   "bootstrap_host",
		"/api/v1/bootstrap/tasks":   "bootstrap_task",
		"/api/v1/users":             "user",
		"/api/v1/roles":             "role",
		"/api/v1/environments":      "environment",
		"/api/v1/agent-policies":    "agent_policy",
		"/api/v1/agent-access-keys": "agent_access_key",
	}
	for prefix, resType := range resourceMap {
		if strings.HasPrefix(path, prefix) {
			// Try to get the resource ID from the :id param
			if idStr := c.Param("id"); idStr != "" {
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					return resType, uint(id)
				}
			}
			return resType, 0
		}
	}
	return "", 0
}

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

		resType, resID := resolveAuditResource(c)
		detail := c.Request.URL.String()
		if auditDetail, ok := c.Get(auditDetailKey); ok {
			if detailText, ok := auditDetail.(string); ok && strings.TrimSpace(detailText) != "" {
				detail = detailText
			}
		}

		log := models.AuditLog{
			UserID:       uid,
			Username:     uname,
			Action:       method,
			Resource:     c.FullPath(),
			ResourceType: resType,
			ResourceID:   resID,
			Detail:       detail,
			IP:           c.ClientIP(),
		}
		models.DB.Create(&log)
	}
}
