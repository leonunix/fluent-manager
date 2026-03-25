package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAgentArtifactRoutes(authed *gin.RouterGroup, h *allHandlers) {
	artifacts := authed.Group("/agent-artifacts")
	{
		artifacts.GET("", middleware.RequirePermission("nodes", "read"), h.AgentArtifact.List)
		artifacts.POST("", middleware.RequirePermission("nodes", "update"), h.AgentArtifact.Create)
		artifacts.DELETE("/:id", middleware.RequirePermission("nodes", "update"), h.AgentArtifact.Delete)
	}
}
