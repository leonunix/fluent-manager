package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerMetricsRoutes(authed *gin.RouterGroup, h *allHandlers) {
	metrics := authed.Group("/metrics")
	{
		metrics.GET("/overview", middleware.RequirePermission("nodes", "read"), h.Metrics.Overview)
		metrics.GET("/top-nodes", middleware.RequirePermission("nodes", "read"), h.Metrics.TopNodes)
		metrics.GET("/by-datacenter", middleware.RequirePermission("nodes", "read"), h.Metrics.ByDatacenter)
	}
}
