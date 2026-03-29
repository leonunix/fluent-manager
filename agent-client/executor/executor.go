package executor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
	"github.com/fluent-manager/fluent-manager-agent/transport"
)

// Executor handles config application, validation, backup/rollback, and remote commands.
type Executor struct {
	cfg *config.Config
	mu  sync.Mutex
}

type agentUpgradeCommandArgs struct {
	PackageURL    string `json:"package_url"`
	Checksum      string `json:"checksum,omitempty"`
	TargetVersion string `json:"target_version,omitempty"`
	ServiceUnit   string `json:"service_unit,omitempty"`
	BinaryPath    string `json:"binary_path,omitempty"`
}

type scanLogsArgs struct {
	Paths []string `json:"paths"`
}

type scanLogsFileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type scanLogsResult struct {
	Files []scanLogsFileInfo `json:"files"`
}

type fetchLogSampleArgs struct {
	Files []string `json:"files"`
	Lines int      `json:"lines"`
}

type fetchLogSampleResult struct {
	Samples map[string]string `json:"samples"`
}

func New(cfg *config.Config) *Executor {
	return &Executor{cfg: cfg}
}

// CurrentConfigHash returns the SHA-256 of the current fluent config on disk.
func (e *Executor) CurrentConfigHash() string {
	data, err := os.ReadFile(e.cfg.FluentConfigPath)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// Apply writes a new config, validates it, reloads or restarts the service, and rolls back on failure.
// Returns (success, message).
func (e *Executor) Apply(content string, configID uint) (bool, string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Printf("[executor] applying config (id=%d, size=%d bytes)", configID, len(content))

	backupPath, err := e.backup()
	if err != nil {
		log.Printf("[executor] backup warning: %v", err)
	}

	if err := os.WriteFile(e.cfg.FluentConfigPath, []byte(content), 0o644); err != nil {
		return false, fmt.Sprintf("write config failed: %v", err)
	}
	e.refreshRuntimeProfile()

	if e.cfg.FluentDryRunCmd != "" {
		if err := e.validate(); err != nil {
			log.Printf("[executor] validation failed, rolling back: %v", err)
			e.rollback(backupPath)
			return false, fmt.Sprintf("config validation failed: %v", err)
		}
		log.Printf("[executor] config validation passed")
	}

	action, output, err := e.reloadOrRestart()
	if err != nil {
		log.Printf("[executor] %s failed, rolling back: %v", action, err)
		e.rollback(backupPath)
		return false, fmt.Sprintf("%s failed: %v\noutput: %s", action, err, output)
	}

	time.Sleep(3 * time.Second)
	if !e.isFluentRunning() {
		log.Printf("[executor] fluent process not running after %s, rolling back", action)
		e.rollback(backupPath)
		return false, fmt.Sprintf("fluent process died after %s, rolled back", action)
	}

	e.refreshRuntimeProfile()
	e.pruneBackups()
	log.Printf("[executor] config applied successfully via %s", action)
	return true, fmt.Sprintf("config applied successfully via %s", action)
}

// validate runs the dry-run command to check config syntax.
func (e *Executor) validate() error {
	if strings.TrimSpace(e.cfg.FluentDryRunCmd) == "" {
		return nil
	}
	output, err := e.runShell(e.cfg.FluentDryRunCmd)
	if err != nil {
		return fmt.Errorf("%s\n%s", err, output)
	}
	return nil
}

// backup creates a timestamped copy of the current config.
func (e *Executor) backup() (string, error) {
	data, err := os.ReadFile(e.cfg.FluentConfigPath)
	if err != nil {
		return "", fmt.Errorf("read current config: %w", err)
	}

	_ = os.MkdirAll(e.cfg.BackupDir, 0o755)
	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("config-%s.bak", ts)
	backupPath := filepath.Join(e.cfg.BackupDir, name)

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	log.Printf("[executor] backed up to %s", backupPath)
	return backupPath, nil
}

// rollback restores config from backup and restarts.
func (e *Executor) rollback(backupPath string) {
	if backupPath == "" {
		backupPath = e.latestBackup()
		if backupPath == "" {
			log.Printf("[executor] no backup found for rollback")
			return
		}
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		log.Printf("[executor] rollback read failed: %v", err)
		return
	}

	if err := os.WriteFile(e.cfg.FluentConfigPath, data, 0o644); err != nil {
		log.Printf("[executor] rollback write failed: %v", err)
		return
	}
	e.refreshRuntimeProfile()

	action, output, err := e.restartOnly()
	if err != nil {
		log.Printf("[executor] rollback %s failed: %v, output: %s", action, err, output)
		return
	}

	e.refreshRuntimeProfile()
	log.Printf("[executor] rollback complete, reverted to %s via %s", backupPath, action)
}

// latestBackup returns the path to the most recent backup file.
func (e *Executor) latestBackup() string {
	entries, err := os.ReadDir(e.cfg.BackupDir)
	if err != nil || len(entries) == 0 {
		return ""
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})
	return filepath.Join(e.cfg.BackupDir, entries[0].Name())
}

