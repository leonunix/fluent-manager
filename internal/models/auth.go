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
