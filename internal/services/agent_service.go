package services

import (
	"errors"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/logwriter"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type HeartbeatResponse struct {
	Status        string             `json:"status"`
	ConfigHash    string             `json:"config_hash,omitempty"`
	ConfigContent string             `json:"config_content,omitempty"`
	ConfigID      uint               `json:"config_id,omitempty"`
	Commands      []HeartbeatCommand `json:"commands,omitempty"`
	AgentSettings *AgentSettings     `json:"agent_settings,omitempty"`
}

type HeartbeatCommand struct {
	ID     uint   `json:"id"`
	Action string `json:"action"`
	Args   string `json:"args"`
}

type AgentSettings struct {
	HeartbeatInterval   int      `json:"heartbeat_interval,omitempty"`
	MetricsInterval     int      `json:"metrics_interval,omitempty"`
	LogUploadInterval   int      `json:"log_upload_interval,omitempty"`
	LogBufferLines      int      `json:"log_buffer_lines,omitempty"`
	HealthPort          int      `json:"health_port,omitempty"`
	MaxRetries          int      `json:"max_retries,omitempty"`
	RetryBaseDelay      int      `json:"retry_base_delay,omitempty"`
	FluentType          string   `json:"fluent_type,omitempty"`
	FluentConfigPath    string   `json:"fluent_config_path,omitempty"`
	FluentConfigDir     string   `json:"fluent_config_dir,omitempty"`
	FluentBinary        string   `json:"fluent_binary,omitempty"`
	FluentServiceUnit   string   `json:"fluent_service_unit,omitempty"`
	FluentRestartCmd    string   `json:"fluent_restart_cmd,omitempty"`
	FluentReloadCmd     string   `json:"fluent_reload_cmd,omitempty"`
	FluentDryRunCmd     string   `json:"fluent_dry_run_cmd,omitempty"`
	FluentLogPath       string   `json:"fluent_log_path,omitempty"`
	FluentExtraFiles    []string `json:"fluent_extra_files,omitempty"`
	FluentMetricsURL    string   `json:"fluent_metrics_url,omitempty"`
	FluentMetricsFormat string   `json:"fluent_metrics_format,omitempty"`
	BackupDir           string   `json:"backup_dir,omitempty"`
	MaxBackups          int      `json:"max_backups,omitempty"`
}

type AgentFluentProfileReport struct {
	LoadedPlugins        string `json:"loaded_plugins"`
	SupportsHotReload    bool   `json:"supports_hot_reload"`
	SupportsMultiline    bool   `json:"supports_multiline"`
	SupportsStorageLayer bool   `json:"supports_storage_layer"`
	SupportsForwardTLS   bool   `json:"supports_forward_tls"`
	SupportsMetricsAPI   bool   `json:"supports_metrics_api"`
	Metadata             string `json:"metadata"`
}

type AgentService interface {
	Register(nodeUID, hostname, ipAddress, os, agentVersion, fluentType, fluentVersion, labels string, profile *AgentFluentProfileReport, preferredClusterID *uint, agentAccessKeyID *uint) (uint, error)
	Heartbeat(nodeUID, configHash string, metrics map[string]interface{}, profile *AgentFluentProfileReport) (*HeartbeatResponse, error)
	GetSettingsForNodeID(nodeID uint) (*AgentSettings, error)
	ReportStatus(nodeUID string, configID uint, success bool, message string) error
	ReportCommandResult(commandID uint, status, output string) error
	UploadLogs(nodeUID string, lines []string) error
	GetNodeMetrics(nodeID string) (*models.NodeMetrics, error)
	GetNodeLogs(nodeID string) ([]models.NodeLog, error)
	SendCommand(nodeID string, userID uint, action, args string) (*models.RemoteCommand, error)
	ListNodeCommands(nodeID string) ([]models.RemoteCommand, error)
	GetCommand(nodeID, commandID string) (*models.RemoteCommand, error)
}

type agentService struct {
	db        *gorm.DB
	policySvc AgentPolicyService
	logWriter *logwriter.FileLogger
}

func NewAgentService(db *gorm.DB, policySvc AgentPolicyService, logWriter *logwriter.FileLogger) AgentService {
	return &agentService{db: db, policySvc: policySvc, logWriter: logWriter}
}

