package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	NodeRoleEdgeCollector = "edge_collector"
	NodeRoleAggregator    = "aggregator"
	NodeRoleGateway       = "gateway"
	NodeRoleStandalone    = "standalone"
)

var validNodeRoles = map[string]bool{
	NodeRoleEdgeCollector: true,
	NodeRoleAggregator:    true,
	NodeRoleGateway:       true,
	NodeRoleStandalone:    true,
}

func IsValidNodeRole(role string) bool {
	return validNodeRoles[role]
}

type AggregationGroup struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null;uniqueIndex:idx_aggregation_group_name_active,priority:1" json:"name"`
	Alias        string         `gorm:"size:128" json:"alias"`
	Description  string         `gorm:"size:512" json:"description"`
	FluentType   string         `gorm:"size:32;not null;default:fluentd" json:"fluent_type"`
	Mode         string         `gorm:"size:32;not null;default:forward" json:"mode"`
	EndpointHost string         `gorm:"size:256" json:"endpoint_host"`
	EndpointPort int            `json:"endpoint_port"`
	EnableTLS    bool           `gorm:"default:false" json:"enable_tls"`
	SharedKey    string         `gorm:"size:512" json:"-"`
	HasSharedKey bool           `gorm:"-" json:"has_shared_key"`
	ClusterID    *uint          `gorm:"index" json:"cluster_id"`
	Cluster      *Cluster       `json:"cluster,omitempty"`
	Nodes        []Node         `gorm:"foreignKey:AggregationGroupID" json:"nodes,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;uniqueIndex:idx_aggregation_group_name_active,priority:2" json:"-"`
}

type NodeFluentProfile struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	NodeID               uint       `gorm:"uniqueIndex;not null" json:"node_id"`
	Node                 *Node      `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	LoadedPlugins        string     `gorm:"type:text" json:"loaded_plugins"`
	SupportsHotReload    bool       `gorm:"default:false" json:"supports_hot_reload"`
	SupportsMultiline    bool       `gorm:"default:false" json:"supports_multiline"`
	SupportsStorageLayer bool       `gorm:"default:false" json:"supports_storage_layer"`
	SupportsForwardTLS   bool       `gorm:"default:false" json:"supports_forward_tls"`
	SupportsMetricsAPI   bool       `gorm:"default:false" json:"supports_metrics_api"`
	Metadata             string     `gorm:"type:text" json:"metadata"`
	LastReportedAt       *time.Time `json:"last_reported_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
