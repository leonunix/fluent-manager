package handlers

import (
	"errors"
	"net/http"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FluentHandler struct {
	Svc     services.FluentService
	NodeSvc services.NodeService
}

func (h *FluentHandler) ListAggregationGroups(c *gin.Context) {
	groups, err := h.Svc.ListAggregationGroups(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func (h *FluentHandler) ListDeletedAggregationGroups(c *gin.Context) {
	groups, err := h.Svc.ListDeletedAggregationGroups(middleware.GetAllowedClusters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func (h *FluentHandler) GetAggregationGroup(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	group, err := h.Svc.GetAggregationGroup(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAggregationGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *FluentHandler) CreateAggregationGroup(c *gin.Context) {
	var req aggregationGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := req.toInput()
	if input.ClusterID == nil && middleware.GetAllowedClusters(c) != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "scoped users must create aggregation groups inside an allowed cluster"})
		return
	}
	if input.ClusterID != nil && !checkClusterInScope(c, *input.ClusterID) {
		return
	}

	group, err := h.Svc.CreateAggregationGroup(input)
	if err != nil {
		writeAggregationGroupError(c, err)
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *FluentHandler) UpdateAggregationGroup(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.Svc.GetAggregationGroup(id, middleware.GetAllowedClusters(c)); err != nil {
		writeAggregationGroupError(c, err)
		return
	}

	var req aggregationGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := req.toInput()
	if input.ClusterID == nil && middleware.GetAllowedClusters(c) != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "scoped users cannot move aggregation groups outside their scope"})
		return
	}
	if input.ClusterID != nil && !checkClusterInScope(c, *input.ClusterID) {
		return
	}

	group, err := h.Svc.UpdateAggregationGroup(id, input)
	if err != nil {
		writeAggregationGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *FluentHandler) DeleteAggregationGroup(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.Svc.GetAggregationGroup(id, middleware.GetAllowedClusters(c)); err != nil {
		writeAggregationGroupError(c, err)
		return
	}

	if err := h.Svc.DeleteAggregationGroup(id); err != nil {
		writeAggregationGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "aggregation group deleted"})
}

func (h *FluentHandler) RestoreAggregationGroup(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	group, err := h.Svc.RestoreAggregationGroup(id, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAggregationGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *FluentHandler) GetNodeProfile(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	node, err := h.NodeSvc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}
	profile, err := h.Svc.GetNodeProfile(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "fluent profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *FluentHandler) UpdateNodeProfile(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	node, err := h.NodeSvc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	var req services.NodeFluentProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AggregationGroupID != nil && !checkNodeAggregationGroupScope(c, h.Svc, *req.AggregationGroupID) {
		return
	}

	profile, err := h.Svc.UpsertNodeProfile(id, &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidArgument) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func checkNodeAggregationGroupScope(c *gin.Context, svc services.FluentService, groupID uint) bool {
	group, err := svc.GetAggregationGroup(groupID, middleware.GetAllowedClusters(c))
	if err != nil {
		writeAggregationGroupError(c, err)
		return false
	}
	if group.ClusterID != nil {
		return checkClusterInScope(c, *group.ClusterID)
	}
	return true
}

type aggregationGroupRequest struct {
	Name         string  `json:"name" binding:"required"`
	Alias        string  `json:"alias"`
	Description  string  `json:"description"`
	FluentType   string  `json:"fluent_type"`
	Mode         string  `json:"mode"`
	EndpointHost string  `json:"endpoint_host"`
	EndpointPort int     `json:"endpoint_port"`
	EnableTLS    bool    `json:"enable_tls"`
	SharedKey    *string `json:"shared_key"`
	ClusterID    *uint   `json:"cluster_id"`
}

func (r aggregationGroupRequest) toInput() *services.AggregationGroupInput {
	return &services.AggregationGroupInput{
		Name:         r.Name,
		Alias:        r.Alias,
		Description:  r.Description,
		FluentType:   r.FluentType,
		Mode:         r.Mode,
		EndpointHost: r.EndpointHost,
		EndpointPort: r.EndpointPort,
		EnableTLS:    r.EnableTLS,
		SharedKey:    r.SharedKey,
		ClusterID:    r.ClusterID,
	}
}

func writeAggregationGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "aggregation group not in your scope"})
	case errors.Is(err, services.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrHasChildren):
		c.JSON(http.StatusConflict, gin.H{"error": "aggregation group still has assigned nodes"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "aggregation group not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
