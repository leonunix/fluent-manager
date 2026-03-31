package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fluent-manager/fluent-manager-agent/collector"
	agentcfg "github.com/fluent-manager/fluent-manager-agent/config"
	"github.com/fluent-manager/fluent-manager-agent/executor"
	"github.com/fluent-manager/fluent-manager-agent/health"
	"github.com/fluent-manager/fluent-manager-agent/logwatch"
	"github.com/fluent-manager/fluent-manager-agent/transport"
)

const Version = "2.0.1"

func main() {
	cfgPath := flag.String("config", "/etc/fluent-manager/agent.yaml", "optional agent config path")
	serverURL := flag.String("server-url", "", "Fluent Manager server URL")
	apiKey := flag.String("api-key", "", "agent API key")
	nodeUID := flag.String("node-uid", "", "optional stable node UID override")
	showVersion := flag.Bool("version", false, "show agent version")
	flag.Parse()

	if *showVersion {
		log.Printf("fluent-manager-agent %s", Version)
		os.Exit(0)
	}

	cfg, err := agentcfg.Load(*cfgPath, agentcfg.Bootstrap{
		ServerURL: *serverURL,
		APIKey:    *apiKey,
		NodeUID:   *nodeUID,
	})
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// --- Transport: handles all server communication with retry ---
	client := transport.New(cfg)

	// --- Register with server ---
	registerResp, err := client.Register(Version)
	if err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}
	if registerResp != nil && registerResp.AgentSettings != nil {
		if err := cfg.ApplyServerSettings(registerResp.AgentSettings); err != nil {
			log.Fatalf("Failed to apply server-delivered agent settings: %v", err)
		}
	}
	log.Printf("Registered as node %s", cfg.Snapshot().NodeUID)

	// --- Metrics Collector: CPU/mem/disk/fluent process ---
	mc := collector.New(cfg)
	mc.Start()
	defer mc.Stop()

	// --- Config Executor: applies & validates configs, runs remote commands ---
	exec := executor.New(cfg)
	exec.EnsureMetricsEnabled()

	// --- Log Watcher: tails fluent logs and buffers for upload ---
	lw := logwatch.New(cfg, client)
	lw.Start()
	defer lw.Stop()

	// --- Local Health API ---
	healthSrv := health.New(cfg, mc)
	healthSrv.Start()
	defer healthSrv.Stop()

	// --- Main heartbeat loop ---
	hb := transport.NewHeartbeat(cfg, client, mc, exec)
	hb.Start()
	defer hb.Stop()

	snapshot := cfg.Snapshot()
	log.Printf("Agent fully started (heartbeat=%ds, metrics=%ds, health=:%d)",
		snapshot.HeartbeatInterval, snapshot.MetricsInterval, snapshot.HealthPort)

	// Wait for shutdown signal
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh
	log.Println("Shutting down agent...")
}
