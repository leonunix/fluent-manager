package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type RuntimeProfile struct {
	LoadedPlugins        string `json:"loaded_plugins"`
	SupportsHotReload    bool   `json:"supports_hot_reload"`
	SupportsMultiline    bool   `json:"supports_multiline"`
	SupportsStorageLayer bool   `json:"supports_storage_layer"`
	SupportsForwardTLS   bool   `json:"supports_forward_tls"`
	SupportsMetricsAPI   bool   `json:"supports_metrics_api"`
	Metadata             string `json:"metadata"`
}

type runtimeMetadata struct {
	RuntimeType       string   `json:"runtime_type"`
	BinaryPath        string   `json:"binary_path"`
	ConfigPath        string   `json:"config_path"`
	ConfigDir         string   `json:"config_dir,omitempty"`
	LogPath           string   `json:"log_path,omitempty"`
	ServiceUnit       string   `json:"service_unit,omitempty"`
	RestartCommand    string   `json:"restart_command,omitempty"`
	ReloadCommand     string   `json:"reload_command,omitempty"`
	DryRunCommand     string   `json:"dry_run_command,omitempty"`
	MetricsURL        string   `json:"metrics_url,omitempty"`
	MetricsFormat     string   `json:"metrics_format,omitempty"`
	MetricsAPIEnabled bool     `json:"metrics_api_enabled"`
	Plugins           []string `json:"plugins,omitempty"`
}

type Bootstrap struct {
	ServerURL string
	APIKey    string
	NodeUID   string
}

type ServerSettings struct {
	HeartbeatInterval   int      `json:"heartbeat_interval,omitempty"`
	MetricsInterval     int      `json:"metrics_interval,omitempty"`
	LogUploadInterval   int      `json:"log_upload_interval,omitempty"`
	LogBufferLines      int      `json:"log_buffer_lines,omitempty"`
	HealthPort          int      `json:"health_port,omitempty"`
	MaxRetries          int      `json:"max_retries,omitempty"`
	RetryBaseDelay      int      `json:"retry_base_delay,omitempty"`
	FluentType          string   `json:"fluent_type,omitempty"`
	FluentConfigPath    string   `json:"fluent_config_path,omitempty"`
	FluentConfigDir     string   `json:"fluent_config_dir,omitempty"`
	FluentBinary        string   `json:"fluent_binary,omitempty"`
	FluentServiceUnit   string   `json:"fluent_service_unit,omitempty"`
	FluentRestartCmd    string   `json:"fluent_restart_cmd,omitempty"`
	FluentReloadCmd     string   `json:"fluent_reload_cmd,omitempty"`
	FluentDryRunCmd     string   `json:"fluent_dry_run_cmd,omitempty"`
	FluentLogPath       string   `json:"fluent_log_path,omitempty"`
	FluentExtraFiles    []string `json:"fluent_extra_files,omitempty"`
	FluentMetricsURL    string   `json:"fluent_metrics_url,omitempty"`
	FluentMetricsFormat string   `json:"fluent_metrics_format,omitempty"`
	BackupDir           string   `json:"backup_dir,omitempty"`
	MaxBackups          int      `json:"max_backups,omitempty"`
}

type Snapshot struct {
	ServerURL           string
	APIKey              string
	NodeUID             string
	ConfigPath          string
	FluentType          string
	FluentConfigPath    string
	FluentConfigDir     string
	FluentBinary        string
	FluentServiceUnit   string
	FluentRestartCmd    string
	FluentReloadCmd     string
	FluentDryRunCmd     string
	FluentLogPath       string
	FluentExtraFiles    []string
	FluentMetricsURL    string
	FluentMetricsFormat string
	HeartbeatInterval   int
	MetricsInterval     int
	LogUploadInterval   int
	LogBufferLines      int
	HealthPort          int
	MaxRetries          int
	RetryBaseDelay      int
	Labels              string
	StateDir            string
	BackupDir           string
	MaxBackups          int
	RuntimeProfile      RuntimeProfile
}

