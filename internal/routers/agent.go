package routers

import (
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerAgentRoutes(api *gin.RouterGroup, cfg *config.Config, h *allHandlers) {
	agentAPI := api.Group("/agent")
	agentAPI.Use(middleware.AgentAuth(cfg.Agent.APIKey))
	{
		agentAPI.POST("/register", h.Agent.Register)
		agentAPI.POST("/heartbeat", h.Agent.Heartbeat)
		agentAPI.POST("/report", h.Agent.ReportStatus)
		agentAPI.POST("/command-result", h.Agent.ReportCommandResult)
		agentAPI.POST("/logs", h.Agent.UploadLogs)
	}
}
