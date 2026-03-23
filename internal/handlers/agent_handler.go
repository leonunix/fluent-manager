package handlers

import (
	"net/http"
	"strconv"

	"github.com/fluent-manager/fluent-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	Svc     services.AgentService
	NodeSvc services.NodeService // for scope checks on user-facing endpoints
}

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

	nodeID, err := h.Svc.Register(req.NodeUID, req.Hostname, req.IPAddress, req.OS, req.AgentVersion, req.FluentType, req.FluentVersion, req.Labels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"node_id": nodeID,
		"message": "registered",
	})
}

// --- Heartbeat ---

type HeartbeatRequest struct {
	NodeUID    string                 `json:"node_uid" binding:"required"`
	ConfigHash string                `json:"config_hash"`
	Metrics    map[string]interface{} `json:"metrics"`
}

func (h *AgentHandler) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Svc.Heartbeat(req.NodeUID, req.ConfigHash, req.Metrics)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}

	result := gin.H{"status": resp.Status}
	if resp.Status == "update_config" {
		result["config_hash"] = resp.ConfigHash
		result["config_content"] = resp.ConfigContent
		result["config_id"] = resp.ConfigID
	}
	if resp.Commands != nil {
		result["commands"] = resp.Commands
	}
	c.JSON(http.StatusOK, result)
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
	if err := h.Svc.ReportStatus(req.NodeUID, req.ConfigID, req.Success, req.Message); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status reported"})
}

// --- Report Command Result ---

type CommandResultRequest struct {
	NodeUID   string `json:"node_uid" binding:"required"`
	CommandID uint   `json:"command_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Output    string `json:"output"`
}

func (h *AgentHandler) ReportCommandResult(c *gin.Context) {
	var req CommandResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Svc.ReportCommandResult(req.CommandID, req.Status, req.Output); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	if err := h.Svc.UploadLogs(req.NodeUID, req.Lines); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not registered"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logs received", "line_count": len(req.Lines)})
}

// --- Get Node Metrics (for UI) ---

func (h *AgentHandler) GetNodeMetrics(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Scope check
	node, err := h.NodeSvc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	m, err := h.Svc.GetNodeMetrics(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no metrics for this node"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// --- Get Node Logs (for UI) ---

func (h *AgentHandler) GetNodeLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Scope check
	node, err := h.NodeSvc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	logs, err := h.Svc.GetNodeLogs(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// --- Send Command to Node (from UI) ---

type SendCommandRequest struct {
	Action string `json:"action" binding:"required"`
	Args   string `json:"args"`
}

// allowedActions is the whitelist of permitted remote command actions.
var allowedActions = map[string]bool{
	"restart":     true,
	"reload":      true,
	"stop":        true,
	"start":       true,
	"status":      true,
	"validate":    true,
	"rollback":    true,
	"show_config": true,
}

func (h *AgentHandler) SendCommand(c *gin.Context) {
	var req SendCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !allowedActions[req.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action not allowed; permitted actions: restart, reload, stop, start, status, validate, rollback, show_config"})
		return
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Scope check
	node, err := h.NodeSvc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	userID := c.GetUint("user_id")
	cmd, err := h.Svc.SendCommand(c.Param("id"), userID, req.Action, req.Args)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusCreated, cmd)
}

// --- List Commands for a Node ---

func (h *AgentHandler) ListNodeCommands(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Scope check
	node, err := h.NodeSvc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !checkNodeScope(c, node) {
		return
	}

	cmds, err := h.Svc.ListNodeCommands(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cmds})
}
