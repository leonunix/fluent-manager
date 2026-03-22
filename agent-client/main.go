package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	agentcfg "github.com/fluent-manager/fluent-manager-agent/config"
	"github.com/fluent-manager/fluent-manager-agent/collector"
	"github.com/fluent-manager/fluent-manager-agent/executor"
	"github.com/fluent-manager/fluent-manager-agent/health"
	"github.com/fluent-manager/fluent-manager-agent/logwatch"
	"github.com/fluent-manager/fluent-manager-agent/transport"
)

const Version = "2.0.0"

func main() {
	cfgPath := flag.String("config", "/etc/fluent-manager/agent.yaml", "agent config path")
	showVersion := flag.Bool("version", false, "show agent version")
	flag.Parse()

	if *showVersion {
		log.Printf("fluent-manager-agent %s", Version)
		os.Exit(0)
	}

	cfg, err := agentcfg.Load(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// --- Transport: handles all server communication with retry ---
	client := transport.New(cfg)

	// --- Register with server ---
	if err := client.Register(Version); err != nil {
		log.Fatalf("Failed to register with server: %v", err)
	}
	log.Printf("Registered as node %s", cfg.NodeUID)

	// --- Metrics Collector: CPU/mem/disk/fluent process ---
	mc := collector.New(cfg)
	mc.Start()
	defer mc.Stop()

	// --- Config Executor: applies & validates configs, runs remote commands ---
	exec := executor.New(cfg)

	// --- Log Watcher: tails fluent logs and buffers for upload ---
	lw := logwatch.New(cfg, client)
	lw.Start()
	defer lw.Stop()

	// --- Local Health API ---
	healthSrv := health.New(cfg, mc)
	go healthSrv.Start()

	// --- Main heartbeat loop ---
	hb := transport.NewHeartbeat(cfg, client, mc, exec)
	hb.Start()
	defer hb.Stop()

	log.Printf("Agent fully started (heartbeat=%ds, metrics=%ds, health=:%d)",
		cfg.HeartbeatInterval, cfg.MetricsInterval, cfg.HealthPort)

	// Wait for shutdown signal
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh
	log.Println("Shutting down agent...")
}
