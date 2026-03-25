package models

import "time"

// AgentUpgradeTask tracks a managed agent upgrade rollout across one or more nodes.
type AgentUpgradeTask struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:128;not null" json:"name"`
	Status        string         `gorm:"size:32;default:pending" json:"status"`
	ArtifactID    *uint          `gorm:"index" json:"artifact_id"`
	Artifact      *AgentArtifact `gorm:"foreignKey:ArtifactID" json:"artifact,omitempty"`
	PackageURL    string         `gorm:"size:1024;not null" json:"package_url"`
	Checksum      string         `gorm:"size:256" json:"checksum"`
	TargetVersion string         `gorm:"size:64" json:"target_version"`
	TargetSummary string         `gorm:"type:text" json:"target_summary"`
	TotalNodes    int            `json:"total_nodes"`
	SuccessCount  int            `json:"success_count"`
	FailCount     int            `json:"fail_count"`
	CreatedBy     uint           `json:"created_by"`
	Creator       *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// AgentUpgradeRecord tracks the per-node status of an agent upgrade rollout.
type AgentUpgradeRecord struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	AgentUpgradeTaskID uint           `gorm:"index" json:"agent_upgrade_task_id"`
	NodeID             uint           `gorm:"index" json:"node_id"`
	Node               *Node          `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	RemoteCommandID    *uint          `gorm:"index" json:"remote_command_id"`
	RemoteCommand      *RemoteCommand `gorm:"foreignKey:RemoteCommandID" json:"remote_command,omitempty"`
	Status             string         `gorm:"size:32;default:pending" json:"status"`
	Message            string         `gorm:"type:text" json:"message"`
	OutputExcerpt      string         `gorm:"type:text" json:"output_excerpt"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}
