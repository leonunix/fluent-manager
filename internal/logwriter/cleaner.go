package logwriter

import (
	"log"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

// Cleaner periodically deletes old NodeLog and AuditLog records from the database.
type Cleaner struct {
	db        *gorm.DB
	retention time.Duration
	stopCh    chan struct{}
}

// NewCleaner creates a cleaner that removes records older than retention.
func NewCleaner(db *gorm.DB, retention time.Duration) *Cleaner {
	return &Cleaner{
		db:        db,
		retention: retention,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the periodic cleanup in a background goroutine.
func (c *Cleaner) Start() {
	go func() {
		// Run once on startup, then every 10 minutes.
		c.cleanup()
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop signals the cleaner goroutine to exit.
func (c *Cleaner) Stop() {
	close(c.stopCh)
}

func (c *Cleaner) cleanup() {
	cutoff := time.Now().Add(-c.retention)

	if result := c.db.Where("created_at < ?", cutoff).Delete(&models.NodeLog{}); result.Error != nil {
		log.Printf("logwriter/cleaner: failed to clean node_logs: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("logwriter/cleaner: deleted %d node_logs older than %v", result.RowsAffected, c.retention)
	}

	if result := c.db.Where("created_at < ?", cutoff).Delete(&models.AuditLog{}); result.Error != nil {
		log.Printf("logwriter/cleaner: failed to clean audit_logs: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("logwriter/cleaner: deleted %d audit_logs older than %v", result.RowsAffected, c.retention)
	}
}