type Config struct {
	mu sync.RWMutex `yaml:"-"`

	ServerURL  string `yaml:"server_url"`
	APIKey     string `yaml:"api_key"`
	NodeUID    string `yaml:"node_uid"`
	ConfigPath string `yaml:"-"` // path to this config file

	// Fluent settings
	FluentType          string   `yaml:"fluent_type"`         // fluentbit, fluentd, auto
	FluentConfigPath    string   `yaml:"fluent_config_path"`  // main config file
	FluentConfigDir     string   `yaml:"fluent_config_dir"`   // extra config directory (parsers, plugins, etc.)
	FluentBinary        string   `yaml:"fluent_binary"`       // path to fluent-bit or fluentd binary
	FluentServiceUnit   string   `yaml:"fluent_service_unit"` // optional explicit service unit
	FluentRestartCmd    string   `yaml:"fluent_restart_cmd"`
	FluentReloadCmd     string   `yaml:"fluent_reload_cmd"`  // graceful reload (SIGHUP)
	FluentDryRunCmd     string   `yaml:"fluent_dry_run_cmd"` // validate config without starting
	FluentLogPath       string   `yaml:"fluent_log_path"`    // fluent log file to tail
	FluentExtraFiles    []string `yaml:"fluent_extra_files"` // additional managed config files
	FluentMetricsURL    string   `yaml:"fluent_metrics_url"`
	FluentMetricsFormat string   `yaml:"fluent_metrics_format"` // prometheus, fluentd_monitor_agent

	// Intervals
	HeartbeatInterval int `yaml:"heartbeat_interval"` // seconds
	MetricsInterval   int `yaml:"metrics_interval"`   // seconds
	LogUploadInterval int `yaml:"log_upload_interval"`
	LogBufferLines    int `yaml:"log_buffer_lines"`

	// Local health check
	HealthPort int `yaml:"health_port"`

	// Retry settings
	MaxRetries     int `yaml:"max_retries"`
	RetryBaseDelay int `yaml:"retry_base_delay"` // milliseconds

	// Labels
	Labels string `yaml:"labels"` // JSON key-value

	// State directory for persistent data (node_uid, backups, etc.)
	StateDir string `yaml:"state_dir"`

	// Backup
	BackupDir  string `yaml:"backup_dir"`
	MaxBackups int    `yaml:"max_backups"`

	// Log discovery — paths the server is permitted to scan and read via remote commands.
	// Defaults to ["/var/log"] if empty. Extend in agent.yaml for non-standard log directories.
	AllowedLogPaths []string `yaml:"allowed_log_paths"`

	// Derived runtime metadata
	RuntimeProfile     RuntimeProfile `yaml:"-"`
	metricsURLExplicit bool           `yaml:"-"`
}

