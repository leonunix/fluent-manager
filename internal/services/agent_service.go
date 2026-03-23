package services

import (
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type HeartbeatResponse struct {
	Status        string             `json:"status"`
	ConfigHash    string             `json:"config_hash,omitempty"`
	ConfigContent string             `json:"config_content,omitempty"`
	ConfigID      uint               `json:"config_id,omitempty"`
	Commands      []HeartbeatCommand `json:"commands,omitempty"`
}

type HeartbeatCommand struct {
	ID     uint   `json:"id"`
	Action string `json:"action"`
	Args   string `json:"args"`
}

type AgentService interface {
	Register(nodeUID, hostname, ipAddress, os, agentVersion, fluentType, fluentVersion, labels string) (uint, error)
	Heartbeat(nodeUID, configHash string, metrics map[string]interface{}) (*HeartbeatResponse, error)
	ReportStatus(nodeUID string, configID uint, success bool, message string) error
	ReportCommandResult(commandID uint, status, output string) error
	UploadLogs(nodeUID string, lines []string) error
	GetNodeMetrics(nodeID string) (*models.NodeMetrics, error)
	GetNodeLogs(nodeID string) ([]models.NodeLog, error)
	SendCommand(nodeID string, userID uint, action, args string) (*models.RemoteCommand, error)
	ListNodeCommands(nodeID string) ([]models.RemoteCommand, error)
}

type agentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) AgentService {
	return &agentService{db: db}
}

func (s *agentService) Register(nodeUID, hostname, ipAddress, os, agentVersion, fluentType, fluentVersion, labels string) (uint, error) {
	now := time.Now()
	var node models.Node
	result := s.db.Where("node_uid = ?", nodeUID).First(&node)
	if result.RowsAffected == 0 {
		clusterID := models.AutoAssignCluster(hostname, ipAddress, fluentType, os, labels)
		node = models.Node{
			NodeUID:       nodeUID,
			Hostname:      hostname,
			IPAddress:     ipAddress,
			OS:            os,
			AgentVersion:  agentVersion,
			FluentType:    fluentType,
			FluentVersion: fluentVersion,
			Labels:        labels,
			ClusterID:     clusterID,
			Status:        "online",
			LastHeartbeat: &now,
		}
		if err := s.db.Create(&node).Error; err != nil {
			return 0, err
		}
	} else {
		s.db.Model(&node).Updates(map[string]interface{}{
			"hostname":       hostname,
			"ip_address":     ipAddress,
			"os":             os,
			"agent_version":  agentVersion,
			"fluent_type":    fluentType,
			"fluent_version": fluentVersion,
			"status":         "online",
			"last_heartbeat": &now,
		})
	}
	return node.ID, nil
}

func (s *agentService) Heartbeat(nodeUID, configHash string, metrics map[string]interface{}) (*HeartbeatResponse, error) {
	now := time.Now()
	var node models.Node
	if err := s.db.Where("node_uid = ?", nodeUID).Preload("Config").Preload("Cluster").First(&node).Error; err != nil {
		return nil, err
	}

	s.db.Model(&node).Updates(map[string]interface{}{
		"status":         "online",
		"last_heartbeat": &now,
	})

	if metrics != nil {
		s.storeMetrics(node.ID, metrics)
	}

	resp := &HeartbeatResponse{Status: "ok"}

	// Use EffectiveConfigID for cluster-level config inheritance
	effectiveConfigID := node.EffectiveConfigID()
	if effectiveConfigID != nil {
		var cv models.ConfigVersion
		if err := s.db.First(&cv, *effectiveConfigID).Error; err == nil {
			if cv.Hash != configHash {
				resp.Status = "update_config"
				resp.ConfigHash = cv.Hash
				resp.ConfigContent = cv.Content
				resp.ConfigID = cv.ID
			}
		}
	}

	var pendingCmds []models.RemoteCommand
	s.db.Where("node_id = ? AND status = ?", node.ID, "pending").Find(&pendingCmds)
	if len(pendingCmds) > 0 {
		cmds := make([]HeartbeatCommand, len(pendingCmds))
		ids := make([]uint, len(pendingCmds))
		for i, cmd := range pendingCmds {
			cmds[i] = HeartbeatCommand{ID: cmd.ID, Action: cmd.Action, Args: cmd.Args}
			ids[i] = cmd.ID
		}
		resp.Commands = cmds
		s.db.Model(&models.RemoteCommand{}).Where("id IN ?", ids).Update("status", "delivered")
	}

	return resp, nil
}

