package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAgentUpgradeRoutes(authed *gin.RouterGroup, h *allHandlers) {
	upgrades := authed.Group("/agent-upgrades")
	{
		upgrades.GET("", middleware.RequirePermission("nodes", "read"), h.AgentUpgrade.List)
		upgrades.GET("/:id", middleware.RequirePermission("nodes", "read"), h.AgentUpgrade.Get)
		upgrades.POST("", middleware.RequirePermission("nodes", "update"), h.AgentUpgrade.Create)
	}
}
