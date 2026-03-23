package services

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

// Response types for topology queries with counts.

type DCWithCounts struct {
	models.DataCenter
	RegionCount  int64 `json:"region_count"`
	ClusterCount int64 `json:"cluster_count"`
	NodeCount    int64 `json:"node_count"`
}

type RegionWithCounts struct {
	models.Region
	ClusterCount int64 `json:"cluster_count"`
	NodeCount    int64 `json:"node_count"`
}

type ClusterWithCount struct {
	models.Cluster
	NodeCount    int64 `json:"node_count"`
	OnlineCount  int64 `json:"online_count"`
	OfflineCount int64 `json:"offline_count"`
}

type ClusterTreeInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	IsDefault   bool   `json:"is_default"`
	Environment string `json:"environment"`
	EnvColor    string `json:"env_color"`
	NodeCount   int64  `json:"node_count"`
	OnlineCount int64  `json:"online_count"`
}

type RegionTreeInfo struct {
	ID       uint              `json:"id"`
	Name     string            `json:"name"`
	Alias    string            `json:"alias"`
	Clusters []ClusterTreeInfo `json:"clusters"`
}

type DCTreeInfo struct {
	ID       uint             `json:"id"`
	Name     string           `json:"name"`
	Alias    string           `json:"alias"`
	Provider string           `json:"provider"`
	Regions  []RegionTreeInfo `json:"regions"`
}

type TopologyService interface {
	// DataCenter
	ListDataCenters(allowedDCIDs []uint) ([]DCWithCounts, error)
	GetDataCenter(id uint) (*models.DataCenter, error)
	CreateDataCenter(name, alias, provider, location, description, tags string) (*models.DataCenter, error)
	UpdateDataCenter(id uint, name, alias, provider, location, description, tags string) (*models.DataCenter, error)
	DeleteDataCenter(id uint) error

	// Region
	ListRegions(allowedDCIDs []uint, datacenterID string) ([]RegionWithCounts, error)
	GetRegion(id uint) (*models.Region, error)
	CreateRegion(name, alias string, dataCenterID uint, description, tags string) (*models.Region, error)
	UpdateRegion(id uint, name, alias string, dataCenterID uint, description, tags string) (*models.Region, error)
	DeleteRegion(id uint) error

	// Cluster
	ListClusters(allowedClusters []uint, regionID, envID string) ([]ClusterWithCount, error)
	GetCluster(id uint) (*models.Cluster, error)
	CreateCluster(name, alias string, regionID uint, environmentID *uint, isDefault bool, configID *uint, description, tags string) (*models.Cluster, error)
	UpdateCluster(id uint, name, alias string, regionID uint, environmentID *uint, isDefault bool, configID *uint, description, tags string) (*models.Cluster, error)
	DeleteCluster(id uint) error

	// Match Rules
	ListMatchRules(clusterID string) ([]models.ClusterMatchRule, error)
	CreateMatchRule(clusterID uint, rule *models.ClusterMatchRule) (*models.ClusterMatchRule, error)
	UpdateMatchRule(ruleID uint, rule *models.ClusterMatchRule) (*models.ClusterMatchRule, error)
	DeleteMatchRule(ruleID uint) error

	// User Scopes
	ListUserScopes(userID string) ([]models.UserScope, error)
	SetUserScopes(userID uint, scopes []ScopeInput) error

	// Environment
	ListEnvironments() ([]models.Environment, error)
	CreateEnvironment(env *models.Environment) (*models.Environment, error)
	UpdateEnvironment(id uint, env *models.Environment) (*models.Environment, error)
	DeleteEnvironment(id uint) error

	// Tree
	GetTree(allowedClusters []uint, allowedDCIDs []uint) ([]DCTreeInfo, error)

	// Helpers
	AllowedDCIDs(allowedClusters []uint) []uint
}

type ScopeInput struct {
	ScopeType string `json:"scope_type"`
	ScopeID   uint   `json:"scope_id"`
}

type topologyService struct {
	db *gorm.DB
}

func NewTopologyService(db *gorm.DB) TopologyService {
	return &topologyService{db: db}
}

// ---------- DataCenter ----------

