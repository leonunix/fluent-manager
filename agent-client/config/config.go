package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL  string `yaml:"server_url"`
	APIKey     string `yaml:"api_key"`
	NodeUID    string `yaml:"node_uid"`
	ConfigPath string `yaml:"-"` // path to this config file

	// Fluent settings
	FluentType       string   `yaml:"fluent_type"`        // fluentbit, fluentd
	FluentConfigPath string   `yaml:"fluent_config_path"` // main config file
	FluentConfigDir  string   `yaml:"fluent_config_dir"`  // extra config directory (parsers, plugins, etc.)
	FluentBinary     string   `yaml:"fluent_binary"`      // path to fluent-bit or fluentd binary
	FluentRestartCmd string   `yaml:"fluent_restart_cmd"`
	FluentReloadCmd  string   `yaml:"fluent_reload_cmd"`  // graceful reload (SIGHUP)
	FluentDryRunCmd  string   `yaml:"fluent_dry_run_cmd"` // validate config without starting
	FluentLogPath    string   `yaml:"fluent_log_path"`    // fluent log file to tail
	FluentExtraFiles []string `yaml:"fluent_extra_files"` // additional managed config files

	// Intervals
	HeartbeatInterval int `yaml:"heartbeat_interval"` // seconds
	MetricsInterval   int `yaml:"metrics_interval"`   // seconds
	LogUploadInterval int `yaml:"log_upload_interval"` // seconds
	LogBufferLines    int `yaml:"log_buffer_lines"`    // max lines to buffer

	// Local health check
	HealthPort int `yaml:"health_port"`

	// Retry settings
	MaxRetries     int `yaml:"max_retries"`
	RetryBaseDelay int `yaml:"retry_base_delay"` // milliseconds

	// Labels
	Labels string `yaml:"labels"` // JSON key-value

	// Backup
	BackupDir    string `yaml:"backup_dir"`     // directory for config backups
	MaxBackups   int    `yaml:"max_backups"`     // keep N most recent backups
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		FluentType:        "fluentbit",
		FluentConfigPath:  "/etc/fluent-bit/fluent-bit.conf",
		FluentConfigDir:   "",
		FluentBinary:      "",
		FluentRestartCmd:  "systemctl restart fluent-bit",
		FluentReloadCmd:   "systemctl reload fluent-bit",
		FluentDryRunCmd:   "",
		FluentLogPath:     "/var/log/fluent-bit.log",
		HeartbeatInterval: 30,
		MetricsInterval:   60,
		LogUploadInterval: 120,
		LogBufferLines:    500,
		HealthPort:        9880,
		MaxRetries:        5,
		RetryBaseDelay:    1000,
		BackupDir:         "/var/lib/fluent-manager/backups",
		MaxBackups:        10,
		ConfigPath:        path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Auto-detect fluent binary
	if cfg.FluentBinary == "" {
		if cfg.FluentType == "fluentbit" {
			cfg.FluentBinary = "fluent-bit"
		} else {
			cfg.FluentBinary = "fluentd"
		}
	}

	// Auto-detect dry-run command
	if cfg.FluentDryRunCmd == "" {
		if cfg.FluentType == "fluentbit" {
			cfg.FluentDryRunCmd = cfg.FluentBinary + " --dry-run -c " + cfg.FluentConfigPath
		} else {
			cfg.FluentDryRunCmd = cfg.FluentBinary + " --dry-run -c " + cfg.FluentConfigPath
		}
	}

	// Generate or load persistent node UID
	if cfg.NodeUID == "" {
		cfg.NodeUID = loadOrCreateUID(filepath.Dir(path))
	}

	// Ensure backup dir exists
	_ = os.MkdirAll(cfg.BackupDir, 0755)

	return cfg, nil
}

func loadOrCreateUID(dir string) string {
	uidFile := filepath.Join(dir, ".node_uid")
	data, err := os.ReadFile(uidFile)
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	uid := uuid.New().String()
	_ = os.WriteFile(uidFile, []byte(uid), 0600)
	return uid
}
