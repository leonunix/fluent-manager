package services

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type NodeListFilters struct {
	Status        string
	ClusterID     string
	EnvironmentID string
	FluentType    string
	AgentVersion  string
	DataCenterID  string
	RegionID      string
	Search        string
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type NodeService interface {
	List(filters NodeListFilters, allowedClusters []uint, page, pageSize int) ([]models.Node, int64, error)
	Get(id uint) (*models.Node, error)
	Update(id uint, updates map[string]interface{}) (*models.Node, error)
	Delete(id uint) error
	BatchMoveCluster(nodeIDs []uint, clusterID uint) error
	Stats(allowedClusters []uint) ([]StatusCount, int64, error)
}

type nodeService struct {
	db *gorm.DB
}

func NewNodeService(db *gorm.DB) NodeService {
	return &nodeService{db: db}
}

func (s *nodeService) List(filters NodeListFilters, allowedClusters []uint, page, pageSize int) ([]models.Node, int64, error) {
	query := s.db.Preload("Cluster.Region.DataCenter").
		Preload("Cluster.Environment").
		Preload("Environment").
		Preload("Config.Template").
		Preload("AggregationGroup").
		Preload("FluentProfile")
	query = applyNodeScopeAndFilters(query, filters, allowedClusters)

	var total int64
	query.Model(&models.Node{}).Count(&total)

	var nodes []models.Node
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&nodes).Error
	return nodes, total, err
}

func (s *nodeService) Get(id uint) (*models.Node, error) {
	var node models.Node
	if err := s.db.Preload("Cluster.Region.DataCenter").
		Preload("Cluster.Environment").
		Preload("Environment").
		Preload("Config.Template").
		Preload("AggregationGroup").
		Preload("FluentProfile").
		First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *nodeService) Update(id uint, updates map[string]interface{}) (*models.Node, error) {
	var node models.Node
	if err := s.db.First(&node, id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&node).Updates(updates)
	s.db.Preload("Cluster.Region.DataCenter").
		Preload("Environment").
		Preload("Config").
		Preload("AggregationGroup").
		Preload("FluentProfile").
		First(&node, node.ID)
	return &node, nil
}

func (s *nodeService) Delete(id uint) error {
	var node models.Node
	if err := s.db.First(&node, id).Error; err != nil {
		return err
	}
	return s.db.Delete(&node).Error
}

func (s *nodeService) BatchMoveCluster(nodeIDs []uint, clusterID uint) error {
	return s.db.Model(&models.Node{}).Where("id IN ?", nodeIDs).Update("cluster_id", clusterID).Error
}

func (s *nodeService) Stats(allowedClusters []uint) ([]StatusCount, int64, error) {
	baseQuery := s.db.Model(&models.Node{})
	if allowedClusters != nil {
		baseQuery = baseQuery.Where("cluster_id IN ?", allowedClusters)
	}

	var counts []StatusCount
	baseQuery.Select("status, count(*) as count").Group("status").Scan(&counts)

	var total int64
	countQuery := s.db.Model(&models.Node{})
	if allowedClusters != nil {
		countQuery = countQuery.Where("cluster_id IN ?", allowedClusters)
	}
	countQuery.Count(&total)

	return counts, total, nil
}

func applyNodeScopeAndFilters(query *gorm.DB, filters NodeListFilters, allowedClusters []uint) *gorm.DB {
	if allowedClusters != nil {
		if len(allowedClusters) == 0 {
			return query.Where("1 = 0")
		}
		query = query.Where("nodes.cluster_id IN ?", allowedClusters)
	}
	if filters.Status != "" {
		query = query.Where("nodes.status = ?", filters.Status)
	}
	if filters.ClusterID != "" {
		query = query.Where("nodes.cluster_id = ?", filters.ClusterID)
	}
	if filters.EnvironmentID != "" {
		query = query.Where("nodes.environment_id = ? OR nodes.cluster_id IN (SELECT id FROM clusters WHERE environment_id = ?)", filters.EnvironmentID, filters.EnvironmentID)
	}
	if filters.FluentType != "" {
		query = query.Where("nodes.fluent_type = ?", filters.FluentType)
	}
	if filters.AgentVersion != "" {
		query = query.Where("nodes.agent_version = ?", filters.AgentVersion)
	}
	if filters.DataCenterID != "" {
		query = query.Joins("JOIN clusters c2 ON c2.id = nodes.cluster_id").
			Joins("JOIN regions r2 ON r2.id = c2.region_id").
			Where("r2.data_center_id = ?", filters.DataCenterID)
	}
	if filters.RegionID != "" {
		query = query.Joins("JOIN clusters c3 ON c3.id = nodes.cluster_id").
			Where("c3.region_id = ?", filters.RegionID)
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		query = query.Where(
			"nodes.hostname LIKE ? OR nodes.ip_address LIKE ? OR nodes.node_uid LIKE ? OR nodes.labels LIKE ?",
			term, term, term, term,
		)
	}
	return query
}
