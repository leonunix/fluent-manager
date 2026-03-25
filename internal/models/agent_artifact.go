package models

import "time"

// AgentArtifact stores an uploaded agent binary that can be used for managed upgrades.
type AgentArtifact struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:256;not null" json:"name"`
	Version     string    `gorm:"size:64" json:"version"`
	Description string    `gorm:"size:512" json:"description"`
	FileName    string    `gorm:"size:512;not null" json:"file_name"`
	ContentType string    `gorm:"size:128" json:"content_type"`
	FileSize    int64     `json:"file_size"`
	SHA256      string    `gorm:"size:128;not null" json:"sha256"`
	StoragePath string    `gorm:"size:1024;not null" json:"-"`
	CreatedBy   uint      `json:"created_by"`
	Creator     *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
