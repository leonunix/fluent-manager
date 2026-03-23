package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerGroupRoutes(authed *gin.RouterGroup, h *allHandlers) {
	// Groups
	groups := authed.Group("/groups")
	{
		groups.GET("", middleware.RequirePermission("groups", "read"), h.Group.List)
		groups.GET("/:id", middleware.RequirePermission("groups", "read"), h.Group.Get)
		groups.POST("", middleware.RequirePermission("groups", "create"), h.Group.Create)
		groups.PUT("/:id", middleware.RequirePermission("groups", "update"), h.Group.Update)
		groups.DELETE("/:id", middleware.RequirePermission("groups", "delete"), h.Group.Delete)
		groups.PUT("/:id/users", middleware.RequirePermission("groups", "update"), h.Group.SetUsers)
	}

	// Auth Settings
	authSettings := authed.Group("/auth-settings")
	{
		authSettings.GET("/ldap", middleware.RequirePermission("auth_settings", "read"), h.AuthSettings.GetLDAPSettings)
		authSettings.PUT("/ldap", middleware.RequirePermission("auth_settings", "update"), h.AuthSettings.UpdateLDAPSettings)
		authSettings.POST("/ldap/test", middleware.RequirePermission("auth_settings", "update"), h.AuthSettings.TestLDAPConnection)
		authSettings.GET("/saml", middleware.RequirePermission("auth_settings", "read"), h.AuthSettings.GetSAMLSettings)
		authSettings.PUT("/saml", middleware.RequirePermission("auth_settings", "update"), h.AuthSettings.UpdateSAMLSettings)
		authSettings.GET("/group-mappings", middleware.RequirePermission("auth_settings", "read"), h.AuthSettings.ListGroupMappings)
		authSettings.PUT("/group-mappings", middleware.RequirePermission("auth_settings", "update"), h.AuthSettings.SetGroupMappings)
	}
}
