package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	ServerURL         string `yaml:"server_url"`
	APIKey            string `yaml:"api_key"`
	NodeUID           string `yaml:"node_uid"`
	FluentType        string `yaml:"fluent_type"`    // fluentbit, fluentd
	FluentConfigPath  string `yaml:"fluent_config_path"`
	FluentRestartCmd  string `yaml:"fluent_restart_cmd"`
	HeartbeatInterval int    `yaml:"heartbeat_interval"` // seconds
	Labels            string `yaml:"labels"`             // JSON key-value
}

func main() {
	cfgPath := flag.String("config", "/etc/fluent-manager/agent.yaml", "agent config file path")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load agent config: %v", err)
	}

	agent := &Agent{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}

	// Register with server
	if err := agent.register(); err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}
	log.Printf("Registered with server as %s", cfg.NodeUID)

	// Start heartbeat loop
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("Agent started (heartbeat every %ds)", cfg.HeartbeatInterval)
	for {
		select {
		case <-ticker.C:
			agent.heartbeat()
		case <-stopCh:
			log.Println("Agent shutting down")
			return
		}
	}
}

func loadConfig(path string) (*AgentConfig, error) {
	cfg := &AgentConfig{
		FluentType:        "fluentbit",
		FluentConfigPath:  "/etc/fluent-bit/fluent-bit.conf",
		FluentRestartCmd:  "systemctl restart fluent-bit",
		HeartbeatInterval: 30,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Generate node UID if not set
	if cfg.NodeUID == "" {
		uidFile := filepath.Join(filepath.Dir(path), ".node_uid")
		data, err := os.ReadFile(uidFile)
		if err == nil {
			cfg.NodeUID = strings.TrimSpace(string(data))
		} else {
			cfg.NodeUID = uuid.New().String()
			_ = os.WriteFile(uidFile, []byte(cfg.NodeUID), 0600)
		}
	}

	return cfg, nil
}

type Agent struct {
	cfg        *AgentConfig
	client     *http.Client
}

func (a *Agent) register() error {
	hostname, _ := os.Hostname()

	body := map[string]string{
		"node_uid":       a.cfg.NodeUID,
		"hostname":       hostname,
		"os":             runtime.GOOS + "/" + runtime.GOARCH,
		"agent_version":  "1.0.0",
		"fluent_type":    a.cfg.FluentType,
		"fluent_version": a.getFluentVersion(),
		"labels":         a.cfg.Labels,
	}

	_, err := a.apiCall("POST", "/api/v1/agent/register", body)
	return err
}

func (a *Agent) heartbeat() {
	currentHash := a.getCurrentConfigHash()

	body := map[string]string{
		"node_uid":    a.cfg.NodeUID,
		"config_hash": currentHash,
	}

	resp, err := a.apiCall("POST", "/api/v1/agent/heartbeat", body)
	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
		return
	}

	var result struct {
		Status        string `json:"status"`
		ConfigHash    string `json:"config_hash"`
		ConfigContent string `json:"config_content"`
		ConfigID      uint   `json:"config_id"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		log.Printf("Failed to parse heartbeat response: %v", err)
		return
	}

	if result.Status == "update_config" {
		log.Printf("New config received (hash=%s), applying...", result.ConfigHash)
		success, msg := a.applyConfig(result.ConfigContent)
		a.reportStatus(result.ConfigID, success, msg)
	}
}

func (a *Agent) applyConfig(content string) (bool, string) {
	// Backup current config
	backupPath := a.cfg.FluentConfigPath + ".bak"
	if data, err := os.ReadFile(a.cfg.FluentConfigPath); err == nil {
		_ = os.WriteFile(backupPath, data, 0644)
	}

	// Write new config
	if err := os.WriteFile(a.cfg.FluentConfigPath, []byte(content), 0644); err != nil {
		return false, fmt.Sprintf("failed to write config: %v", err)
	}

	// Restart fluent service
	parts := strings.Fields(a.cfg.FluentRestartCmd)
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Rollback on failure
		if data, readErr := os.ReadFile(backupPath); readErr == nil {
			_ = os.WriteFile(a.cfg.FluentConfigPath, data, 0644)
			exec.Command(parts[0], parts[1:]...).Run()
		}
		return false, fmt.Sprintf("restart failed: %v, output: %s", err, string(output))
	}

	log.Println("Config applied and service restarted successfully")
	return true, "config applied successfully"
}

func (a *Agent) reportStatus(configID uint, success bool, message string) {
	body := map[string]interface{}{
		"node_uid":  a.cfg.NodeUID,
		"config_id": configID,
		"success":   success,
		"message":   message,
	}
	if _, err := a.apiCall("POST", "/api/v1/agent/report", body); err != nil {
		log.Printf("Failed to report status: %v", err)
	}
}

func (a *Agent) getCurrentConfigHash() string {
	data, err := os.ReadFile(a.cfg.FluentConfigPath)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func (a *Agent) getFluentVersion() string {
	var cmd *exec.Cmd
	if a.cfg.FluentType == "fluentbit" {
		cmd = exec.Command("fluent-bit", "--version")
	} else {
		cmd = exec.Command("fluentd", "--version")
	}
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func (a *Agent) apiCall(method, path string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, a.cfg.ServerURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Key", a.cfg.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
