package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerBootstrapRoutes(authed *gin.RouterGroup, h *allHandlers) {
	bootstrap := authed.Group("/bootstrap")
	{
		bootstrap.GET("/capability", middleware.RequirePermission("nodes", "read"), h.Bootstrap.Capability)
		bootstrap.GET("/hosts", middleware.RequirePermission("nodes", "read"), h.Bootstrap.ListHosts)
		bootstrap.POST("/hosts", middleware.RequirePermission("nodes", "create"), h.Bootstrap.CreateHost)
		bootstrap.POST("/hosts/bulk", middleware.RequirePermission("nodes", "create"), h.Bootstrap.CreateHosts)
		bootstrap.PUT("/hosts/:id", middleware.RequirePermission("nodes", "update"), h.Bootstrap.UpdateHost)
		bootstrap.DELETE("/hosts/:id", middleware.RequirePermission("nodes", "delete"), h.Bootstrap.DeleteHost)
		bootstrap.GET("/tasks", middleware.RequirePermission("nodes", "read"), h.Bootstrap.ListTasks)
		bootstrap.GET("/tasks/:id", middleware.RequirePermission("nodes", "read"), h.Bootstrap.GetTask)
		bootstrap.POST("/tasks", middleware.RequirePermission("nodes", "create"), h.Bootstrap.CreateTask)
	}
}