func Load(path string, bootstrap Bootstrap) (*Config, error) {
	cfg := &Config{
		FluentType:        "fluentbit",
		FluentConfigPath:  "",
		FluentConfigDir:   "",
		FluentBinary:      "",
		FluentServiceUnit: "",
		FluentRestartCmd:  "",
		FluentReloadCmd:   "",
		FluentDryRunCmd:   "",
		FluentLogPath:     "",
		HeartbeatInterval: 30,
		MetricsInterval:   60,
		LogUploadInterval: 120,
		LogBufferLines:    500,
		HealthPort:        9880,
		MaxRetries:        5,
		RetryBaseDelay:    1000,
		MaxBackups:        10,
		ConfigPath:        path,
	}

	if strings.TrimSpace(path) != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read config: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		}
	}
	cfg.applyBootstrap(bootstrap)
	cfg.metricsURLExplicit = strings.TrimSpace(cfg.FluentMetricsURL) != ""

	// Resolve state directory: explicit > config dir fallback > $HOME fallback
	if strings.TrimSpace(cfg.StateDir) == "" {
		cfg.StateDir = resolveStateDir(path)
	}

	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	if cfg.NodeUID == "" {
		cfg.NodeUID = loadOrCreateUID(cfg.StateDir)
	}

	if strings.TrimSpace(cfg.BackupDir) == "" {
		cfg.BackupDir = filepath.Join(cfg.StateDir, "backups")
	}
	if err := os.MkdirAll(cfg.BackupDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	if err := cfg.ResolveRuntime(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Snapshot() Snapshot {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	extraFiles := make([]string, len(cfg.FluentExtraFiles))
	copy(extraFiles, cfg.FluentExtraFiles)

	return Snapshot{
		ServerURL:           cfg.ServerURL,
		APIKey:              cfg.APIKey,
		NodeUID:             cfg.NodeUID,
		ConfigPath:          cfg.ConfigPath,
		FluentType:          cfg.FluentType,
		FluentConfigPath:    cfg.FluentConfigPath,
		FluentConfigDir:     cfg.FluentConfigDir,
		FluentBinary:        cfg.FluentBinary,
		FluentServiceUnit:   cfg.FluentServiceUnit,
		FluentRestartCmd:    cfg.FluentRestartCmd,
		FluentReloadCmd:     cfg.FluentReloadCmd,
		FluentDryRunCmd:     cfg.FluentDryRunCmd,
		FluentLogPath:       cfg.FluentLogPath,
		FluentExtraFiles:    extraFiles,
		FluentMetricsURL:    cfg.FluentMetricsURL,
		FluentMetricsFormat: cfg.FluentMetricsFormat,
		HeartbeatInterval:   cfg.HeartbeatInterval,
		MetricsInterval:     cfg.MetricsInterval,
		LogUploadInterval:   cfg.LogUploadInterval,
		LogBufferLines:      cfg.LogBufferLines,
		HealthPort:          cfg.HealthPort,
		MaxRetries:          cfg.MaxRetries,
		RetryBaseDelay:      cfg.RetryBaseDelay,
		Labels:              cfg.Labels,
		StateDir:            cfg.StateDir,
		BackupDir:           cfg.BackupDir,
		MaxBackups:          cfg.MaxBackups,
		RuntimeProfile:      cfg.RuntimeProfile,
	}
}

func (cfg *Config) applyBootstrap(bootstrap Bootstrap) {
	if value := strings.TrimSpace(bootstrap.ServerURL); value != "" {
		cfg.ServerURL = value
	}
	if value := strings.TrimSpace(bootstrap.APIKey); value != "" {
		cfg.APIKey = value
	}
	if value := strings.TrimSpace(bootstrap.NodeUID); value != "" {
		cfg.NodeUID = value
	}
}

func (cfg *Config) ApplyServerSettings(settings *ServerSettings) error {
	if settings == nil {
		return nil
	}
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.applyServerSettingsLocked(settings)
}

func (cfg *Config) applyServerSettingsLocked(settings *ServerSettings) error {
	if settings.HeartbeatInterval > 0 {
		cfg.HeartbeatInterval = settings.HeartbeatInterval
	}
	if settings.MetricsInterval > 0 {
		cfg.MetricsInterval = settings.MetricsInterval
	}
	if settings.LogUploadInterval > 0 {
		cfg.LogUploadInterval = settings.LogUploadInterval
	}
	if settings.LogBufferLines > 0 {
		cfg.LogBufferLines = settings.LogBufferLines
	}
	if settings.HealthPort > 0 {
		cfg.HealthPort = settings.HealthPort
	}
	if settings.MaxRetries > 0 {
		cfg.MaxRetries = settings.MaxRetries
	}
	if settings.RetryBaseDelay > 0 {
		cfg.RetryBaseDelay = settings.RetryBaseDelay
	}
	if value := strings.TrimSpace(settings.FluentType); value != "" {
		cfg.FluentType = value
	}
	if value := strings.TrimSpace(settings.FluentConfigPath); value != "" {
		cfg.FluentConfigPath = value
	}
	if value := strings.TrimSpace(settings.FluentConfigDir); value != "" {
		cfg.FluentConfigDir = value
	}
	if value := strings.TrimSpace(settings.FluentBinary); value != "" {
		cfg.FluentBinary = value
	}
	if value := strings.TrimSpace(settings.FluentServiceUnit); value != "" {
		cfg.FluentServiceUnit = value
	}
	if value := strings.TrimSpace(settings.FluentRestartCmd); value != "" {
		cfg.FluentRestartCmd = value
	}
	if value := strings.TrimSpace(settings.FluentReloadCmd); value != "" {
		cfg.FluentReloadCmd = value
	}
	if value := strings.TrimSpace(settings.FluentDryRunCmd); value != "" {
		cfg.FluentDryRunCmd = value
	}
	if value := strings.TrimSpace(settings.FluentLogPath); value != "" {
		cfg.FluentLogPath = value
	}
	if len(settings.FluentExtraFiles) > 0 {
		cfg.FluentExtraFiles = append([]string(nil), settings.FluentExtraFiles...)
	}
	if value := strings.TrimSpace(settings.FluentMetricsURL); value != "" {
		cfg.FluentMetricsURL = value
		cfg.metricsURLExplicit = true
	}
	if value := strings.TrimSpace(settings.FluentMetricsFormat); value != "" {
		cfg.FluentMetricsFormat = value
	}
	if value := strings.TrimSpace(settings.BackupDir); value != "" {
		cfg.BackupDir = value
	}
	if settings.MaxBackups > 0 {
		cfg.MaxBackups = settings.MaxBackups
	}

	if err := os.MkdirAll(cfg.BackupDir, 0o755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}
	return cfg.resolveRuntimeLocked()
}

func (cfg *Config) ResolveRuntime() error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.resolveRuntimeLocked()
}

func (cfg *Config) resolveRuntimeLocked() error {
	cfg.FluentType = normalizeRuntimeType(cfg.FluentType)

	resolvedType, binaryPath, err := cfg.resolveRuntimeTypeAndBinary()
	if err != nil {
		return err
	}
	cfg.FluentType = resolvedType
	cfg.FluentBinary = binaryPath
	cfg.FluentConfigPath = resolvePath(cfg.FluentConfigPath, defaultConfigPathCandidates(cfg.FluentType))
	if cfg.FluentConfigPath == "" {
		// Fallback to the first candidate even if it doesn't exist yet;
		// config path is required for dry-run, reload, and deploy operations.
		if candidates := defaultConfigPathCandidates(cfg.FluentType); len(candidates) > 0 {
			cfg.FluentConfigPath = candidates[0]
		}
	}
	cfg.FluentConfigDir = resolveOptionalDir(cfg.FluentConfigDir, defaultConfigDirCandidates(cfg.FluentType, cfg.FluentConfigPath))
	cfg.FluentLogPath = resolvePath(cfg.FluentLogPath, defaultLogPathCandidates(cfg.FluentType))
	cfg.FluentServiceUnit = cfg.resolveServiceUnit()
	if cfg.FluentRestartCmd == "" && cfg.FluentServiceUnit != "" {
		cfg.FluentRestartCmd = "systemctl restart " + cfg.FluentServiceUnit
	}
	if cfg.FluentReloadCmd == "" && cfg.FluentServiceUnit != "" {
		cfg.FluentReloadCmd = "systemctl reload " + cfg.FluentServiceUnit
	}
	if cfg.FluentDryRunCmd == "" {
		cfg.FluentDryRunCmd = buildDryRunCommand(cfg.FluentType, cfg.FluentBinary, cfg.FluentConfigPath)
	}
	if cfg.FluentMetricsFormat == "" {
		cfg.FluentMetricsFormat = defaultMetricsFormat(cfg.FluentType)
	}
	if cfg.FluentMetricsURL == "" {
		cfg.FluentMetricsURL = defaultMetricsURL(cfg.FluentType)
	}

	return cfg.refreshRuntimeProfileLocked()
}

func (cfg *Config) RefreshRuntimeProfile() error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.refreshRuntimeProfileLocked()
}

