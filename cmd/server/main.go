package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/agent"
	"github.com/fluent-manager/fluent-manager/internal/auth"
	"github.com/fluent-manager/fluent-manager/internal/cache"
	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/routers"
	"github.com/fluent-manager/fluent-manager/internal/services"
)

// samlDTOToConfig converts a SAMLSettingsDTO to a SAMLConfig, routing
// cert/key data to PEM or file path fields based on content.
func samlDTOToConfig(dto *services.SAMLSettingsDTO) config.SAMLConfig {
	cfg := config.SAMLConfig{
		Enabled:        dto.Enabled,
		IDPMetadata:    dto.IDPMetadata,
		EntityID:       dto.EntityID,
		ACSURL:         dto.ACSURL,
		GroupAttribute: dto.GroupAttribute,
	}
	if strings.HasPrefix(strings.TrimSpace(dto.CertData), "-----BEGIN") {
		cfg.CertPEM = dto.CertData
	} else {
		cfg.CertFile = dto.CertData
	}
	if strings.HasPrefix(strings.TrimSpace(dto.KeyData), "-----BEGIN") {
		cfg.KeyPEM = dto.KeyData
	} else {
		cfg.KeyFile = dto.KeyData
	}
	return cfg
}

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

	// SAMLProvider is initialized empty; it will be loaded from DB after service registry is ready.
	samlProvider := auth.NewSAMLProvider(nil)

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
	disableBootstrapHostKeyChecking := !strings.EqualFold(strings.TrimSpace(cfg.Server.Mode), "release")
	if raw, ok := os.LookupEnv("FM_BOOTSTRAP_DISABLE_HOST_KEY_CHECKING"); ok {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "1", "true", "yes", "on":
			disableBootstrapHostKeyChecking = true
		case "0", "false", "no", "off":
			disableBootstrapHostKeyChecking = false
		default:
			log.Printf("WARNING: ignoring invalid FM_BOOTSTRAP_DISABLE_HOST_KEY_CHECKING value %q", raw)
		}
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
	}, services.BootstrapSettings{
		DefaultAgentAPIKey:     cfg.Agent.APIKey,
		Secret:                 cfg.Auth.JWTSecret,
		DisableHostKeyChecking: disableBootstrapHostKeyChecking,
	})

	// Seed auth settings from config.yaml (only if DB has no settings yet)
	svc.AuthSettings.SeedFromConfig(cfg.Auth)

	// Load SAML from DB settings (covers both seeded-from-yaml and UI-configured scenarios)
	if samlDTO, err := svc.AuthSettings.GetSAMLSettings(); err == nil && samlDTO.Enabled {
		samlCfg := samlDTOToConfig(samlDTO)
		if err := samlProvider.Reload(samlCfg); err != nil {
			log.Printf("WARNING: SAML init from DB settings failed: %v", err)
		} else {
			log.Println("SAML authentication enabled (from DB settings)")
		}
	}

	// Restart channel — setup handler sends on this to trigger server restart
	restartCh := make(chan struct{}, 1)

	// Router
	r := routers.SetupRouter(routers.Deps{
		Cfg:          cfg,
		Svc:          svc,
		JWTSvc:       jwtSvc,
		SAMLProvider: samlProvider,
		CfgPath:      cfgPath,
		RestartCh:    restartCh,
		FrontendFS:   frontendFS,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Fluent Manager server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal or restart request
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	restart := false
	select {
	case <-quit:
		log.Println("Received shutdown signal")
	case <-restartCh:
		log.Println("Received restart request from setup wizard")
		restart = true
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	svc.Bootstrap.Close()
	log.Println("Server stopped")

	if restart {
		log.Println("Re-executing server process...")
		execPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}
		if err := syscall.Exec(execPath, os.Args, os.Environ()); err != nil {
			log.Fatalf("Failed to restart: %v", err)
		}
	}
}
