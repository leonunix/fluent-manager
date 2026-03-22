package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

// Client handles all communication with the Fluent Manager server.
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register sends the initial registration to the server.
func (c *Client) Register(agentVersion string) error {
	hostname, _ := os.Hostname()
	body := map[string]string{
		"node_uid":       c.cfg.NodeUID,
		"hostname":       hostname,
		"ip_address":     getOutboundIP(),
		"os":             runtime.GOOS + "/" + runtime.GOARCH,
		"agent_version":  agentVersion,
		"fluent_type":    c.cfg.FluentType,
		"fluent_version": getFluentVersion(c.cfg),
		"labels":         c.cfg.Labels,
	}
	_, err := c.APICall("POST", "/api/v1/agent/register", body)
	return err
}

// APICall makes an HTTP request to the server with automatic retries.
func (c *Client) APICall(method, path string, body interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(c.cfg.RetryBaseDelay)*math.Pow(2, float64(attempt-1))) * time.Millisecond
			log.Printf("Retry %d/%d after %v...", attempt, c.cfg.MaxRetries, delay)
			time.Sleep(delay)
		}

		req, err := http.NewRequest(method, c.cfg.ServerURL+path, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Key", c.cfg.APIKey)
		req.Header.Set("X-Node-UID", c.cfg.NodeUID)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response: %w", err)
			continue
		}

		// Server errors are retriable; client errors (4xx) are not
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
			continue
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}

	return nil, fmt.Errorf("all %d retries exhausted: %w", c.cfg.MaxRetries, lastErr)
}

// getOutboundIP tries to discover the node's primary outbound IP.
func getOutboundIP() string {
	// Read from /etc/hostname or similar; simplified approach
	interfaces, _ := os.ReadFile("/proc/net/route")
	_ = interfaces // In production, parse routing table
	return ""
}

func getFluentVersion(cfg *config.Config) string {
	out, err := runCmd(cfg.FluentBinary + " --version")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out)
}

func runCmd(cmdStr string) (string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := newCommand(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