func (cfg *Config) refreshRuntimeProfileLocked() error {
	content, err := cfg.readManagedConfig()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read managed config: %w", err)
	}

	plugins := cfg.discoverPlugins(content)
	metricsAPIEnabled := cfg.detectMetricsAPIEnabled(content)
	supportsMetricsAPI := cfg.metricsURLExplicit || metricsAPIEnabled

	metadataPayload := runtimeMetadata{
		RuntimeType:       cfg.FluentType,
		BinaryPath:        cfg.FluentBinary,
		ConfigPath:        cfg.FluentConfigPath,
		ConfigDir:         cfg.FluentConfigDir,
		LogPath:           cfg.FluentLogPath,
		ServiceUnit:       cfg.FluentServiceUnit,
		RestartCommand:    cfg.FluentRestartCmd,
		ReloadCommand:     cfg.FluentReloadCmd,
		DryRunCommand:     cfg.FluentDryRunCmd,
		MetricsURL:        cfg.FluentMetricsURL,
		MetricsFormat:     cfg.FluentMetricsFormat,
		MetricsAPIEnabled: supportsMetricsAPI,
		Plugins:           plugins,
	}
	metadataBytes, err := json.Marshal(metadataPayload)
	if err != nil {
		return fmt.Errorf("marshal runtime metadata: %w", err)
	}

	cfg.RuntimeProfile = RuntimeProfile{
		LoadedPlugins:        strings.Join(plugins, ","),
		SupportsHotReload:    true,
		SupportsMultiline:    cfg.FluentType == "fluentbit" || cfg.FluentType == "fluentd",
		SupportsStorageLayer: cfg.FluentType == "fluentbit" || cfg.FluentType == "fluentd",
		SupportsForwardTLS:   cfg.FluentType == "fluentbit" || cfg.FluentType == "fluentd",
		SupportsMetricsAPI:   supportsMetricsAPI,
		Metadata:             string(metadataBytes),
	}
	return nil
}

