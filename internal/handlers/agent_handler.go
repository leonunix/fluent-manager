package handlers

import (
	"net/http"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct{}

// RegisterRequest is sent by agent on first contact.
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
		node = models.Node{
			NodeUID:       req.NodeUID,
			Hostname:      req.Hostname,
			IPAddress:     req.IPAddress,
			OS:            req.OS,
			AgentVersion:  req.AgentVersion,
			FluentType:    req.FluentType,
			FluentVersion: req.FluentVersion,
			Labels:        req.Labels,
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

// HeartbeatRequest is sent periodically by the agent.
type HeartbeatRequest struct {
	NodeUID    string `json:"node_uid" binding:"required"`
	ConfigHash string `json:"config_hash"` // current config hash on agent
}

// HeartbeatResponse tells the agent what to do.
type HeartbeatResponse struct {
	Status       string `json:"status"`                  // ok, update_config
	ConfigHash   string `json:"config_hash,omitempty"`   // new config hash if changed
	ConfigContent string `json:"config_content,omitempty"` // new config content
	ConfigID     uint   `json:"config_id,omitempty"`
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

	models.DB.Model(&node).Updates(map[string]interface{}{
		"status":         "online",
		"last_heartbeat": &now,
	})

	resp := HeartbeatResponse{Status: "ok"}

	// Check if config needs updating
	if node.Config != nil && node.Config.Hash != req.ConfigHash {
		resp.Status = "update_config"
		resp.ConfigHash = node.Config.Hash
		resp.ConfigContent = node.Config.Content
		resp.ConfigID = node.Config.ID
	}

	c.JSON(http.StatusOK, resp)
}

// ReportStatus is sent by agent after applying a config.
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

	// Update deploy records
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
		var successCount, failCount int64
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "success").Count(&successCount)
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "failed").Count(&failCount)

		updates := map[string]interface{}{
			"success_count": successCount,
			"fail_count":    failCount,
		}

		var totalPending int64
		models.DB.Model(&models.DeployRecord{}).Where("deploy_task_id = ? AND status = ?", taskID, "pending").Count(&totalPending)
		if totalPending == 0 {
			updates["status"] = "completed"
		}

		models.DB.Model(&models.DeployTask{}).Where("id = ?", taskID).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"message": "status reported"})
}
