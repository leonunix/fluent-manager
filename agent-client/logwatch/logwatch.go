package logwatch

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fluent-manager/fluent-manager-agent/config"
	"github.com/fluent-manager/fluent-manager-agent/transport"
)

// Watcher tails the fluent log file and periodically uploads buffered lines to the server.
type Watcher struct {
	cfg    *config.Config
	client *transport.Client
	mu     sync.Mutex
	buffer []string
	stopCh chan struct{}
}

func New(cfg *config.Config, client *transport.Client) *Watcher {
	snapshot := cfg.Snapshot()
	return &Watcher{
		cfg:    cfg,
		client: client,
		buffer: make([]string, 0, snapshot.LogBufferLines),
		stopCh: make(chan struct{}),
	}
}

func (w *Watcher) Start() {
	snapshot := w.cfg.Snapshot()
	if snapshot.FluentLogPath == "" {
		log.Printf("[logwatch] no fluent log path configured, log tailing disabled")
	} else {
		go w.tailLoop()
	}
	go w.uploadLoop()
}

func (w *Watcher) Stop() {
	close(w.stopCh)
}

// tailLoop continuously tails the fluent log file.
func (w *Watcher) tailLoop() {
	for {
		select {
		case <-w.stopCh:
			return
		default:
		}

		if err := w.tailFile(); err != nil {
			log.Printf("[logwatch] tail error: %v, retrying in 10s", err)
			select {
			case <-time.After(10 * time.Second):
			case <-w.stopCh:
				return
			}
		}
	}
}

func (w *Watcher) tailFile() error {
	path := w.cfg.Snapshot().FluentLogPath
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Seek to end of file (tail -f behavior)
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	for {
		select {
		case <-w.stopCh:
			return nil
		default:
		}

		if scanner.Scan() {
			line := scanner.Text()
			w.appendLine(line)
		} else {
			if err := scanner.Err(); err != nil {
				return err
			}
			// No new data; wait briefly and check again
			select {
			case <-time.After(1 * time.Second):
			case <-w.stopCh:
				return nil
			}

			// Check if file was rotated
			currentPath := w.cfg.Snapshot().FluentLogPath
			if currentPath != path {
				log.Printf("[logwatch] log path changed from %s to %s, reopening", path, currentPath)
				return nil
			}
			newInfo, err := os.Stat(path)
			if err != nil {
				return err
			}
			currentInfo, err := f.Stat()
			if err != nil {
				return err
			}
			if !os.SameFile(newInfo, currentInfo) {
				log.Printf("[logwatch] log file rotated, reopening")
				return nil // Will reopen in tailLoop
			}
		}
	}
}

func (w *Watcher) appendLine(line string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer = append(w.buffer, line)
	// Keep buffer bounded
	maxLines := w.cfg.Snapshot().LogBufferLines
	if maxLines <= 0 {
		maxLines = 500
	}
	if len(w.buffer) > maxLines {
		w.buffer = w.buffer[len(w.buffer)-maxLines:]
	}
}

// uploadLoop periodically sends buffered log lines to the server.
func (w *Watcher) uploadLoop() {
	for {
		wait := time.Duration(w.cfg.Snapshot().LogUploadInterval) * time.Second
		if wait <= 0 {
			wait = 120 * time.Second
		}
		select {
		case <-time.After(wait):
			w.upload()
		case <-w.stopCh:
			return
		}
	}
}

func (w *Watcher) upload() {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return
	}
	lines := make([]string, len(w.buffer))
	copy(lines, w.buffer)
	w.buffer = w.buffer[:0]
	w.mu.Unlock()

	snapshot := w.cfg.Snapshot()
	body := map[string]interface{}{
		"node_uid": snapshot.NodeUID,
		"lines":    lines,
	}
	if _, err := w.client.APICall("POST", "/api/v1/agent/logs", body); err != nil {
		log.Printf("[logwatch] upload failed: %v (buffered %d lines lost)", err, len(lines))
	} else {
		log.Printf("[logwatch] uploaded %d log lines", len(lines))
	}
}

// GetRecentLines returns the current buffer (for local health endpoint).
func (w *Watcher) GetRecentLines() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	lines := make([]string, len(w.buffer))
	copy(lines, w.buffer)
	return lines
}
