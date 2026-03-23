package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	Svc services.NodeService
}

// checkNodeScope verifies the node belongs to the user's allowed clusters.
// Returns the node and true if access is allowed, or writes an HTTP error and returns false.
func checkNodeScope(c *gin.Context, node *models.Node) bool {
	allowed := middleware.GetAllowedClusters(c)
	if allowed == nil {
		return true // global access
	}
	if node.ClusterID != nil {
		for _, cid := range allowed {
			if cid == *node.ClusterID {
				return true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "node not in your scope"})
	return false
}

// checkClusterInScope verifies a cluster ID is within the user's allowed clusters.
func checkClusterInScope(c *gin.Context, clusterID uint) bool {
	allowed := middleware.GetAllowedClusters(c)
	if allowed == nil {
		return true
	}
	for _, cid := range allowed {
		if cid == clusterID {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "target cluster not in your scope"})
	return false
}

func (h *NodeHandler) List(c *gin.Context) {
	filters := services.NodeListFilters{
		Status:        c.Query("status"),
		ClusterID:     c.Query("cluster_id"),
		EnvironmentID: c.Query("environment_id"),
		FluentType:    c.Query("fluent_type"),
		DataCenterID:  c.Query("datacenter_id"),
		RegionID:      c.Query("region_id"),
		Search:        c.Query("search"),
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	allowed := middleware.GetAllowedClusters(c)
	nodes, total, err := h.Svc.List(filters, allowed, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      nodes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *NodeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	node, err := h.Svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
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
	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Scope check: verify user can access the current node
	node, err := h.Svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	// If moving to a different cluster, verify target cluster is also in scope
	if req.ClusterID != nil {
		if !checkClusterInScope(c, *req.ClusterID) {
			return
		}
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

	updated, err := h.Svc.Update(uint(id), updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *NodeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Scope check
	node, err := h.Svc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	if err := h.Svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}

func (h *NodeHandler) BatchMoveCluster(c *gin.Context) {
	var req struct {
		NodeIDs   []uint `json:"node_ids" binding:"required"`
		ClusterID uint   `json:"cluster_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Scope check: verify target cluster is in scope
	if !checkClusterInScope(c, req.ClusterID) {
		return
	}

	// Scope check: verify all source nodes are in scope
	for _, nodeID := range req.NodeIDs {
		node, err := h.Svc.Get(nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}
		if !checkNodeScope(c, node) {
			return
		}
	}

	if err := h.Svc.BatchMoveCluster(req.NodeIDs, req.ClusterID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "nodes moved", "count": len(req.NodeIDs)})
}

func (h *NodeHandler) Stats(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	counts, total, err := h.Svc.Stats(allowed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"statuses": counts,
	})
}
