package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct{}

func (h *NodeHandler) List(c *gin.Context) {
	var nodes []models.Node
	query := models.DB.Preload("Group").Preload("Config")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if groupID := c.Query("group_id"); groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	if fluentType := c.Query("fluent_type"); fluentType != "" {
		query = query.Where("fluent_type = ?", fluentType)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("hostname LIKE ? OR ip_address LIKE ? OR node_uid LIKE ?",
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
	if err := models.DB.Preload("Group").Preload("Config.Template").First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, node)
}

type UpdateNodeRequest struct {
	GroupID  *uint  `json:"group_id"`
	ConfigID *uint  `json:"config_id"`
	Labels  string `json:"labels"`
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
	if req.GroupID != nil {
		updates["group_id"] = req.GroupID
	}
	if req.ConfigID != nil {
		updates["config_id"] = req.ConfigID
	}
	if req.Labels != "" {
		updates["labels"] = req.Labels
	}

	models.DB.Model(&node).Updates(updates)
	models.DB.Preload("Group").Preload("Config").First(&node, node.ID)
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

// BatchUpdateGroup assigns multiple nodes to a group.
func (h *NodeHandler) BatchUpdateGroup(c *gin.Context) {
	var req struct {
		NodeIDs []uint `json:"node_ids" binding:"required"`
		GroupID uint   `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&models.Node{}).Where("id IN ?", req.NodeIDs).Update("group_id", req.GroupID)
	c.JSON(http.StatusOK, gin.H{"message": "nodes updated", "count": len(req.NodeIDs)})
}

// Stats returns node status summary.
func (h *NodeHandler) Stats(c *gin.Context) {
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var counts []StatusCount
	models.DB.Model(&models.Node{}).Select("status, count(*) as count").Group("status").Scan(&counts)

	var total int64
	models.DB.Model(&models.Node{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"statuses": counts,
	})
}
