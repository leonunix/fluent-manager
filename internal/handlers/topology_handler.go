package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type TopologyHandler struct{}

// ============================================================================
// DataCenter
// ============================================================================

func (h *TopologyHandler) ListDataCenters(c *gin.Context) {
	var dcs []models.DataCenter
	models.DB.Preload("Regions").Find(&dcs)

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
	// Verify datacenter exists
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
	query := models.DB.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template")
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
	if err := models.DB.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template").Preload("Nodes").First(&cluster, id).Error; err != nil {
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
	cluster := models.Cluster{
		Name: req.Name, Alias: req.Alias, RegionID: req.RegionID,
		EnvironmentID: req.EnvironmentID, ConfigID: req.ConfigID,
		Description: req.Description, Tags: req.Tags,
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
	models.DB.Model(&cluster).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "region_id": req.RegionID,
		"environment_id": req.EnvironmentID, "config_id": req.ConfigID,
		"description": req.Description, "tags": req.Tags,
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
// Topology Tree (full tree for sidebar/navigation)
// ============================================================================

func (h *TopologyHandler) GetTree(c *gin.Context) {
	type ClusterInfo struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Alias       string `json:"alias"`
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

	var dcs []models.DataCenter
	models.DB.Find(&dcs)

	var tree []DCInfo
	for _, dc := range dcs {
		dcInfo := DCInfo{ID: dc.ID, Name: dc.Name, Alias: dc.Alias, Provider: dc.Provider}

		var regions []models.Region
		models.DB.Where("data_center_id = ?", dc.ID).Find(&regions)

		for _, r := range regions {
			rInfo := RegionInfo{ID: r.ID, Name: r.Name, Alias: r.Alias}

			var clusters []models.Cluster
			models.DB.Where("region_id = ?", r.ID).Preload("Environment").Find(&clusters)

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
					ID: cl.ID, Name: cl.Name, Alias: cl.Alias,
					Environment: envName, EnvColor: envColor,
					NodeCount: nodeCount, OnlineCount: onlineCount,
				})
			}
			dcInfo.Regions = append(dcInfo.Regions, rInfo)
		}
		tree = append(tree, dcInfo)
	}

	c.JSON(http.StatusOK, gin.H{"data": tree})
}