func (s *agentService) storeMetrics(nodeID uint, raw map[string]interface{}) {
	var m models.NodeMetrics
	result := s.db.Where("node_id = ?", nodeID).First(&m)
	if result.Error != nil {
		m = models.NodeMetrics{NodeID: nodeID}
		s.db.Create(&m)
	}

	if v, ok := raw["cpu_usage_percent"].(float64); ok {
		m.CPUUsagePercent = v
	}
	if v, ok := raw["mem_total_mb"].(float64); ok {
		m.MemTotalMB = uint64(v)
	}
	if v, ok := raw["mem_used_mb"].(float64); ok {
		m.MemUsedMB = uint64(v)
	}
	if v, ok := raw["mem_usage_percent"].(float64); ok {
		m.MemUsagePercent = v
	}
	if v, ok := raw["disk_total_gb"].(float64); ok {
		m.DiskTotalGB = uint64(v)
	}
	if v, ok := raw["disk_used_gb"].(float64); ok {
		m.DiskUsedGB = uint64(v)
	}
	if v, ok := raw["disk_usage_percent"].(float64); ok {
		m.DiskUsagePercent = v
	}
	if v, ok := raw["load_avg_1"].(float64); ok {
		m.LoadAvg1 = v
	}
	if v, ok := raw["load_avg_5"].(float64); ok {
		m.LoadAvg5 = v
	}
	if v, ok := raw["load_avg_15"].(float64); ok {
		m.LoadAvg15 = v
	}
	if v, ok := raw["fluent_running"].(bool); ok {
		m.FluentRunning = v
	}
	if v, ok := raw["fluent_pid"].(float64); ok {
		m.FluentPID = int(v)
	}
	if v, ok := raw["fluent_cpu_percent"].(float64); ok {
		m.FluentCPUPercent = v
	}
	if v, ok := raw["fluent_mem_mb"].(float64); ok {
		m.FluentMemMB = v
	}
	if v, ok := raw["fluent_open_fds"].(float64); ok {
		m.FluentOpenFDs = int(v)
	}

	s.db.Save(&m)
}

func (s *agentService) ReportStatus(nodeUID string, configID uint, success bool, message string) error {
	var node models.Node
	if err := s.db.Where("node_uid = ?", nodeUID).First(&node).Error; err != nil {
		return err
	}

	status := "success"
	if !success {
		status = "failed"
		s.db.Model(&node).Update("status", "error")
	}

	// Use configID to target the specific deploy record(s) for this config deployment
	result := s.db.Model(&models.DeployRecord{}).
		Where("node_id = ? AND status = ? AND deploy_task_id IN (SELECT id FROM deploy_tasks WHERE config_id = ?)",
			node.ID, "pending", configID).
		Updates(map[string]interface{}{
			"status":  status,
			"message": message,
		})
	if result.Error != nil {
		return result.Error
	}

	// Find the specific deploy task(s) affected and update their stats
	var affectedRecords []models.DeployRecord
	s.db.Where("node_id = ? AND deploy_task_id IN (SELECT id FROM deploy_tasks WHERE config_id = ?)",
		node.ID, configID).
		Order("created_at DESC").Limit(1).Find(&affectedRecords)

	if len(affectedRecords) > 0 {
		taskID := affectedRecords[0].DeployTaskID
		var successCount, failCount, totalPending int64
		s.db.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "success").Count(&successCount)
		s.db.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "failed").Count(&failCount)
		s.db.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "pending").Count(&totalPending)

		updates := map[string]interface{}{
			"success_count": successCount,
			"fail_count":    failCount,
		}
		if totalPending == 0 {
			updates["status"] = "completed"
		}
		if err := s.db.Model(&models.DeployTask{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *agentService) ReportCommandResult(commandID uint, status, output string) error {
	return s.db.Model(&models.RemoteCommand{}).Where("id = ?", commandID).Updates(map[string]interface{}{
		"status": status,
		"output": output,
	}).Error
}

func (s *agentService) UploadLogs(nodeUID string, lines []string) error {
	var node models.Node
	if err := s.db.Where("node_uid = ?", nodeUID).First(&node).Error; err != nil {
		return err
	}
	logEntry := models.NodeLog{
		NodeID:    node.ID,
		Lines:     strings.Join(lines, "\n"),
		LineCount: len(lines),
	}
	return s.db.Create(&logEntry).Error
}

func (s *agentService) GetNodeMetrics(nodeID string) (*models.NodeMetrics, error) {
	var m models.NodeMetrics
	if err := s.db.Where("node_id = ?", nodeID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *agentService) GetNodeLogs(nodeID string) ([]models.NodeLog, error) {
	var logs []models.NodeLog
	err := s.db.Where("node_id = ?", nodeID).Order("created_at DESC").Limit(20).Find(&logs).Error
	return logs, err
}

func (s *agentService) SendCommand(nodeID string, userID uint, action, args string) (*models.RemoteCommand, error) {
	var node models.Node
	if err := s.db.First(&node, nodeID).Error; err != nil {
		return nil, err
	}
	cmd := models.RemoteCommand{
		NodeID:    node.ID,
		Action:    action,
		Args:      args,
		Status:    "pending",
		CreatedBy: userID,
	}
	if err := s.db.Create(&cmd).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (s *agentService) ListNodeCommands(nodeID string) ([]models.RemoteCommand, error) {
	var cmds []models.RemoteCommand
	err := s.db.Where("node_id = ?", nodeID).Preload("Creator").Order("created_at DESC").Limit(50).Find(&cmds).Error
	return cmds, err
}
