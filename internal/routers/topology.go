package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerTopologyRoutes(authed *gin.RouterGroup, h *allHandlers) {
	authed.GET("/topology/tree", middleware.RequirePermission("topology", "read"), h.Topology.GetTree)

	// Environments
	envs := authed.Group("/environments")
	{
		envs.GET("", middleware.RequirePermission("topology", "read"), h.Topology.ListEnvironments)
		envs.POST("", middleware.RequirePermission("topology", "create"), h.Topology.CreateEnvironment)
		envs.PUT("/:id", middleware.RequirePermission("topology", "update"), h.Topology.UpdateEnvironment)
		envs.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.Topology.DeleteEnvironment)
	}

	// DataCenters
	dcs := authed.Group("/datacenters")
	{
		dcs.GET("", middleware.RequirePermission("topology", "read"), h.Topology.ListDataCenters)
		dcs.GET("/:id", middleware.RequirePermission("topology", "read"), h.Topology.GetDataCenter)
		dcs.POST("", middleware.RequirePermission("topology", "create"), h.Topology.CreateDataCenter)
		dcs.PUT("/:id", middleware.RequirePermission("topology", "update"), h.Topology.UpdateDataCenter)
		dcs.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.Topology.DeleteDataCenter)
	}

	// Regions
	regions := authed.Group("/regions")
	{
		regions.GET("", middleware.RequirePermission("topology", "read"), h.Topology.ListRegions)
		regions.GET("/:id", middleware.RequirePermission("topology", "read"), h.Topology.GetRegion)
		regions.POST("", middleware.RequirePermission("topology", "create"), h.Topology.CreateRegion)
		regions.PUT("/:id", middleware.RequirePermission("topology", "update"), h.Topology.UpdateRegion)
		regions.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.Topology.DeleteRegion)
	}

	// Clusters
	clusters := authed.Group("/clusters")
	{
		clusters.GET("", middleware.RequirePermission("topology", "read"), h.Topology.ListClusters)
		clusters.GET("/:id", middleware.RequirePermission("topology", "read"), h.Topology.GetCluster)
		clusters.POST("", middleware.RequirePermission("topology", "create"), h.Topology.CreateCluster)
		clusters.PUT("/:id", middleware.RequirePermission("topology", "update"), h.Topology.UpdateCluster)
		clusters.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.Topology.DeleteCluster)
		// Match rules
		clusters.GET("/:id/rules", middleware.RequirePermission("topology", "read"), h.Topology.ListMatchRules)
		clusters.POST("/:id/rules", middleware.RequirePermission("topology", "create"), h.Topology.CreateMatchRule)
		clusters.PUT("/:id/rules/:rule_id", middleware.RequirePermission("topology", "update"), h.Topology.UpdateMatchRule)
		clusters.DELETE("/:id/rules/:rule_id", middleware.RequirePermission("topology", "delete"), h.Topology.DeleteMatchRule)
	}
}
