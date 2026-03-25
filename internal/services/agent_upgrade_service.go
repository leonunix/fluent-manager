package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type AgentUpgradeTaskInput struct {
	Name          string          `json:"name"`
	ArtifactID    *uint           `json:"artifact_id"`
	PackageURL    string          `json:"package_url"`
	Checksum      string          `json:"checksum"`
	TargetVersion string          `json:"target_version"`
	ServiceUnit   string          `json:"service_unit"`
	BinaryPath    string          `json:"binary_path"`
	AllMatching   bool            `json:"all_matching"`
	NodeIDs       []uint          `json:"node_ids"`
	Filters       NodeListFilters `json:"filters"`
}

type AgentUpgradeService interface {
	Create(input AgentUpgradeTaskInput, userID uint, allowedClusters []uint) (*models.AgentUpgradeTask, error)
	List(page, pageSize int, allowedClusters []uint) ([]models.AgentUpgradeTask, int64, error)
	Get(id uint, allowedClusters []uint) (*models.AgentUpgradeTask, []models.AgentUpgradeRecord, error)
}

type agentUpgradeService struct {
	db *gorm.DB
}

type agentUpgradeCommandArgs struct {
	PackageURL    string `json:"package_url"`
	Checksum      string `json:"checksum,omitempty"`
	TargetVersion string `json:"target_version,omitempty"`
	ServiceUnit   string `json:"service_unit,omitempty"`
	BinaryPath    string `json:"binary_path,omitempty"`
}

func NewAgentUpgradeService(db *gorm.DB) AgentUpgradeService {
	return &agentUpgradeService{db: db}
}

