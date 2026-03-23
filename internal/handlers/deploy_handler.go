package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type DeployHandler struct {
	Svc services.DeployService
}

type CreateDeployRequest struct {
	ConfigVersionID uint   `json:"config_version_id" binding:"required"`
	NodeIDs         []uint `json:"node_ids"`
	ClusterID       *uint  `json:"cluster_id"`
	RegionID        *uint  `json:"region_id"`
	DataCenterID    *uint  `json:"datacenter_id"`
	EnvironmentID   *uint  `json:"environment_id"`
}

func (h *DeployHandler) Create(c *gin.Context) {
	var req CreateDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	allowed := middleware.GetAllowedClusters(c)
	task, err := h.Svc.Create(req.ConfigVersionID, req.NodeIDs, req.ClusterID, req.RegionID, req.DataCenterID, req.EnvironmentID, userID, allowed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *DeployHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	allowed := middleware.GetAllowedClusters(c)
	tasks, total, err := h.Svc.List(page, pageSize, allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *DeployHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	allowed := middleware.GetAllowedClusters(c)
	task, records, err := h.Svc.Get(uint(id), allowed)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deploy task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"records": records,
	})
}

func (h *DeployHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	allowed := middleware.GetAllowedClusters(c)
	logs, total, err := h.Svc.GetAuditLogs(page, pageSize, allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
