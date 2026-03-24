package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbCounter atomic.Int64

// NewTestDB creates an isolated SQLite database with all tables migrated.
// Uses temp files to avoid SQLite in-memory multi-connection issues.
// It also sets models.DB so scope.go functions work correctly.
func NewTestDB() *gorm.DB {
	n := dbCounter.Add(1)
	dsn := filepath.Join(os.TempDir(), fmt.Sprintf("fluent_test_%d.db", n))
	// Clean up any leftover file
	os.Remove(dsn)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserScope{},
		&models.AuthSettings{},
		&models.Environment{},
		&models.DataCenter{},
		&models.Region{},
		&models.Cluster{},
		&models.ClusterMatchRule{},
		&models.AggregationGroup{},
		&models.OutputTarget{},
		&models.AgentAccessKey{},
		&models.Node{},
		&models.NodeFluentProfile{},
		&models.AgentPolicy{},
		&models.NodeMetrics{},
		&models.RemoteCommand{},
		&models.NodeLog{},
		&models.ConfigTemplate{},
		&models.ConfigVersion{},
		&models.ConfigModule{},
		&models.ConfigModuleVersion{},
		&models.RenderedConfig{},
		&models.LogPipeline{},
		&models.ConfigAnalysisResult{},
		&models.ConfigAnalysisFinding{},
		&models.NodeRuntimeState{},
		&models.DeployTask{},
		&models.DeployRecord{},
		&models.BootstrapHost{},
		&models.BootstrapTask{},
		&models.BootstrapRecord{},
		&models.AuditLog{},
	); err != nil {
		panic("failed to migrate test db: " + err.Error())
	}
	if err := models.MigrateSoftDeleteUniqueIndexes(db); err != nil {
		panic("failed to migrate soft delete unique indexes: " + err.Error())
	}

	// Set global DB for scope.go functions
	models.DB = db
	return db
}
