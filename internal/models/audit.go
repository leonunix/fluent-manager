package models

import "time"

// AuditLog records all important operations.
type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Username     string    `gorm:"size:128" json:"username"`
	Action       string    `gorm:"size:64;not null" json:"action"`
	Resource     string    `gorm:"size:64" json:"resource"`
	ResourceType string    `gorm:"size:64;index" json:"resource_type"`
	ResourceID   uint      `gorm:"index" json:"resource_id"`
	Detail       string    `gorm:"type:text" json:"detail"`
	IP           string    `gorm:"size:64" json:"ip"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}
