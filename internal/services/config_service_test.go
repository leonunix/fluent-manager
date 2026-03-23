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

	tpl, err := svc.CreateTemplate("nginx-fb", "Nginx FluentBit config", "fluentbit", "[INPUT]\n  Name tail", "", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Name != "nginx-fb" || tpl.FluentType != "fluentbit" {
		t.Error("template not created correctly")
	}
}

func TestListTemplates(t *testing.T) {
	_, svc := setupConfigTest(t)

	svc.CreateTemplate("tpl-fb", "", "fluentbit", "c1", "", 1)
	svc.CreateTemplate("tpl-fd", "", "fluentd", "c2", "", 1)

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

	created, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)
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

	tpl, _ := svc.CreateTemplate("tpl1", "old desc", "fluentbit", "old", "", 1)

	updated, err := svc.UpdateTemplate(tpl.ID, "tpl1-updated", "new desc", "fluentd", "new content", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "tpl1-updated" || updated.FluentType != "fluentd" {
		t.Error("template not updated")
	}
}

func TestDeleteTemplate(t *testing.T) {
	db, svc := setupConfigTest(t)

	tpl, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)

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

	tpl, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)

	v1, err := svc.CreateVersion(tpl.ID, 1, "config v1", "initial version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v1.Version != 1 || v1.Hash == "" {
		t.Error("version not created correctly")
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

	tpl, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)
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

	tpl, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)
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

	tpl, _ := svc.CreateTemplate("tpl1", "", "fluentbit", "content", "", 1)

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
