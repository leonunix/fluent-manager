package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fluent-manager/fluent-manager/internal/agent"
	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/cache"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/routers"
	"github.com/fluent-manager/fluent-manager/internal/services"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database
	if err := models.InitDB(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Auth services
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenExpireHours)

	var samlAuth *auth.SAMLAuth
	if cfg.Auth.SAML.Enabled {
		samlAuth, err = auth.NewSAMLAuth(cfg.Auth.SAML)
		if err != nil {
			log.Printf("WARNING: SAML init failed: %v", err)
		} else {
			log.Println("SAML authentication enabled")
		}
	}

	// Cache (Redis)
	cache.Init(&cfg.Cache)

	// Node monitor
	monitor := agent.NewMonitor(cfg.Agent.HeartbeatInterval)
	monitor.Start()
	defer monitor.Stop()

	// Service layer
	fluentSharedKeySecret := cfg.Fluent.SharedKeySecret
	if fluentSharedKeySecret == "" {
		fluentSharedKeySecret = cfg.Auth.JWTSecret
	}
	svc := services.NewRegistry(models.DB, fluentSharedKeySecret, services.AgentSettings{
		HeartbeatInterval:   cfg.Agent.HeartbeatInterval,
		MetricsInterval:     cfg.Agent.MetricsInterval,
		LogUploadInterval:   cfg.Agent.LogUploadInterval,
		LogBufferLines:      cfg.Agent.LogBufferLines,
		HealthPort:          cfg.Agent.HealthPort,
		MaxRetries:          cfg.Agent.MaxRetries,
		RetryBaseDelay:      cfg.Agent.RetryBaseDelay,
		FluentType:          cfg.Agent.FluentType,
		FluentConfigPath:    cfg.Agent.FluentConfigPath,
		FluentConfigDir:     cfg.Agent.FluentConfigDir,
		FluentBinary:        cfg.Agent.FluentBinary,
		FluentServiceUnit:   cfg.Agent.FluentServiceUnit,
		FluentRestartCmd:    cfg.Agent.FluentRestartCmd,
		FluentReloadCmd:     cfg.Agent.FluentReloadCmd,
		FluentDryRunCmd:     cfg.Agent.FluentDryRunCmd,
		FluentLogPath:       cfg.Agent.FluentLogPath,
		FluentExtraFiles:    cfg.Agent.FluentExtraFiles,
		FluentMetricsURL:    cfg.Agent.FluentMetricsURL,
		FluentMetricsFormat: cfg.Agent.FluentMetricsFormat,
		BackupDir:           cfg.Agent.BackupDir,
		MaxBackups:          cfg.Agent.MaxBackups,
	})

	// Router
	r := routers.SetupRouter(routers.Deps{
		Cfg:      cfg,
		Svc:      svc,
		JWTSvc:   jwtSvc,
		SAMLAuth: samlAuth,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Fluent Manager server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
