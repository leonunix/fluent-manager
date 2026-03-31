package models

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedBuiltinConfigModulesIncludesFluentBitParsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "config-seed.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := migrateAll(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	seedBuiltinConfigModules(db)

	var modules []ConfigModule
	if err := db.Where("module_type = ? AND fluent_type = ?", "parser", "fluentbit").Order("name ASC").Find(&modules).Error; err != nil {
		t.Fatalf("query parser modules: %v", err)
	}
	if len(modules) < 2 {
		t.Fatalf("expected builtin fluentbit parser modules to be seeded, got %d", len(modules))
	}

	contentByName := map[string]string{}
	for _, module := range modules {
		contentByName[module.Name] = module.Content
	}
	if contentByName["guided-fb-parser-docker"] == "" {
		t.Fatalf("expected guided-fb-parser-docker to be seeded, got %#v", contentByName)
	}
	if contentByName["guided-fb-parser-nginx"] == "" {
		t.Fatalf("expected guided-fb-parser-nginx to be seeded, got %#v", contentByName)
	}
}