// pruneBackups removes old backups exceeding MaxBackups.
func (e *Executor) pruneBackups() {
	entries, err := os.ReadDir(e.cfg.BackupDir)
	if err != nil {
		return
	}
	if len(entries) <= e.cfg.MaxBackups {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	toRemove := len(entries) - e.cfg.MaxBackups
	for i := 0; i < toRemove; i++ {
		path := filepath.Join(e.cfg.BackupDir, entries[i].Name())
		_ = os.Remove(path)
		log.Printf("[executor] pruned old backup: %s", entries[i].Name())
	}
}

// RunCommand executes a remote command from the server.
func (e *Executor) RunCommand(action, args string) (transport.CommandResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch action {
	case "restart":
		_, output, err := e.restartOnly()
		return transport.CommandResult{Output: output}, err
	case "reload":
		e.refreshRuntimeProfile()
		_, output, err := e.reloadOrRestart()
		e.refreshRuntimeProfile()
		return transport.CommandResult{Output: output}, err
	case "stop":
		output, err := e.runShell(e.systemctlCommand("stop"))
		return transport.CommandResult{Output: output}, err
	case "start":
		output, err := e.runShell(e.systemctlCommand("start"))
		return transport.CommandResult{Output: output}, err
	case "status":
		output, err := e.runShell(e.systemctlCommand("status"))
		return transport.CommandResult{Output: output}, err
	case "validate":
		if err := e.validate(); err != nil {
			return transport.CommandResult{}, err
		}
		return transport.CommandResult{Output: "config is valid"}, nil
	case "rollback":
		e.rollback("")
		return transport.CommandResult{Output: "rolled back to latest backup"}, nil
	case "show_config":
		data, err := os.ReadFile(e.cfg.FluentConfigPath)
		if err != nil {
			return transport.CommandResult{}, err
		}
		return transport.CommandResult{Output: string(data)}, nil
	case "agent_upgrade":
		output, err := e.prepareAgentUpgrade(args)
		if err != nil {
			return transport.CommandResult{}, err
		}
		return transport.CommandResult{Output: output, RestartAgent: true}, nil
	case "scan_logs":
		output, err := e.scanLogs(args)
		if err != nil {
			return transport.CommandResult{}, err
		}
		return transport.CommandResult{Output: output}, nil
	case "fetch_log_sample":
		output, err := e.fetchLogSample(args)
		if err != nil {
			return transport.CommandResult{}, err
		}
		return transport.CommandResult{Output: output}, nil
	default:
		return transport.CommandResult{}, fmt.Errorf("unknown command: %s", action)
	}
}

const (
	// maxOutputBytes is the hard limit stored in RemoteCommand.Output (mirrored from agent_service).
	// Log content is sanitized (control characters replaced with space) before budget accounting,
	// so the worst-case JSON expansion is 2× (all-quote or all-backslash printable content).
	// Without sanitization, NUL bytes expand to \u0000 (6×), breaking the budget.
	// Invariant: maxTotalBytes × 2 + JSON framing overhead < maxOutputBytes
	//            250 KB × 2 + ~2 KB = ~502 KB < 512 KB ✓
	maxOutputBytes  = 512 * 1024
	maxBytesPerFile = 25 * 1024        // per-file raw content cap  (25 KB × 2 = 50 KB encoded)
	maxTotalBytes   = 250 * 1024       // total raw content cap     (250 KB × 2 = 500 KB encoded < 512 KB)
	maxCollect      = int64(512 * 1024) // tailFile internal read buffer cap
)

// allowedLogPaths returns the configured whitelist, defaulting to ["/var/log"].
func (e *Executor) allowedLogPaths() []string {
	if paths := e.cfg.AllowedLogPaths; len(paths) > 0 {
		return paths
	}
	return []string{"/var/log"}
}

// isPathAllowed resolves ALL symlinks on both sides — the candidate path and each allowed
// root — before comparing. This prevents two classes of escape:
//   - intermediate-directory symlinks in the candidate path (/allowed/linkdir/f where linkdir→/etc)
//   - allowed roots that are themselves symlinks: without resolving the root,
//     EvalSymlinks(candidate) returns a real path that no longer matches the unresolved root prefix.
//
// Fails closed: unresolvable paths (non-existent) return false; unresolvable roots are skipped.
func isPathAllowed(path string, allowed []string) bool {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	resolved = filepath.Clean(resolved)
	for _, root := range allowed {
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue // root doesn't exist or can't be resolved; skip this entry
		}
		resolvedRoot = filepath.Clean(resolvedRoot)
		if resolved == resolvedRoot || strings.HasPrefix(resolved, resolvedRoot+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func (e *Executor) scanLogs(rawArgs string) (string, error) {
	var args scanLogsArgs
	if rawArgs != "" {
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", fmt.Errorf("parse scan_logs args: %w", err)
		}
	}
	allowed := e.allowedLogPaths()
	if len(args.Paths) == 0 {
		args.Paths = allowed
	}

	const maxFiles = 2000
	seen := map[string]bool{}
	var files []scanLogsFileInfo
	for _, root := range args.Paths {
		root = filepath.Clean(root)
		if !isPathAllowed(root, allowed) {
			continue
		}
		// Resolve symlinks before walking: filepath.Walk does not follow symlink directories,
		// so a symlink allowed root would return empty results without this step.
		realRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		_ = filepath.Walk(realRoot, func(path string, info os.FileInfo, err error) error {
			if len(files) >= maxFiles {
				return filepath.SkipAll
			}
			if err != nil || info == nil || !info.Mode().IsRegular() || seen[path] {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext != "" && ext != ".log" && ext != ".txt" && ext != ".out" && ext != ".err" {
				return nil
			}
			if info.Size() < 10 {
				return nil
			}
			seen[path] = true
			files = append(files, scanLogsFileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
			})
			return nil
		})
	}

	// RFC3339 strings are lexicographically sortable, so string comparison gives correct chronological order
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime > files[j].ModTime
	})

	data, err := json.Marshal(scanLogsResult{Files: files})
	if err != nil {
		return "", fmt.Errorf("marshal scan_logs result: %w", err)
	}
	return string(data), nil
}

