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
	Log      LogConfig      `yaml:"log"`
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
	JWTSecret       string     `yaml:"jwt_secret"`
	TokenExpireHours int       `yaml:"token_expire_hours"`
	LDAP            LDAPConfig `yaml:"ldap"`
	SAML            SAMLConfig `yaml:"saml"`
}

type LDAPConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	UseTLS       bool   `yaml:"use_tls"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	BaseDN       string `yaml:"base_dn"`
	UserFilter   string `yaml:"user_filter"`
	GroupFilter  string `yaml:"group_filter"`
	Attributes   struct {
		Username string `yaml:"username"`
		Email    string `yaml:"email"`
		Name     string `yaml:"name"`
	} `yaml:"attributes"`
}

type SAMLConfig struct {
	Enabled     bool   `yaml:"enabled"`
	IDPMetadata string `yaml:"idp_metadata"` // URL or file path
	EntityID    string `yaml:"entity_id"`
	ACSURL      string `yaml:"acs_url"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
}

type AgentConfig struct {
	HeartbeatInterval int    `yaml:"heartbeat_interval"` // seconds
	SyncInterval      int    `yaml:"sync_interval"`      // seconds
	APIKey            string `yaml:"api_key"`
}

type LogConfig struct {
	Level string `yaml:"level"` // debug, info, warn, error
	File  string `yaml:"file"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Driver: "sqlite", DSN: "fluent_manager.db"},
		Auth: AuthConfig{JWTSecret: "change-me-in-production", TokenExpireHours: 24},
		Agent: AgentConfig{HeartbeatInterval: 30, SyncInterval: 60},
		Log: LogConfig{Level: "info"},
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
