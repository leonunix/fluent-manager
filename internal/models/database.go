package models

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		dialector = sqlite.Open(cfg.DSN)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	if err := DB.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&NodeGroup{},
		&Node{},
		&ConfigTemplate{},
		&ConfigVersion{},
		&DeployTask{},
		&DeployRecord{},
		&AuditLog{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	seedDefaults()
	return nil
}

func seedDefaults() {
	// Seed permissions
	permissions := []Permission{
		{Name: "nodes:create", Resource: "nodes", Action: "create"},
		{Name: "nodes:read", Resource: "nodes", Action: "read"},
		{Name: "nodes:update", Resource: "nodes", Action: "update"},
		{Name: "nodes:delete", Resource: "nodes", Action: "delete"},
		{Name: "configs:create", Resource: "configs", Action: "create"},
		{Name: "configs:read", Resource: "configs", Action: "read"},
		{Name: "configs:update", Resource: "configs", Action: "update"},
		{Name: "configs:delete", Resource: "configs", Action: "delete"},
		{Name: "configs:deploy", Resource: "configs", Action: "deploy"},
		{Name: "users:create", Resource: "users", Action: "create"},
		{Name: "users:read", Resource: "users", Action: "read"},
		{Name: "users:update", Resource: "users", Action: "update"},
		{Name: "users:delete", Resource: "users", Action: "delete"},
		{Name: "roles:create", Resource: "roles", Action: "create"},
		{Name: "roles:read", Resource: "roles", Action: "read"},
		{Name: "roles:update", Resource: "roles", Action: "update"},
		{Name: "roles:delete", Resource: "roles", Action: "delete"},
		{Name: "groups:create", Resource: "groups", Action: "create"},
		{Name: "groups:read", Resource: "groups", Action: "read"},
		{Name: "groups:update", Resource: "groups", Action: "update"},
		{Name: "groups:delete", Resource: "groups", Action: "delete"},
		{Name: "audit:read", Resource: "audit", Action: "read"},
	}
	for _, p := range permissions {
		DB.FirstOrCreate(&p, Permission{Name: p.Name})
	}

	// Seed admin role with all permissions
	var allPerms []Permission
	DB.Find(&allPerms)

	var adminRole Role
	result := DB.Where("name = ?", "admin").First(&adminRole)
	if result.RowsAffected == 0 {
		adminRole = Role{Name: "admin", Description: "System administrator with full access"}
		DB.Create(&adminRole)
	}
	DB.Model(&adminRole).Association("Permissions").Replace(allPerms)

	// Seed viewer role
	var viewerRole Role
	result = DB.Where("name = ?", "viewer").First(&viewerRole)
	if result.RowsAffected == 0 {
		var readPerms []Permission
		DB.Where("action = ?", "read").Find(&readPerms)
		viewerRole = Role{Name: "viewer", Description: "Read-only access"}
		DB.Create(&viewerRole)
		DB.Model(&viewerRole).Association("Permissions").Replace(readPerms)
	}

	// Seed operator role
	var operatorRole Role
	result = DB.Where("name = ?", "operator").First(&operatorRole)
	if result.RowsAffected == 0 {
		var opPerms []Permission
		DB.Where("resource IN ? AND action IN ?",
			[]string{"nodes", "configs", "groups"},
			[]string{"create", "read", "update", "deploy"},
		).Find(&opPerms)
		operatorRole = Role{Name: "operator", Description: "Can manage nodes and configs"}
		DB.Create(&operatorRole)
		DB.Model(&operatorRole).Association("Permissions").Replace(opPerms)
	}

	// Seed default admin user
	var adminUser User
	result = DB.Where("username = ?", "admin").First(&adminUser)
	if result.RowsAffected == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("WARNING: failed to hash admin password: %v", err)
			return
		}
		adminUser = User{
			Username:     "admin",
			Email:        "admin@localhost",
			DisplayName:  "Administrator",
			PasswordHash: string(hash),
			AuthSource:   "local",
			IsActive:     true,
		}
		DB.Create(&adminUser)
		DB.Model(&adminUser).Association("Roles").Append(&adminRole)
		log.Println("Default admin user created (admin / admin123)")
	}
}

func HashConfig(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}
