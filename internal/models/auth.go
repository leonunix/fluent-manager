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
	Groups       []Group        `gorm:"many2many:user_groups;" json:"groups"`
	Scopes       []UserScope    `gorm:"foreignKey:UserID" json:"scopes"`
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

// UserScope binds a user to specific topology resources they can access.
// If a user has no UserScope records, they are treated as having global access (admin).
type UserScope struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ScopeType string         `gorm:"size:32;not null" json:"scope_type"` // datacenter, region, cluster
	ScopeID   uint           `gorm:"not null" json:"scope_id"`
	ScopeName string         `gorm:"size:128" json:"scope_name"` // denormalized for display
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Group represents a user group that can be mapped from LDAP/SAML external groups.
type Group struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Description string         `gorm:"size:256" json:"description"`
	MemberCount int64          `gorm:"-" json:"member_count"`
	Roles       []Role         `gorm:"many2many:group_roles;" json:"roles"`
	Scopes      []GroupScope   `gorm:"foreignKey:GroupID" json:"scopes"`
	Users       []User         `gorm:"many2many:user_groups;" json:"users,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// GroupScope binds a group to specific topology resources. Members inherit these scopes.
type GroupScope struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"index;not null" json:"group_id"`
	ScopeType string         `gorm:"size:32;not null" json:"scope_type"` // datacenter, region, cluster
	ScopeID   uint           `gorm:"not null" json:"scope_id"`
	ScopeName string         `gorm:"size:128" json:"scope_name"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ExternalGroupMapping maps an external LDAP/SAML group name to a system group.
type ExternalGroupMapping struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Source            string         `gorm:"size:32;not null;uniqueIndex:idx_source_extname" json:"source"` // ldap, saml
	ExternalGroupName string         `gorm:"size:512;not null;uniqueIndex:idx_source_extname" json:"external_group_name"`
	GroupID           uint           `gorm:"not null" json:"group_id"`
	Group             *Group         `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// AuthSettings stores runtime LDAP/SAML configuration in the database.
type AuthSettings struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Provider          string    `gorm:"size:32;uniqueIndex;not null" json:"provider"` // ldap, saml
	Config            string    `gorm:"type:text" json:"config"`                      // JSON blob
	GroupSyncStrategy string    `gorm:"size:32;default:always" json:"group_sync_strategy"` // always, first_login
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
