package logwriter

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
)

// FileLogger writes structured JSON lines to log files for Fluent Bit collection.
type FileLogger struct {
	nodeFile  *os.File
	auditFile *os.File
	mu        sync.Mutex
}

type nodeLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	NodeID    uint      `json:"node_id"`
	LineCount int       `json:"line_count"`
	Lines     string    `json:"lines"`
}

type auditLogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	ResourceType string    `json:"resource_type"`
	ResourceID   uint      `json:"resource_id"`
	Detail       string    `json:"detail"`
	IP           string    `json:"ip"`
}

// New creates a FileLogger that writes to the given directory.
// It creates the directory if it does not exist.
func New(dir string) (*FileLogger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	nf, err := os.OpenFile(filepath.Join(dir, "node-logs.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	af, err := os.OpenFile(filepath.Join(dir, "audit-logs.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		nf.Close()
		return nil, err
	}
	return &FileLogger{nodeFile: nf, auditFile: af}, nil
}

// WriteNodeLog appends a JSON line for a node log entry.
func (f *FileLogger) WriteNodeLog(nodeID uint, lines string, lineCount int) {
	entry := nodeLogEntry{
		Timestamp: time.Now(),
		NodeID:    nodeID,
		LineCount: lineCount,
		Lines:     lines,
	}
	f.writeLine(f.nodeFile, entry)
}

// WriteAuditLog appends a JSON line for an audit log entry.
func (f *FileLogger) WriteAuditLog(a models.AuditLog) {
	entry := auditLogEntry{
		Timestamp:    time.Now(),
		UserID:       a.UserID,
		Username:     a.Username,
		Action:       a.Action,
		Resource:     a.Resource,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		Detail:       a.Detail,
		IP:           a.IP,
	}
	f.writeLine(f.auditFile, entry)
}

func (f *FileLogger) writeLine(file *os.File, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("logwriter: marshal error: %v", err)
		return
	}
	data = append(data, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()
	if _, err := file.Write(data); err != nil {
		log.Printf("logwriter: write error: %v", err)
	}
}

// Close releases file handles.
func (f *FileLogger) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodeFile.Close()
	f.auditFile.Close()
}