func (s *topologyService) ListDataCenters(allowedDCIDs []uint) ([]DCWithCounts, error) {
	var dcs []models.DataCenter
	query := s.db.Preload("Regions")
	if allowedDCIDs != nil {
		query = query.Where("id IN ?", allowedDCIDs)
	}
	query.Find(&dcs)

	var result []DCWithCounts
	for _, dc := range dcs {
		var regionCount, clusterCount, nodeCount int64
		s.db.Model(&models.Region{}).Where("data_center_id = ?", dc.ID).Count(&regionCount)
		s.db.Model(&models.Cluster{}).
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", dc.ID).Count(&clusterCount)
		s.db.Model(&models.Node{}).
			Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", dc.ID).Count(&nodeCount)
		result = append(result, DCWithCounts{DataCenter: dc, RegionCount: regionCount, ClusterCount: clusterCount, NodeCount: nodeCount})
	}
	return result, nil
}

func (s *topologyService) GetDataCenter(id uint) (*models.DataCenter, error) {
	var dc models.DataCenter
	if err := s.db.Preload("Regions.Clusters").First(&dc, id).Error; err != nil {
		return nil, err
	}
	return &dc, nil
}

func (s *topologyService) CreateDataCenter(name, alias, provider, location, description, tags string) (*models.DataCenter, error) {
	dc := models.DataCenter{
		Name: name, Alias: alias, Provider: provider,
		Location: location, Description: description, Tags: tags,
	}
	if err := s.db.Create(&dc).Error; err != nil {
		return nil, err
	}
	return &dc, nil
}

func (s *topologyService) UpdateDataCenter(id uint, name, alias, provider, location, description, tags string) (*models.DataCenter, error) {
	var dc models.DataCenter
	if err := s.db.First(&dc, id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&dc).Updates(map[string]interface{}{
		"name": name, "alias": alias, "provider": provider,
		"location": location, "description": description, "tags": tags,
	}).Error; err != nil {
		return nil, err
	}
	return &dc, nil
}

func (s *topologyService) DeleteDataCenter(id uint) error {
	var regionCount int64
	s.db.Model(&models.Region{}).Where("data_center_id = ?", id).Count(&regionCount)
	if regionCount > 0 {
		return ErrHasChildren
	}
	return s.db.Delete(&models.DataCenter{}, id).Error
}

// ---------- Region ----------

func (s *topologyService) ListRegions(allowedDCIDs []uint, datacenterID string) ([]RegionWithCounts, error) {
	var regions []models.Region
	query := s.db.Preload("DataCenter")
	if allowedDCIDs != nil {
		query = query.Where("data_center_id IN ?", allowedDCIDs)
	}
	if datacenterID != "" {
		query = query.Where("data_center_id = ?", datacenterID)
	}
	query.Find(&regions)

	var result []RegionWithCounts
	for _, r := range regions {
		var clusterCount, nodeCount int64
		s.db.Model(&models.Cluster{}).Where("region_id = ?", r.ID).Count(&clusterCount)
		s.db.Model(&models.Node{}).
			Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Where("clusters.region_id = ?", r.ID).Count(&nodeCount)
		result = append(result, RegionWithCounts{Region: r, ClusterCount: clusterCount, NodeCount: nodeCount})
	}
	return result, nil
}

