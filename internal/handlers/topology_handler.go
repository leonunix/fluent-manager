package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type TopologyHandler struct {
	Svc services.TopologyService
}

func (h *TopologyHandler) allowedDCIDs(c *gin.Context) []uint {
	allowed := middleware.GetAllowedClusters(c)
	return h.Svc.AllowedDCIDs(allowed)
}

// checkDCScope returns true if the user has access to the given datacenter ID.
// nil allowedDCs means global access.
func (h *TopologyHandler) checkDCScope(c *gin.Context, dcID uint) bool {
	allowedDCs := h.allowedDCIDs(c)
	if allowedDCs == nil {
		return true
	}
	for _, id := range allowedDCs {
		if id == dcID {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "datacenter not in your scope"})
	return false
}

// checkClusterScope returns true if the user has access to the given cluster ID.
func (h *TopologyHandler) checkClusterScope(c *gin.Context, clusterID uint) bool {
	allowed := middleware.GetAllowedClusters(c)
	if allowed == nil {
		return true
	}
	for _, id := range allowed {
		if id == clusterID {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "cluster not in your scope"})
	return false
}

// regionDCID resolves the datacenter ID for a region.
func (h *TopologyHandler) regionDCID(regionID uint) (uint, error) {
	region, err := h.Svc.GetRegion(regionID)
	if err != nil {
		return 0, err
	}
	return region.DataCenterID, nil
}

// ============================================================================
// DataCenter
// ============================================================================

func (h *TopologyHandler) ListDataCenters(c *gin.Context) {
	result, err := h.Svc.ListDataCenters(h.allowedDCIDs(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	dc, err := h.Svc.GetDataCenter(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "datacenter not found"})
		return
	}
	// Scope check
	if allowedDCs := h.allowedDCIDs(c); allowedDCs != nil {
		found := false
		for _, dcID := range allowedDCs {
			if dcID == dc.ID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "datacenter not in your scope"})
			return
		}
	}
	c.JSON(http.StatusOK, dc)
}

