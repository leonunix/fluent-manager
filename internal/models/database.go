package models

import (
	"crypto/sha256"
	"fmt"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// OpenDB opens a GORM connection for the given driver and DSN without running migrations.
func OpenDB(driver, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	return db, nil
}

// InitDBWithConn opens a connection, runs migrations and seeds defaults.
// Returns the *gorm.DB without setting the global DB variable.
func InitDBWithConn(driver, dsn string) (*gorm.DB, error) {
	db, err := OpenDB(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := migrateAll(db); err != nil {
		return nil, err
	}

	seedDefaultsOn(db)
	return db, nil
}

// InitDB initializes the global DB variable with migrations and seed data.
func InitDB(cfg *config.DatabaseConfig) error {
	db, err := InitDBWithConn(cfg.Driver, cfg.DSN)
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func migrateAll(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&UserScope{},
		&Group{},
		&GroupScope{},
		&ExternalGroupMapping{},
		&AuthSettings{},
		&Environment{},
		&DataCenter{},
		&Region{},
		&Cluster{},
		&ClusterMatchRule{},
		&AggregationGroup{},
		&OutputTarget{},
		&AgentAccessKey{},
		&AgentArtifact{},
		&Node{},
		&NodeFluentProfile{},
		&AgentPolicy{},
		&NodeMetrics{},
		&NodeThroughputHour{},
		&RemoteCommand{},
		&NodeLog{},
		&ConfigTemplate{},
		&ConfigVersion{},
		&ConfigModule{},
		&ConfigModuleVersion{},
		&ConfigPipeline{},
		&RenderedConfig{},
		&LogPipeline{},
		&ConfigAnalysisResult{},
		&ConfigAnalysisFinding{},
		&NodeRuntimeState{},
		&DeployTask{},
		&DeployRecord{},
		&AgentUpgradeTask{},
		&AgentUpgradeRecord{},
		&BootstrapHost{},
		&BootstrapTask{},
		&BootstrapRecord{},
		&AuditLog{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	if err := MigrateSoftDeleteUniqueIndexes(db); err != nil {
		return fmt.Errorf("failed to migrate soft delete unique indexes: %w", err)
	}
	return nil
}

func MigrateSoftDeleteUniqueIndexes(db *gorm.DB) error {
	indexes := []struct {
		model  interface{}
		legacy string
		target string
	}{
		{
			model:  &AggregationGroup{},
			legacy: "idx_aggregation_groups_name",
			target: "idx_aggregation_group_name_active",
		},
		{
			model:  &LogPipeline{},
			legacy: "idx_log_pipelines_name",
			target: "idx_log_pipeline_name_active",
		},
		{
			model:  &OutputTarget{},
			legacy: "idx_output_targets_name",
			target: "idx_output_target_name_active",
		},
	}

	for _, item := range indexes {
		if db.Migrator().HasIndex(item.model, item.legacy) {
			if err := db.Migrator().DropIndex(item.model, item.legacy); err != nil {
				return err
			}
		}
		if !db.Migrator().HasIndex(item.model, item.target) {
			if err := db.Migrator().CreateIndex(item.model, item.target); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedDefaultsOn seeds environments, permissions, and roles on the given db.
func seedDefaultsOn(db *gorm.DB) {
	defaultEnvs := []Environment{
		{Name: "production", Alias: "生产环境", Color: "#dc3545", SortOrder: 1, Description: "Production environment"},
		{Name: "staging", Alias: "预发布环境", Color: "#ffc107", SortOrder: 2, Description: "Staging / pre-production"},
		{Name: "development", Alias: "开发环境", Color: "#17a2b8", SortOrder: 3, Description: "Development environment"},
		{Name: "testing", Alias: "测试环境", Color: "#6c757d", SortOrder: 4, Description: "Testing / QA environment"},
	}
	for _, env := range defaultEnvs {
		db.FirstOrCreate(&env, Environment{Name: env.Name})
	}
	seedBuiltinConfigModules(db)

	permissions := []Permission{
		{Name: "nodes:create", Resource: "nodes", Action: "create"},
		{Name: "nodes:read", Resource: "nodes", Action: "read"},
		{Name: "nodes:update", Resource: "nodes", Action: "update"},
		{Name: "nodes:delete", Resource: "nodes", Action: "delete"},
		{Name: "topology:create", Resource: "topology", Action: "create"},
		{Name: "topology:read", Resource: "topology", Action: "read"},
		{Name: "topology:update", Resource: "topology", Action: "update"},
		{Name: "topology:delete", Resource: "topology", Action: "delete"},
		{Name: "configs:create", Resource: "configs", Action: "create"},
		{Name: "configs:read", Resource: "configs", Action: "read"},
		{Name: "configs:update", Resource: "configs", Action: "update"},
		{Name: "configs:delete", Resource: "configs", Action: "delete"},
		{Name: "configs:deploy", Resource: "configs", Action: "deploy"},
		{Name: "agent_policies:create", Resource: "agent_policies", Action: "create"},
		{Name: "agent_policies:read", Resource: "agent_policies", Action: "read"},
		{Name: "agent_policies:update", Resource: "agent_policies", Action: "update"},
		{Name: "agent_policies:delete", Resource: "agent_policies", Action: "delete"},
		{Name: "agent_keys:create", Resource: "agent_keys", Action: "create"},
		{Name: "agent_keys:read", Resource: "agent_keys", Action: "read"},
		{Name: "agent_keys:update", Resource: "agent_keys", Action: "update"},
		{Name: "agent_keys:delete", Resource: "agent_keys", Action: "delete"},
		{Name: "users:create", Resource: "users", Action: "create"},
		{Name: "users:read", Resource: "users", Action: "read"},
		{Name: "users:update", Resource: "users", Action: "update"},
		{Name: "users:delete", Resource: "users", Action: "delete"},
		{Name: "roles:create", Resource: "roles", Action: "create"},
		{Name: "roles:read", Resource: "roles", Action: "read"},
		{Name: "roles:update", Resource: "roles", Action: "update"},
		{Name: "roles:delete", Resource: "roles", Action: "delete"},
		{Name: "audit:read", Resource: "audit", Action: "read"},
		{Name: "groups:create", Resource: "groups", Action: "create"},
		{Name: "groups:read", Resource: "groups", Action: "read"},
		{Name: "groups:update", Resource: "groups", Action: "update"},
		{Name: "groups:delete", Resource: "groups", Action: "delete"},
		{Name: "auth_settings:read", Resource: "auth_settings", Action: "read"},
		{Name: "auth_settings:update", Resource: "auth_settings", Action: "update"},
		{Name: "ai_settings:read", Resource: "ai_settings", Action: "read"},
		{Name: "ai_settings:update", Resource: "ai_settings", Action: "update"},
	}
	for _, p := range permissions {
		db.FirstOrCreate(&p, Permission{Name: p.Name})
	}

	var allPerms []Permission
	db.Find(&allPerms)

	var adminRole Role
	result := db.Where("name = ?", "admin").First(&adminRole)
	if result.RowsAffected == 0 {
		adminRole = Role{Name: "admin", Description: "System administrator with full access"}
		db.Create(&adminRole)
	}
	db.Model(&adminRole).Association("Permissions").Replace(allPerms)

	var viewerRole Role
	result = db.Where("name = ?", "viewer").First(&viewerRole)
	var readPerms []Permission
	db.Where("action = ? AND resource NOT IN ?", "read", []string{"agent_keys"}).Find(&readPerms)
	if result.RowsAffected == 0 {
		viewerRole = Role{Name: "viewer", Description: "Read-only access"}
		db.Create(&viewerRole)
	}
	db.Model(&viewerRole).Association("Permissions").Replace(readPerms)

	var operatorRole Role
	result = db.Where("name = ?", "operator").First(&operatorRole)
	var opPerms []Permission
	db.Where("resource IN ? AND action IN ?",
		[]string{"nodes", "configs", "topology", "agent_policies"},
		[]string{"create", "read", "update", "deploy"},
	).Find(&opPerms)
	if result.RowsAffected == 0 {
		operatorRole = Role{Name: "operator", Description: "Can manage nodes and configs"}
		db.Create(&operatorRole)
	}
	db.Model(&operatorRole).Association("Permissions").Replace(opPerms)

	// Admin user is created via the setup page on first run.
}

func HashConfig(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}
