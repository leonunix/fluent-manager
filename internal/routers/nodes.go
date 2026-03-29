package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerNodeRoutes(authed *gin.RouterGroup, h *allHandlers) {
	nodes := authed.Group("/nodes")
	{
		nodes.GET("", middleware.RequirePermission("nodes", "read"), h.Node.List)
		nodes.GET("/stats", middleware.RequirePermission("nodes", "read"), h.Node.Stats)
		nodes.GET("/:id", middleware.RequirePermission("nodes", "read"), h.Node.Get)
		nodes.PUT("/:id", middleware.RequirePermission("nodes", "update"), h.Node.Update)
		nodes.DELETE("/:id", middleware.RequirePermission("nodes", "delete"), h.Node.Delete)
		nodes.POST("/batch-move", middleware.RequirePermission("nodes", "update"), h.Node.BatchMoveCluster)
		// Node metrics, logs, remote commands
		nodes.GET("/:id/metrics", middleware.RequirePermission("nodes", "read"), h.Agent.GetNodeMetrics)
		nodes.GET("/:id/logs", middleware.RequirePermission("nodes", "read"), h.Agent.GetNodeLogs)
		nodes.POST("/:id/commands", middleware.RequirePermission("nodes", "update"), h.Agent.SendCommand)
		nodes.GET("/:id/commands", middleware.RequirePermission("nodes", "read"), h.Agent.ListNodeCommands)
		nodes.GET("/:id/commands/:cmdID", middleware.RequirePermission("nodes", "read"), h.Agent.GetNodeCommand)
		nodes.GET("/:id/fluent-profile", middleware.RequirePermission("nodes", "read"), h.Fluent.GetNodeProfile)
		nodes.PUT("/:id/fluent-profile", middleware.RequirePermission("nodes", "update"), h.Fluent.UpdateNodeProfile)
	}
}
