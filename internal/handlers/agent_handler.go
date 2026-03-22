package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct{}

// --- Register ---

type RegisterRequest struct {
	NodeUID       string `json:"node_uid" binding:"required"`
	Hostname      string `json:"hostname" binding:"required"`
	IPAddress     string `json:"ip_address"`
	OS            string `json:"os"`
	AgentVersion  string `json:"agent_version"`
	FluentType    string `json:"fluent_type"`
	FluentVersion string `json:"fluent_version"`
	Labels        string `json:"labels"`
}

func (h *AgentHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	var node models.Node
	result := models.DB.Where("node_uid = ?", req.NodeUID).First(&node)
	if result.RowsAffected == 0 {
		// Auto-assign to cluster based on match rules
		clusterID := models.AutoAssignCluster(req.Hostname, req.IPAddress, req.FluentType, req.OS, req.Labels)

		node = models.Node{
			NodeUID:       req.NodeUID,
			Hostname:      req.Hostname,
			IPAddress:     req.IPAddress,
			OS:            req.OS,
			AgentVersion:  req.AgentVersion,
			FluentType:    req.FluentType,
			FluentVersion: req.FluentVersion,
			Labels:        req.Labels,
			ClusterID:     clusterID,
			Status:        "online",
			LastHeartbeat: &now,
		}
		models.DB.Create(&node)
	} else {
		models.DB.Model(&node).Updates(map[string]interface{}{
			"hostname":       req.Hostname,
			"ip_address":     req.IPAddress,
			"os":             req.OS,
			"agent_version":  req.AgentVersion,
			"fluent_type":    req.FluentType,
			"fluent_version": req.FluentVersion,
			"status":         "online",
			"last_heartbeat": &now,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id": node.ID,
		"message": "registered",
	})
}

// --- Heartbeat (enhanced: receives metrics, returns pending commands) ---

type HeartbeatRequest struct {
	NodeUID    string                 `json:"node_uid" binding:"required"`
	ConfigHash string                `json:"config_hash"`
	Metrics    map[string]interface{} `json:"metrics"` // system+fluent metrics
}

type HeartbeatCommandResp struct {
	ID     uint   `json:"id"`
	Action string `json:"action"`
	Args   string `json:"args"`
}

func (h *AgentHandler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	var node models.Node
	if err := models.DB.Where("node_uid = ?", req.NodeUID).Preload("Config").First(&node).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}

	// Update heartbeat
	models.DB.Model(&node).Updates(map[string]interface{}{
		"status":         "online",
		"last_heartbeat": &now,
	})

	// Store metrics
	if req.Metrics != nil {
		h.storeMetrics(node.ID, req.Metrics)
	}

	// Build response
	resp := gin.H{"status": "ok"}

	// Check if config needs updating
	if node.Config != nil && node.Config.Hash != req.ConfigHash {
		resp["status"] = "update_config"
		resp["config_hash"] = node.Config.Hash
		resp["config_content"] = node.Config.Content
		resp["config_id"] = node.Config.ID
	}

	// Fetch and deliver pending commands
	var pendingCmds []models.RemoteCommand
	models.DB.Where("node_id = ? AND status = ?", node.ID, "pending").Find(&pendingCmds)
	if len(pendingCmds) > 0 {
		cmds := make([]HeartbeatCommandResp, len(pendingCmds))
		ids := make([]uint, len(pendingCmds))
		for i, cmd := range pendingCmds {
			cmds[i] = HeartbeatCommandResp{ID: cmd.ID, Action: cmd.Action, Args: cmd.Args}
			ids[i] = cmd.ID
		}
		resp["commands"] = cmds
		// Mark as delivered
		models.DB.Model(&models.RemoteCommand{}).Where("id IN ?", ids).Update("status", "delivered")
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AgentHandler) storeMetrics(nodeID uint, raw map[string]interface{}) {
	m := models.NodeMetrics{NodeID: nodeID}

	// Upsert: find or create
	models.DB.Where("node_id = ?", nodeID).FirstOrCreate(&m)

	updates := map[string]interface{}{}
	if v, ok := raw["cpu_usage_percent"].(float64); ok {
		updates["cpu_usage_percent"] = v
	}
	if v, ok := raw["mem_total_mb"].(float64); ok {
		updates["mem_total_mb"] = uint64(v)
	}
	if v, ok := raw["mem_used_mb"].(float64); ok {
		updates["mem_used_mb"] = uint64(v)
	}
	if v, ok := raw["mem_usage_percent"].(float64); ok {
		updates["mem_usage_percent"] = v
	}
	if v, ok := raw["disk_total_gb"].(float64); ok {
		updates["disk_total_gb"] = uint64(v)
	}
	if v, ok := raw["disk_used_gb"].(float64); ok {
		updates["disk_used_gb"] = uint64(v)
	}
	if v, ok := raw["disk_usage_percent"].(float64); ok {
		updates["disk_usage_percent"] = v
	}
	if v, ok := raw["load_avg_1"].(float64); ok {
		updates["load_avg_1"] = v
	}
	if v, ok := raw["load_avg_5"].(float64); ok {
		updates["load_avg_5"] = v
	}
	if v, ok := raw["load_avg_15"].(float64); ok {
		updates["load_avg_15"] = v
	}
	if v, ok := raw["fluent_running"].(bool); ok {
		updates["fluent_running"] = v
	}
	if v, ok := raw["fluent_pid"].(float64); ok {
		updates["fluent_pid"] = int(v)
	}
	if v, ok := raw["fluent_cpu_percent"].(float64); ok {
		updates["fluent_cpu_percent"] = v
	}
	if v, ok := raw["fluent_mem_mb"].(float64); ok {
		updates["fluent_mem_mb"] = v
	}
	if v, ok := raw["fluent_open_fds"].(float64); ok {
		updates["fluent_open_fds"] = int(v)
	}

	if len(updates) > 0 {
		models.DB.Model(&m).Updates(updates)
	}
}

// --- Report Config Status ---

type ReportStatusRequest struct {
	NodeUID  string `json:"node_uid" binding:"required"`
	ConfigID uint   `json:"config_id"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

func (h *AgentHandler) ReportStatus(c *gin.Context) {
	var req ReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var node models.Node
	if err := models.DB.Where("node_uid = ?", req.NodeUID).First(&node).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}

	status := "success"
	if !req.Success {
		status = "failed"
		models.DB.Model(&node).Update("status", "error")
	}

	models.DB.Model(&models.DeployRecord{}).
		Where("node_id = ? AND status = ?", node.ID, "pending").
		Updates(map[string]interface{}{
			"status":  status,
			"message": req.Message,
		})

	// Update deploy task counts
	var records []models.DeployRecord
	models.DB.Where("node_id = ?", node.ID).Order("created_at DESC").Limit(1).Find(&records)
	if len(records) > 0 {
		taskID := records[0].DeployTaskID
		var successCount, failCount, totalPending int64
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "success").Count(&successCount)
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "failed").Count(&failCount)
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "pending").Count(&totalPending)

		updates := map[string]interface{}{
			"success_count": successCount,
			"fail_count":    failCount,
		}
		if totalPending == 0 {
			updates["status"] = "completed"
		}
		models.DB.Model(&models.DeployTask{}).Where("id = ?", taskID).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"message": "status reported"})
}

// --- Report Command Result ---

type CommandResultRequest struct {
	NodeUID   string `json:"node_uid" binding:"required"`
	CommandID uint   `json:"command_id" binding:"required"`
	Status    string `json:"status" binding:"required"` // success, failed
	Output    string `json:"output"`
}

func (h *AgentHandler) ReportCommandResult(c *gin.Context) {
	var req CommandResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&models.RemoteCommand{}).Where("id = ?", req.CommandID).Updates(map[string]interface{}{
		"status": req.Status,
		"output": req.Output,
	})

	c.JSON(http.StatusOK, gin.H{"message": "command result recorded"})
}

// --- Upload Logs ---

type UploadLogsRequest struct {
	NodeUID string   `json:"node_uid" binding:"required"`
	Lines   []string `json:"lines" binding:"required"`
}

func (h *AgentHandler) UploadLogs(c *gin.Context) {
	var req UploadLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var node models.Node
	if err := models.DB.Where("node_uid = ?", req.NodeUID).First(&node).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}

	logEntry := models.NodeLog{
		NodeID:    node.ID,
		Lines:     strings.Join(req.Lines, "\n"),
		LineCount: len(req.Lines),
	}
	models.DB.Create(&logEntry)

	c.JSON(http.StatusOK, gin.H{"message": "logs received", "line_count": len(req.Lines)})
}

// --- Get Node Metrics (for UI) ---

func (h *AgentHandler) GetNodeMetrics(c *gin.Context) {
	nodeID := c.Param("id")
	var m models.NodeMetrics
	if err := models.DB.Where("node_id = ?", nodeID).First(&m).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no metrics for this node"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// --- Get Node Logs (for UI) ---

func (h *AgentHandler) GetNodeLogs(c *gin.Context) {
	nodeID := c.Param("id")
	var logs []models.NodeLog
	models.DB.Where("node_id = ?", nodeID).Order("created_at DESC").Limit(20).Find(&logs)
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// --- Send Command to Node (from UI) ---

type SendCommandRequest struct {
	Action string `json:"action" binding:"required"`
	Args   string `json:"args"`
}

func (h *AgentHandler) SendCommand(c *gin.Context) {
	nodeID := c.Param("id")
	var req SendCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var node models.Node
	if err := models.DB.First(&node, nodeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	userID := c.GetUint("user_id")
	cmd := models.RemoteCommand{
		NodeID:    node.ID,
		Action:    req.Action,
		Args:      req.Args,
		Status:    "pending",
		CreatedBy: userID,
	}
	models.DB.Create(&cmd)

	c.JSON(http.StatusCreated, cmd)
}

// --- List Commands for a Node ---

func (h *AgentHandler) ListNodeCommands(c *gin.Context) {
	nodeID := c.Param("id")
	var cmds []models.RemoteCommand
	models.DB.Where("node_id = ?", nodeID).Preload("Creator").Order("created_at DESC").Limit(50).Find(&cmds)
	c.JSON(http.StatusOK, gin.H{"data": cmds})
}
