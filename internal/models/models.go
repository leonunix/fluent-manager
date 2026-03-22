package models

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// Auth & RBAC
// ============================================================================

// User represents a system user (local, LDAP, or SAML).
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:128;not null" json:"username"`
	Email        string         `gorm:"size:256" json:"email"`
	DisplayName  string         `gorm:"size:256" json:"display_name"`
	PasswordHash string         `gorm:"size:256" json:"-"`
	AuthSource   string         `gorm:"size:32;default:local" json:"auth_source"` // local, ldap, saml
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	Roles        []Role         `gorm:"many2many:user_roles;" json:"roles"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Role represents a RBAC role.
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string         `gorm:"size:256" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Permission represents a fine-grained permission.
type Permission struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Resource string `gorm:"size:64;not null" json:"resource"`
	Action   string `gorm:"size:32;not null" json:"action"`
	Name     string `gorm:"uniqueIndex;size:128;not null" json:"name"`
}

// ============================================================================
// Resource Scope (user ↔ topology binding for RBAC)
// ============================================================================

// UserScope binds a user to specific topology resources they can access.
// If a user has no UserScope records, they are treated as having global access (admin).
type UserScope struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	User         *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ScopeType    string         `gorm:"size:32;not null" json:"scope_type"` // datacenter, region, cluster
	ScopeID      uint           `gorm:"not null" json:"scope_id"`
	ScopeName    string         `gorm:"size:128" json:"scope_name"` // denormalized for display
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// ============================================================================
// Environment
// ============================================================================

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

// ============================================================================
// Topology: DataCenter → Region → Cluster → Node
// ============================================================================

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
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:128;not null" json:"name"` // e.g. "log-cluster-01"
	Alias         string         `gorm:"size:128" json:"alias"`
	RegionID      uint           `gorm:"index;not null" json:"region_id"`
	Region        *Region        `json:"region,omitempty"`
	EnvironmentID *uint          `gorm:"index" json:"environment_id"`
	Environment   *Environment   `json:"environment,omitempty"`
	IsDefault     bool           `gorm:"default:false" json:"is_default"` // default cluster for unmatched nodes
	Nodes         []Node         `gorm:"foreignKey:ClusterID" json:"nodes,omitempty"`
	MatchRules    []ClusterMatchRule `gorm:"foreignKey:ClusterID" json:"match_rules,omitempty"`
	ConfigID      *uint          `json:"config_id"` // inherited config for all nodes in cluster
	Config        *ConfigVersion `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	Description   string         `gorm:"size:512" json:"description"`
	Tags          string         `gorm:"type:text" json:"tags"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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

// ============================================================================
// Node
// ============================================================================

// Node represents a Fluent Bit / Fluentd agent node.
type Node struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	NodeUID       string         `gorm:"uniqueIndex;size:128;not null" json:"node_uid"`
	Hostname      string         `gorm:"size:256;not null" json:"hostname"`
	IPAddress     string         `gorm:"size:64" json:"ip_address"`
	OS            string         `gorm:"size:64" json:"os"`
	AgentVersion  string         `gorm:"size:32" json:"agent_version"`
	FluentType    string         `gorm:"size:32" json:"fluent_type"`    // fluentbit, fluentd
	FluentVersion string         `gorm:"size:32" json:"fluent_version"`
	Status        string         `gorm:"size:32;default:offline" json:"status"` // online, offline, error
	ClusterID     *uint          `gorm:"index" json:"cluster_id"`
	Cluster       *Cluster       `json:"cluster,omitempty"`
	EnvironmentID *uint          `gorm:"index" json:"environment_id"` // can override cluster's environment
	Environment   *Environment   `json:"environment,omitempty"`
	Labels        string         `gorm:"type:text" json:"labels"` // JSON key-value pairs
	ConfigID      *uint          `json:"config_id"`                // node-level config (overrides cluster config)
	Config        *ConfigVersion `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	LastHeartbeat *time.Time     `json:"last_heartbeat"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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

// ============================================================================
// Configuration Management
// ============================================================================

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

// ============================================================================
// Deployment
// ============================================================================

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

// ============================================================================
// Agent Communication
// ============================================================================

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

// ============================================================================
// Audit
// ============================================================================

// AuditLog records all important operations.
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:128" json:"username"`
	Action    string    `gorm:"size:64;not null" json:"action"`
	Resource  string    `gorm:"size:64" json:"resource"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"size:64" json:"ip"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
