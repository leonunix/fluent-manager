package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerFluentRoutes(authed *gin.RouterGroup, h *allHandlers) {
	// Aggregation groups
	aggGroups := authed.Group("/aggregation-groups")
	{
		aggGroups.GET("", middleware.RequirePermission("topology", "read"), h.Fluent.ListAggregationGroups)
		aggGroups.GET("/deleted", middleware.RequirePermission("topology", "read"), h.Fluent.ListDeletedAggregationGroups)
		aggGroups.GET("/:id", middleware.RequirePermission("topology", "read"), h.Fluent.GetAggregationGroup)
		aggGroups.POST("", middleware.RequirePermission("topology", "create"), h.Fluent.CreateAggregationGroup)
		aggGroups.PUT("/:id", middleware.RequirePermission("topology", "update"), h.Fluent.UpdateAggregationGroup)
		aggGroups.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.Fluent.DeleteAggregationGroup)
		aggGroups.POST("/:id/restore", middleware.RequirePermission("topology", "update"), h.Fluent.RestoreAggregationGroup)
		aggGroups.GET("/:id/metrics", middleware.RequirePermission("topology", "read"), h.FluentOps.AggregationGroupMetrics)
	}

	// Log pipelines and flow graph
	outputTargets := authed.Group("/output-targets")
	{
		outputTargets.GET("", middleware.RequirePermission("topology", "read"), h.FluentOps.ListOutputTargets)
		outputTargets.GET("/:id", middleware.RequirePermission("topology", "read"), h.FluentOps.GetOutputTarget)
		outputTargets.POST("", middleware.RequirePermission("topology", "create"), h.FluentOps.CreateOutputTarget)
		outputTargets.PUT("/:id", middleware.RequirePermission("topology", "update"), h.FluentOps.UpdateOutputTarget)
		outputTargets.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.FluentOps.DeleteOutputTarget)
	}

	// Log pipelines and flow graph
	pipelines := authed.Group("/log-pipelines")
	{
		pipelines.GET("", middleware.RequirePermission("topology", "read"), h.FluentOps.ListPipelines)
		pipelines.GET("/graph", middleware.RequirePermission("topology", "read"), h.FluentOps.PipelineGraph)
		pipelines.GET("/:id", middleware.RequirePermission("topology", "read"), h.FluentOps.GetPipeline)
		pipelines.POST("", middleware.RequirePermission("topology", "create"), h.FluentOps.CreatePipeline)
		pipelines.PUT("/:id", middleware.RequirePermission("topology", "update"), h.FluentOps.UpdatePipeline)
		pipelines.DELETE("/:id", middleware.RequirePermission("topology", "delete"), h.FluentOps.DeletePipeline)
	}

	// Runtime
	runtime := authed.Group("/runtime")
	{
		runtime.GET("/drift", middleware.RequirePermission("nodes", "read"), h.FluentOps.RuntimeDrift)
		runtime.GET("/health/graph", middleware.RequirePermission("nodes", "read"), h.FluentOps.RuntimeHealthGraph)
		runtime.GET("/recommendations", middleware.RequirePermission("nodes", "read"), h.FluentOps.RuntimeRecommendations)
	}

	// Agent policies
	agentPolicies := authed.Group("/agent-policies")
	{
		agentPolicies.GET("", middleware.RequirePermission("agent_policies", "read"), h.AgentPolicy.List)
		agentPolicies.GET("/resolve/:node_id", middleware.RequirePermission("nodes", "read"), h.AgentPolicy.ResolveForNode)
		agentPolicies.GET("/:id", middleware.RequirePermission("agent_policies", "read"), h.AgentPolicy.Get)
		agentPolicies.POST("", middleware.RequirePermission("agent_policies", "create"), h.AgentPolicy.Create)
		agentPolicies.PUT("/:id", middleware.RequirePermission("agent_policies", "update"), h.AgentPolicy.Update)
		agentPolicies.DELETE("/:id", middleware.RequirePermission("agent_policies", "delete"), h.AgentPolicy.Delete)
	}
}
