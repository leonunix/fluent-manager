package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

// DeployConflictError is returned when active deploy tasks exist for the same target and force=false.
type DeployConflictError struct {
	Count    int
	TaskIDs  []uint
}

func (e *DeployConflictError) Error() string {
	return fmt.Sprintf("there are %d active deploy task(s) targeting the same scope", e.Count)
}

type DeployService interface {
	Create(configVersionID uint, nodeIDs []uint, clusterID, regionID, dataCenterID, environmentID *uint, userID uint, allowedClusters []uint, force bool) (*models.DeployTask, error)
	List(page, pageSize int, allowedClusters []uint) ([]models.DeployTask, int64, error)
	Get(id uint, page, pageSize int, allowedClusters []uint) (*models.DeployTask, []models.DeployRecord, int64, error)
	GetAuditLogs(page, pageSize int, allowedClusters []uint) ([]models.AuditLog, int64, error)
}

type deployService struct {
	db        *gorm.DB
	fluentOps FluentOpsService
}

func NewDeployService(db *gorm.DB, fluentOps FluentOpsService) DeployService {
	return &deployService{db: db, fluentOps: fluentOps}
}

func (s *deployService) Create(configVersionID uint, nodeIDs []uint, clusterID, regionID, dataCenterID, environmentID *uint, userID uint, allowedClusters []uint, force bool) (*models.DeployTask, error) {
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

	// Scope check: verify all target nodes are in user's allowed clusters
	if allowedClusters != nil {
		allowedSet := map[uint]bool{}
		for _, cid := range allowedClusters {
			allowedSet[cid] = true
		}
		var nodes []models.Node
		s.db.Where("id IN ?", uniqueIDs).Find(&nodes)
		for _, n := range nodes {
			if n.ClusterID != nil && !allowedSet[*n.ClusterID] {
				return nil, errors.New("some target nodes are not in your scope")
			}
		}
	}

	// Conflict check: warn if there are already pending/running tasks for the same target.
	if !force {
		var conflictTasks []models.DeployTask
		if scopeID != 0 {
			s.db.Where("scope = ? AND scope_id = ? AND status IN ('pending','running')", scope, scopeID).
				Find(&conflictTasks)
		} else {
			// node scope: check if any target nodes appear in active deploy records
			s.db.Joins("JOIN deploy_records ON deploy_records.deploy_task_id = deploy_tasks.id").
				Where("deploy_tasks.status IN ('pending','running') AND deploy_records.node_id IN ?", uniqueIDs).
				Distinct("deploy_tasks.id").
				Find(&conflictTasks)
		}
		if len(conflictTasks) > 0 {
			ids := make([]uint, len(conflictTasks))
			for i, t := range conflictTasks {
				ids[i] = t.ID
			}
			return nil, &DeployConflictError{Count: len(conflictTasks), TaskIDs: ids}
		}
	}

	now := time.Now()
	task := models.DeployTask{
		ConfigID:   configVersionID,
		Scope:      scope,
		ScopeID:    scopeID,
		Status:     "running",
		TotalNodes: len(uniqueIDs),
		StartedAt:  &now,
		CreatedBy:  userID,
	}
	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}

	for _, nodeID := range uniqueIDs {
		if err := s.db.Create(&models.DeployRecord{
			DeployTaskID: task.ID,
			NodeID:       nodeID,
			Status:       "pending",
		}).Error; err != nil {
			return nil, err
		}
	}

	s.db.Model(&models.Node{}).Where("id IN ?", uniqueIDs).Update("config_id", configVersionID)

	// Derive affected cluster IDs and sync LogPipelines from the deployed config's flow_layout.
	if s.fluentOps != nil {
		affectedClusterIDs := s.resolveAffectedClusterIDs(clusterID, regionID, dataCenterID, environmentID, uniqueIDs)
		if len(affectedClusterIDs) > 0 {
			s.fluentOps.SyncPipelinesFromDeploy(affectedClusterIDs, configVersionID, userID)
		}
	}

	return &task, nil
}

func (s *deployService) resolveAffectedClusterIDs(clusterID, regionID, dataCenterID, environmentID *uint, nodeIDs []uint) []uint {
	seen := map[uint]bool{}
	if clusterID != nil {
		seen[*clusterID] = true
	}
	if len(nodeIDs) > 0 && (regionID != nil || dataCenterID != nil || environmentID != nil || clusterID == nil) {
		var nodes []models.Node
		s.db.Select("cluster_id").Where("id IN ? AND cluster_id IS NOT NULL", nodeIDs).Find(&nodes)
		for _, n := range nodes {
			if n.ClusterID != nil {
				seen[*n.ClusterID] = true
			}
		}
	}
	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

func (s *deployService) List(page, pageSize int, allowedClusters []uint) ([]models.DeployTask, int64, error) {
	query := s.db.Preload("Config.Template").Preload("Creator")

	// Scope filter: only show tasks that have records targeting nodes in allowed clusters
	if allowedClusters != nil {
		query = query.Where("id IN (SELECT DISTINCT deploy_task_id FROM deploy_records WHERE node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?))", allowedClusters)
	}

	var total int64
	query.Model(&models.DeployTask{}).Count(&total)

	var tasks []models.DeployTask
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (s *deployService) Get(id uint, page, pageSize int, allowedClusters []uint) (*models.DeployTask, []models.DeployRecord, int64, error) {
	var task models.DeployTask
	if err := s.db.Preload("Config.Template").Preload("Creator").First(&task, id).Error; err != nil {
		return nil, nil, 0, err
	}

	// Scope check: verify the task has at least one record targeting a node in the user's allowed clusters
	if allowedClusters != nil {
		var count int64
		s.db.Model(&models.DeployRecord{}).
			Where("deploy_task_id = ? AND node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?)", id, allowedClusters).
			Count(&count)
		if count == 0 {
			return nil, nil, 0, errors.New("deploy task not found")
		}
	}

	countQuery := s.db.Model(&models.DeployRecord{}).Where("deploy_task_id = ?", id)
	if allowedClusters != nil {
		countQuery = countQuery.Where("node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?)", allowedClusters)
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	var records []models.DeployRecord
	recordQuery := s.db.Where("deploy_task_id = ?", id).Preload("Node")
	if allowedClusters != nil {
		recordQuery = recordQuery.Where("node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?)", allowedClusters)
	}
	if err := recordQuery.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, nil, 0, err
	}

	return &task, records, total, nil
}

func (s *deployService) GetAuditLogs(page, pageSize int, allowedClusters []uint) ([]models.AuditLog, int64, error) {
	query := s.db.Model(&models.AuditLog{})

	// Scope filter: only show audit logs for resources within the user's scope
	if allowedClusters != nil {
		query = query.Where(
			"(resource_type = 'node' AND resource_id IN (SELECT id FROM nodes WHERE cluster_id IN ?)) "+
				"OR (resource_type = 'cluster' AND resource_id IN ?) "+
				"OR (resource_type = 'agent_policy' AND resource_id IN (SELECT id FROM agent_policies WHERE scope_type = 'cluster' AND cluster_id IN ?)) "+
				"OR (resource_type NOT IN ('node', 'cluster', 'agent_policy') OR resource_type = '')",
			allowedClusters, allowedClusters, allowedClusters,
		)
	}

	var total int64
	query.Count(&total)

	var logs []models.AuditLog
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}