func (s *topologyService) GetRegion(id uint) (*models.Region, error) {
	var region models.Region
	if err := s.db.Preload("DataCenter").Preload("Clusters.Environment").First(&region, id).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (s *topologyService) CreateRegion(name, alias string, dataCenterID uint, description, tags string) (*models.Region, error) {
	var dc models.DataCenter
	if err := s.db.First(&dc, dataCenterID).Error; err != nil {
		return nil, err
	}
	region := models.Region{
		Name: name, Alias: alias, DataCenterID: dataCenterID,
		Description: description, Tags: tags,
	}
	if err := s.db.Create(&region).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (s *topologyService) UpdateRegion(id uint, name, alias string, dataCenterID uint, description, tags string) (*models.Region, error) {
	var region models.Region
	if err := s.db.First(&region, id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&region).Updates(map[string]interface{}{
		"name": name, "alias": alias, "data_center_id": dataCenterID,
		"description": description, "tags": tags,
	}).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (s *topologyService) DeleteRegion(id uint) error {
	var clusterCount int64
	s.db.Model(&models.Cluster{}).Where("region_id = ?", id).Count(&clusterCount)
	if clusterCount > 0 {
		return ErrHasChildren
	}
	return s.db.Delete(&models.Region{}, id).Error
}

// ---------- Cluster ----------

func (s *topologyService) ListClusters(allowedClusters []uint, regionID, envID string) ([]ClusterWithCount, error) {
	var clusters []models.Cluster
	query := s.db.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template").Preload("MatchRules")
	if allowedClusters != nil {
		query = query.Where("clusters.id IN ?", allowedClusters)
	}
	if regionID != "" {
		query = query.Where("region_id = ?", regionID)
	}
	if envID != "" {
		query = query.Where("environment_id = ?", envID)
	}
	query.Find(&clusters)

	var result []ClusterWithCount
	for _, cl := range clusters {
		var nodeCount, onlineCount, offlineCount int64
		s.db.Model(&models.Node{}).Where("cluster_id = ?", cl.ID).Count(&nodeCount)
		s.db.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "online").Count(&onlineCount)
		s.db.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "offline").Count(&offlineCount)
		result = append(result, ClusterWithCount{Cluster: cl, NodeCount: nodeCount, OnlineCount: onlineCount, OfflineCount: offlineCount})
	}
	return result, nil
}

func (s *topologyService) GetCluster(id uint) (*models.Cluster, error) {
	var cluster models.Cluster
	if err := s.db.Preload("Region.DataCenter").Preload("Environment").Preload("Config.Template").Preload("Nodes").Preload("MatchRules").First(&cluster, id).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (s *topologyService) CreateCluster(name, alias string, regionID uint, environmentID *uint, isDefault bool, configID *uint, description, tags string) (*models.Cluster, error) {
	var region models.Region
	if err := s.db.First(&region, regionID).Error; err != nil {
		return nil, err
	}
	if isDefault {
		s.db.Model(&models.Cluster{}).Where("is_default = ?", true).Update("is_default", false)
	}
	cluster := models.Cluster{
		Name: name, Alias: alias, RegionID: regionID,
		EnvironmentID: environmentID, IsDefault: isDefault,
		ConfigID: configID, Description: description, Tags: tags,
	}
	if err := s.db.Create(&cluster).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (s *topologyService) UpdateCluster(id uint, name, alias string, regionID uint, environmentID *uint, isDefault bool, configID *uint, description, tags string) (*models.Cluster, error) {
	var cluster models.Cluster
	if err := s.db.First(&cluster, id).Error; err != nil {
		return nil, err
	}
	if isDefault {
		s.db.Model(&models.Cluster{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false)
	}
	if err := s.db.Model(&cluster).Updates(map[string]interface{}{
		"name": name, "alias": alias, "region_id": regionID,
		"environment_id": environmentID, "is_default": isDefault,
		"config_id": configID, "description": description, "tags": tags,
	}).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (s *topologyService) DeleteCluster(id uint) error {
	var nodeCount int64
	s.db.Model(&models.Node{}).Where("cluster_id = ?", id).Count(&nodeCount)
	if nodeCount > 0 {
		return ErrHasChildren
	}
	return s.db.Delete(&models.Cluster{}, id).Error
}

// ---------- Match Rules ----------

func (s *topologyService) ListMatchRules(clusterID string) ([]models.ClusterMatchRule, error) {
	var rules []models.ClusterMatchRule
	err := s.db.Where("cluster_id = ?", clusterID).Order("priority ASC").Find(&rules).Error
	return rules, err
}

func (s *topologyService) CreateMatchRule(clusterID uint, rule *models.ClusterMatchRule) (*models.ClusterMatchRule, error) {
	rule.ClusterID = clusterID
	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *topologyService) UpdateMatchRule(ruleID uint, req *models.ClusterMatchRule) (*models.ClusterMatchRule, error) {
	var rule models.ClusterMatchRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return nil, err
	}
	s.db.Model(&rule).Updates(map[string]interface{}{
		"name": req.Name, "priority": req.Priority,
		"hostname_pattern": req.HostnamePattern, "ip_pattern": req.IPPattern,
		"fluent_type": req.FluentType, "label_selector": req.LabelSelector,
		"os_pattern": req.OSPattern, "is_active": req.IsActive,
	})
	return &rule, nil
}

func (s *topologyService) DeleteMatchRule(ruleID uint) error {
	return s.db.Delete(&models.ClusterMatchRule{}, ruleID).Error
}

// ---------- User Scopes ----------

func (s *topologyService) ListUserScopes(userID string) ([]models.UserScope, error) {
	var scopes []models.UserScope
	err := s.db.Where("user_id = ?", userID).Find(&scopes).Error
	return scopes, err
}

func (s *topologyService) SetUserScopes(userID uint, scopes []ScopeInput) error {
	s.db.Where("user_id = ?", userID).Delete(&models.UserScope{})
	for _, sc := range scopes {
		scopeName := s.resolveScopeName(sc.ScopeType, sc.ScopeID)
		s.db.Create(&models.UserScope{
			UserID:    userID,
			ScopeType: sc.ScopeType,
			ScopeID:   sc.ScopeID,
			ScopeName: scopeName,
		})
	}
	return nil
}

func (s *topologyService) resolveScopeName(scopeType string, scopeID uint) string {
	switch scopeType {
	case "datacenter":
		var dc models.DataCenter
		if s.db.First(&dc, scopeID).Error == nil {
			if dc.Alias != "" {
				return dc.Alias
			}
			return dc.Name
		}
	case "region":
		var r models.Region
		if s.db.First(&r, scopeID).Error == nil {
			if r.Alias != "" {
				return r.Alias
			}
			return r.Name
		}
	case "cluster":
		var cl models.Cluster
		if s.db.First(&cl, scopeID).Error == nil {
			if cl.Alias != "" {
				return cl.Alias
			}
			return cl.Name
		}
	}
	return ""
}

// ---------- Environment ----------

func (s *topologyService) ListEnvironments() ([]models.Environment, error) {
	var envs []models.Environment
	err := s.db.Order("sort_order").Find(&envs).Error
	return envs, err
}

func (s *topologyService) CreateEnvironment(env *models.Environment) (*models.Environment, error) {
	if err := s.db.Create(env).Error; err != nil {
		return nil, err
	}
	return env, nil
}

func (s *topologyService) UpdateEnvironment(id uint, req *models.Environment) (*models.Environment, error) {
	var env models.Environment
	if err := s.db.First(&env, id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&env).Updates(map[string]interface{}{
		"name": req.Name, "alias": req.Alias, "color": req.Color,
		"sort_order": req.SortOrder, "description": req.Description,
	})
	return &env, nil
}

func (s *topologyService) DeleteEnvironment(id uint) error {
	return s.db.Delete(&models.Environment{}, id).Error
}

// ---------- Tree ----------

func (s *topologyService) GetTree(allowedClusters []uint, allowedDCIDs []uint) ([]DCTreeInfo, error) {
	var dcs []models.DataCenter
	dcQuery := s.db.Model(&models.DataCenter{})
	if allowedDCIDs != nil {
		dcQuery = dcQuery.Where("id IN ?", allowedDCIDs)
	}
	dcQuery.Find(&dcs)

	var tree []DCTreeInfo
	for _, dc := range dcs {
		dcInfo := DCTreeInfo{ID: dc.ID, Name: dc.Name, Alias: dc.Alias, Provider: dc.Provider}

		var regions []models.Region
		s.db.Where("data_center_id = ?", dc.ID).Find(&regions)

		for _, r := range regions {
			rInfo := RegionTreeInfo{ID: r.ID, Name: r.Name, Alias: r.Alias}

			var clusters []models.Cluster
			clQuery := s.db.Where("region_id = ?", r.ID).Preload("Environment")
			if allowedClusters != nil {
				clQuery = clQuery.Where("id IN ?", allowedClusters)
			}
			clQuery.Find(&clusters)

			for _, cl := range clusters {
				var nodeCount, onlineCount int64
				s.db.Model(&models.Node{}).Where("cluster_id = ?", cl.ID).Count(&nodeCount)
				s.db.Model(&models.Node{}).Where("cluster_id = ? AND status = ?", cl.ID, "online").Count(&onlineCount)

				envName := ""
				envColor := ""
				if cl.Environment != nil {
					envName = cl.Environment.Alias
					envColor = cl.Environment.Color
				}
				rInfo.Clusters = append(rInfo.Clusters, ClusterTreeInfo{
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
	return tree, nil
}

// ---------- Helpers ----------

func (s *topologyService) AllowedDCIDs(allowedClusters []uint) []uint {
	if allowedClusters == nil {
		return nil
	}
	dcSet := map[uint]bool{}
	var clusters []models.Cluster
	s.db.Where("id IN ?", allowedClusters).Preload("Region").Find(&clusters)
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
