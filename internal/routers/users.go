package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerUserRoutes(authed *gin.RouterGroup, h *allHandlers) {
	// Profile
	authed.GET("/auth/profile", h.Auth.GetProfile)
	authed.PUT("/auth/password", h.Auth.ChangePassword)

	// Users
	users := authed.Group("/users")
	{
		users.GET("", middleware.RequirePermission("users", "read"), h.User.List)
		users.GET("/:id", middleware.RequirePermission("users", "read"), h.User.Get)
		users.POST("", middleware.RequirePermission("users", "create"), h.User.Create)
		users.PUT("/:id", middleware.RequirePermission("users", "update"), h.User.Update)
		users.DELETE("/:id", middleware.RequirePermission("users", "delete"), h.User.Delete)
	}

	// User Scopes (admin only)
	authed.GET("/users/:id/scopes", middleware.RequirePermission("users", "read"), h.Topology.ListUserScopes)
	authed.PUT("/users/:id/scopes", middleware.RequirePermission("users", "update"), h.Topology.SetUserScopes)

	// Roles & Permissions
	roles := authed.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission("roles", "read"), h.Role.List)
		roles.GET("/:id", middleware.RequirePermission("roles", "read"), h.Role.Get)
		roles.POST("", middleware.RequirePermission("roles", "create"), h.Role.Create)
		roles.PUT("/:id", middleware.RequirePermission("roles", "update"), h.Role.Update)
		roles.DELETE("/:id", middleware.RequirePermission("roles", "delete"), h.Role.Delete)
	}
	authed.GET("/permissions", middleware.RequirePermission("roles", "read"), h.Role.ListPermissions)
}
