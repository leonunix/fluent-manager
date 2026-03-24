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

// ConfigModule is a reusable, typed Fluent config building block.
type ConfigModule struct {
	ID          uint                  `gorm:"primaryKey" json:"id"`
	Name        string                `gorm:"size:128;not null;uniqueIndex:idx_config_module_identity,priority:1" json:"name"`
	Description string                `gorm:"size:512" json:"description"`
	ModuleType  string                `gorm:"size:32;not null;uniqueIndex:idx_config_module_identity,priority:2" json:"module_type"`
	FluentType  string                `gorm:"size:32;not null;uniqueIndex:idx_config_module_identity,priority:3" json:"fluent_type"`
	Content     string                `gorm:"type:text;not null" json:"content"`
	Variables   string                `gorm:"type:text" json:"variables"`
	IsBuiltin   bool                  `gorm:"default:false" json:"is_builtin"`
	PresetKind  string                `gorm:"size:16" json:"preset_kind"`
	PresetKey   string                `gorm:"size:64;index" json:"preset_key"`
	Versions    []ConfigModuleVersion `gorm:"foreignKey:ModuleID" json:"versions,omitempty"`
	CreatedBy   uint                  `json:"created_by"`
	Creator     *User                 `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"-"`
}

// ConfigModuleVersion is an immutable version of a config module.
type ConfigModuleVersion struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ModuleID  uint           `json:"module_id"`
	Module    *ConfigModule  `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Version   int            `gorm:"not null" json:"version"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Variables string         `gorm:"type:text" json:"variables"`
	Hash      string         `gorm:"size:64;not null" json:"hash"`
	Comment   string         `gorm:"size:512" json:"comment"`
	CreatedBy uint           `json:"created_by"`
	Creator   *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// RenderedConfig is a persisted render preview / snapshot for a runtime target.
type RenderedConfig struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:128" json:"name"`
	FluentType     string         `gorm:"size:32;not null" json:"fluent_type"`
	RuntimeVersion string         `gorm:"size:64" json:"runtime_version"`
	SourceModules  string         `gorm:"type:text" json:"source_modules"`
	Variables      string         `gorm:"type:text" json:"variables"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Hash           string         `gorm:"size:64;not null" json:"hash"`
	CreatedBy      uint           `json:"created_by"`
	Creator        *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