func (s *agentService) Register(nodeUID, hostname, ipAddress, os, agentVersion, fluentType, fluentVersion, labels string, profile *AgentFluentProfileReport, preferredClusterID *uint, agentAccessKeyID *uint) (uint, error) {
	now := time.Now()
	var node models.Node
	result := s.db.Where("node_uid = ?", nodeUID).First(&node)
	if result.RowsAffected == 0 && strings.TrimSpace(hostname) != "" {
		// Fallback: match by hostname to prevent duplicate registrations
		// when the agent cannot persist its node_uid (e.g., read-only filesystem).
		result = s.db.Where("hostname = ?", hostname).First(&node)
		if result.RowsAffected > 0 {
			// Adopt the new UID so future lookups use it directly.
			s.db.Model(&node).Update("node_uid", nodeUID)
		}
	}
	if result.RowsAffected == 0 {
		clusterID := preferredClusterID
		if clusterID == nil {
			clusterID = models.AutoAssignCluster(hostname, ipAddress, fluentType, os, labels)
		}
		node = models.Node{
			NodeUID:          nodeUID,
			Hostname:         hostname,
			IPAddress:        ipAddress,
			OS:               os,
			AgentVersion:     agentVersion,
			FluentType:       fluentType,
			FluentVersion:    fluentVersion,
			NodeRole:         models.NodeRoleStandalone,
			Labels:           labels,
			ClusterID:        clusterID,
			AgentAccessKeyID: agentAccessKeyID,
			Status:           "online",
			LastHeartbeat:    &now,
		}
		if err := s.db.Create(&node).Error; err != nil {
			return 0, err
		}
	} else {
		updates := map[string]interface{}{
			"hostname":       hostname,
			"ip_address":     ipAddress,
			"os":             os,
			"agent_version":  agentVersion,
			"fluent_type":    fluentType,
			"fluent_version": fluentVersion,
			"labels":         labels,
			"status":         "online",
			"last_heartbeat": &now,
		}
		if agentAccessKeyID != nil {
			updates["agent_access_key_id"] = *agentAccessKeyID
		}
		if node.ClusterID == nil && preferredClusterID != nil {
			updates["cluster_id"] = *preferredClusterID
		}
		s.db.Model(&node).Updates(updates)
	}
	if err := s.upsertFluentProfile(node.ID, profile, now); err != nil {
		return 0, err
	}
	return node.ID, nil
}

