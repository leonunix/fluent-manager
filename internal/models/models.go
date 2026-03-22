package models

import (
	"time"

	"gorm.io/gorm"
)

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
	Resource string `gorm:"size:64;not null" json:"resource"` // nodes, configs, users, roles, groups
	Action   string `gorm:"size:32;not null" json:"action"`   // create, read, update, delete, deploy
	Name     string `gorm:"uniqueIndex;size:128;not null" json:"name"`
}

// NodeGroup allows organizing nodes into logical groups.
type NodeGroup struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	Nodes       []Node         `gorm:"foreignKey:GroupID" json:"nodes,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Node represents a Fluent Bit / Fluentd agent node.
type Node struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	NodeUID       string         `gorm:"uniqueIndex;size:128;not null" json:"node_uid"` // unique agent identifier
	Hostname      string         `gorm:"size:256;not null" json:"hostname"`
	IPAddress     string         `gorm:"size:64" json:"ip_address"`
	OS            string         `gorm:"size:64" json:"os"`
	AgentVersion  string         `gorm:"size:32" json:"agent_version"`
	FluentType    string         `gorm:"size:32" json:"fluent_type"` // fluentbit, fluentd
	FluentVersion string         `gorm:"size:32" json:"fluent_version"`
	Status        string         `gorm:"size:32;default:offline" json:"status"` // online, offline, error
	GroupID       *uint          `json:"group_id"`
	Group         *NodeGroup     `json:"group,omitempty"`
	Labels        string         `gorm:"type:text" json:"labels"` // JSON key-value pairs
	ConfigID      *uint          `json:"config_id"`
	Config        *ConfigVersion `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	LastHeartbeat *time.Time     `json:"last_heartbeat"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// ConfigTemplate is a reusable Fluent configuration template.
type ConfigTemplate struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Description string          `gorm:"size:512" json:"description"`
	FluentType  string          `gorm:"size:32;not null" json:"fluent_type"` // fluentbit, fluentd
	Content     string          `gorm:"type:text;not null" json:"content"`   // template with variables
	Variables   string          `gorm:"type:text" json:"variables"`          // JSON schema of variables
	Versions    []ConfigVersion `gorm:"foreignKey:TemplateID" json:"versions,omitempty"`
	CreatedBy   uint            `json:"created_by"`
	Creator     *User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

// ConfigVersion is a versioned, immutable snapshot of a rendered configuration.
type ConfigVersion struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	TemplateID uint           `json:"template_id"`
	Template   *ConfigTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Version    int            `gorm:"not null" json:"version"`
	Content    string         `gorm:"type:text;not null" json:"content"` // rendered config
	Hash       string         `gorm:"size:64;not null" json:"hash"`     // SHA-256 of content
	Comment    string         `gorm:"size:512" json:"comment"`
	CreatedBy  uint           `json:"created_by"`
	Creator    *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// DeployTask tracks a configuration deployment to nodes.
type DeployTask struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ConfigID     uint           `json:"config_id"`
	Config       *ConfigVersion `gorm:"foreignKey:ConfigID" json:"config,omitempty"`
	Status       string         `gorm:"size:32;default:pending" json:"status"` // pending, running, completed, failed
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
	Status       string    `gorm:"size:32;default:pending" json:"status"` // pending, success, failed
	Message      string    `gorm:"type:text" json:"message"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

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
