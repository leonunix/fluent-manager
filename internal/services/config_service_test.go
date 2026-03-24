package services

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"github.com/fluent-manager/fluent-manager/internal/testutil"
	"gorm.io/gorm"
)

func setupConfigTest(t *testing.T) (*gorm.DB, ConfigService) {
	t.Helper()
	db := testutil.NewTestDB()
	svc := NewConfigService(db)
	return db, svc
}

func TestCreateTemplate(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:        "nginx-fb",
		Description: "Nginx FluentBit config",
		FluentType:  "fluentbit",
		Content:     "[INPUT]\n  Name tail",
	}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Name != "nginx-fb" || tpl.FluentType != "fluentbit" {
		t.Error("template not created correctly")
	}
	if tpl.SourceType != "manual" {
		t.Fatalf("expected manual source type, got %q", tpl.SourceType)
	}
}

func TestListTemplates(t *testing.T) {
	_, svc := setupConfigTest(t)

	svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl-fb", FluentType: "fluentbit", Content: "c1"}, 1)
	svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl-fd", FluentType: "fluentd", Content: "c2"}, 1)

	// All
	templates, total, _ := svc.ListTemplates("", "", 1, 10)
	if total != 2 {
		t.Errorf("expected 2 templates, got %d", total)
	}

	// Filter by type
	templates, total, _ = svc.ListTemplates("fluentbit", "", 1, 10)
	if total != 1 || templates[0].Name != "tpl-fb" {
		t.Error("fluent type filter not working")
	}

	// Search
	templates, total, _ = svc.ListTemplates("", "fb", 1, 10)
	if total != 1 {
		t.Error("search filter not working")
	}
}

func TestGetTemplate(t *testing.T) {
	_, svc := setupConfigTest(t)

	created, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", FluentType: "fluentbit", Content: "content"}, 1)
	got, err := svc.GetTemplate(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "tpl1" {
		t.Error("wrong template returned")
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	_, svc := setupConfigTest(t)
	_, err := svc.GetTemplate(999)
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestUpdateTemplate(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", Description: "old desc", FluentType: "fluentbit", Content: "old"}, 1)

	updated, err := svc.UpdateTemplate(tpl.ID, &ConfigTemplateInput{
		Name:        "tpl1-updated",
		Description: "new desc",
		FluentType:  "fluentd",
		Content:     "new content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "tpl1-updated" || updated.FluentType != "fluentd" {
		t.Error("template not updated")
	}
}

func TestCreateTemplate_WithAssemblyMetadata(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:          "edge-assembled",
		Description:   "created from modules",
		FluentType:    "fluentbit",
		Content:       "[INPUT]\n  Name tail",
		SourceType:    "module_assembly",
		SourceModules: `[{"module_id":1,"module_name":"tail-input","module_type":"input"}]`,
		FlowLayout:    `{"builder":"wizard","goal":"edge_collection","destinations":[{"name":"central-forward"}]}`,
	}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.SourceType != "module_assembly" {
		t.Fatalf("expected module_assembly, got %q", tpl.SourceType)
	}
	if !strings.Contains(tpl.SourceModules, `"module_name":"tail-input"`) {
		t.Fatalf("expected source modules to be persisted, got %s", tpl.SourceModules)
	}
	if !strings.Contains(tpl.FlowLayout, `"builder":"wizard"`) {
		t.Fatalf("expected flow layout to be persisted, got %s", tpl.FlowLayout)
	}
}

func TestDeleteTemplate(t *testing.T) {
	db, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", FluentType: "fluentbit", Content: "content"}, 1)

	// Add a version
	db.Create(&models.ConfigVersion{TemplateID: tpl.ID, Version: 1, Content: "v1", Hash: "h1"})

	if err := svc.DeleteTemplate(tpl.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify versions also deleted
	var count int64
	db.Model(&models.ConfigVersion{}).Where("template_id = ?", tpl.ID).Count(&count)
	if count != 0 {
		t.Error("versions should be deleted with template")
	}
}

func TestDeleteTemplateAllowsRecreateSameName(t *testing.T) {
	db, svc := setupConfigTest(t)

	created, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:       "reusable-template",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail",
	}, 1)
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	if err := svc.DeleteTemplate(created.ID); err != nil {
		t.Fatalf("delete template: %v", err)
	}

	recreated, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:       "reusable-template",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail\n  Tag app.logs",
	}, 1)
	if err != nil {
		t.Fatalf("recreate template with same name: %v", err)
	}
	if recreated.ID == created.ID {
		t.Fatalf("expected recreated template to be a fresh row, got same id %d", recreated.ID)
	}

	var total int64
	db.Unscoped().Model(&models.ConfigTemplate{}).Where("name = ?", "reusable-template").Count(&total)
	if total != 1 {
		t.Fatalf("expected exactly one persisted template row, got %d", total)
	}
}

