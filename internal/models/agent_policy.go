package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	AgentPolicyScopeGlobal        = "global"
	AgentPolicyScopeEnvironment   = "environment"
	AgentPolicyScopeCluster       = "cluster"
	AgentPolicyScopeLabelSelector = "label_selector"
)

type AgentPolicy struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:128;not null" json:"name"`
	Description   string         `gorm:"size:512" json:"description"`
	ScopeType     string         `gorm:"size:32;not null;index" json:"scope_type"`
	EnvironmentID *uint          `gorm:"index" json:"environment_id"`
	Environment   *Environment   `json:"environment,omitempty"`
	ClusterID     *uint          `gorm:"index" json:"cluster_id"`
	Cluster       *Cluster       `json:"cluster,omitempty"`
	LabelSelector string         `gorm:"type:text" json:"label_selector"`
	Priority      int            `gorm:"default:100;index" json:"priority"`
	IsEnabled     bool           `gorm:"default:true;index" json:"is_enabled"`
	Settings      string         `gorm:"type:text" json:"settings"`
	CreatedBy     uint           `json:"created_by"`
	Creator       *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