// fetchLogSample reads the last N lines from each specified file and returns a JSON map.
// Errors per file are inlined rather than failing the whole request so partial results remain usable.
func (e *Executor) fetchLogSample(rawArgs string) (string, error) {
	var args fetchLogSampleArgs
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return "", fmt.Errorf("parse fetch_log_sample args: %w", err)
	}
	if len(args.Files) == 0 {
		return "", fmt.Errorf("files list is empty")
	}
	// Hard caps — prevent a single request from reading unbounded data
	const maxFiles = 10
	if len(args.Files) > maxFiles {
		args.Files = args.Files[:maxFiles]
	}
	if args.Lines <= 0 {
		args.Lines = 100
	}
	if args.Lines > 500 {
		args.Lines = 500
	}

	allowed := e.allowedLogPaths()
	samples := make(map[string]string, len(args.Files))
	totalBytes := 0
	for _, path := range args.Files {
		if totalBytes >= maxTotalBytes {
			break
		}
		path = filepath.Clean(path)
		if !isPathAllowed(path, allowed) {
			samples[path] = "[error: path not in allowed log directories]"
			continue
		}
		// isPathAllowed already resolved all symlinks; os.Stat confirms the resolved target is a regular file.
		fi, err := os.Stat(path)
		if err != nil || !fi.Mode().IsRegular() {
			samples[path] = "[error: not a regular file]"
			continue
		}
		lines, err := tailFile(path, args.Lines)
		if err != nil {
			samples[path] = fmt.Sprintf("[error reading file: %v]", err)
			continue
		}
		content := sanitizeLogContent(strings.Join(lines, "\n"))
		// Trim to the per-file cap at the last complete line so content stays valid UTF-8 text.
		if len(content) > maxBytesPerFile {
			trimmed := content[:maxBytesPerFile]
			if idx := strings.LastIndexByte(trimmed, '\n'); idx > 0 {
				trimmed = trimmed[:idx]
			}
			content = trimmed
		}
		// Apply remaining total budget.
		if totalBytes+len(content) > maxTotalBytes {
			remaining := maxTotalBytes - totalBytes
			trimmed := content[:remaining]
			if idx := strings.LastIndexByte(trimmed, '\n'); idx > 0 {
				trimmed = trimmed[:idx]
			}
			content = trimmed
		}
		samples[path] = content
		totalBytes += len(content)
	}

	data, err := json.Marshal(fetchLogSampleResult{Samples: samples})
	if err != nil {
		return "", fmt.Errorf("marshal fetch_log_sample result: %w", err)
	}
	return string(data), nil
}