func TestCreateTemplatePurgesLegacySoftDeletedName(t *testing.T) {
	db, svc := setupConfigTest(t)

	legacy := models.ConfigTemplate{
		Name:       "legacy-imported-template",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 1",
	}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("seed legacy template: %v", err)
	}
	if err := db.Delete(&legacy).Error; err != nil {
		t.Fatalf("soft delete legacy template: %v", err)
	}

	created, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:       "legacy-imported-template",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 5",
	}, 1)
	if err != nil {
		t.Fatalf("create template after soft delete: %v", err)
	}

	var rows []models.ConfigTemplate
	if err := db.Unscoped().
		Where("name = ?", "legacy-imported-template").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		t.Fatalf("query rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected legacy tombstone to be purged, got %d rows", len(rows))
	}
	if rows[0].ID != created.ID || rows[0].DeletedAt.Valid {
		t.Fatalf("expected only the new active template row to remain, got %+v", rows[0])
	}
}

func TestCreateVersion(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{
		Name:          "tpl1",
		FluentType:    "fluentbit",
		Content:       "content",
		SourceType:    "module_assembly",
		SourceModules: `[{"module_id":1,"module_name":"tail-input","module_type":"input"}]`,
		FlowLayout:    `{"builder":"wizard","goal":"edge_collection"}`,
	}, 1)

	v1, err := svc.CreateVersion(tpl.ID, 1, "config v1", "initial version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v1.Version != 1 || v1.Hash == "" {
		t.Error("version not created correctly")
	}
	if v1.SourceType != "module_assembly" {
		t.Fatalf("expected source type to be copied into version, got %q", v1.SourceType)
	}
	if !strings.Contains(v1.SourceModules, `"module_type":"input"`) {
		t.Fatalf("expected source modules to be copied into version, got %s", v1.SourceModules)
	}

	v2, _ := svc.CreateVersion(tpl.ID, 1, "config v2", "second version")
	if v2.Version != 2 {
		t.Errorf("expected version 2, got %d", v2.Version)
	}
}

func TestCreateVersion_InvalidTemplate(t *testing.T) {
	_, svc := setupConfigTest(t)
	_, err := svc.CreateVersion(999, 1, "content", "comment")
	if err == nil {
		t.Error("expected error for non-existent template")
	}
}

