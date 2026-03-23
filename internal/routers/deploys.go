package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerDeployRoutes(authed *gin.RouterGroup, h *allHandlers) {
	deploys := authed.Group("/deploys")
	{
		deploys.GET("", middleware.RequirePermission("configs", "read"), h.Deploy.List)
		deploys.GET("/:id", middleware.RequirePermission("configs", "read"), h.Deploy.Get)
		deploys.POST("", middleware.RequirePermission("configs", "deploy"), h.Deploy.Create)
	}

	// Audit Logs
	authed.GET("/audit-logs", middleware.RequirePermission("audit", "read"), h.Deploy.GetAuditLogs)
}