type CreateDataCenterRequest struct {
	Name        string `json:"name" binding:"required"`
	Alias       string `json:"alias"`
	Provider    string `json:"provider"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

func (h *TopologyHandler) CreateDataCenter(c *gin.Context) {
	// Scoped users cannot create new datacenters
	if allowedDCs := h.allowedDCIDs(c); allowedDCs != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "scoped users cannot create datacenters"})
		return
	}
	var req CreateDataCenterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc, err := h.Svc.CreateDataCenter(req.Name, req.Alias, req.Provider, req.Location, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "datacenter name already exists"})
		return
	}
	c.JSON(http.StatusCreated, dc)
}

func (h *TopologyHandler) UpdateDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkDCScope(c, uint(id)) {
		return
	}
	var req CreateDataCenterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc, err := h.Svc.UpdateDataCenter(uint(id), req.Name, req.Alias, req.Provider, req.Location, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "datacenter not found"})
		return
	}
	c.JSON(http.StatusOK, dc)
}

func (h *TopologyHandler) DeleteDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkDCScope(c, uint(id)) {
		return
	}
	if err := h.Svc.DeleteDataCenter(uint(id)); err != nil {
		if errors.Is(err, services.ErrHasChildren) {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete datacenter with existing regions; delete regions first"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "datacenter deleted"})
}

// ============================================================================
// Region
// ============================================================================

func (h *TopologyHandler) ListRegions(c *gin.Context) {
	result, err := h.Svc.ListRegions(h.allowedDCIDs(c), c.Query("datacenter_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	region, err := h.Svc.GetRegion(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	// Scope check: region's DC must be in allowed DCs
	if allowedDCs := h.allowedDCIDs(c); allowedDCs != nil {
		found := false
		for _, dcID := range allowedDCs {
			if dcID == region.DataCenterID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "region not in your scope"})
			return
		}
	}
	c.JSON(http.StatusOK, region)
}

type CreateRegionRequest struct {
	Name         string `json:"name" binding:"required"`
	Alias        string `json:"alias"`
	DataCenterID uint   `json:"datacenter_id" binding:"required"`
	Description  string `json:"description"`
	Tags         string `json:"tags"`
}

func (h *TopologyHandler) CreateRegion(c *gin.Context) {
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.checkDCScope(c, req.DataCenterID) {
		return
	}
	region, err := h.Svc.CreateRegion(req.Name, req.Alias, req.DataCenterID, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datacenter not found"})
		return
	}
	c.JSON(http.StatusCreated, region)
}

func (h *TopologyHandler) UpdateRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	// Check scope on current region's DC
	currentDCID, err := h.regionDCID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	if !h.checkDCScope(c, currentDCID) {
		return
	}
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Also check scope on target DC if it's being moved
	if req.DataCenterID != currentDCID {
		if !h.checkDCScope(c, req.DataCenterID) {
			return
		}
	}
	region, err := h.Svc.UpdateRegion(uint(id), req.Name, req.Alias, req.DataCenterID, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	c.JSON(http.StatusOK, region)
}

func (h *TopologyHandler) DeleteRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	// Scope check on region's DC
	dcID, err := h.regionDCID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	if !h.checkDCScope(c, dcID) {
		return
	}
	if err := h.Svc.DeleteRegion(uint(id)); err != nil {
		if errors.Is(err, services.ErrHasChildren) {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete region with existing clusters; delete clusters first"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "region deleted"})
}

// ============================================================================
// Cluster
// ============================================================================

func (h *TopologyHandler) ListClusters(c *gin.Context) {
	allowed := middleware.GetAllowedClusters(c)
	result, err := h.Svc.ListClusters(allowed, c.Query("region_id"), c.Query("environment_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	cluster, err := h.Svc.GetCluster(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	// Scope check
	if allowed := middleware.GetAllowedClusters(c); allowed != nil {
		found := false
		for _, cid := range allowed {
			if cid == cluster.ID {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusForbidden, gin.H{"error": "cluster not in your scope"})
			return
		}
	}
	c.JSON(http.StatusOK, cluster)
}

type CreateClusterRequest struct {
	Name          string `json:"name" binding:"required"`
	Alias         string `json:"alias"`
	RegionID      uint   `json:"region_id" binding:"required"`
	EnvironmentID *uint  `json:"environment_id"`
	IsDefault     bool   `json:"is_default"`
	ConfigID      *uint  `json:"config_id"`
	Description   string `json:"description"`
	Tags          string `json:"tags"`
}

func (h *TopologyHandler) CreateCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Scope check: verify user has access to the region's DC
	dcID, err := h.regionDCID(req.RegionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region not found"})
		return
	}
	if !h.checkDCScope(c, dcID) {
		return
	}
	cluster, err := h.Svc.CreateCluster(req.Name, req.Alias, req.RegionID, req.EnvironmentID, req.IsDefault, req.ConfigID, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region not found"})
		return
	}
	c.JSON(http.StatusCreated, cluster)
}

func (h *TopologyHandler) UpdateCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	// Scope check on existing cluster
	if !h.checkClusterScope(c, uint(id)) {
		return
	}
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cluster, err := h.Svc.UpdateCluster(uint(id), req.Name, req.Alias, req.RegionID, req.EnvironmentID, req.IsDefault, req.ConfigID, req.Description, req.Tags)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	c.JSON(http.StatusOK, cluster)
}

func (h *TopologyHandler) DeleteCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	// Scope check
	if !h.checkClusterScope(c, uint(id)) {
		return
	}
	if err := h.Svc.DeleteCluster(uint(id)); err != nil {
		if errors.Is(err, services.ErrHasChildren) {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete cluster with existing nodes; move or delete nodes first"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted"})
}

// ============================================================================
// Cluster Match Rules
// ============================================================================

func (h *TopologyHandler) ListMatchRules(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkClusterScope(c, uint(clusterID)) {
		return
	}
	rules, err := h.Svc.ListMatchRules(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *TopologyHandler) CreateMatchRule(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkClusterScope(c, uint(clusterID)) {
		return
	}
	var rule models.ClusterMatchRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.CreateMatchRule(uint(clusterID), &rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *TopologyHandler) UpdateMatchRule(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkClusterScope(c, uint(clusterID)) {
		return
	}
	ruleID, _ := strconv.ParseUint(c.Param("rule_id"), 10, 32)
	var req models.ClusterMatchRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.Svc.UpdateMatchRule(uint(ruleID), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *TopologyHandler) DeleteMatchRule(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if !h.checkClusterScope(c, uint(clusterID)) {
		return
	}
	ruleID, _ := strconv.ParseUint(c.Param("rule_id"), 10, 32)
	if err := h.Svc.DeleteMatchRule(uint(ruleID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
}

// ============================================================================
// User Scope Management
// ============================================================================

func (h *TopologyHandler) ListUserScopes(c *gin.Context) {
	scopes, err := h.Svc.ListUserScopes(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scopes})
}

func (h *TopologyHandler) SetUserScopes(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Scopes []services.ScopeInput `json:"scopes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Svc.SetUserScopes(uint(userID), req.Scopes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "scopes updated"})
}

// ============================================================================
// Environment
// ============================================================================

func (h *TopologyHandler) ListEnvironments(c *gin.Context) {
	envs, err := h.Svc.ListEnvironments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": envs})
}

func (h *TopologyHandler) CreateEnvironment(c *gin.Context) {
	var env models.Environment
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.Svc.CreateEnvironment(&env)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "environment name already exists"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *TopologyHandler) UpdateEnvironment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req models.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	env, err := h.Svc.UpdateEnvironment(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}
	c.JSON(http.StatusOK, env)
}

func (h *TopologyHandler) DeleteEnvironment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.Svc.DeleteEnvironment(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "environment deleted"})
}

// ============================================================================
// Topology Tree
// ============================================================================

func (h *TopologyHandler) GetTree(c *gin.Context) {
	allowedClusters := middleware.GetAllowedClusters(c)
	allowedDCIDs := h.allowedDCIDs(c)

	tree, err := h.Svc.GetTree(allowedClusters, allowedDCIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tree})
}
