package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct{}

func (h *NodeHandler) List(c *gin.Context) {
	var nodes []models.Node
	query := models.DB.Preload("Cluster.Region.DataCenter").Preload("Cluster.Environment").Preload("Environment").Preload("Config.Template")

	// Scope filtering
	if allowed := middleware.GetAllowedClusters(c); allowed != nil {
		query = query.Where("nodes.cluster_id IN ?", allowed)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("nodes.status = ?", status)
	}
	if clusterID := c.Query("cluster_id"); clusterID != "" {
		query = query.Where("nodes.cluster_id = ?", clusterID)
	}
	if envID := c.Query("environment_id"); envID != "" {
		query = query.Where("nodes.environment_id = ? OR nodes.cluster_id IN (SELECT id FROM clusters WHERE environment_id = ?)", envID, envID)
	}
	if fluentType := c.Query("fluent_type"); fluentType != "" {
		query = query.Where("nodes.fluent_type = ?", fluentType)
	}
	if dcID := c.Query("datacenter_id"); dcID != "" {
		query = query.Joins("JOIN clusters c2 ON c2.id = nodes.cluster_id").
			Joins("JOIN regions r2 ON r2.id = c2.region_id").
			Where("r2.data_center_id = ?", dcID)
	}
	if regionID := c.Query("region_id"); regionID != "" {
		query = query.Joins("JOIN clusters c3 ON c3.id = nodes.cluster_id").
			Where("c3.region_id = ?", regionID)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("nodes.hostname LIKE ? OR nodes.ip_address LIKE ? OR nodes.node_uid LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Model(&models.Node{}).Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&nodes)

	c.JSON(http.StatusOK, gin.H{
		"data":      nodes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *NodeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var node models.Node
	if err := models.DB.Preload("Cluster.Region.DataCenter").Preload("Cluster.Environment").Preload("Environment").Preload("Config.Template").First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	// Scope check
	if allowed := middleware.GetAllowedClusters(c); allowed != nil && node.ClusterID != nil {
		found := false
		for _, cid := range allowed {
			if cid == *node.ClusterID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "node not in your scope"})
			return
		}
	}
	c.JSON(http.StatusOK, node)
}

type UpdateNodeRequest struct {
	ClusterID     *uint  `json:"cluster_id"`
	EnvironmentID *uint  `json:"environment_id"`
	ConfigID      *uint  `json:"config_id"`
	Labels        string `json:"labels"`
}

func (h *NodeHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var node models.Node
	if err := models.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.ClusterID != nil {
		updates["cluster_id"] = req.ClusterID
	}
	if req.EnvironmentID != nil {
		updates["environment_id"] = req.EnvironmentID
	}
	if req.ConfigID != nil {
		updates["config_id"] = req.ConfigID
	}
	if req.Labels != "" {
		updates["labels"] = req.Labels
	}

	models.DB.Model(&node).Updates(updates)
	models.DB.Preload("Cluster.Region.DataCenter").Preload("Environment").Preload("Config").First(&node, node.ID)
	c.JSON(http.StatusOK, node)
}

func (h *NodeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var node models.Node
	if err := models.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	models.DB.Delete(&node)
	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}

// BatchMoveCluster assigns multiple nodes to a cluster.
func (h *NodeHandler) BatchMoveCluster(c *gin.Context) {
	var req struct {
		NodeIDs   []uint `json:"node_ids" binding:"required"`
		ClusterID uint   `json:"cluster_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Model(&models.Node{}).Where("id IN ?", req.NodeIDs).Update("cluster_id", req.ClusterID)
	c.JSON(http.StatusOK, gin.H{"message": "nodes moved", "count": len(req.NodeIDs)})
}

// Stats returns node status summary (scope-aware).
func (h *NodeHandler) Stats(c *gin.Context) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	baseQuery := models.DB.Model(&models.Node{})
	if allowed := middleware.GetAllowedClusters(c); allowed != nil {
		baseQuery = baseQuery.Where("cluster_id IN ?", allowed)
	}

	var counts []StatusCount
	baseQuery.Select("status, count(*) as count").Group("status").Scan(&counts)

	var total int64
	baseQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"statuses": counts,
	})
}
