package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/middleware"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type TopologyHandler struct{}

// scopeFilterDCs filters datacenter list by user's allowed clusters (returns full list for admins).
func scopeFilterDCs(c *gin.Context) []uint {
	allowed := middleware.GetAllowedClusters(c)
	if allowed == nil {
		return nil // global
	}
	// Find which DCs contain allowed clusters
	dcSet := map[uint]bool{}
	var clusters []models.Cluster
	models.DB.Where("id IN ?", allowed).Preload("Region").Find(&clusters)
	for _, cl := range clusters {
		if cl.Region != nil {
			dcSet[cl.Region.DataCenterID] = true
		}
	}
	ids := make([]uint, 0, len(dcSet))
	for id := range dcSet {
		ids = append(ids, id)
	}
	return ids
}

// ============================================================================
// DataCenter
// ============================================================================

func (h *TopologyHandler) ListDataCenters(c *gin.Context) {
	var dcs []models.DataCenter
	query := models.DB.Preload("Regions")
	if dcIDs := scopeFilterDCs(c); dcIDs != nil {
		query = query.Where("id IN ?", dcIDs)
	}
	query.Find(&dcs)

	type DCWithCounts struct {
		models.DataCenter
		RegionCount  int64 `json:"region_count"`
		ClusterCount int64 `json:"cluster_count"`
		NodeCount    int64 `json:"node_count"`
	}
	var result []DCWithCounts
	for _, dc := range dcs {
		var regionCount, clusterCount, nodeCount int64
		models.DB.Model(&models.Region{}).Where("data_center_id = ?", dc.ID).Count(&regionCount)
		models.DB.Model(&models.Cluster{}).
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", dc.ID).Count(&clusterCount)
		models.DB.Model(&models.Node{}).
			Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", dc.ID).Count(&nodeCount)
		result = append(result, DCWithCounts{DataCenter: dc, RegionCount: regionCount, ClusterCount: clusterCount, NodeCount: nodeCount})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var dc models.DataCenter
	if err := models.DB.Preload("Regions.Clusters").First(&dc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "datacenter not found"})
		return
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
	var req CreateDataCenterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dc := models.DataCenter{
		Name: req.Name, Alias: req.Alias, Provider: req.Provider,
		Location: req.Location, Description: req.Description, Tags: req.Tags,
	}
	if err := models.DB.Create(&dc).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "datacenter name already exists"})
		return
	}
	c.JSON(http.StatusCreated, dc)
}

func (h *TopologyHandler) UpdateDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var dc models.DataCenter
	if err := models.DB.First(&dc, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "datacenter not found"})
		return
	}
	var req CreateDataCenterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Model(&dc).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "provider": req.Provider,
		"location": req.Location, "description": req.Description, "tags": req.Tags,
	})
	c.JSON(http.StatusOK, dc)
}

func (h *TopologyHandler) DeleteDataCenter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var regionCount int64
	models.DB.Model(&models.Region{}).Where("data_center_id = ?", id).Count(&regionCount)
	if regionCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete datacenter with existing regions; delete regions first"})
		return
	}
	models.DB.Delete(&models.DataCenter{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "datacenter deleted"})
}

// ============================================================================
// Region
// ============================================================================

