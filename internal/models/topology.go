package models

import (
	"time"

	"gorm.io/gorm"
)

// Environment separates infrastructure (production, staging, development, etc.).
type Environment struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:64;not null" json:"name"` // production, staging, dev
	Alias       string         `gorm:"size:64" json:"alias"`                     // 生产, 预发布, 开发
	Color       string         `gorm:"size:16" json:"color"`                     // #28a745, #ffc107, #17a2b8
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Description string         `gorm:"size:256" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DataCenter represents a physical data center or cloud provider.
type DataCenter struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:128;not null" json:"name"` // e.g. "aws-us", "ali-cn-hz"
	Alias       string         `gorm:"size:128" json:"alias"`                     // 阿里云-杭州
	Provider    string         `gorm:"size:64" json:"provider"`                   // aws, aliyun, azure, idc
	Location    string         `gorm:"size:256" json:"location"`                  // physical location
	Description string         `gorm:"size:512" json:"description"`
	Regions     []Region       `gorm:"foreignKey:DataCenterID" json:"regions,omitempty"`
	Tags        string         `gorm:"type:text" json:"tags"` // JSON key-value
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Region represents a logical region within a data center (e.g. cn-east-1, us-west-2).
type Region struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"` // e.g. "cn-east-1"
	Alias        string         `gorm:"size:128" json:"alias"`         // 华东一区
	DataCenterID uint           `gorm:"index;not null" json:"datacenter_id"`
	DataCenter   *DataCenter    `json:"datacenter,omitempty"`
	Clusters     []Cluster      `gorm:"foreignKey:RegionID" json:"clusters,omitempty"`
	Description  string         `gorm:"size:512" json:"description"`
	Tags         string         `gorm:"type:text" json:"tags"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Cluster represents a HA cluster / availability group within a region.
type Cluster struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	Name          string             `gorm:"size:128;not null" json:"name"` // e.g. "log-cluster-01"
	Alias         string             `gorm:"size:128" json:"alias"`
	RegionID      uint               `gorm:"index;not null" json:"region_id"`
	Region        *Region            `json:"region,omitempty"`
	EnvironmentID *uint              `gorm:"index" json:"environment_id"`
	Environment   *Environment       `json:"environment,omitempty"`
	IsDefault     bool               `gorm:"default:false" json:"is_default"` // default cluster for unmatched nodes
	Nodes         []Node             `gorm:"foreignKey:ClusterID" json:"nodes,omitempty"`
	MatchRules    []ClusterMatchRule `gorm:"foreignKey:ClusterID" json:"match_rules,omitempty"`
	ConfigID      *uint              `json:"config_id"` // inherited config for all nodes in cluster
	Config        *ConfigVersion     `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	Description   string             `gorm:"size:512" json:"description"`
	Tags          string             `gorm:"type:text" json:"tags"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `gorm:"index" json:"-"`
}

// ClusterMatchRule defines an auto-assignment rule for incoming nodes.
// When a new node registers, rules are evaluated in priority order.
// If all conditions in a rule match, the node is assigned to that cluster.
type ClusterMatchRule struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ClusterID       uint           `gorm:"index;not null" json:"cluster_id"`
	Name            string         `gorm:"size:128;not null" json:"name"`
	Priority        int            `gorm:"default:0" json:"priority"` // lower = higher priority
	HostnamePattern string         `gorm:"size:256" json:"hostname_pattern"` // glob: web-*, db-cn-*
	IPPattern       string         `gorm:"size:256" json:"ip_pattern"`       // CIDR or prefix: 10.0.1.*, 192.168.0.0/16
	FluentType      string         `gorm:"size:32" json:"fluent_type"`       // fluentbit, fluentd, or empty=any
	LabelSelector   string         `gorm:"type:text" json:"label_selector"`  // JSON: {"env":"prod","role":"web"}
	OSPattern       string         `gorm:"size:128" json:"os_pattern"`       // linux, windows, or glob
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