// sanitizeLogContent replaces ASCII control characters (except \t, \n, \r) with a space.
// Control bytes such as NUL (\x00) expand to \u0000 (6 bytes) when JSON-encoded, which
// breaks the ≤2× raw→encoded size assumption used for the content budget calculations.
func sanitizeLogContent(s string) string {
	b := []byte(s)
	for i, c := range b {
		if (c < 0x20 && c != 0x09 && c != 0x0A && c != 0x0D) || c == 0x7F {
			b[i] = ' '
		}
	}
	return string(b)
}

// tailFile returns the last n lines of a file by reading backwards in 32 KB chunks,
// stopping as soon as enough newlines are found — avoids reading GB-sized logs in full.
func tailFile(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}

	const chunkSize = int64(32 * 1024)
	var collected []byte
	pos := size
	for pos > 0 {
		read := chunkSize
		if pos < read {
			read = pos
		}
		pos -= read
		chunk := make([]byte, read)
		if _, err := f.ReadAt(chunk, pos); err != nil {
			return nil, err
		}
		collected = append(chunk, collected...)
		if bytes.Count(collected, []byte{'\n'}) > n {
			break
		}
		if int64(len(collected)) >= maxCollect {
			break
		}
	}

	lines := strings.Split(strings.TrimRight(string(collected), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

func (e *Executor) ReexecAgent() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return syscall.Exec(exePath, os.Args, os.Environ())
}

func (e *Executor) reloadOrRestart() (string, string, error) {
	if e.cfg.RuntimeProfile.SupportsHotReload && strings.TrimSpace(e.cfg.FluentReloadCmd) != "" {
		output, err := e.runShell(e.cfg.FluentReloadCmd)
		if err == nil {
			return "reload", output, nil
		}
		log.Printf("[executor] reload failed, falling back to restart: %v", err)
	}
	return e.restartOnly()
}

func (e *Executor) restartOnly() (string, string, error) {
	output, err := e.runShell(e.cfg.FluentRestartCmd)
	return "restart", output, err
}

func (e *Executor) systemctlCommand(action string) string {
	unit := e.serviceUnit()
	if unit == "" {
		return ""
	}
	return fmt.Sprintf("systemctl %s %s", action, unit)
}

func (e *Executor) serviceUnit() string {
	if unit := strings.TrimSpace(e.cfg.FluentServiceUnit); unit != "" {
		return unit
	}
	if e.cfg.FluentType == "fluentd" {
		return "fluentd"
	}
	return "fluent-bit"
}

func (e *Executor) isFluentRunning() bool {
	procName := "fluent-bit"
	if e.cfg.FluentType == "fluentd" {
		procName = "fluentd"
	}
	err := exec.Command("pgrep", "-x", procName).Run()
	return err == nil
}

func (e *Executor) refreshRuntimeProfile() {
	if err := e.cfg.RefreshRuntimeProfile(); err != nil {
		log.Printf("[executor] refresh runtime profile warning: %v", err)
	}
}

func (e *Executor) runShell(cmdStr string) (string, error) {
	cmdStr = strings.TrimSpace(cmdStr)
	if cmdStr == "" {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command("sh", "-lc", cmdStr)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func (e *Executor) prepareAgentUpgrade(rawArgs string) (string, error) {
	var args agentUpgradeCommandArgs
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return "", fmt.Errorf("parse agent upgrade args: %w", err)
	}

	packageURL := strings.TrimSpace(args.PackageURL)
	if packageURL == "" {
		return "", fmt.Errorf("package_url is required")
	}
	resolvedURL, includeAPIKey, err := e.resolvePackageURL(packageURL)
	if err != nil {
		return "", err
	}

	exePath := strings.TrimSpace(args.BinaryPath)
	if exePath == "" {
		exePath, err = os.Executable()
		if err != nil {
			return "", fmt.Errorf("resolve current executable: %w", err)
		}
	}

	upgradePath := filepath.Join(filepath.Dir(exePath), ".fluent-manager-agent.upgrade."+time.Now().Format("20060102150405"))
	if err := downloadAgentBinary(resolvedURL, upgradePath, e.cfg.Snapshot().APIKey, includeAPIKey); err != nil {
		return "", err
	}

	if err := verifyAgentBinaryChecksum(upgradePath, args.Checksum); err != nil {
		_ = os.Remove(upgradePath)
		return "", err
	}

	if err := os.Chmod(upgradePath, 0o755); err != nil {
		_ = os.Remove(upgradePath)
		return "", fmt.Errorf("chmod downloaded binary: %w", err)
	}
	if err := os.Rename(upgradePath, exePath); err != nil {
		_ = os.Remove(upgradePath)
		return "", fmt.Errorf("replace agent binary: %w", err)
	}

	versionText := strings.TrimSpace(args.TargetVersion)
	if versionText == "" {
		versionText = "new package"
	}
	return fmt.Sprintf("Agent binary replaced from %s (%s). Restarting agent process.", resolvedURL, versionText), nil
}

func (e *Executor) resolvePackageURL(packageURL string) (string, bool, error) {
	snapshot := e.cfg.Snapshot()
	serverURL := strings.TrimSpace(snapshot.ServerURL)

	if strings.HasPrefix(packageURL, "/") {
		if serverURL == "" {
			return "", false, fmt.Errorf("server_url is required for relative package URLs")
		}
		return strings.TrimRight(serverURL, "/") + packageURL, true, nil
	}

	parsedPackageURL, err := url.Parse(packageURL)
	if err != nil {
		return "", false, fmt.Errorf("parse package_url: %w", err)
	}
	if !parsedPackageURL.IsAbs() {
		if serverURL == "" {
			return "", false, fmt.Errorf("server_url is required for relative package URLs")
		}
		baseURL, err := url.Parse(strings.TrimRight(serverURL, "/") + "/")
		if err != nil {
			return "", false, fmt.Errorf("parse server_url: %w", err)
		}
		return baseURL.ResolveReference(parsedPackageURL).String(), true, nil
	}

	includeAPIKey := false
	if serverURL != "" {
		if parsedServerURL, err := url.Parse(serverURL); err == nil {
			includeAPIKey = parsedServerURL.Host == parsedPackageURL.Host && parsedServerURL.Scheme == parsedPackageURL.Scheme
		}
	}
	return parsedPackageURL.String(), includeAPIKey, nil
}

func downloadAgentBinary(packageURL, targetPath, apiKey string, includeAPIKey bool) error {
	req, err := http.NewRequest(http.MethodGet, packageURL, nil)
	if err != nil {
		return fmt.Errorf("download agent binary: %w", err)
	}
	if includeAPIKey && strings.TrimSpace(apiKey) != "" {
		req.Header.Set("X-Agent-Key", strings.TrimSpace(apiKey))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download agent binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download agent binary: unexpected HTTP status %s", resp.Status)
	}

	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700)
	if err != nil {
		return fmt.Errorf("create temporary agent binary: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write temporary agent binary: %w", err)
	}
	return nil
}

func verifyAgentBinaryChecksum(path, rawChecksum string) error {
	checksum := normalizeChecksum(rawChecksum)
	if checksum == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open downloaded binary: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash downloaded binary: %w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != checksum {
		return fmt.Errorf("checksum mismatch: expected %s got %s", checksum, actual)
	}
	return nil
}

func normalizeChecksum(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "sha256:")
	return value
}