func (h *TopologyHandler) ListRegions(c *gin.Context) {
	var regions []models.Region
	query := models.DB.Preload("DataCenter")
	if dcIDs := scopeFilterDCs(c); dcIDs != nil {
		query = query.Where("data_center_id IN ?", dcIDs)
	}
	if dcID := c.Query("datacenter_id"); dcID != "" {
		query = query.Where("data_center_id = ?", dcID)
	}
	query.Find(&regions)

	type RegionWithCounts struct {
		models.Region
		ClusterCount int64 `json:"cluster_count"`
		NodeCount    int64 `json:"node_count"`
	}
	var result []RegionWithCounts
	for _, r := range regions {
		var clusterCount, nodeCount int64
		models.DB.Model(&models.Cluster{}).Where("region_id = ?", r.ID).Count(&clusterCount)
		models.DB.Model(&models.Node{}).
			Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Where("clusters.region_id = ?", r.ID).Count(&nodeCount)
		result = append(result, RegionWithCounts{Region: r, ClusterCount: clusterCount, NodeCount: nodeCount})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var region models.Region
	if err := models.DB.Preload("DataCenter").Preload("Clusters.Environment").First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
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
	var dc models.DataCenter
	if err := models.DB.First(&dc, req.DataCenterID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datacenter not found"})
		return
	}
	region := models.Region{
		Name: req.Name, Alias: req.Alias, DataCenterID: req.DataCenterID,
		Description: req.Description, Tags: req.Tags,
	}
	models.DB.Create(&region)
	c.JSON(http.StatusCreated, region)
}

func (h *TopologyHandler) UpdateRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var region models.Region
	if err := models.DB.First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
		return
	}
	var req CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Model(&region).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "datacenter_id": req.DataCenterID,
		"description": req.Description, "tags": req.Tags,
	})
	c.JSON(http.StatusOK, region)
}

func (h *TopologyHandler) DeleteRegion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var clusterCount int64
	models.DB.Model(&models.Cluster{}).Where("region_id = ?", id).Count(&clusterCount)
	if clusterCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete region with existing clusters; delete clusters first"})
		return
	}
	models.DB.Delete(&models.Region{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "region deleted"})
}

// ============================================================================
// Cluster
// ============================================================================

func (h *TopologyHandler) ListClusters(c *gin.Context) {
	var clusters []models.Cluster
	query := models.DB.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template").Preload("MatchRules")

	// Scope filtering
	if allowed := middleware.GetAllowedClusters(c); allowed != nil {
		query = query.Where("clusters.id IN ?", allowed)
	}
	if regionID := c.Query("region_id"); regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}
	if envID := c.Query("environment_id"); envID != "" {
		query = query.Where("environment_id = ?", envID)
	}
	query.Find(&clusters)

	type ClusterWithCount struct {
		models.Cluster
		NodeCount    int64 `json:"node_count"`
		OnlineCount  int64 `json:"online_count"`
		OfflineCount int64 `json:"offline_count"`
	}
	var result []ClusterWithCount
	for _, cl := range clusters {
		var nodeCount, onlineCount, offlineCount int64
		models.DB.Model(&models.Node{}).Where("cluster_id = ?", cl.ID).Count(&nodeCount)
		models.DB.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "online").Count(&onlineCount)
		models.DB.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "offline").Count(&offlineCount)
		result = append(result, ClusterWithCount{Cluster: cl, NodeCount: nodeCount, OnlineCount: onlineCount, OfflineCount: offlineCount})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *TopologyHandler) GetCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var cluster models.Cluster
	if err := models.DB.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template").Preload("Nodes").Preload("MatchRules").First(&cluster, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
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
	var region models.Region
	if err := models.DB.First(&region, req.RegionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region not found"})
		return
	}
	// If setting as default, unset any existing default
	if req.IsDefault {
		models.DB.Model(&models.Cluster{}).Where("is_default = ?", true).Update("is_default", false)
	}
	cluster := models.Cluster{
		Name: req.Name, Alias: req.Alias, RegionID: req.RegionID,
		EnvironmentID: req.EnvironmentID, IsDefault: req.IsDefault,
		ConfigID: req.ConfigID, Description: req.Description, Tags: req.Tags,
	}
	models.DB.Create(&cluster)
	c.JSON(http.StatusCreated, cluster)
}

func (h *TopologyHandler) UpdateCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var cluster models.Cluster
	if err := models.DB.First(&cluster, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IsDefault {
		models.DB.Model(&models.Cluster{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false)
	}
	models.DB.Model(&cluster).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "region_id": req.RegionID,
		"environment_id": req.EnvironmentID, "is_default": req.IsDefault,
		"config_id": req.ConfigID, "description": req.Description, "tags": req.Tags,
	})
	c.JSON(http.StatusOK, cluster)
}

