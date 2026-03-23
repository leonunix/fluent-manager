package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Agent    AgentConfig    `yaml:"agent"`
	Fluent   FluentConfig   `yaml:"fluent"`
	Cache    CacheConfig    `yaml:"cache"`
	Log      LogConfig      `yaml:"log"`
}

type CacheConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TTL      int    `yaml:"ttl"` // seconds
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"` // sqlite, mysql, postgres
	DSN    string `yaml:"dsn"`
}

type AuthConfig struct {
	JWTSecret        string     `yaml:"jwt_secret"`
	TokenExpireHours int        `yaml:"token_expire_hours"`
	LDAP             LDAPConfig `yaml:"ldap"`
	SAML             SAMLConfig `yaml:"saml"`
}

type LDAPConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	UseTLS            bool   `yaml:"use_tls"`
	BindDN            string `yaml:"bind_dn"`
	BindPassword      string `yaml:"bind_password"`
	BaseDN            string `yaml:"base_dn"`
	UserFilter        string `yaml:"user_filter"`
	GroupFilter       string `yaml:"group_filter"`
	GroupSyncStrategy string `yaml:"group_sync_strategy"` // always, first_login
	Attributes        struct {
		Username string `yaml:"username"`
		Email    string `yaml:"email"`
		Name     string `yaml:"name"`
	} `yaml:"attributes"`
}

type SAMLConfig struct {
	Enabled           bool   `yaml:"enabled"`
	IDPMetadata       string `yaml:"idp_metadata"` // URL or file path
	EntityID          string `yaml:"entity_id"`
	ACSURL            string `yaml:"acs_url"`
	CertFile          string `yaml:"cert_file"`
	KeyFile           string `yaml:"key_file"`
	GroupAttribute    string `yaml:"group_attribute"`     // SAML assertion attribute for groups
	GroupSyncStrategy string `yaml:"group_sync_strategy"` // always, first_login
}

type AgentConfig struct {
	HeartbeatInterval   int      `yaml:"heartbeat_interval"` // seconds
	SyncInterval        int      `yaml:"sync_interval"`      // seconds
	APIKey              string   `yaml:"api_key"`
	MetricsInterval     int      `yaml:"metrics_interval"`
	LogUploadInterval   int      `yaml:"log_upload_interval"`
	LogBufferLines      int      `yaml:"log_buffer_lines"`
	HealthPort          int      `yaml:"health_port"`
	MaxRetries          int      `yaml:"max_retries"`
	RetryBaseDelay      int      `yaml:"retry_base_delay"`
	FluentType          string   `yaml:"fluent_type"`
	FluentConfigPath    string   `yaml:"fluent_config_path"`
	FluentConfigDir     string   `yaml:"fluent_config_dir"`
	FluentBinary        string   `yaml:"fluent_binary"`
	FluentServiceUnit   string   `yaml:"fluent_service_unit"`
	FluentRestartCmd    string   `yaml:"fluent_restart_cmd"`
	FluentReloadCmd     string   `yaml:"fluent_reload_cmd"`
	FluentDryRunCmd     string   `yaml:"fluent_dry_run_cmd"`
	FluentLogPath       string   `yaml:"fluent_log_path"`
	FluentExtraFiles    []string `yaml:"fluent_extra_files"`
	FluentMetricsURL    string   `yaml:"fluent_metrics_url"`
	FluentMetricsFormat string   `yaml:"fluent_metrics_format"`
	BackupDir           string   `yaml:"backup_dir"`
	MaxBackups          int      `yaml:"max_backups"`
}

type FluentConfig struct {
	SharedKeySecret string `yaml:"shared_key_secret"`
}

type LogConfig struct {
	Level string `yaml:"level"` // debug, info, warn, error
	File  string `yaml:"file"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server:   ServerConfig{Host: "0.0.0.0", Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Driver: "sqlite", DSN: "fluent_manager.db"},
		Auth:     AuthConfig{JWTSecret: "change-me-in-production", TokenExpireHours: 24},
		Agent: AgentConfig{
			HeartbeatInterval: 30,
			SyncInterval:      60,
			MetricsInterval:   60,
			LogUploadInterval: 120,
			LogBufferLines:    500,
			HealthPort:        9880,
			MaxRetries:        5,
			RetryBaseDelay:    1000,
			MaxBackups:        10,
		},
		Fluent: FluentConfig{},
		Cache:  CacheConfig{Addr: "localhost:6379", TTL: 30},
		Log:    LogConfig{Level: "info"},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
