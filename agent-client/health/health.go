package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/collector"
	"github.com/fluent-manager/fluent-manager-agent/config"
)

// Server provides a local HTTP endpoint for health checks and diagnostics.
type Server struct {
	cfg     *config.Config
	metrics *collector.Collector
}

func New(cfg *config.Config, metrics *collector.Collector) *Server {
	return &Server{cfg: cfg, metrics: metrics}
}

// Start runs the health check HTTP server (blocking).
func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/config", s.handleConfig)
	mux.HandleFunc("/info", s.handleInfo)

	addr := fmt.Sprintf("127.0.0.1:%d", s.cfg.HealthPort)
	log.Printf("[health] listening on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("[health] server error: %v", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	m := s.metrics.Snapshot()
	status := "healthy"
	httpCode := http.StatusOK
	if !m.FluentRunning {
		status = "unhealthy"
		httpCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         status,
		"fluent_running": m.FluentRunning,
		"fluent_pid":     m.FluentPID,
		"timestamp":      time.Now(),
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.metrics.Snapshot())
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(s.cfg.FluentConfigPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"node_uid":          s.cfg.NodeUID,
		"hostname":          hostname,
		"fluent_type":       s.cfg.FluentType,
		"fluent_config":     s.cfg.FluentConfigPath,
		"server_url":        s.cfg.ServerURL,
		"heartbeat_interval": s.cfg.HeartbeatInterval,
		"metrics_interval":  s.cfg.MetricsInterval,
		"health_port":       s.cfg.HealthPort,
	})
}
