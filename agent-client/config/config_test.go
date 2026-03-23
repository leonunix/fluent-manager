package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithoutConfigFileUsesBootstrapOnly(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	fakeBinary := filepath.Join(binDir, "fluent-bit")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPath)
	t.Setenv("XDG_STATE_HOME", tempDir)

	cfg, err := Load(filepath.Join(tempDir, "missing-agent.yaml"), Bootstrap{
		ServerURL: "http://127.0.0.1:8080",
		APIKey:    "bootstrap-key",
		NodeUID:   "node-from-flag",
	})
	if err != nil {
		t.Fatalf("load bootstrap-only config: %v", err)
	}

	if cfg.ServerURL != "http://127.0.0.1:8080" {
		t.Fatalf("expected server_url override, got %q", cfg.ServerURL)
	}
	if cfg.APIKey != "bootstrap-key" {
		t.Fatalf("expected api_key override, got %q", cfg.APIKey)
	}
	if cfg.NodeUID != "node-from-flag" {
		t.Fatalf("expected node_uid override, got %q", cfg.NodeUID)
	}
	if cfg.FluentBinary != fakeBinary {
		t.Fatalf("expected auto-detected fluent binary %q, got %q", fakeBinary, cfg.FluentBinary)
	}
}

func TestLoadBootstrapOverridesConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	fakeBinary := filepath.Join(binDir, "fluent-bit")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	configPath := filepath.Join(tempDir, "agent.yaml")
	configBody := []byte("server_url: http://from-file:8080\napi_key: file-key\nnode_uid: file-uid\nfluent_type: fluentbit\nfluent_binary: " + fakeBinary + "\n")
	if err := os.WriteFile(configPath, configBody, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv("XDG_STATE_HOME", tempDir)
	cfg, err := Load(configPath, Bootstrap{
		ServerURL: "http://from-flag:8080",
		APIKey:    "flag-key",
		NodeUID:   "flag-uid",
	})
	if err != nil {
		t.Fatalf("load config with overrides: %v", err)
	}

	if cfg.ServerURL != "http://from-flag:8080" {
		t.Fatalf("expected server_url override, got %q", cfg.ServerURL)
	}
	if cfg.APIKey != "flag-key" {
		t.Fatalf("expected api_key override, got %q", cfg.APIKey)
	}
	if cfg.NodeUID != "flag-uid" {
		t.Fatalf("expected node_uid override, got %q", cfg.NodeUID)
	}
}

func TestApplyServerSettingsOverridesRuntimeOptions(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}

	fakeBinary := filepath.Join(binDir, "fluent-bit")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_STATE_HOME", tempDir)

	cfg, err := Load("", Bootstrap{
		ServerURL: "http://127.0.0.1:8080",
		APIKey:    "bootstrap-key",
		NodeUID:   "node-bootstrap",
	})
	if err != nil {
		t.Fatalf("load initial config: %v", err)
	}

	overridePath := filepath.Join(tempDir, "managed.conf")
	if err := os.WriteFile(overridePath, []byte("[SERVICE]\n"), 0o644); err != nil {
		t.Fatalf("write managed config: %v", err)
	}

	err = cfg.ApplyServerSettings(&ServerSettings{
		HeartbeatInterval: 10,
		MetricsInterval:   20,
		FluentType:        "fluentbit",
		FluentBinary:      fakeBinary,
		FluentConfigPath:  overridePath,
		FluentMetricsURL:  "http://127.0.0.1:2020/metrics",
	})
	if err != nil {
		t.Fatalf("apply server settings: %v", err)
	}

	if cfg.HeartbeatInterval != 10 || cfg.MetricsInterval != 20 {
		t.Fatalf("expected server intervals to be applied, got heartbeat=%d metrics=%d", cfg.HeartbeatInterval, cfg.MetricsInterval)
	}
	if cfg.FluentConfigPath != overridePath {
		t.Fatalf("expected server config path to be applied, got %q", cfg.FluentConfigPath)
	}
	if cfg.FluentMetricsURL != "http://127.0.0.1:2020/metrics" {
		t.Fatalf("expected metrics URL override, got %q", cfg.FluentMetricsURL)
	}
	if !cfg.RuntimeProfile.SupportsMetricsAPI {
		t.Fatal("expected runtime profile to be refreshed after server settings")
	}
}