func (cfg *Config) resolveRuntimeTypeAndBinary() (string, string, error) {
	hint := cfg.FluentType
	candidates := []struct {
		runtimeType string
		values      []string
	}{
		{runtimeType: "fluentbit", values: binaryCandidates("fluentbit", cfg.FluentBinary)},
		{runtimeType: "fluentd", values: binaryCandidates("fluentd", cfg.FluentBinary)},
	}
	if hint == "fluentd" {
		candidates[0], candidates[1] = candidates[1], candidates[0]
	}

	for _, candidate := range candidates {
		path := resolveBinary(candidate.values)
		if path == "" {
			continue
		}
		if hint == "" || hint == "auto" || hint == candidate.runtimeType || runtimeTypeFromBinary(path) == hint {
			return candidate.runtimeType, path, nil
		}
	}

	if hint == "fluentbit" || hint == "fluentd" {
		path := resolveBinary(binaryCandidates(hint, cfg.FluentBinary))
		if path != "" {
			return hint, path, nil
		}
	}

	return "", "", fmt.Errorf("unable to detect a Fluent Bit or Fluentd runtime; set fluent_type and fluent_binary explicitly")
}

func (cfg *Config) resolveServiceUnit() string {
	if unit := parseServiceUnitFromCommand(cfg.FluentRestartCmd); unit != "" {
		return unit
	}
	if unit := parseServiceUnitFromCommand(cfg.FluentReloadCmd); unit != "" {
		return unit
	}
	if cfg.FluentServiceUnit != "" {
		return strings.TrimSpace(cfg.FluentServiceUnit)
	}

	candidates := defaultServiceUnits(cfg.FluentType)
	for _, unit := range candidates {
		if serviceUnitExists(unit) {
			return unit
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func (cfg *Config) readManagedConfig() (string, error) {
	contents := make([]string, 0, 4)
	seen := map[string]bool{}
	appendFile := func(path string) error {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			return nil
		}
		seen[path] = true
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contents = append(contents, string(data))
		return nil
	}

	if err := appendFile(cfg.FluentConfigPath); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	files := make([]string, 0, len(cfg.FluentExtraFiles))
	files = append(files, cfg.FluentExtraFiles...)
	if cfg.FluentConfigDir != "" {
		entries, err := os.ReadDir(cfg.FluentConfigDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				if strings.HasSuffix(name, ".conf") || strings.HasSuffix(name, ".cfg") {
					files = append(files, filepath.Join(cfg.FluentConfigDir, name))
				}
			}
		}
	}
	sort.Strings(files)
	for _, file := range files {
		if err := appendFile(file); err != nil && !os.IsNotExist(err) {
			return "", err
		}
	}
	return strings.Join(contents, "\n"), nil
}

func (cfg *Config) discoverPlugins(content string) []string {
	switch cfg.FluentType {
	case "fluentbit":
		return parseFluentBitPlugins(content)
	case "fluentd":
		return parseFluentdPlugins(content)
	default:
		return nil
	}
}

func (cfg *Config) detectMetricsAPIEnabled(content string) bool {
	switch cfg.FluentType {
	case "fluentbit":
		return matchFluentBitDirective(content, "http_server", "on")
	case "fluentd":
		return strings.Contains(strings.ToLower(content), "monitor_agent")
	default:
		return false
	}
}

// matchFluentBitDirective checks whether a Fluent Bit config contains a
// directive like "http_server  On" (case-insensitive, whitespace-tolerant).
// It avoids false positives such as matching "Off" as containing "on".
func matchFluentBitDirective(content, key, value string) bool {
	lower := strings.ToLower(content)
	key = strings.ToLower(key)
	value = strings.ToLower(value)
	for _, line := range strings.Split(lower, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[0] == key && parts[1] == value {
			return true
		}
	}
	return false
}

func normalizeRuntimeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return "auto"
	case "fluent-bit", "fluentbit":
		return "fluentbit"
	case "fluentd", "td-agent", "tdagent":
		return "fluentd"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func binaryCandidates(runtimeType, configured string) []string {
	out := []string{}
	if configured != "" {
		out = append(out, configured)
	}
	switch runtimeType {
	case "fluentbit":
		out = append(out,
			"fluent-bit",
			"/opt/fluent-bit/bin/fluent-bit",
			"/usr/bin/fluent-bit",
			"/usr/local/bin/fluent-bit",
		)
	case "fluentd":
		out = append(out,
			"fluentd",
			"td-agent",
			"/usr/sbin/td-agent",
			"/opt/td-agent/bin/td-agent",
			"/usr/bin/fluentd",
			"/usr/local/bin/fluentd",
		)
	}
	return uniqueStrings(out)
}

func resolveBinary(candidates []string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, "/") {
			if fileExists(candidate) {
				return candidate
			}
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved
		}
	}
	return ""
}

