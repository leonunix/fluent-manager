package services

import (
	"errors"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type DeployService interface {
	Create(configVersionID uint, nodeIDs []uint, clusterID, regionID, dataCenterID, environmentID *uint, userID uint) (*models.DeployTask, error)
	List(page, pageSize int) ([]models.DeployTask, int64, error)
	Get(id uint) (*models.DeployTask, []models.DeployRecord, error)
	GetAuditLogs(page, pageSize int) ([]models.AuditLog, int64, error)
}

type deployService struct {
	db *gorm.DB
}

func NewDeployService(db *gorm.DB) DeployService {
	return &deployService{db: db}
}

func (s *deployService) Create(configVersionID uint, nodeIDs []uint, clusterID, regionID, dataCenterID, environmentID *uint, userID uint) (*models.DeployTask, error) {
	var configVersion models.ConfigVersion
	if err := s.db.First(&configVersion, configVersionID).Error; err != nil {
		return nil, errors.New("config version not found")
	}

	scope := "node"
	scopeID := uint(0)
	var targetNodeIDs []uint

	if dataCenterID != nil {
		scope = "datacenter"
		scopeID = *dataCenterID
		var nodes []models.Node
		s.db.Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Joins("JOIN regions ON regions.id = clusters.region_id").
			Where("regions.data_center_id = ?", *dataCenterID).Find(&nodes)
		for _, n := range nodes {
			targetNodeIDs = append(targetNodeIDs, n.ID)
		}
	}
	if regionID != nil {
		scope = "region"
		scopeID = *regionID
		var nodes []models.Node
		s.db.Joins("JOIN clusters ON clusters.id = nodes.cluster_id").
			Where("clusters.region_id = ?", *regionID).Find(&nodes)
		for _, n := range nodes {
			targetNodeIDs = append(targetNodeIDs, n.ID)
		}
	}
	if clusterID != nil {
		scope = "cluster"
		scopeID = *clusterID
		var nodes []models.Node
		s.db.Where("cluster_id = ?", *clusterID).Find(&nodes)
		for _, n := range nodes {
			targetNodeIDs = append(targetNodeIDs, n.ID)
		}
		s.db.Model(&models.Cluster{}).Where("id = ?", *clusterID).Update("config_id", configVersionID)
	}
	if environmentID != nil {
		scope = "environment"
		scopeID = *environmentID
		var nodes []models.Node
		s.db.Where("environment_id = ? OR cluster_id IN (SELECT id FROM clusters WHERE environment_id = ?)",
			*environmentID, *environmentID).Find(&nodes)
		for _, n := range nodes {
			targetNodeIDs = append(targetNodeIDs, n.ID)
		}
	}
	targetNodeIDs = append(targetNodeIDs, nodeIDs...)

	if len(targetNodeIDs) == 0 {
		return nil, errors.New("no target nodes found")
	}

	// Deduplicate
	seen := map[uint]bool{}
	var uniqueIDs []uint
	for _, id := range targetNodeIDs {
		if !seen[id] {
			seen[id] = true
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	task := models.DeployTask{
		ConfigID:   configVersionID,
		Scope:      scope,
		ScopeID:    scopeID,
		Status:     "pending",
		TotalNodes: len(uniqueIDs),
		CreatedBy:  userID,
	}
	s.db.Create(&task)

	for _, nodeID := range uniqueIDs {
		s.db.Create(&models.DeployRecord{
			DeployTaskID: task.ID,
			NodeID:       nodeID,
			Status:       "pending",
		})
	}

	s.db.Model(&models.Node{}).Where("id IN ?", uniqueIDs).Update("config_id", configVersionID)
	s.db.Model(&task).Update("status", "running")

	return &task, nil
}

func (s *deployService) List(page, pageSize int) ([]models.DeployTask, int64, error) {
	query := s.db.Preload("Config.Template").Preload("Creator")
	var total int64
	query.Model(&models.DeployTask{}).Count(&total)

	var tasks []models.DeployTask
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (s *deployService) Get(id uint) (*models.DeployTask, []models.DeployRecord, error) {
	var task models.DeployTask
	if err := s.db.Preload("Config.Template").Preload("Creator").First(&task, id).Error; err != nil {
		return nil, nil, err
	}

	var records []models.DeployRecord
	s.db.Where("deploy_task_id = ?", id).Preload("Node").Find(&records)
	return &task, records, nil
}

func (s *deployService) GetAuditLogs(page, pageSize int) ([]models.AuditLog, int64, error) {
	var total int64
	s.db.Model(&models.AuditLog{}).Count(&total)

	var logs []models.AuditLog
	err := s.db.Order("created_at DESC").
		Offset((page-1)*pageSize).Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}