func (s *agentUpgradeService) Create(input AgentUpgradeTaskInput, userID uint, allowedClusters []uint) (*models.AgentUpgradeTask, error) {
	packageURL := strings.TrimSpace(input.PackageURL)
	switch {
	case input.ArtifactID == nil && packageURL == "":
		return nil, fmt.Errorf("%w: either artifact_id or package_url is required", ErrInvalidArgument)
	case input.ArtifactID != nil && packageURL != "":
		return nil, fmt.Errorf("%w: choose either artifact_id or package_url, not both", ErrInvalidArgument)
	}
	if !input.AllMatching && len(input.NodeIDs) == 0 {
		return nil, fmt.Errorf("%w: select at least one node or enable all_matching", ErrInvalidArgument)
	}

	if input.ArtifactID != nil {
		var record models.AgentArtifact
		if err := s.db.First(&record, *input.ArtifactID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: agent artifact not found", ErrInvalidArgument)
			}
			return nil, err
		}
		packageURL = fmt.Sprintf("/api/v1/agent/artifacts/%d/download", record.ID)
		if strings.TrimSpace(input.Checksum) == "" {
			input.Checksum = record.SHA256
		}
		if strings.TrimSpace(input.TargetVersion) == "" {
			input.TargetVersion = record.Version
		}
	}

	nodes, err := s.resolveTargetNodes(input, allowedClusters)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("%w: no matching target nodes found", ErrInvalidArgument)
	}

	commandArgs, err := json.Marshal(agentUpgradeCommandArgs{
		PackageURL:    packageURL,
		Checksum:      strings.TrimSpace(input.Checksum),
		TargetVersion: strings.TrimSpace(input.TargetVersion),
		ServiceUnit:   strings.TrimSpace(input.ServiceUnit),
		BinaryPath:    strings.TrimSpace(input.BinaryPath),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal upgrade command: %w", err)
	}

	targetSummary, err := buildAgentUpgradeTargetSummary(input, len(nodes))
	if err != nil {
		return nil, fmt.Errorf("marshal target summary: %w", err)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		if version := strings.TrimSpace(input.TargetVersion); version != "" {
			name = fmt.Sprintf("Agent Upgrade %s (%d nodes)", version, len(nodes))
		} else {
			name = fmt.Sprintf("Agent Upgrade (%d nodes)", len(nodes))
		}
	}

	task := &models.AgentUpgradeTask{
		Name:          name,
		Status:        "pending",
		ArtifactID:    input.ArtifactID,
		PackageURL:    packageURL,
		Checksum:      strings.TrimSpace(input.Checksum),
		TargetVersion: strings.TrimSpace(input.TargetVersion),
		TargetSummary: targetSummary,
		TotalNodes:    len(nodes),
		CreatedBy:     userID,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for _, node := range nodes {
			command := models.RemoteCommand{
				NodeID:    node.ID,
				Action:    "agent_upgrade",
				Args:      string(commandArgs),
				Status:    "pending",
				CreatedBy: userID,
			}
			if err := tx.Create(&command).Error; err != nil {
				return err
			}
			record := models.AgentUpgradeRecord{
				AgentUpgradeTaskID: task.ID,
				NodeID:             node.ID,
				RemoteCommandID:    &command.ID,
				Status:             "pending",
				Message:            "Upgrade queued. Waiting for the next agent heartbeat.",
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := s.db.Preload("Creator").Preload("Artifact").First(task, task.ID).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *agentUpgradeService) List(page, pageSize int, allowedClusters []uint) ([]models.AgentUpgradeTask, int64, error) {
	query := s.db.Preload("Creator").Preload("Artifact")
	if allowedClusters != nil {
		if len(allowedClusters) == 0 {
			return []models.AgentUpgradeTask{}, 0, nil
		}
		query = query.Where("id IN (SELECT DISTINCT agent_upgrade_task_id FROM agent_upgrade_records WHERE node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?))", allowedClusters)
	}

	var total int64
	query.Model(&models.AgentUpgradeTask{}).Count(&total)

	var tasks []models.AgentUpgradeTask
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (s *agentUpgradeService) Get(id uint, allowedClusters []uint) (*models.AgentUpgradeTask, []models.AgentUpgradeRecord, error) {
	var task models.AgentUpgradeTask
	if err := s.db.Preload("Creator").Preload("Artifact").First(&task, id).Error; err != nil {
		return nil, nil, err
	}

	query := s.db.Where("agent_upgrade_task_id = ?", id).
		Preload("Node.Cluster.Region.DataCenter").
		Preload("Node.Cluster.Environment").
		Preload("Node.Environment").
		Preload("RemoteCommand.Creator")
	if allowedClusters != nil {
		if len(allowedClusters) == 0 {
			return nil, nil, gorm.ErrRecordNotFound
		}
		query = query.Where("node_id IN (SELECT id FROM nodes WHERE cluster_id IN ?)", allowedClusters)
	}

	var records []models.AgentUpgradeRecord
	if err := query.Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, nil, err
	}
	if allowedClusters != nil && len(records) == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}
	return &task, records, nil
}

func (s *agentUpgradeService) resolveTargetNodes(input AgentUpgradeTaskInput, allowedClusters []uint) ([]models.Node, error) {
	seen := map[uint]struct{}{}
	result := make([]models.Node, 0)
	appendUnique := func(items []models.Node) {
		for _, item := range items {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			result = append(result, item)
		}
	}

	if len(input.NodeIDs) > 0 {
		query := s.db.Model(&models.Node{}).
			Preload("Cluster.Region.DataCenter").
			Preload("Cluster.Environment").
			Preload("Environment")
		query = applyAllowedClusterScope(query, "nodes.cluster_id", allowedClusters)

		var explicit []models.Node
		if err := query.Where("nodes.id IN ?", uniqueUintSlice(input.NodeIDs)).Find(&explicit).Error; err != nil {
			return nil, err
		}
		if len(explicit) != len(uniqueUintSlice(input.NodeIDs)) {
			return nil, fmt.Errorf("%w: some selected nodes were not found or are outside your scope", ErrForbidden)
		}
		appendUnique(explicit)
	}

	if input.AllMatching {
		query := s.db.Model(&models.Node{}).
			Preload("Cluster.Region.DataCenter").
			Preload("Cluster.Environment").
			Preload("Environment")
		query = applyNodeScopeAndFilters(query, input.Filters, allowedClusters)

		var matched []models.Node
		if err := query.Find(&matched).Error; err != nil {
			return nil, err
		}
		appendUnique(matched)
	}

	return result, nil
}

func buildAgentUpgradeTargetSummary(input AgentUpgradeTaskInput, nodeCount int) (string, error) {
	summary := map[string]interface{}{
		"all_matching": input.AllMatching,
		"node_ids":     uniqueUintSlice(input.NodeIDs),
		"artifact_id":  input.ArtifactID,
		"filters": map[string]string{
			"status":         strings.TrimSpace(input.Filters.Status),
			"cluster_id":     strings.TrimSpace(input.Filters.ClusterID),
			"environment_id": strings.TrimSpace(input.Filters.EnvironmentID),
			"fluent_type":    strings.TrimSpace(input.Filters.FluentType),
			"agent_version":  strings.TrimSpace(input.Filters.AgentVersion),
			"datacenter_id":  strings.TrimSpace(input.Filters.DataCenterID),
			"region_id":      strings.TrimSpace(input.Filters.RegionID),
			"search":         strings.TrimSpace(input.Filters.Search),
		},
		"matched_nodes": nodeCount,
	}
	data, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func uniqueUintSlice(values []uint) []uint {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
