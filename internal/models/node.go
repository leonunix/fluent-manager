package models

import (
	"time"

	"gorm.io/gorm"
)

// Node represents a Fluent Bit / Fluentd agent node.
type Node struct {
	ID                 uint               `gorm:"primaryKey" json:"id"`
	NodeUID            string             `gorm:"uniqueIndex;size:128;not null" json:"node_uid"`
	Hostname           string             `gorm:"size:256;not null" json:"hostname"`
	IPAddress          string             `gorm:"size:64" json:"ip_address"`
	OS                 string             `gorm:"size:64" json:"os"`
	AgentVersion       string             `gorm:"size:32" json:"agent_version"`
	FluentType         string             `gorm:"size:32" json:"fluent_type"` // fluentbit, fluentd
	FluentVersion      string             `gorm:"size:32" json:"fluent_version"`
	NodeRole           string             `gorm:"size:32;default:standalone" json:"node_role"`
	Status             string             `gorm:"size:32;default:offline" json:"status"` // online, offline, error
	ClusterID          *uint              `gorm:"index" json:"cluster_id"`
	Cluster            *Cluster           `json:"cluster,omitempty"`
	AggregationGroupID *uint              `gorm:"index" json:"aggregation_group_id"`
	AggregationGroup   *AggregationGroup  `json:"aggregation_group,omitempty"`
	EnvironmentID      *uint              `gorm:"index" json:"environment_id"` // can override cluster's environment
	Environment        *Environment       `json:"environment,omitempty"`
	Labels             string             `gorm:"type:text" json:"labels"` // JSON key-value pairs
	ConfigID           *uint              `json:"config_id"`               // node-level config (overrides cluster config)
	Config             *ConfigVersion     `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	FluentProfile      *NodeFluentProfile `gorm:"foreignKey:NodeID" json:"fluent_profile,omitempty"`
	LastHeartbeat      *time.Time         `json:"last_heartbeat"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
}

// EffectiveConfigID returns the config to use: node-level overrides cluster-level.
func (n *Node) EffectiveConfigID() *uint {
	if n.ConfigID != nil {
		return n.ConfigID
	}
	if n.Cluster != nil && n.Cluster.ConfigID != nil {
		return n.Cluster.ConfigID
	}
	return nil
}

// EffectiveEnvironmentID returns the environment: node-level overrides cluster-level.
func (n *Node) EffectiveEnvironmentID() *uint {
	if n.EnvironmentID != nil {
		return n.EnvironmentID
	}
	if n.Cluster != nil && n.Cluster.EnvironmentID != nil {
		return n.Cluster.EnvironmentID
	}
	return nil
}

// NodeMetrics stores the latest metrics snapshot from an agent.
type NodeMetrics struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	NodeID           uint      `gorm:"uniqueIndex" json:"node_id"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	MemTotalMB       uint64    `json:"mem_total_mb"`
	MemUsedMB        uint64    `json:"mem_used_mb"`
	MemUsagePercent  float64   `json:"mem_usage_percent"`
	DiskTotalGB      uint64    `json:"disk_total_gb"`
	DiskUsedGB       uint64    `json:"disk_used_gb"`
	DiskUsagePercent float64   `json:"disk_usage_percent"`
	LoadAvg1         float64   `json:"load_avg_1"`
	LoadAvg5         float64   `json:"load_avg_5"`
	LoadAvg15        float64   `json:"load_avg_15"`
	FluentRunning    bool      `json:"fluent_running"`
	FluentPID        int       `json:"fluent_pid"`
	FluentCPUPercent float64   `json:"fluent_cpu_percent"`
	FluentMemMB      float64   `json:"fluent_mem_mb"`
	FluentOpenFDs    int       `json:"fluent_open_fds"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// RemoteCommand is a pending command to be delivered to a node via heartbeat.
type RemoteCommand struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `gorm:"index" json:"node_id"`
	Node      *Node     `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	Action    string    `gorm:"size:64;not null" json:"action"`
	Args      string    `gorm:"type:text" json:"args"`
	Status    string    `gorm:"size:32;default:pending" json:"status"`
	Output    string    `gorm:"type:text" json:"output"`
	CreatedBy uint      `json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NodeLog stores log snippets uploaded by agents.
type NodeLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `gorm:"index" json:"node_id"`
	Lines     string    `gorm:"type:text" json:"lines"`
	LineCount int       `json:"line_count"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
