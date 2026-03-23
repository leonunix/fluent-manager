package models

import (
	"time"

	"gorm.io/gorm"
)

// ConfigTemplate is a reusable Fluent configuration template.
type ConfigTemplate struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Description string          `gorm:"size:512" json:"description"`
	FluentType  string          `gorm:"size:32;not null" json:"fluent_type"`
	Content     string          `gorm:"type:text;not null" json:"content"`
	Variables   string          `gorm:"type:text" json:"variables"`
	Versions    []ConfigVersion `gorm:"foreignKey:TemplateID" json:"versions,omitempty"`
	CreatedBy   uint            `json:"created_by"`
	Creator     *User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ConfigVersion is a versioned, immutable snapshot of a rendered configuration.
type ConfigVersion struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	TemplateID uint            `json:"template_id"`
	Template   *ConfigTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Version    int             `gorm:"not null" json:"version"`
	Content    string          `gorm:"type:text;not null" json:"content"`
	Hash       string          `gorm:"size:64;not null" json:"hash"`
	Comment    string          `gorm:"size:512" json:"comment"`
	CreatedBy  uint            `json:"created_by"`
	Creator    *User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}
