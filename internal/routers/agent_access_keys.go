package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAgentAccessKeyRoutes(authed *gin.RouterGroup, h *allHandlers) {
	keys := authed.Group("/agent-access-keys")
	{
		keys.GET("", middleware.RequirePermission("agent_keys", "read"), h.AgentAccessKey.List)
		keys.POST("", middleware.RequirePermission("agent_keys", "create"), h.AgentAccessKey.Create)
		keys.PUT("/:id", middleware.RequirePermission("agent_keys", "update"), h.AgentAccessKey.Update)
		keys.DELETE("/:id", middleware.RequirePermission("agent_keys", "delete"), h.AgentAccessKey.Delete)
	}
}