func TestListVersions(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", FluentType: "fluentbit", Content: "content"}, 1)
	svc.CreateVersion(tpl.ID, 1, "v1", "")
	svc.CreateVersion(tpl.ID, 1, "v2", "")

	versions, err := svc.ListVersions(tpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
	// Should be ordered DESC
	if versions[0].Version != 2 {
		t.Error("versions should be ordered by version DESC")
	}
}

func TestGetVersion(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", FluentType: "fluentbit", Content: "content"}, 1)
	v, _ := svc.CreateVersion(tpl.ID, 1, "config content", "test")

	got, err := svc.GetVersion(v.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Content != "config content" {
		t.Error("wrong version content")
	}
}

func TestCreateVersion_HashConsistency(t *testing.T) {
	_, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate(&ConfigTemplateInput{Name: "tpl1", FluentType: "fluentbit", Content: "content"}, 1)

	v1, _ := svc.CreateVersion(tpl.ID, 1, "same content", "")
	v2, _ := svc.CreateVersion(tpl.ID, 1, "same content", "")
	v3, _ := svc.CreateVersion(tpl.ID, 1, "different", "")

	if v1.Hash != v2.Hash {
		t.Error("same content should produce same hash")
	}
	if v1.Hash == v3.Hash {
		t.Error("different content should produce different hash")
	}
}

func TestCreateModuleAndListModules(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "tail-input",
		ModuleType: "input",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail",
		Variables:  `{"path":"/var/log/*.log"}`,
	}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if module.Name != "tail-input" || module.ModuleType != "input" {
		t.Fatalf("module not created correctly: %+v", module)
	}

	modules, total, err := svc.ListModules("fluentbit", "input", "", 1, 10)
	if err != nil {
		t.Fatalf("list modules: %v", err)
	}
	if total != 1 || len(modules) != 1 {
		t.Fatalf("expected 1 module, got total=%d len=%d", total, len(modules))
	}
}

func TestListModulesSearchesOutputFieldsAndReturnsNewestFirst(t *testing.T) {
	_, svc := setupConfigTest(t)

	older, err := svc.CreateModule(&ConfigModuleInput{
		Name:        "shared-opensearch-output",
		Description: "OpenSearch output preset",
		ModuleType:  "output",
		FluentType:  "shared",
		Content:     "[OUTPUT]\n  Name opensearch\n  Host search.internal",
		Variables:   `{"host":"search.internal","match":"nova.*"}`,
		PresetKind:  "output",
		PresetKey:   "opensearch",
	}, 1)
	if err != nil {
		t.Fatalf("create older output module: %v", err)
	}

	newer, err := svc.CreateModule(&ConfigModuleInput{
		Name:        "tail-nova-input",
		Description: "Nova tail input",
		ModuleType:  "input",
		FluentType:  "fluentbit",
		Content:     "[INPUT]\n  Name tail\n  Path /var/log/nova/*.log",
	}, 1)
	if err != nil {
		t.Fatalf("create newer input module: %v", err)
	}

	modules, total, err := svc.ListModules("", "", "opensearch", 1, 10)
	if err != nil {
		t.Fatalf("list modules by preset key search: %v", err)
	}
	if total != 1 || len(modules) != 1 || modules[0].ID != older.ID {
		t.Fatalf("expected opensearch output module to be found, total=%d len=%d first=%+v", total, len(modules), modules)
	}

	modules, total, err = svc.ListModules("", "", "search.internal", 1, 10)
	if err != nil {
		t.Fatalf("list modules by content search: %v", err)
	}
	if total != 1 || len(modules) != 1 || modules[0].ID != older.ID {
		t.Fatalf("expected content search to find output module, total=%d len=%d first=%+v", total, len(modules), modules)
	}

	modules, total, err = svc.ListModules("", "", "", 1, 10)
	if err != nil {
		t.Fatalf("list modules newest first: %v", err)
	}
	if total != 2 || len(modules) != 2 {
		t.Fatalf("expected 2 modules, got total=%d len=%d", total, len(modules))
	}
	if modules[0].ID != newer.ID || modules[1].ID != older.ID {
		t.Fatalf("expected newest module first, got ids %d then %d", modules[0].ID, modules[1].ID)
	}
}

func TestCreateModuleRejectsInvalidVariables(t *testing.T) {
	_, svc := setupConfigTest(t)

	_, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "bad-module",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "<filter **>",
		Variables:  `not-json`,
	}, 1)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestCreateModuleVersion(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "shared-parser",
		ModuleType: "parser",
		FluentType: "shared",
		Content:    "<parse>@type json</parse>",
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	v1, err := svc.CreateModuleVersion(module.ID, 1, "<parse>@type json</parse>", `{"keep_time_key":true}`, "initial")
	if err != nil {
		t.Fatalf("create module version: %v", err)
	}
	v2, err := svc.CreateModuleVersion(module.ID, 1, "<parse>@type regexp</parse>", `{"expression":"^foo$"}`, "regexp")
	if err != nil {
		t.Fatalf("create module version 2: %v", err)
	}
	if v1.Version != 1 || v2.Version != 2 {
		t.Fatalf("expected versions 1 and 2, got %d and %d", v1.Version, v2.Version)
	}
}

func TestDeleteModuleRejectsBuiltin(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "builtin-tail",
		ModuleType: "input",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail",
		IsBuiltin:  true,
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	err = svc.DeleteModule(module.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestDeleteModulesBatch(t *testing.T) {
	db, svc := setupConfigTest(t)

	first, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "batch-filter-1",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "[FILTER]\n  Name modify",
	}, 1)
	if err != nil {
		t.Fatalf("create first module: %v", err)
	}
	second, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "batch-filter-2",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "[FILTER]\n  Name modify",
	}, 1)
	if err != nil {
		t.Fatalf("create second module: %v", err)
	}

	if err := svc.DeleteModules([]uint{first.ID, second.ID, first.ID}); err != nil {
		t.Fatalf("delete modules: %v", err)
	}

	var count int64
	db.Model(&models.ConfigModule{}).Where("id IN ?", []uint{first.ID, second.ID}).Count(&count)
	if count != 0 {
		t.Fatalf("expected modules to be deleted, got %d remaining", count)
	}
}