func runtimeTypeFromBinary(path string) string {
	base := strings.ToLower(filepath.Base(path))
	if strings.Contains(base, "fluent-bit") {
		return "fluentbit"
	}
	if strings.Contains(base, "fluentd") || strings.Contains(base, "td-agent") {
		return "fluentd"
	}
	return ""
}

func defaultConfigPathCandidates(runtimeType string) []string {
	if runtimeType == "fluentd" {
		return []string{
			"/etc/fluent/fluentd.conf",
			"/etc/td-agent/td-agent.conf",
			"/etc/fluentd/fluent.conf",
		}
	}
	return []string{
		"/etc/fluent-bit/fluent-bit.conf",
		"/etc/fluent-bit.conf",
	}
}

func defaultConfigDirCandidates(runtimeType, configPath string) []string {
	dir := filepath.Dir(strings.TrimSpace(configPath))
	if runtimeType == "fluentd" {
		return []string{
			filepath.Join(dir, "conf.d"),
			"/etc/fluent/conf.d",
			"/etc/td-agent/conf.d",
			"/etc/fluentd/conf.d",
		}
	}
	return []string{
		filepath.Join(dir, "conf.d"),
		"/etc/fluent-bit/conf.d",
	}
}

func defaultLogPathCandidates(runtimeType string) []string {
	if runtimeType == "fluentd" {
		return []string{
			"/var/log/fluentd.log",
			"/var/log/td-agent/td-agent.log",
			"/var/log/td-agent.log",
		}
	}
	return []string{
		"/var/log/fluent-bit.log",
		"/var/log/fluent-bit/fluent-bit.log",
	}
}

