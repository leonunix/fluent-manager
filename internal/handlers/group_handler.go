package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct{}

func (h *GroupHandler) List(c *gin.Context) {
	var groups []models.NodeGroup
	models.DB.Find(&groups)

	// Attach node count per group
	type GroupWithCount struct {
		models.NodeGroup
		NodeCount int64 `json:"node_count"`
	}
	var result []GroupWithCount
	for _, g := range groups {
		var count int64
		models.DB.Model(&models.Node{}).Where("group_id = ?", g.ID).Count(&count)
		result = append(result, GroupWithCount{NodeGroup: g, NodeCount: count})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *GroupHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var group models.NodeGroup
	if err := models.DB.Preload("Nodes").First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.NodeGroup{Name: req.Name, Description: req.Description}
	if err := models.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "group name already exists"})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *GroupHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var group models.NodeGroup
	if err := models.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&group).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	})
	c.JSON(http.StatusOK, group)
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var group models.NodeGroup
	if err := models.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	// Unassign nodes from this group
	models.DB.Model(&models.Node{}).Where("group_id = ?", id).Update("group_id", nil)
	models.DB.Delete(&group)
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}