func TestDeleteModulesBatchRejectsBuiltin(t *testing.T) {
	db, svc := setupConfigTest(t)

	regular, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "regular-filter",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "[FILTER]\n  Name modify",
	}, 1)
	if err != nil {
		t.Fatalf("create regular module: %v", err)
	}
	builtin, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "builtin-filter",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "[FILTER]\n  Name modify",
		IsBuiltin:  true,
	}, 1)
	if err != nil {
		t.Fatalf("create builtin module: %v", err)
	}

	err = svc.DeleteModules([]uint{regular.ID, builtin.ID})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	var count int64
	db.Model(&models.ConfigModule{}).Where("id IN ?", []uint{regular.ID, builtin.ID}).Count(&count)
	if count != 2 {
		t.Fatalf("expected no modules to be deleted, got %d remaining", count)
	}
}

func TestDeleteModuleAllowsRecreateSameIdentity(t *testing.T) {
	db, svc := setupConfigTest(t)

	created, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "reusable-service-module",
		ModuleType: "service",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 1",
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	if err := svc.DeleteModule(created.ID); err != nil {
		t.Fatalf("delete module: %v", err)
	}

	recreated, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "reusable-service-module",
		ModuleType: "service",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 5",
	}, 1)
	if err != nil {
		t.Fatalf("recreate module with same identity: %v", err)
	}
	if recreated.ID == created.ID {
		t.Fatalf("expected recreated module to be a fresh row, got same id %d", recreated.ID)
	}

	var total int64
	db.Unscoped().Model(&models.ConfigModule{}).
		Where("name = ? AND module_type = ? AND fluent_type = ?", "reusable-service-module", "service", "fluentbit").
		Count(&total)
	if total != 1 {
		t.Fatalf("expected exactly one persisted row after recreate, got %d", total)
	}
}

func TestDeleteModuleRejectsTemplateReference(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "protected-filter",
		ModuleType: "filter",
		FluentType: "shared",
		Content:    "[FILTER]\n  Name modify",
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	_, err = svc.CreateTemplate(&ConfigTemplateInput{
		Name:          "assembled-template",
		FluentType:    "fluentbit",
		Content:       "[FILTER]\n  Name modify",
		SourceType:    "module_assembly",
		SourceModules: fmt.Sprintf(`[{"module_id":%d,"module_name":"%s","module_type":"filter"}]`, module.ID, module.Name),
	}, 1)
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	err = svc.DeleteModule(module.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if !strings.Contains(err.Error(), "assembled-template") {
		t.Fatalf("expected usage detail in error, got %v", err)
	}
}

func TestDeleteModuleRejectsTemplateVersionReference(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "protected-service",
		ModuleType: "service",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 1",
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	template, err := svc.CreateTemplate(&ConfigTemplateInput{
		Name:          "versioned-template",
		FluentType:    "fluentbit",
		Content:       "[SERVICE]\n  Flush 1",
		SourceType:    "module_assembly",
		SourceModules: fmt.Sprintf(`[{"module_id":%d,"module_name":"%s","module_type":"service"}]`, module.ID, module.Name),
	}, 1)
	if err != nil {
		t.Fatalf("create template: %v", err)
	}

	if _, err := svc.CreateVersion(template.ID, 1, "[SERVICE]\n  Flush 1", "initial"); err != nil {
		t.Fatalf("create version: %v", err)
	}

	if _, err := svc.UpdateTemplate(template.ID, &ConfigTemplateInput{
		Name:          "versioned-template",
		FluentType:    "fluentbit",
		Content:       "[SERVICE]\n  Flush 5",
		SourceType:    "module_assembly",
		SourceModules: `[]`,
	}); err != nil {
		t.Fatalf("update template: %v", err)
	}

	err = svc.DeleteModule(module.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if !strings.Contains(err.Error(), "versioned-template v1") {
		t.Fatalf("expected version usage detail in error, got %v", err)
	}
}

func TestCreateModulePurgesLegacySoftDeletedIdentity(t *testing.T) {
	db, svc := setupConfigTest(t)

	legacy := models.ConfigModule{
		Name:       "legacy-imported-service",
		ModuleType: "service",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 1",
	}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("seed legacy module: %v", err)
	}
	if err := db.Delete(&legacy).Error; err != nil {
		t.Fatalf("soft delete legacy module: %v", err)
	}

	created, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "legacy-imported-service",
		ModuleType: "service",
		FluentType: "fluentbit",
		Content:    "[SERVICE]\n  Flush 10",
	}, 1)
	if err != nil {
		t.Fatalf("create module after legacy soft delete: %v", err)
	}

	var rows []models.ConfigModule
	if err := db.Unscoped().
		Where("name = ? AND module_type = ? AND fluent_type = ?", "legacy-imported-service", "service", "fluentbit").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		t.Fatalf("query rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected legacy tombstone to be purged, got %d rows", len(rows))
	}
	if rows[0].ID != created.ID || rows[0].DeletedAt.Valid {
		t.Fatalf("expected only the new active row to remain, got %+v", rows[0])
	}
}