func (h *TopologyHandler) DeleteCluster(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var nodeCount int64
	models.DB.Model(&models.Node{}).Where("cluster_id = ?", id).Count(&nodeCount)
	if nodeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete cluster with existing nodes; move or delete nodes first"})
		return
	}
	models.DB.Delete(&models.Cluster{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted"})
}

// ============================================================================
// Cluster Match Rules
// ============================================================================

func (h *TopologyHandler) ListMatchRules(c *gin.Context) {
	clusterID := c.Param("id")
	var rules []models.ClusterMatchRule
	models.DB.Where("cluster_id = ?", clusterID).Order("priority ASC").Find(&rules)
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *TopologyHandler) CreateMatchRule(c *gin.Context) {
	clusterID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var rule models.ClusterMatchRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.ClusterID = uint(clusterID)
	models.DB.Create(&rule)
	c.JSON(http.StatusCreated, rule)
}

func (h *TopologyHandler) UpdateMatchRule(c *gin.Context) {
	ruleID, _ := strconv.ParseUint(c.Param("rule_id"), 10, 32)
	var rule models.ClusterMatchRule
	if err := models.DB.First(&rule, ruleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	var req models.ClusterMatchRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Model(&rule).Updates(map[string]interface{}{
		"name": req.Name, "priority": req.Priority,
		"hostname_pattern": req.HostnamePattern, "ip_pattern": req.IPPattern,
		"fluent_type": req.FluentType, "label_selector": req.LabelSelector,
		"os_pattern": req.OSPattern, "is_active": req.IsActive,
	})
	c.JSON(http.StatusOK, rule)
}

func (h *TopologyHandler) DeleteMatchRule(c *gin.Context) {
	ruleID, _ := strconv.ParseUint(c.Param("rule_id"), 10, 32)
	models.DB.Delete(&models.ClusterMatchRule{}, ruleID)
	c.JSON(http.StatusOK, gin.H{"message": "rule deleted"})
}

// ============================================================================
// User Scope Management
// ============================================================================

func (h *TopologyHandler) ListUserScopes(c *gin.Context) {
	userID := c.Param("id")
	var scopes []models.UserScope
	models.DB.Where("user_id = ?", userID).Find(&scopes)
	c.JSON(http.StatusOK, gin.H{"data": scopes})
}

func (h *TopologyHandler) SetUserScopes(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Scopes []struct {
			ScopeType string `json:"scope_type" binding:"required"` // datacenter, region, cluster
			ScopeID   uint   `json:"scope_id" binding:"required"`
		} `json:"scopes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete existing scopes and replace
	models.DB.Where("user_id = ?", userID).Delete(&models.UserScope{})

	for _, s := range req.Scopes {
		scopeName := resolveScopeName(s.ScopeType, s.ScopeID)
		models.DB.Create(&models.UserScope{
			UserID:    uint(userID),
			ScopeType: s.ScopeType,
			ScopeID:   s.ScopeID,
			ScopeName: scopeName,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "scopes updated"})
}

func resolveScopeName(scopeType string, scopeID uint) string {
	switch scopeType {
	case "datacenter":
		var dc models.DataCenter
		if models.DB.First(&dc, scopeID).Error == nil {
			if dc.Alias != "" {
				return dc.Alias
			}
			return dc.Name
		}
	case "region":
		var r models.Region
		if models.DB.First(&r, scopeID).Error == nil {
			if r.Alias != "" {
				return r.Alias
			}
			return r.Name
		}
	case "cluster":
		var cl models.Cluster
		if models.DB.First(&cl, scopeID).Error == nil {
			if cl.Alias != "" {
				return cl.Alias
			}
			return cl.Name
		}
	}
	return ""
}

// ============================================================================
// Environment
// ============================================================================

func (h *TopologyHandler) ListEnvironments(c *gin.Context) {
	var envs []models.Environment
	models.DB.Order("sort_order").Find(&envs)
	c.JSON(http.StatusOK, gin.H{"data": envs})
}

func (h *TopologyHandler) CreateEnvironment(c *gin.Context) {
	var env models.Environment
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := models.DB.Create(&env).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "environment name already exists"})
		return
	}
	c.JSON(http.StatusCreated, env)
}

func (h *TopologyHandler) UpdateEnvironment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var env models.Environment
	if err := models.DB.First(&env, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "environment not found"})
		return
	}
	var req models.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	models.DB.Model(&env).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "color": req.Color,
		"sort_order": req.SortOrder, "description": req.Description,
	})
	c.JSON(http.StatusOK, env)
}

func (h *TopologyHandler) DeleteEnvironment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	models.DB.Delete(&models.Environment{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "environment deleted"})
}

// ============================================================================
// Topology Tree (scope-aware)
// ============================================================================

func (h *TopologyHandler) GetTree(c *gin.Context) {
	type ClusterInfo struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Alias       string `json:"alias"`
		IsDefault   bool   `json:"is_default"`
		Environment string `json:"environment"`
		EnvColor    string `json:"env_color"`
		NodeCount   int64  `json:"node_count"`
		OnlineCount int64  `json:"online_count"`
	}
	type RegionInfo struct {
		ID       uint          `json:"id"`
		Name     string        `json:"name"`
		Alias    string        `json:"alias"`
		Clusters []ClusterInfo `json:"clusters"`
	}
	type DCInfo struct {
		ID       uint         `json:"id"`
		Name     string       `json:"name"`
		Alias    string       `json:"alias"`
		Provider string       `json:"provider"`
		Regions  []RegionInfo `json:"regions"`
	}

	allowedClusters := middleware.GetAllowedClusters(c)

	var dcs []models.DataCenter
	dcQuery := models.DB.Model(&models.DataCenter{})
	if dcIDs := scopeFilterDCs(c); dcIDs != nil {
		dcQuery = dcQuery.Where("id IN ?", dcIDs)
	}
	dcQuery.Find(&dcs)

	var tree []DCInfo
	for _, dc := range dcs {
		dcInfo := DCInfo{ID: dc.ID, Name: dc.Name, Alias: dc.Alias, Provider: dc.Provider}

		var regions []models.Region
		models.DB.Where("data_center_id = ?", dc.ID).Find(&regions)

		for _, r := range regions {
			rInfo := RegionInfo{ID: r.ID, Name: r.Name, Alias: r.Alias}

			var clusters []models.Cluster
			clQuery := models.DB.Where("region_id = ?", r.ID).Preload("Environment")
			if allowedClusters != nil {
				clQuery = clQuery.Where("id IN ?", allowedClusters)
			}
			clQuery.Find(&clusters)

			for _, cl := range clusters {
				var nodeCount, onlineCount int64
				models.DB.Model(&models.Node{}).Where("cluster_id = ?", cl.ID).Count(&nodeCount)
				models.DB.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "online").Count(&onlineCount)

				envName := ""
				envColor := ""
				if cl.Environment != nil {
					envName = cl.Environment.Alias
					envColor = cl.Environment.Color
				}
				rInfo.Clusters = append(rInfo.Clusters, ClusterInfo{
					ID: cl.ID, Name: cl.Name, Alias: cl.Alias, IsDefault: cl.IsDefault,
					Environment: envName, EnvColor: envColor,
					NodeCount: nodeCount, OnlineCount: onlineCount,
				})
			}
			if len(rInfo.Clusters) > 0 || allowedClusters == nil {
				dcInfo.Regions = append(dcInfo.Regions, rInfo)
			}
		}
		tree = append(tree, dcInfo)
	}

	c.JSON(http.StatusOK, gin.H{"data": tree})
}
