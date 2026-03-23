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

// Apply writes a new config, validates it, restarts the service, and rolls back on failure.
// Returns (success, message).
func (e *Executor) Apply(content string, configID uint) (bool, string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Printf("[executor] applying config (id=%d, size=%d bytes)", configID, len(content))

	// Step 1: Create timestamped backup
	backupPath, err := e.backup()
	if err != nil {
		log.Printf("[executor] backup warning: %v", err)
		// Non-fatal, continue
	}

	// Step 2: Write new config
	if err := os.WriteFile(e.cfg.FluentConfigPath, []byte(content), 0644); err != nil {
		return false, fmt.Sprintf("write config failed: %v", err)
	}

	// Step 3: Validate (dry-run) before restarting
	if e.cfg.FluentDryRunCmd != "" {
		if err := e.validate(); err != nil {
			log.Printf("[executor] validation failed, rolling back: %v", err)
			e.rollback(backupPath)
			return false, fmt.Sprintf("config validation failed: %v", err)
		}
		log.Printf("[executor] config validation passed")
	}

	// Step 4: Restart service
	output, err := e.runShell(e.cfg.FluentRestartCmd)
	if err != nil {
		log.Printf("[executor] restart failed, rolling back: %v", err)
		e.rollback(backupPath)
		return false, fmt.Sprintf("restart failed: %v\noutput: %s", err, output)
	}

	// Step 5: Brief health check — wait and verify process is running
	time.Sleep(3 * time.Second)
	if !e.isFluentRunning() {
		log.Printf("[executor] fluent process not running after restart, rolling back")
		e.rollback(backupPath)
		return false, "fluent process died after restart, rolled back"
	}

	log.Printf("[executor] config applied successfully")
	e.pruneBackups()
	return true, "config applied and service restarted successfully"
}

// validate runs the dry-run command to check config syntax.
func (e *Executor) validate() error {
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

	_ = os.MkdirAll(e.cfg.BackupDir, 0755)
	ts := time.Now().Format("20060102-150405")
	name := fmt.Sprintf("config-%s.bak", ts)
	backupPath := filepath.Join(e.cfg.BackupDir, name)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	log.Printf("[executor] backed up to %s", backupPath)
	return backupPath, nil
}

// rollback restores config from backup and restarts.
func (e *Executor) rollback(backupPath string) {
	if backupPath == "" {
		// Try to find the most recent backup
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

	if err := os.WriteFile(e.cfg.FluentConfigPath, data, 0644); err != nil {
		log.Printf("[executor] rollback write failed: %v", err)
		return
	}

	if output, err := e.runShell(e.cfg.FluentRestartCmd); err != nil {
		log.Printf("[executor] rollback restart failed: %v, output: %s", err, output)
	} else {
		log.Printf("[executor] rollback complete, reverted to %s", backupPath)
	}
}

// latestBackup returns the path to the most recent backup file.
func (e *Executor) latestBackup() string {
	entries, err := os.ReadDir(e.cfg.BackupDir)
	if err != nil || len(entries) == 0 {
		return ""
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name() // descending
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

	// Sort ascending by name (which contains timestamp)
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
		return e.runShell(e.cfg.FluentRestartCmd)

	case "reload":
		if e.cfg.FluentReloadCmd != "" {
			return e.runShell(e.cfg.FluentReloadCmd)
		}
		return e.runShell(e.cfg.FluentRestartCmd)

	case "stop":
		return e.runShell("systemctl stop " + e.serviceUnit())

	case "start":
		return e.runShell("systemctl start " + e.serviceUnit())

	case "status":
		return e.runShell("systemctl status " + e.serviceUnit())

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

func (e *Executor) serviceUnit() string {
	if e.cfg.FluentType == "fluentbit" {
		return "fluent-bit"
	}
	return "fluentd"
}

func (e *Executor) isFluentRunning() bool {
	procName := "fluent-bit"
	if e.cfg.FluentType == "fluentd" {
		procName = "fluentd"
	}
	err := exec.Command("pgrep", "-x", procName).Run()
	return err == nil
}

func (e *Executor) runShell(cmdStr string) (string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}
