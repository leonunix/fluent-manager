package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAIRoutes(authed *gin.RouterGroup, h *allHandlers) {
	ai := authed.Group("/ai-settings")
	{
		ai.GET("", middleware.RequirePermission("ai_settings", "read"), h.AI.GetSettings)
		ai.PUT("", middleware.RequirePermission("ai_settings", "update"), h.AI.UpdateSettings)
		ai.POST("/log-sample-analysis", middleware.RequirePermission("ai_settings", "read"), h.AI.AnalyzeLogSample)
	}
}
