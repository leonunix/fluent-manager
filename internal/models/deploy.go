package models

import "time"

// DeployTask tracks a configuration deployment.
type DeployTask struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ConfigID     uint           `json:"config_id"`
	Config       *ConfigVersion `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	Scope        string         `gorm:"size:32" json:"scope"` // node, cluster, region, datacenter
	ScopeID      uint           `json:"scope_id"`             // ID of the target scope entity
	Status       string         `gorm:"size:32;default:pending" json:"status"`
	TotalNodes   int            `json:"total_nodes"`
	SuccessCount int            `json:"success_count"`
	FailCount    int            `json:"fail_count"`
	CreatedBy    uint           `json:"created_by"`
	Creator      *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// DeployRecord tracks the deployment result for individual nodes.
type DeployRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeployTaskID uint      `gorm:"index" json:"deploy_task_id"`
	NodeID       uint      `gorm:"index" json:"node_id"`
	Node         *Node     `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	Status       string    `gorm:"size:32;default:pending" json:"status"`
	Message      string    `gorm:"type:text" json:"message"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
