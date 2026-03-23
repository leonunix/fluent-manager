package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/collector"
	"github.com/fluent-manager/fluent-manager-agent/config"
)

// Server provides a local HTTP endpoint for health checks and diagnostics.
type Server struct {
	cfg     *config.Config
	metrics *collector.Collector

	mu          sync.Mutex
	currentPort int
	currentSrv  *http.Server
	stopCh      chan struct{}
}

func New(cfg *config.Config, metrics *collector.Collector) *Server {
	return &Server{
		cfg:     cfg,
		metrics: metrics,
		stopCh:  make(chan struct{}),
	}
}

// Start runs the health check HTTP server manager.
func (s *Server) Start() {
	go s.loop()
}

func (s *Server) Stop() {
	close(s.stopCh)
	s.shutdownCurrent()
}

func (s *Server) loop() {
	for {
		desiredPort := s.cfg.Snapshot().HealthPort
		s.ensureServer(desiredPort)

		select {
		case <-time.After(2 * time.Second):
		case <-s.stopCh:
			return
		}
	}
}

func (s *Server) ensureServer(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if port == s.currentPort && s.currentSrv != nil {
		return
	}

	if s.currentSrv != nil {
		_ = s.currentSrv.Close()
		s.currentSrv = nil
		s.currentPort = 0
	}

	if port <= 0 {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/config", s.handleConfig)
	mux.HandleFunc("/info", s.handleInfo)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.currentSrv = srv
	s.currentPort = port
	log.Printf("[health] listening on %s", addr)

	go func(port int, server *http.Server) {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[health] server error on port %d: %v", port, err)
		}
	}(port, srv)
}

func (s *Server) shutdownCurrent() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentSrv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.currentSrv.Shutdown(ctx)
	s.currentSrv = nil
	s.currentPort = 0
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
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         status,
		"fluent_running": m.FluentRunning,
		"fluent_pid":     m.FluentPID,
		"timestamp":      time.Now(),
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.metrics.Snapshot())
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	snapshot := s.cfg.Snapshot()
	data, err := os.ReadFile(snapshot.FluentConfigPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(data)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	snapshot := s.cfg.Snapshot()
	runtimeMetadata := map[string]interface{}{}
	if err := json.Unmarshal([]byte(snapshot.RuntimeProfile.Metadata), &runtimeMetadata); err != nil {
		runtimeMetadata["raw"] = snapshot.RuntimeProfile.Metadata
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"node_uid":           snapshot.NodeUID,
		"hostname":           hostname,
		"fluent_type":        snapshot.FluentType,
		"fluent_config":      snapshot.FluentConfigPath,
		"server_url":         snapshot.ServerURL,
		"heartbeat_interval": snapshot.HeartbeatInterval,
		"metrics_interval":   snapshot.MetricsInterval,
		"health_port":        snapshot.HealthPort,
		"metrics_url":        snapshot.FluentMetricsURL,
		"metrics_format":     snapshot.FluentMetricsFormat,
		"runtime_profile":    snapshot.RuntimeProfile,
		"runtime_metadata":   runtimeMetadata,
	})
}