func defaultServiceUnits(runtimeType string) []string {
	if runtimeType == "fluentd" {
		return []string{"fluentd", "td-agent"}
	}
	return []string{"fluent-bit", "fluent-bitd"}
}

func buildDryRunCommand(runtimeType, binaryPath, configPath string) string {
	if binaryPath == "" || configPath == "" {
		return ""
	}
	if runtimeType == "fluentd" {
		return fmt.Sprintf("%s --dry-run -c %s", binaryPath, configPath)
	}
	return fmt.Sprintf("%s --dry-run -c %s", binaryPath, configPath)
}

func defaultMetricsURL(runtimeType string) string {
	if runtimeType == "fluentd" {
		return "http://127.0.0.1:24220/api/plugins.json"
	}
	return "http://127.0.0.1:2020/api/v1/metrics/prometheus"
}

func defaultMetricsFormat(runtimeType string) string {
	if runtimeType == "fluentd" {
		return "fluentd_monitor_agent"
	}
	return "prometheus"
}

func resolvePath(current string, candidates []string) string {
	current = strings.TrimSpace(current)
	if current != "" {
		return current
	}
	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func resolveOptionalDir(current string, candidates []string) string {
	current = strings.TrimSpace(current)
	if current != "" {
		return current
	}
	for _, candidate := range candidates {
		if dirExists(candidate) {
			return candidate
		}
	}
	return ""
}

func parseServiceUnitFromCommand(command string) string {
	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) >= 3 && parts[0] == "systemctl" {
		unit := strings.TrimSuffix(parts[len(parts)-1], ".service")
		if unit != "" {
			return unit
		}
	}
	if len(parts) >= 3 && parts[0] == "service" {
		return parts[1]
	}
	return ""
}

func serviceUnitExists(unit string) bool {
	for _, dir := range []string{"/etc/systemd/system", "/lib/systemd/system", "/usr/lib/systemd/system"} {
		if fileExists(filepath.Join(dir, unit+".service")) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func parseFluentBitPlugins(content string) []string {
	currentSection := ""
	plugins := make([]string, 0, 8)
	lines := strings.Split(content, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToUpper(strings.Trim(line, "[]"))
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "name ") {
			switch currentSection {
			case "INPUT", "FILTER", "OUTPUT", "PARSER":
				plugins = append(plugins, strings.TrimSpace(line[len("Name"):]))
			}
		}
	}
	return uniqueSortedStrings(plugins)
}

func parseFluentdPlugins(content string) []string {
	plugins := make([]string, 0, 8)
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "@type ") {
			plugins = append(plugins, strings.TrimSpace(strings.TrimPrefix(line, "@type ")))
		}
	}
	return uniqueSortedStrings(plugins)
}

func uniqueStrings(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || set[value] {
			continue
		}
		set[value] = true
		out = append(out, value)
	}
	return out
}

func uniqueSortedStrings(values []string) []string {
	out := uniqueStrings(values)
	sort.Strings(out)
	return out
}

func resolveStateDir(configPath string) string {
	configPath = strings.TrimSpace(configPath)
	if configPath != "" {
		dir := filepath.Dir(configPath)
		if dir != "" && dir != "." {
			return dir
		}
	}
	return defaultStateDir()
}

func defaultStateDir() string {
	if stateHome := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); stateHome != "" {
		return filepath.Join(stateHome, "fluent-manager-agent")
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".local", "state", "fluent-manager-agent")
	}
	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		return filepath.Join(cwd, ".fluent-manager-agent")
	}
	return ".fluent-manager-agent"
}

func loadOrCreateUID(dir string) string {
	_ = os.MkdirAll(dir, 0o755)
	uidFile := filepath.Join(dir, ".node_uid")
	data, err := os.ReadFile(uidFile)
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	uid := uuid.New().String()
	_ = os.WriteFile(uidFile, []byte(uid), 0o600)
	return uid
}
