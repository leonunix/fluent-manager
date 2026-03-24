package services

import (
	"errors"
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