func (s *agentService) Heartbeat(nodeUID, configHash string, metrics map[string]interface{}, profile *AgentFluentProfileReport) (*HeartbeatResponse, error) {
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
	if err := s.upsertFluentProfile(node.ID, profile, now); err != nil {
		return nil, err
	}
	s.updateRuntimeStateFromHeartbeat(&node, configHash, metrics, now)

	settings, err := s.resolveSettingsForNode(&node)
	if err != nil {
		return nil, err
	}
	resp := &HeartbeatResponse{Status: "ok", AgentSettings: settings}

	// Use EffectiveConfigID for cluster-level config inheritance
	effectiveConfigID := node.EffectiveConfigID()
	if effectiveConfigID != nil {
		var cv models.ConfigVersion
		if err := s.db.Preload("Template").First(&cv, *effectiveConfigID).Error; err == nil {
			deliveredContent, deliveredHash := s.resolveDeliveredConfig(&cv, node.FluentType)
			if deliveredHash != configHash {
				resp.Status = "update_config"
				resp.ConfigHash = deliveredHash
				resp.ConfigContent = deliveredContent
				resp.ConfigID = cv.ID
				// Mark stale pending deploy records as failed so they don't stay pending forever.
				// This covers cases where the agent reported failure but the HTTP call was lost,
				// or where the agent keeps failing to apply the config across heartbeats.
				s.expireStalePendingDeployRecords(node.ID, cv.ID, now)
			} else {
				// Hash matches — config is confirmed applied. Close any lingering pending records
				// that were not resolved by the /report endpoint (e.g. agent restarted mid-flight).
				s.resolveDeployRecordsOnHashMatch(node.ID, cv.ID)
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
		s.markUpgradeCommandsDelivered(ids)
	}

	return resp, nil
}

func (s *agentService) GetSettingsForNodeID(nodeID uint) (*AgentSettings, error) {
	var node models.Node
	if err := s.db.Preload("Cluster.Environment").Preload("Environment").First(&node, nodeID).Error; err != nil {
		return nil, err
	}
	return s.resolveSettingsForNode(&node)
}

func (s *agentService) resolveSettingsForNode(node *models.Node) (*AgentSettings, error) {
	if s.policySvc == nil {
		settings := AgentSettings{}
		return &settings, nil
	}
	resolved, err := s.policySvc.ResolveForNode(node)
	if err != nil {
		return nil, err
	}
	return &resolved.Settings, nil
}

func (s *agentService) upsertFluentProfile(nodeID uint, profile *AgentFluentProfileReport, reportedAt time.Time) error {
	if profile == nil {
		return nil
	}

	var existing models.NodeFluentProfile
	err := s.db.Where("node_id = ?", nodeID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = models.NodeFluentProfile{NodeID: nodeID}
		if err := s.db.Create(&existing).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"loaded_plugins":         strings.TrimSpace(profile.LoadedPlugins),
		"supports_hot_reload":    profile.SupportsHotReload,
		"supports_multiline":     profile.SupportsMultiline,
		"supports_storage_layer": profile.SupportsStorageLayer,
		"supports_forward_tls":   profile.SupportsForwardTLS,
		"supports_metrics_api":   profile.SupportsMetricsAPI,
		"metadata":               profile.Metadata,
		"last_reported_at":       &reportedAt,
	}
	return s.db.Model(&existing).Updates(updates).Error
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
	s.updateRuntimeStateFromDeployResult(node.ID, configID, success, message)

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

func (s *agentService) updateRuntimeStateFromHeartbeat(node *models.Node, configHash string, metrics map[string]interface{}, now time.Time) {
	if node == nil {
		return
	}

	desiredHash := ""
	effectiveConfigID := node.EffectiveConfigID()
	if effectiveConfigID != nil {
		var version models.ConfigVersion
		if err := s.db.Preload("Template").First(&version, *effectiveConfigID).Error; err == nil {
			_, desiredHash = s.resolveDeliveredConfig(&version, node.FluentType)
		}
	}

	state := models.NodeRuntimeState{
		NodeID:              node.ID,
		DesiredConfigHash:   desiredHash,
		EffectiveConfigHash: configHash,
		LastSyncAt:          &now,
	}
	if metrics != nil {
		if v, ok := metrics["queue_depth"].(float64); ok {
			state.QueueDepth = int(v)
		}
		if v, ok := metrics["retry_count"].(float64); ok {
			state.RetryCount = int(v)
		}
		if v, ok := metrics["flush_latency_ms"].(float64); ok {
			state.FlushLatencyMS = int(v)
		}
		if v, ok := metrics["input_status"].(string); ok {
			state.InputStatus = v
		}
		if v, ok := metrics["output_status"].(string); ok {
			state.OutputStatus = v
		}
	}

	var existing models.NodeRuntimeState
	err := s.db.Where("node_id = ?", node.ID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_ = s.db.Create(&state).Error
		}
		return
	}

	updates := map[string]interface{}{
		"desired_config_hash":   state.DesiredConfigHash,
		"effective_config_hash": state.EffectiveConfigHash,
		"last_sync_at":          state.LastSyncAt,
	}
	if state.QueueDepth > 0 {
		updates["queue_depth"] = state.QueueDepth
	}
	if state.RetryCount > 0 {
		updates["retry_count"] = state.RetryCount
	}
	if state.FlushLatencyMS > 0 {
		updates["flush_latency_ms"] = state.FlushLatencyMS
	}
	if state.InputStatus != "" {
		updates["input_status"] = state.InputStatus
	}
	if state.OutputStatus != "" {
		updates["output_status"] = state.OutputStatus
	}
	_ = s.db.Model(&existing).Updates(updates).Error
}

func (s *agentService) updateRuntimeStateFromDeployResult(nodeID, configID uint, success bool, message string) {
	now := time.Now()
	desiredHash := ""
	if configID != 0 {
		var version models.ConfigVersion
		if err := s.db.Preload("Template").First(&version, configID).Error; err == nil {
			_, desiredHash = s.resolveDeliveredConfig(&version, "")
		}
	}

	var existing models.NodeRuntimeState
	err := s.db.Where("node_id = ?", nodeID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			state := models.NodeRuntimeState{
				NodeID:            nodeID,
				DesiredConfigHash: desiredHash,
				LastReloadAt:      &now,
			}
			if !success {
				state.LastError = message
			}
			_ = s.db.Create(&state).Error
		}
		return
	}

	updates := map[string]interface{}{
		"desired_config_hash": desiredHash,
		"last_reload_at":      &now,
	}
	if success {
		updates["last_error"] = ""
	} else {
		updates["last_error"] = message
	}
	_ = s.db.Model(&existing).Updates(updates).Error
}

func (s *agentService) resolveDeliveredConfig(version *models.ConfigVersion, fallbackFluentType string) (string, string) {
	if version == nil {
		return "", ""
	}

	content := version.Content
	hash := version.Hash
	if version.SourceType != "module_assembly" || strings.TrimSpace(version.SourceModules) == "" {
		return content, hash
	}

	refs, err := parseStoredRenderModuleRefs(version.SourceModules)
	if err != nil || len(refs) == 0 {
		return content, hash
	}

	fluentType := strings.TrimSpace(fallbackFluentType)
	baseVariables := "{}"
	if version.Template != nil {
		if strings.TrimSpace(version.Template.FluentType) != "" {
			fluentType = strings.TrimSpace(version.Template.FluentType)
		}
		if strings.TrimSpace(version.Template.Variables) != "" {
			baseVariables = version.Template.Variables
		}
	}
	if fluentType == "" {
		return content, hash
	}

	renderedContent, _, err := (&configService{db: s.db}).renderModulesForRuntime(fluentType, refs, baseVariables)
	if err != nil || strings.TrimSpace(renderedContent) == "" {
		return content, hash
	}
	return renderedContent, models.HashConfig(renderedContent)
}

func (s *agentService) ReportCommandResult(commandID uint, status, output string) error {
	var cmd models.RemoteCommand
	if err := s.db.First(&cmd, commandID).Error; err != nil {
		return err
	}
	const maxOutputBytes = 512 * 1024
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
	}
	if err := s.db.Model(&cmd).Updates(map[string]interface{}{
		"status": status,
		"output": output,
	}).Error; err != nil {
		return err
	}
	s.updateUpgradeRecordForCommandResult(cmd.ID, status, output)
	return nil
}

