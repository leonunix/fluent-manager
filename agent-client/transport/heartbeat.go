package transport

import (
	"encoding/json"
	"log"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/collector"
	"github.com/fluent-manager/fluent-manager-agent/config"
)

// ConfigApplier applies config and runs commands (implemented by executor.Executor).
type ConfigApplier interface {
	CurrentConfigHash() string
	Apply(content string, configID uint) (bool, string)
	RunCommand(action, args string) (string, error)
}

// HeartbeatResponse is the server's response to a heartbeat.
type HeartbeatResponse struct {
	Status        string    `json:"status"` // ok, update_config, exec_command
	ConfigHash    string    `json:"config_hash,omitempty"`
	ConfigContent string    `json:"config_content,omitempty"`
	ConfigID      uint      `json:"config_id,omitempty"`
	Commands      []Command `json:"commands,omitempty"` // remote commands to execute
}

// Command is a remote command from the server.
type Command struct {
	ID     uint   `json:"id"`
	Action string `json:"action"` // restart, reload, status, custom
	Args   string `json:"args"`
}

// Heartbeat handles the periodic heartbeat loop.
type Heartbeat struct {
	cfg     *config.Config
	client  *Client
	metrics *collector.Collector
	applier ConfigApplier
	stopCh  chan struct{}
}

func NewHeartbeat(cfg *config.Config, client *Client, metrics *collector.Collector, applier ConfigApplier) *Heartbeat {
	return &Heartbeat{
		cfg:     cfg,
		client:  client,
		metrics: metrics,
		applier: applier,
		stopCh:  make(chan struct{}),
	}
}

func (h *Heartbeat) Start() {
	go h.loop()
}

func (h *Heartbeat) Stop() {
	close(h.stopCh)
}

func (h *Heartbeat) loop() {
	// Run one heartbeat immediately
	h.beat()

	ticker := time.NewTicker(time.Duration(h.cfg.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.beat()
		case <-h.stopCh:
			return
		}
	}
}

func (h *Heartbeat) beat() {
	snapshot := h.metrics.Snapshot()

	body := map[string]interface{}{
		"node_uid":    h.cfg.NodeUID,
		"config_hash": h.applier.CurrentConfigHash(),
		"metrics":     snapshot,
	}

	resp, err := h.client.APICall("POST", "/api/v1/agent/heartbeat", body)
	if err != nil {
		log.Printf("[heartbeat] failed: %v", err)
		return
	}

	var result HeartbeatResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		log.Printf("[heartbeat] bad response: %v", err)
		return
	}

	// Handle config update
	if result.Status == "update_config" && result.ConfigContent != "" {
		log.Printf("[heartbeat] new config received (hash=%s)", result.ConfigHash)
		success, msg := h.applier.Apply(result.ConfigContent, result.ConfigID)
		h.reportConfigResult(result.ConfigID, success, msg)
	}

	// Handle remote commands
	for _, cmd := range result.Commands {
		log.Printf("[heartbeat] executing remote command: %s (id=%d)", cmd.Action, cmd.ID)
		output, err := h.applier.RunCommand(cmd.Action, cmd.Args)
		status := "success"
		msg := output
		if err != nil {
			status = "failed"
			msg = err.Error() + ": " + output
		}
		h.reportCommandResult(cmd.ID, status, msg)
	}
}

func (h *Heartbeat) reportConfigResult(configID uint, success bool, message string) {
	body := map[string]interface{}{
		"node_uid":  h.cfg.NodeUID,
		"config_id": configID,
		"success":   success,
		"message":   message,
	}
	if _, err := h.client.APICall("POST", "/api/v1/agent/report", body); err != nil {
		log.Printf("[heartbeat] failed to report config status: %v", err)
	}
}

func (h *Heartbeat) reportCommandResult(commandID uint, status, output string) {
	body := map[string]interface{}{
		"node_uid":   h.cfg.NodeUID,
		"command_id": commandID,
		"status":     status,
		"output":     output,
	}
	if _, err := h.client.APICall("POST", "/api/v1/agent/command-result", body); err != nil {
		log.Printf("[heartbeat] failed to report command result: %v", err)
	}
}
