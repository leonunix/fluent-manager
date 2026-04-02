package models

import (
	"time"

	"gorm.io/gorm"
)

// AgentAccessKey represents a managed server-side credential used by agents.
// The plaintext key is only shown once at creation time; only a SHA-256 hash is persisted.
// KeyEncrypted stores an AES-GCM encrypted copy so bootstrap tasks can reference the key by ID.
type AgentAccessKey struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	KeyHash      string         `gorm:"size:64;uniqueIndex;not null" json:"-"`
	KeyPreview   string         `gorm:"size:32;not null" json:"key_preview"`
	KeyEncrypted string         `gorm:"type:text" json:"-"`
	ClusterID   *uint          `gorm:"index" json:"cluster_id"`
	Cluster     *Cluster       `json:"cluster,omitempty"`
	Description string         `gorm:"size:512" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	CreatedBy   uint           `json:"created_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
