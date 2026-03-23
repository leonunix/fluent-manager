package executor

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

// Executor handles config application, validation, backup/rollback, and remote commands.
type Executor struct {
	cfg *config.Config
	mu  sync.Mutex
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
func (e *Executor) RunCommand(action, args string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch action {
	case "restart":
		_, output, err := e.restartOnly()
		return output, err
	case "reload":
		e.refreshRuntimeProfile()
		_, output, err := e.reloadOrRestart()
		e.refreshRuntimeProfile()
		return output, err
	case "stop":
		return e.runShell(e.systemctlCommand("stop"))
	case "start":
		return e.runShell(e.systemctlCommand("start"))
	case "status":
		return e.runShell(e.systemctlCommand("status"))
	case "validate":
		if err := e.validate(); err != nil {
			return "", err
		}
		return "config is valid", nil
	case "rollback":
		e.rollback("")
		return "rolled back to latest backup", nil
	case "show_config":
		data, err := os.ReadFile(e.cfg.FluentConfigPath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unknown command: %s", action)
	}
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
