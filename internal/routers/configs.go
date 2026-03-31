package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerConfigRoutes(authed *gin.RouterGroup, h *allHandlers) {
	// Config Templates & Modules
	configs := authed.Group("/configs")
	{
		configs.GET("/templates", middleware.RequirePermission("configs", "read"), h.Config.ListTemplates)
		configs.GET("/templates/:id", middleware.RequirePermission("configs", "read"), h.Config.GetTemplate)
		configs.POST("/templates", middleware.RequirePermission("configs", "create"), h.Config.CreateTemplate)
		configs.PUT("/templates/:id", middleware.RequirePermission("configs", "update"), h.Config.UpdateTemplate)
		configs.DELETE("/templates/:id", middleware.RequirePermission("configs", "delete"), h.Config.DeleteTemplate)
		configs.GET("/templates/:id/versions", middleware.RequirePermission("configs", "read"), h.Config.ListVersions)
		configs.POST("/templates/:id/versions", middleware.RequirePermission("configs", "create"), h.Config.CreateVersion)
		configs.GET("/versions/:version_id", middleware.RequirePermission("configs", "read"), h.Config.GetVersion)
		configs.GET("/modules", middleware.RequirePermission("configs", "read"), h.Config.ListModules)
		configs.GET("/modules/:id", middleware.RequirePermission("configs", "read"), h.Config.GetModule)
		configs.POST("/modules", middleware.RequirePermission("configs", "create"), h.Config.CreateModule)
		configs.POST("/modules/batch-delete", middleware.RequirePermission("configs", "delete"), h.Config.DeleteModules)
		configs.PUT("/modules/:id", middleware.RequirePermission("configs", "update"), h.Config.UpdateModule)
		configs.DELETE("/modules/:id", middleware.RequirePermission("configs", "delete"), h.Config.DeleteModule)
		configs.GET("/modules/:id/versions", middleware.RequirePermission("configs", "read"), h.Config.ListModuleVersions)
		configs.POST("/modules/:id/versions", middleware.RequirePermission("configs", "create"), h.Config.CreateModuleVersion)
		configs.POST("/rendered-configs/preview", middleware.RequirePermission("configs", "read"), h.Config.PreviewRenderedConfig)
		configs.GET("/rendered-configs/:id", middleware.RequirePermission("configs", "read"), h.Config.GetRenderedConfig)
		configs.GET("/pipelines", middleware.RequirePermission("configs", "read"), h.Config.ListPipelines)
		configs.GET("/pipelines/:id", middleware.RequirePermission("configs", "read"), h.Config.GetPipeline)
		configs.POST("/pipelines", middleware.RequirePermission("configs", "create"), h.Config.CreatePipeline)
		configs.PUT("/pipelines/:id", middleware.RequirePermission("configs", "update"), h.Config.UpdatePipeline)
		configs.DELETE("/pipelines/:id", middleware.RequirePermission("configs", "delete"), h.Config.DeletePipeline)
	}

	// Config Analysis
	analysis := authed.Group("/config-analysis")
	{
		analysis.POST("/log-sample-assistant", middleware.RequirePermission("configs", "read"), h.AI.AnalyzeLogSample)
		analysis.POST("/import-existing", middleware.RequirePermission("configs", "read"), h.FluentOps.ImportExistingConfig)
		analysis.POST("/lint", middleware.RequirePermission("configs", "read"), h.FluentOps.LintConfig)
		analysis.POST("/replay", middleware.RequirePermission("configs", "read"), h.FluentOps.ReplayConfig)
		analysis.POST("/diff", middleware.RequirePermission("configs", "read"), h.FluentOps.SemanticDiff)
		analysis.POST("/compatibility", middleware.RequirePermission("configs", "read"), h.FluentOps.CheckCompatibility)
		analysis.GET("/:id", middleware.RequirePermission("configs", "read"), h.FluentOps.GetAnalysisResult)
	}
}