func TestPreviewRenderedConfig(t *testing.T) {
	_, svc := setupConfigTest(t)

	inputModule, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "tail-input",
		ModuleType: "input",
		FluentType: "fluentbit",
		Content:    "[INPUT]\n  Name tail\n  Path {{.path}}",
	}, 1)
	if err != nil {
		t.Fatalf("create input module: %v", err)
	}

	outputModule, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "stdout-output",
		ModuleType: "output",
		FluentType: "shared",
		Content:    "[OUTPUT]\n  Name stdout\n  Match {{.match}}",
	}, 1)
	if err != nil {
		t.Fatalf("create output module: %v", err)
	}

	rendered, err := svc.PreviewRenderedConfig(&RenderedConfigPreviewInput{
		Name:       "preview-1",
		FluentType: "fluentbit",
		Modules: []RenderModuleRef{
			{ModuleID: outputModule.ID},
			{ModuleID: inputModule.ID},
		},
		Variables: `{"path":"/var/log/app.log","match":"*"}`,
	}, 1)
	if err != nil {
		t.Fatalf("preview render: %v", err)
	}
	if rendered.Hash == "" {
		t.Fatal("expected rendered config hash")
	}
	if !strings.Contains(rendered.Content, "Path /var/log/app.log") {
		t.Fatalf("expected rendered input content, got %s", rendered.Content)
	}
	if !strings.Contains(rendered.Content, "Match *") {
		t.Fatalf("expected rendered output content, got %s", rendered.Content)
	}
	if strings.Index(rendered.Content, "module:input") > strings.Index(rendered.Content, "module:output") {
		t.Fatalf("expected input module to render before output module, got %s", rendered.Content)
	}
}

func TestPreviewRenderedConfigSupportsPerModuleVariables(t *testing.T) {
	_, svc := setupConfigTest(t)

	outputModule, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "http-output",
		ModuleType: "output",
		FluentType: "shared",
		Content:    "[OUTPUT]\n  Name http\n  Match {{.match}}\n  Host {{.host}}\n  Port {{.port}}",
	}, 1)
	if err != nil {
		t.Fatalf("create output module: %v", err)
	}

	rendered, err := svc.PreviewRenderedConfig(&RenderedConfigPreviewInput{
		Name:       "preview-multi-output",
		FluentType: "fluentbit",
		Modules: []RenderModuleRef{
			{
				ModuleID:  outputModule.ID,
				Variables: `{"host":"primary.internal","port":8080,"match":"app.primary"}`,
			},
			{
				ModuleID:  outputModule.ID,
				Variables: `{"host":"backup.internal","port":8081,"match":"app.backup"}`,
			},
		},
		Variables: `{"match":"*"}`,
	}, 1)
	if err != nil {
		t.Fatalf("preview render with module variables: %v", err)
	}
	if !strings.Contains(rendered.Content, "Host primary.internal") {
		t.Fatalf("expected first output override, got %s", rendered.Content)
	}
	if !strings.Contains(rendered.Content, "Host backup.internal") {
		t.Fatalf("expected second output override, got %s", rendered.Content)
	}
	if !strings.Contains(rendered.Content, "Match app.primary") || !strings.Contains(rendered.Content, "Match app.backup") {
		t.Fatalf("expected per-module match overrides, got %s", rendered.Content)
	}
}

func TestPreviewRenderedConfigRejectsRuntimeMismatch(t *testing.T) {
	_, svc := setupConfigTest(t)

	module, err := svc.CreateModule(&ConfigModuleInput{
		Name:       "fd-service",
		ModuleType: "service",
		FluentType: "fluentd",
		Content:    "<system></system>",
	}, 1)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}

	_, err = svc.PreviewRenderedConfig(&RenderedConfigPreviewInput{
		FluentType: "fluentbit",
		Modules: []RenderModuleRef{
			{ModuleID: module.ID},
		},
	}, 1)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
