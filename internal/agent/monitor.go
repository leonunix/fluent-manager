package agent

import (
	"log"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
)

// Monitor periodically checks node heartbeats and marks stale nodes as offline.
type Monitor struct {
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
}

func NewMonitor(heartbeatIntervalSec int) *Monitor {
	return &Monitor{
		interval: time.Duration(heartbeatIntervalSec) * time.Second,
		timeout:  time.Duration(heartbeatIntervalSec*3) * time.Second, // 3x heartbeat = offline
		stopCh:   make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.checkNodes()
			case <-m.stopCh:
				return
			}
		}
	}()
	log.Printf("Node monitor started (interval=%v, timeout=%v)", m.interval, m.timeout)
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}

func (m *Monitor) checkNodes() {
	cutoff := time.Now().Add(-m.timeout)
	result := models.DB.Model(&models.Node{}).
		Where("status = ? AND last_heartbeat < ?", "online", cutoff).
		Update("status", "offline")
	if result.RowsAffected > 0 {
		log.Printf("Marked %d nodes as offline", result.RowsAffected)
	}
}