func (s *agentService) UploadLogs(nodeUID string, lines []string) error {
	var node models.Node
	if err := s.db.Where("node_uid = ?", nodeUID).First(&node).Error; err != nil {
		return err
	}
	joined := strings.Join(lines, "\n")
	if s.logWriter != nil {
		s.logWriter.WriteNodeLog(node.ID, joined, len(lines))
	}
	logEntry := models.NodeLog{
		NodeID:    node.ID,
		Lines:     joined,
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

func (s *agentService) GetCommand(nodeID, commandID string) (*models.RemoteCommand, error) {
	var cmd models.RemoteCommand
	err := s.db.Where("id = ? AND node_id = ?", commandID, nodeID).Preload("Creator").First(&cmd).Error
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (s *agentService) markUpgradeCommandsDelivered(commandIDs []uint) {
	if len(commandIDs) == 0 {
		return
	}

	result := s.db.Model(&models.AgentUpgradeRecord{}).
		Where("remote_command_id IN ? AND status = ?", commandIDs, "pending").
		Updates(map[string]interface{}{
			"status":  "running",
			"message": "Upgrade command delivered. Waiting for the node to apply and restart.",
		})
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}

	var taskIDs []uint
	s.db.Model(&models.AgentUpgradeRecord{}).
		Where("remote_command_id IN ?", commandIDs).
		Distinct().
		Pluck("agent_upgrade_task_id", &taskIDs)
	s.recalculateUpgradeTasks(taskIDs)
}

func (s *agentService) updateUpgradeRecordForCommandResult(commandID uint, status, output string) {
	recordStatus := "failed"
	message := strings.TrimSpace(output)
	if status == "success" || status == "completed" {
		recordStatus = "completed"
		if message == "" {
			message = "Agent upgrade completed."
		}
	} else if message == "" {
		message = "Agent upgrade failed."
	}

	result := s.db.Model(&models.AgentUpgradeRecord{}).
		Where("remote_command_id = ?", commandID).
		Updates(map[string]interface{}{
			"status":         recordStatus,
			"message":        message,
			"output_excerpt": truncateString(output, 4000),
		})
	if result.Error != nil || result.RowsAffected == 0 {
		return
	}

	var taskIDs []uint
	s.db.Model(&models.AgentUpgradeRecord{}).
		Where("remote_command_id = ?", commandID).
		Distinct().
		Pluck("agent_upgrade_task_id", &taskIDs)
	s.recalculateUpgradeTasks(taskIDs)
}

// resolveDeployRecordsOnHashMatch closes pending deploy records with "success" when the
// agent's reported config hash already matches the desired hash (config confirmed applied).
// This handles cases where the agent applied the config but the /report call was lost.
func (s *agentService) resolveDeployRecordsOnHashMatch(nodeID, configVersionID uint) {
	s.settleDeployRecords(nodeID, configVersionID, "success", "config confirmed via heartbeat")
}

// expireStalePendingDeployRecords marks pending deploy records as failed when the agent
// keeps reporting a mismatched hash, meaning it has not successfully applied the config.
// Only records older than two heartbeat cycles (60s) are expired to avoid racing with
// an in-flight apply+report sequence.
func (s *agentService) expireStalePendingDeployRecords(nodeID, configVersionID uint, now time.Time) {
	cutoff := now.Add(-60 * time.Second)
	s.db.Model(&models.DeployRecord{}).
		Where("node_id = ? AND status = 'pending' AND created_at < ? AND deploy_task_id IN (SELECT id FROM deploy_tasks WHERE config_id = ?)",
			nodeID, cutoff, configVersionID).
		Updates(map[string]interface{}{
			"status":  "failed",
			"message": "config not applied after multiple heartbeats",
		})
	s.recalculateDeployTasksForNode(nodeID, configVersionID)
}

func (s *agentService) settleDeployRecords(nodeID, configVersionID uint, status, message string) {
	result := s.db.Model(&models.DeployRecord{}).
		Where("node_id = ? AND status = 'pending' AND deploy_task_id IN (SELECT id FROM deploy_tasks WHERE config_id = ?)",
			nodeID, configVersionID).
		Updates(map[string]interface{}{
			"status":  status,
			"message": message,
		})
	if result.RowsAffected == 0 {
		return
	}
	s.recalculateDeployTasksForNode(nodeID, configVersionID)
}

func (s *agentService) recalculateDeployTasksForNode(nodeID, configVersionID uint) {
	var taskIDs []uint
	s.db.Model(&models.DeployRecord{}).
		Where("node_id = ? AND deploy_task_id IN (SELECT id FROM deploy_tasks WHERE config_id = ?)", nodeID, configVersionID).
		Distinct().Pluck("deploy_task_id", &taskIDs)

	for _, taskID := range taskIDs {
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
		s.db.Model(&models.DeployTask{}).Where("id = ?", taskID).Updates(updates)
	}
}

func (s *agentService) recalculateUpgradeTasks(taskIDs []uint) {
	for _, taskID := range uniqueUintSlice(taskIDs) {
		if taskID == 0 {
			continue
		}
		var task models.AgentUpgradeTask
		if err := s.db.First(&task, taskID).Error; err != nil {
			continue
		}

		var pendingCount int64
		var runningCount int64
		var successCount int64
		var failCount int64

		s.db.Model(&models.AgentUpgradeRecord{}).Where("agent_upgrade_task_id = ? AND status = ?", taskID, "pending").Count(&pendingCount)
		s.db.Model(&models.AgentUpgradeRecord{}).Where("agent_upgrade_task_id = ? AND status = ?", taskID, "running").Count(&runningCount)
		s.db.Model(&models.AgentUpgradeRecord{}).Where("agent_upgrade_task_id = ? AND status = ?", taskID, "completed").Count(&successCount)
		s.db.Model(&models.AgentUpgradeRecord{}).Where("agent_upgrade_task_id = ? AND status = ?", taskID, "failed").Count(&failCount)

		status := "running"
		switch {
		case successCount == int64(task.TotalNodes) && task.TotalNodes > 0:
			status = "completed"
		case failCount == int64(task.TotalNodes) && task.TotalNodes > 0:
			status = "failed"
		case pendingCount == int64(task.TotalNodes) && task.TotalNodes > 0:
			status = "pending"
		case pendingCount == 0 && runningCount == 0 && failCount > 0 && successCount > 0:
			status = "partial"
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":        status,
			"success_count": int(successCount),
			"fail_count":    int(failCount),
		}
		if task.StartedAt == nil && status != "pending" {
			updates["started_at"] = &now
		}
		if status == "completed" || status == "failed" || status == "partial" {
			updates["finished_at"] = &now
		} else {
			updates["finished_at"] = nil
		}

		s.db.Model(&task).Updates(updates)
	}
}
