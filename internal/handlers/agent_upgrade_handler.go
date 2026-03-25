package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgentUpgradeHandler struct {
	Svc services.AgentUpgradeService
}

func (h *AgentUpgradeHandler) Create(c *gin.Context) {
	var req services.AgentUpgradeTaskInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.Svc.Create(req, c.GetUint("user_id"), middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentUpgradeError(c, err)
		return
	}
	middleware.SetAuditResource(c, "agent_upgrade_task", task.ID)
	c.JSON(http.StatusCreated, task)
}

func (h *AgentUpgradeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.Svc.List(page, pageSize, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentUpgradeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AgentUpgradeHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	task, records, err := h.Svc.Get(uint(id), middleware.GetAllowedClusters(c))
	if err != nil {
		writeAgentUpgradeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"records": records,
	})
}

func writeAgentUpgradeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "agent upgrade task not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
