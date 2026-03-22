package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type DeployHandler struct{}

type CreateDeployRequest struct {
	ConfigVersionID uint   `json:"config_version_id" binding:"required"`
	NodeIDs         []uint `json:"node_ids"`       // specific nodes
	ClusterID       *uint  `json:"cluster_id"`     // all nodes in a cluster
	RegionID        *uint  `json:"region_id"`      // all nodes in a region
	DataCenterID    *uint  `json:"datacenter_id"`  // all nodes in a DC
	EnvironmentID   *uint  `json:"environment_id"` // all nodes in an environment
}

func (h *DeployHandler) Create(c *gin.Context) {
	var req CreateDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var configVersion models.ConfigVersion
	if err := models.DB.First(&configVersion, req.ConfigVersionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config version not found"})
		return
	}

	// Determine scope and target nodes
	scope := "node"
	scopeID := uint(0)
	var nodeIDs []uint

	if req.DataCenterID != nil {
		scope = "datacenter"
		scopeID = *req.DataCenterID
		var nodes []models.Node
		models.DB.Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", *req.DataCenterID).Find(&nodes)
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}
	if req.RegionID != nil {
		scope = "region"
		scopeID = *req.RegionID
		var nodes []models.Node
		models.DB.Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Where("clusters.region_id = ?", *req.RegionID).Find(&nodes)
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}
	if req.ClusterID != nil {
		scope = "cluster"
		scopeID = *req.ClusterID
		var nodes []models.Node
		models.DB.Where("cluster_id = ?", *req.ClusterID).Find(&nodes)
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID)
		}
		// Also set cluster-level config
		models.DB.Model(&models.Cluster{}).Where("id = ?", *req.ClusterID).Update("config_id", req.ConfigVersionID)
	}
	if req.EnvironmentID != nil {
		scope = "environment"
		scopeID = *req.EnvironmentID
		var nodes []models.Node
		models.DB.Where("environment_id = ? OR cluster_id IN (SELECT id FROM clusters WHERE environment_id = ?)",
			*req.EnvironmentID, *req.EnvironmentID).Find(&nodes)
		for _, n := range nodes {
			nodeIDs = append(nodeIDs, n.ID)
		}
	}
	nodeIDs = append(nodeIDs, req.NodeIDs...)

	if len(nodeIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no target nodes found"})
		return
	}

	// Deduplicate
	seen := map[uint]bool{}
	var uniqueIDs []uint
	for _, id := range nodeIDs {
		if !seen[id] {
			seen[id] = true
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	userID := c.GetUint("user_id")
	task := models.DeployTask{
		ConfigID:   req.ConfigVersionID,
		Scope:      scope,
		ScopeID:    scopeID,
		Status:     "pending",
		TotalNodes: len(uniqueIDs),
		CreatedBy:  userID,
	}
	models.DB.Create(&task)

	for _, nodeID := range uniqueIDs {
		models.DB.Create(&models.DeployRecord{
			DeployTaskID: task.ID,
			NodeID:       nodeID,
			Status:       "pending",
		})
	}

	// Set node-level config
	models.DB.Model(&models.Node{}).Where("id IN ?", uniqueIDs).Update("config_id", req.ConfigVersionID)
	models.DB.Model(&task).Update("status", "running")

	c.JSON(http.StatusCreated, task)
}

func (h *DeployHandler) List(c *gin.Context) {
	var tasks []models.DeployTask
	query := models.DB.Preload("Config.Template").Preload("Creator")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Model(&models.DeployTask{}).Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks)

	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *DeployHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var task models.DeployTask
	if err := models.DB.Preload("Config.Template").Preload("Creator").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deploy task not found"})
		return
	}

	var records []models.DeployRecord
	models.DB.Where("deploy_task_id = ?", id).Preload("Node").Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"records": records,
	})
}

func (h *DeployHandler) GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var total int64
	models.DB.Model(&models.AuditLog{}).Count(&total)
	models.DB.Order("created_at DESC").
		Offset((page-1)*pageSize).Limit(pageSize).
		Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
