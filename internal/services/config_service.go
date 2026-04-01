package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

var (
	validConfigFluentTypes = map[string]bool{
		"fluentbit": true,
		"fluentd":   true,
		"shared":    true,
	}
	validTemplateSourceTypes = map[string]bool{
		"manual":          true,
		"module_assembly": true,
	}
	validRenderedConfigTypes = map[string]bool{
		"fluentbit": true,
		"fluentd":   true,
	}
	validConfigModuleTypes = map[string]bool{
		"service": true,
		"input":   true,
		"parser":  true,
		"filter":  true,
		"route":   true,
		"output":  true,
	}
	moduleTypeOrder = map[string]int{
		"service": 0,
		"input":   1,
		"parser":  2,
		"filter":  3,
		"route":   4,
		"output":  5,
	}
)

type ConfigTemplateInput struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	FluentType    string `json:"fluent_type"`
	Content       string `json:"content"`
	Variables     string `json:"variables"`
	SourceType    string `json:"source_type"`
	SourceModules string `json:"source_modules"`
	FlowLayout    string `json:"flow_layout"`
}

type ConfigModuleInput struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	ModuleType     string `json:"module_type"`
	FluentType     string `json:"fluent_type"`
	Content        string `json:"content"`
	ContentFluentd string `json:"content_fluentd"`
	Variables      string `json:"variables"`
	IsBuiltin      bool   `json:"is_builtin"`
	PresetKind     string `json:"preset_kind"`
	PresetKey      string `json:"preset_key"`
}

type RenderModuleRef struct {
	ModuleID  uint   `json:"module_id"`
	VersionID *uint  `json:"version_id"`
	Variables string `json:"variables"`
}

type RenderedConfigPreviewInput struct {
	Name           string            `json:"name"`
	FluentType     string            `json:"fluent_type"`
	RuntimeVersion string            `json:"runtime_version"`
	Modules        []RenderModuleRef `json:"modules"`
	Variables      string            `json:"variables"`
}

// ConfigPipelineInput is the DTO for creating or updating a ConfigPipeline.
type ConfigPipelineInput struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	FluentType      string `json:"fluent_type"`
	InputModuleID   *uint  `json:"input_module_id"`
	FilterModuleIDs []uint `json:"filter_module_ids"`
	OutputTargetIDs []uint `json:"output_target_ids"`
}

// ConfigPipelineDetail extends ConfigPipeline with resolved related objects.
type ConfigPipelineDetail struct {
	models.ConfigPipeline
	InputModule   *models.ConfigModule  `json:"input_module,omitempty"`
	FilterModules []models.ConfigModule `json:"filter_modules"`
	OutputTargets []models.OutputTarget `json:"output_targets"`
}

type ConfigService interface {
	ListTemplates(fluentType, search string, page, pageSize int) ([]models.ConfigTemplate, int64, error)
	GetTemplate(id uint) (*models.ConfigTemplate, error)
	CreateTemplate(input *ConfigTemplateInput, createdBy uint) (*models.ConfigTemplate, error)
	UpdateTemplate(id uint, input *ConfigTemplateInput) (*models.ConfigTemplate, error)
	DeleteTemplate(id uint) error
	ListVersions(templateID uint) ([]models.ConfigVersion, error)
	CreateVersion(templateID, createdBy uint, content, comment string) (*models.ConfigVersion, error)
	GetVersion(versionID uint) (*models.ConfigVersion, error)

	ListModules(fluentType, moduleType, search string, page, pageSize int) ([]models.ConfigModule, int64, error)
	GetModule(id uint) (*models.ConfigModule, error)
	CreateModule(input *ConfigModuleInput, createdBy uint) (*models.ConfigModule, error)
	UpdateModule(id uint, input *ConfigModuleInput) (*models.ConfigModule, error)
	DeleteModule(id uint) error
	DeleteModules(ids []uint) error
	ListModuleVersions(moduleID uint) ([]models.ConfigModuleVersion, error)
	CreateModuleVersion(moduleID, createdBy uint, content, variables, comment string) (*models.ConfigModuleVersion, error)
	PreviewRenderedConfig(input *RenderedConfigPreviewInput, createdBy uint) (*models.RenderedConfig, error)
	GetRenderedConfig(id uint) (*models.RenderedConfig, error)

	ListPipelines(fluentType, search string) ([]ConfigPipelineDetail, error)
	GetPipeline(id uint) (*ConfigPipelineDetail, error)
	CreatePipeline(input *ConfigPipelineInput, createdBy uint) (*ConfigPipelineDetail, error)
	UpdatePipeline(id uint, input *ConfigPipelineInput) (*ConfigPipelineDetail, error)
	DeletePipeline(id uint) error
}

type configService struct {
	db *gorm.DB
}

type sourceModuleRef struct {
	ModuleID   uint   `json:"module_id"`
	ModuleName string `json:"module_name"`
	ModuleType string `json:"module_type"`
}

type moduleUsageRef struct {
	kind string
	name string
	id   uint
}

func NewConfigService(db *gorm.DB) ConfigService {
	return &configService{db: db}
}

func (s *configService) ListTemplates(fluentType, search string, page, pageSize int) ([]models.ConfigTemplate, int64, error) {
	query := s.db.Preload("Creator")
	if fluentType != "" {
		query = query.Where("fluent_type = ?", fluentType)
	}
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Model(&models.ConfigTemplate{}).Count(&total)

	var templates []models.ConfigTemplate
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error
	return templates, total, err
}

func (s *configService) GetTemplate(id uint) (*models.ConfigTemplate, error) {
	var tpl models.ConfigTemplate
	if err := s.db.Preload("Creator").Preload("Versions").First(&tpl, id).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (s *configService) CreateTemplate(input *ConfigTemplateInput, createdBy uint) (*models.ConfigTemplate, error) {
	tpl, err := validateConfigTemplateInput(input)
	if err != nil {
		return nil, err
	}
	tpl.CreatedBy = createdBy
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := purgeSoftDeletedTemplateName(tx, tpl.Name, 0); err != nil {
			return err
		}
		return tx.Create(&tpl).Error
	}); err != nil {
		return nil, err
	}
	return tpl, nil
}

func (s *configService) UpdateTemplate(id uint, input *ConfigTemplateInput) (*models.ConfigTemplate, error) {
	next, err := validateConfigTemplateInput(input)
	if err != nil {
		return nil, err
	}

	var tpl models.ConfigTemplate
	if err := s.db.First(&tpl, id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := purgeSoftDeletedTemplateName(tx, next.Name, tpl.ID); err != nil {
			return err
		}
		return tx.Model(&tpl).Updates(map[string]interface{}{
			"name":           next.Name,
			"description":    next.Description,
			"fluent_type":    next.FluentType,
			"content":        next.Content,
			"variables":      next.Variables,
			"source_type":    next.SourceType,
			"source_modules": next.SourceModules,
			"flow_layout":    next.FlowLayout,
		}).Error
	}); err != nil {
		return nil, err
	}
	return s.GetTemplate(id)
}

func (s *configService) DeleteTemplate(id uint) error {
	var tpl models.ConfigTemplate
	if err := s.db.First(&tpl, id).Error; err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("template_id = ?", id).Delete(&models.ConfigVersion{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&tpl).Error
	})
}

func (s *configService) ListVersions(templateID uint) ([]models.ConfigVersion, error) {
	var versions []models.ConfigVersion
	err := s.db.Where("template_id = ?", templateID).
		Preload("Creator").
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

func (s *configService) CreateVersion(templateID, createdBy uint, content, comment string) (*models.ConfigVersion, error) {
	var tpl models.ConfigTemplate
	if err := s.db.First(&tpl, templateID).Error; err != nil {
		return nil, err
	}

	renderedContent := content
	if tpl.SourceType == "module_assembly" && strings.TrimSpace(tpl.SourceModules) != "" {
		refs, err := parseStoredRenderModuleRefs(tpl.SourceModules)
		if err != nil {
			return nil, err
		}
		renderedContent, _, err = s.renderModulesForRuntime(tpl.FluentType, refs, tpl.Variables)
		if err != nil {
			return nil, err
		}
	}

	var maxVersion int
	s.db.Model(&models.ConfigVersion{}).
		Where("template_id = ?", templateID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	version := models.ConfigVersion{
		TemplateID:    templateID,
		Version:       maxVersion + 1,
		Content:       renderedContent,
		Hash:          models.HashConfig(renderedContent),
		Comment:       comment,
		SourceType:    tpl.SourceType,
		SourceModules: tpl.SourceModules,
		FlowLayout:    tpl.FlowLayout,
		CreatedBy:     createdBy,
	}
	if err := s.db.Create(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (s *configService) GetVersion(versionID uint) (*models.ConfigVersion, error) {
	var version models.ConfigVersion
	if err := s.db.Preload("Template").Preload("Creator").First(&version, versionID).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (s *configService) ListModules(fluentType, moduleType, search string, page, pageSize int) ([]models.ConfigModule, int64, error) {
	query := s.db.Preload("Creator")
	if fluentType != "" {
		query = query.Where("fluent_type = ?", fluentType)
	}
	if moduleType != "" {
		query = query.Where("module_type = ?", moduleType)
	}
	if search != "" {
		keyword := "%" + strings.TrimSpace(search) + "%"
		query = query.Where(
			`name LIKE ? OR description LIKE ? OR module_type LIKE ? OR fluent_type LIKE ? OR preset_kind LIKE ? OR preset_key LIKE ? OR content LIKE ? OR variables LIKE ?`,
			keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword,
		)
	}

	var total int64
	query.Model(&models.ConfigModule{}).Count(&total)

	var modules []models.ConfigModule
	err := query.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&modules).Error
	return modules, total, err
}

func (s *configService) GetModule(id uint) (*models.ConfigModule, error) {
	var module models.ConfigModule
	if err := s.db.Preload("Creator").Preload("Versions").First(&module, id).Error; err != nil {
		return nil, err
	}
	return &module, nil
}

func (s *configService) CreateModule(input *ConfigModuleInput, createdBy uint) (*models.ConfigModule, error) {
	module, err := validateConfigModuleInput(input)
	if err != nil {
		return nil, err
	}
	module.CreatedBy = createdBy
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := purgeSoftDeletedModuleIdentity(tx, module.Name, module.ModuleType, module.FluentType, 0); err != nil {
			return err
		}
		return tx.Create(module).Error
	}); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *configService) UpdateModule(id uint, input *ConfigModuleInput) (*models.ConfigModule, error) {
	module, err := validateConfigModuleInput(input)
	if err != nil {
		return nil, err
	}

	var current models.ConfigModule
	if err := s.db.First(&current, id).Error; err != nil {
		return nil, err
	}
	module.ID = current.ID
	module.CreatedBy = current.CreatedBy
	module.CreatedAt = current.CreatedAt
	if module.PresetKind == "" && module.PresetKey == "" {
		module.PresetKind = current.PresetKind
		module.PresetKey = current.PresetKey
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := purgeSoftDeletedModuleIdentity(tx, module.Name, module.ModuleType, module.FluentType, current.ID); err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]interface{}{
			"name":            module.Name,
			"description":     module.Description,
			"module_type":     module.ModuleType,
			"fluent_type":     module.FluentType,
			"content":         module.Content,
			"content_fluentd": module.ContentFluentd,
			"variables":       module.Variables,
			"is_builtin":      module.IsBuiltin,
			"preset_kind":     module.PresetKind,
			"preset_key":      module.PresetKey,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetModule(id)
}

func (s *configService) DeleteModule(id uint) error {
	return s.DeleteModules([]uint{id})
}

func (s *configService) DeleteModules(ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("%w: at least one module id is required", ErrInvalidArgument)
	}

	uniqueIDs := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return fmt.Errorf("%w: invalid module id 0", ErrInvalidArgument)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var modules []models.ConfigModule
		if err := tx.Where("id IN ?", uniqueIDs).Find(&modules).Error; err != nil {
			return err
		}
		if len(modules) != len(uniqueIDs) {
			return gorm.ErrRecordNotFound
		}

		var builtinNames []string
		for _, module := range modules {
			if module.IsBuiltin {
				builtinNames = append(builtinNames, module.Name)
			}
		}
		if len(builtinNames) > 0 {
			sort.Strings(builtinNames)
			return fmt.Errorf("%w: builtin modules cannot be deleted: %s", ErrForbidden, strings.Join(builtinNames, ", "))
		}

		moduleUsages, err := findModuleUsageRefs(tx, uniqueIDs)
		if err != nil {
			return err
		}
		if len(moduleUsages) > 0 {
			usedNames := make([]string, 0, len(moduleUsages))
			for _, module := range modules {
				usages := moduleUsages[module.ID]
				if len(usages) == 0 {
					continue
				}
				usedNames = append(usedNames, fmt.Sprintf("%s(%s)", module.Name, summarizeModuleUsageRefs(usages)))
			}
			if len(usedNames) > 0 {
				sort.Strings(usedNames)
				return fmt.Errorf("%w: modules are still referenced by templates, versions or pipelines: %s", ErrForbidden, strings.Join(usedNames, ", "))
			}
		}

		if err := tx.Unscoped().Where("module_id IN ?", uniqueIDs).Delete(&models.ConfigModuleVersion{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&models.ConfigModule{}, uniqueIDs).Error
	})
}

func purgeSoftDeletedModuleIdentity(tx *gorm.DB, name, moduleType, fluentType string, excludeID uint) error {
	query := tx.Unscoped().
		Where("name = ? AND module_type = ? AND fluent_type = ? AND deleted_at IS NOT NULL", name, moduleType, fluentType)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var stale []models.ConfigModule
	if err := query.Find(&stale).Error; err != nil {
		return err
	}
	if len(stale) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(stale))
	for _, item := range stale {
		ids = append(ids, item.ID)
	}

	if err := tx.Unscoped().Where("module_id IN ?", ids).Delete(&models.ConfigModuleVersion{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Delete(&models.ConfigModule{}, ids).Error
}

func purgeSoftDeletedTemplateName(tx *gorm.DB, name string, excludeID uint) error {
	query := tx.Unscoped().
		Where("name = ? AND deleted_at IS NOT NULL", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var stale []models.ConfigTemplate
	if err := query.Find(&stale).Error; err != nil {
		return err
	}
	if len(stale) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(stale))
	for _, item := range stale {
		ids = append(ids, item.ID)
	}

	if err := tx.Unscoped().Where("template_id IN ?", ids).Delete(&models.ConfigVersion{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Delete(&models.ConfigTemplate{}, ids).Error
}

func findModuleUsageRefs(tx *gorm.DB, moduleIDs []uint) (map[uint][]moduleUsageRef, error) {
	usageByModuleID := make(map[uint][]moduleUsageRef, len(moduleIDs))
	targets := make(map[uint]struct{}, len(moduleIDs))
	for _, id := range moduleIDs {
		targets[id] = struct{}{}
	}

	var templates []models.ConfigTemplate
	if err := tx.Where("source_type = ? AND source_modules <> ''", "module_assembly").Find(&templates).Error; err != nil {
		return nil, err
	}
	for _, tpl := range templates {
		attachModuleUsageRefs(usageByModuleID, targets, tpl.SourceModules, moduleUsageRef{
			kind: "template",
			name: tpl.Name,
			id:   tpl.ID,
		})
	}

	var versions []models.ConfigVersion
	if err := tx.Preload("Template").Where("source_type = ? AND source_modules <> ''", "module_assembly").Find(&versions).Error; err != nil {
		return nil, err
	}
	for _, version := range versions {
		label := fmt.Sprintf("template version #%d", version.ID)
		if version.Template != nil && strings.TrimSpace(version.Template.Name) != "" {
			label = fmt.Sprintf("%s v%d", version.Template.Name, version.Version)
		}
		attachModuleUsageRefs(usageByModuleID, targets, version.SourceModules, moduleUsageRef{
			kind: "version",
			name: label,
			id:   version.ID,
		})
	}

	var configPipelines []models.ConfigPipeline
	if err := tx.Find(&configPipelines).Error; err != nil {
		return nil, err
	}
	for _, cp := range configPipelines {
		ref := moduleUsageRef{kind: "pipeline", name: cp.Name, id: cp.ID}
		if cp.InputModuleID != nil {
			if _, ok := targets[*cp.InputModuleID]; ok {
				usageByModuleID[*cp.InputModuleID] = append(usageByModuleID[*cp.InputModuleID], ref)
			}
		}
		var filterIDs []uint
		if cp.FilterModuleIDs != "" {
			_ = json.Unmarshal([]byte(cp.FilterModuleIDs), &filterIDs)
		}
		seen := make(map[uint]struct{})
		for _, mid := range filterIDs {
			if _, ok := targets[mid]; !ok {
				continue
			}
			if _, dup := seen[mid]; dup {
				continue
			}
			seen[mid] = struct{}{}
			usageByModuleID[mid] = append(usageByModuleID[mid], ref)
		}
	}

	return usageByModuleID, nil
}

func attachModuleUsageRefs(dst map[uint][]moduleUsageRef, targets map[uint]struct{}, sourceModules string, usage moduleUsageRef) {
	refs := parseSourceModuleRefs(sourceModules)
	if len(refs) == 0 {
		return
	}

	seen := make(map[uint]struct{}, len(refs))
	for _, ref := range refs {
		if ref.ModuleID == 0 {
			continue
		}
		if _, exists := targets[ref.ModuleID]; !exists {
			continue
		}
		if _, exists := seen[ref.ModuleID]; exists {
			continue
		}
		seen[ref.ModuleID] = struct{}{}
		dst[ref.ModuleID] = append(dst[ref.ModuleID], usage)
	}
}

func parseSourceModuleRefs(raw string) []sourceModuleRef {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var refs []sourceModuleRef
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return nil
	}
	return refs
}

func summarizeModuleUsageRefs(usages []moduleUsageRef) string {
	if len(usages) == 0 {
		return ""
	}

	parts := make([]string, 0, len(usages))
	seen := make(map[string]struct{}, len(usages))
	for _, usage := range usages {
		label := usage.kind
		if strings.TrimSpace(usage.name) != "" {
			label = fmt.Sprintf("%s:%s", usage.kind, usage.name)
		}
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		parts = append(parts, label)
	}
	sort.Strings(parts)
	if len(parts) > 3 {
		return strings.Join(parts[:3], "; ") + fmt.Sprintf(" and %d more", len(parts)-3)
	}
	return strings.Join(parts, "; ")
}

func (s *configService) ListModuleVersions(moduleID uint) ([]models.ConfigModuleVersion, error) {
	var versions []models.ConfigModuleVersion
	err := s.db.Where("module_id = ?", moduleID).
		Preload("Creator").
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

func (s *configService) CreateModuleVersion(moduleID, createdBy uint, content, variables, comment string) (*models.ConfigModuleVersion, error) {
	var module models.ConfigModule
	if err := s.db.First(&module, moduleID).Error; err != nil {
		return nil, err
	}

	var maxVersion int
	s.db.Model(&models.ConfigModuleVersion{}).
		Where("module_id = ?", moduleID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	version := models.ConfigModuleVersion{
		ModuleID:  moduleID,
		Version:   maxVersion + 1,
		Content:   content,
		Variables: variables,
		Hash:      models.HashConfig(content + "\n" + variables),
		Comment:   comment,
		CreatedBy: createdBy,
	}
	if err := s.db.Create(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (s *configService) PreviewRenderedConfig(input *RenderedConfigPreviewInput, createdBy uint) (*models.RenderedConfig, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: render preview payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	if len(input.Modules) == 0 {
		return nil, fmt.Errorf("%w: at least one module is required", ErrInvalidArgument)
	}
	content, sourcePayload, err := s.renderModulesForRuntime(fluentType, input.Modules, input.Variables)
	if err != nil {
		return nil, err
	}
	rendered := &models.RenderedConfig{
		Name:           strings.TrimSpace(input.Name),
		FluentType:     fluentType,
		RuntimeVersion: strings.TrimSpace(input.RuntimeVersion),
		SourceModules:  sourcePayload,
		Variables:      normalizeJSONString(input.Variables),
		Content:        content,
		Hash:           models.HashConfig(content),
		CreatedBy:      createdBy,
	}
	if err := s.db.Create(rendered).Error; err != nil {
		return nil, err
	}
	return rendered, nil
}

func (s *configService) GetRenderedConfig(id uint) (*models.RenderedConfig, error) {
	var rendered models.RenderedConfig
	if err := s.db.Preload("Creator").First(&rendered, id).Error; err != nil {
		return nil, err
	}
	return &rendered, nil
}

func (s *configService) resolveRenderModule(ref RenderModuleRef, fluentType string) (*models.ConfigModule, *models.ConfigModuleVersion, string, error) {
	var module models.ConfigModule
	if err := s.db.First(&module, ref.ModuleID).Error; err != nil {
		return nil, nil, "", err
	}
	if module.FluentType != "shared" && module.FluentType != fluentType {
		return nil, nil, "", fmt.Errorf("%w: module %q is for %s, not %s", ErrInvalidArgument, module.Name, module.FluentType, fluentType)
	}

	content := module.Content
	if module.FluentType == "shared" && fluentType == "fluentd" && strings.TrimSpace(module.ContentFluentd) != "" {
		content = module.ContentFluentd
	}
	var version *models.ConfigModuleVersion
	if ref.VersionID != nil {
		var explicit models.ConfigModuleVersion
		if err := s.db.Where("module_id = ?", module.ID).First(&explicit, *ref.VersionID).Error; err != nil {
			return nil, nil, "", err
		}
		version = &explicit
		content = explicit.Content
	}

	return &module, version, content, nil
}

func parseStoredRenderModuleRefs(raw string) ([]RenderModuleRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}

	var refs []RenderModuleRef
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return nil, fmt.Errorf("%w: source_modules must contain renderable module refs", ErrInvalidArgument)
	}
	return refs, nil
}

func (s *configService) renderModulesForRuntime(fluentType string, refs []RenderModuleRef, variablesRaw string) (string, string, error) {
	variables, err := parseRenderVariables(variablesRaw)
	if err != nil {
		return "", "", err
	}

	type renderPart struct {
		module  models.ConfigModule
		version *models.ConfigModuleVersion
		content string
		ref     RenderModuleRef
	}

	parts := make([]renderPart, 0, len(refs))
	sourceRefs := make([]map[string]interface{}, 0, len(refs))

	for _, ref := range refs {
		module, version, content, err := s.resolveRenderModule(ref, fluentType)
		if err != nil {
			return "", "", err
		}
		parts = append(parts, renderPart{
			module:  *module,
			version: version,
			content: content,
			ref:     ref,
		})

		sourceRef := map[string]interface{}{
			"module_id":   module.ID,
			"module_name": module.Name,
			"module_type": module.ModuleType,
		}
		if strings.TrimSpace(ref.Variables) != "" {
			sourceRef["variables"] = normalizeJSONString(ref.Variables)
		}
		if version != nil {
			sourceRef["version_id"] = version.ID
			sourceRef["version"] = version.Version
		}
		sourceRefs = append(sourceRefs, sourceRef)
	}

	sort.SliceStable(parts, func(i, j int) bool {
		left := moduleTypeOrder[parts[i].module.ModuleType]
		right := moduleTypeOrder[parts[j].module.ModuleType]
		return left < right
	})

	sections := make([]string, 0, len(parts))
	for _, part := range parts {
		moduleVariables := cloneRenderVariables(variables)
		overrideVariables, err := parseRenderVariables(part.ref.Variables)
		if err != nil {
			return "", "", err
		}
		for key, value := range overrideVariables {
			moduleVariables[key] = value
		}

		rendered, err := renderModuleTemplate(part.content, moduleVariables)
		if err != nil {
			return "", "", fmt.Errorf("%w: render module %q failed: %v", ErrInvalidArgument, part.module.Name, err)
		}
		header := fmt.Sprintf("# module:%s name:%s runtime:%s", part.module.ModuleType, part.module.Name, fluentType)
		sections = append(sections, header+"\n"+strings.TrimSpace(rendered))
	}

	sourcePayload, _ := json.Marshal(sourceRefs)
	return strings.Join(sections, "\n\n"), string(sourcePayload), nil
}

func validateConfigModuleInput(input *ConfigModuleInput) (*models.ConfigModule, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: module payload is required", ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	moduleType := strings.TrimSpace(input.ModuleType)
	if !validConfigModuleTypes[moduleType] {
		return nil, fmt.Errorf("%w: unsupported module_type %q", ErrInvalidArgument, moduleType)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validConfigFluentTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}
	contentFluentd := strings.TrimSpace(input.ContentFluentd)
	if contentFluentd != "" && fluentType != "shared" {
		return nil, fmt.Errorf("%w: content_fluentd is only allowed for shared modules", ErrInvalidArgument)
	}
	if _, err := parseRenderVariables(input.Variables); err != nil {
		return nil, err
	}
	presetKind := strings.TrimSpace(input.PresetKind)
	if presetKind != "" && presetKind != "input" && presetKind != "output" {
		return nil, fmt.Errorf("%w: unsupported preset_kind %q", ErrInvalidArgument, presetKind)
	}

	return &models.ConfigModule{
		Name:           name,
		Description:    strings.TrimSpace(input.Description),
		ModuleType:     moduleType,
		FluentType:     fluentType,
		Content:        content,
		ContentFluentd: contentFluentd,
		Variables:      normalizeJSONString(input.Variables),
		IsBuiltin:      input.IsBuiltin,
		PresetKind:     presetKind,
		PresetKey:      strings.TrimSpace(input.PresetKey),
	}, nil
}

func validateConfigTemplateInput(input *ConfigTemplateInput) (*models.ConfigTemplate, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: template payload is required", ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if fluentType != "fluentbit" && fluentType != "fluentd" {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}
	if _, err := parseRenderVariables(input.Variables); err != nil {
		return nil, err
	}

	sourceType := strings.TrimSpace(input.SourceType)
	if sourceType == "" {
		sourceType = "manual"
	}
	if !validTemplateSourceTypes[sourceType] {
		return nil, fmt.Errorf("%w: unsupported source_type %q", ErrInvalidArgument, sourceType)
	}

	sourceModules, err := normalizeJSONArrayString(input.SourceModules, "source_modules")
	if err != nil {
		return nil, err
	}
	flowLayout, err := normalizeJSONObjectString(input.FlowLayout, "flow_layout")
	if err != nil {
		return nil, err
	}

	return &models.ConfigTemplate{
		Name:          name,
		Description:   strings.TrimSpace(input.Description),
		FluentType:    fluentType,
		Content:       content,
		Variables:     normalizeJSONString(input.Variables),
		SourceType:    sourceType,
		SourceModules: sourceModules,
		FlowLayout:    flowLayout,
	}, nil
}

func parseRenderVariables(raw string) (map[string]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}, nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("%w: variables must be a JSON object", ErrInvalidArgument)
	}
	if parsed == nil {
		return map[string]interface{}{}, nil
	}
	return parsed, nil
}

func normalizeJSONString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	return raw
}

func normalizeJSONArrayString(raw, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]", nil
	}

	var parsed []interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", fmt.Errorf("%w: %s must be a JSON array", ErrInvalidArgument, field)
	}
	normalized, _ := json.Marshal(parsed)
	return string(normalized), nil
}

func normalizeJSONObjectString(raw, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}", nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", fmt.Errorf("%w: %s must be a JSON object", ErrInvalidArgument, field)
	}
	normalized, _ := json.Marshal(parsed)
	return string(normalized), nil
}

func cloneRenderVariables(source map[string]interface{}) map[string]interface{} {
	if len(source) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func renderModuleTemplate(content string, variables map[string]interface{}) (string, error) {
	tpl, err := template.New("module").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, variables); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *configService) resolvePipelineDetail(p models.ConfigPipeline) *ConfigPipelineDetail {
	detail := &ConfigPipelineDetail{
		ConfigPipeline: p,
		FilterModules:  []models.ConfigModule{},
		OutputTargets:  []models.OutputTarget{},
	}

	if p.InputModuleID != nil {
		var m models.ConfigModule
		if err := s.db.First(&m, *p.InputModuleID).Error; err == nil {
			detail.InputModule = &m
		}
	}

	var filterIDs []uint
	if p.FilterModuleIDs != "" {
		_ = json.Unmarshal([]byte(p.FilterModuleIDs), &filterIDs)
	}
	if len(filterIDs) > 0 {
		var mods []models.ConfigModule
		s.db.Where("id IN ?", filterIDs).Find(&mods)
		byID := make(map[uint]models.ConfigModule, len(mods))
		for _, m := range mods {
			byID[m.ID] = m
		}
		for _, id := range filterIDs {
			if m, ok := byID[id]; ok {
				detail.FilterModules = append(detail.FilterModules, m)
			}
		}
	}

	var outputIDs []uint
	if p.OutputTargetIDs != "" {
		_ = json.Unmarshal([]byte(p.OutputTargetIDs), &outputIDs)
	}
	if len(outputIDs) > 0 {
		var targets []models.OutputTarget
		s.db.Where("id IN ?", outputIDs).Find(&targets)
		byID := make(map[uint]models.OutputTarget, len(targets))
		for _, t := range targets {
			byID[t.ID] = t
		}
		for _, id := range outputIDs {
			if t, ok := byID[id]; ok {
				detail.OutputTargets = append(detail.OutputTargets, t)
			}
		}
	}

	return detail
}

func (s *configService) ListPipelines(fluentType, search string) ([]ConfigPipelineDetail, error) {
	q := s.db.Order("created_at DESC")
	if fluentType != "" {
		q = q.Where("fluent_type = ?", fluentType)
	}
	if search != "" {
		q = q.Where("name LIKE ?", "%"+search+"%")
	}
	var rows []models.ConfigPipeline
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ConfigPipelineDetail, 0, len(rows))
	for _, p := range rows {
		result = append(result, *s.resolvePipelineDetail(p))
	}
	return result, nil
}

func (s *configService) GetPipeline(id uint) (*ConfigPipelineDetail, error) {
	var p models.ConfigPipeline
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return s.resolvePipelineDetail(p), nil
}

func marshalUintSlice(ids []uint) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(ids)
	return string(b)
}

func validatePipelineInput(input *ConfigPipelineInput) (string, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return "", fmt.Errorf("%w: pipeline name is required", ErrInvalidArgument)
	}
	if input.FluentType != "fluentbit" && input.FluentType != "fluentd" {
		return "", fmt.Errorf("%w: fluent_type must be fluentbit or fluentd", ErrInvalidArgument)
	}
	return name, nil
}

func (s *configService) validatePipelineRefs(input *ConfigPipelineInput) error {
	ft := input.FluentType
	if input.InputModuleID != nil {
		var m models.ConfigModule
		if err := s.db.First(&m, *input.InputModuleID).Error; err != nil {
			return fmt.Errorf("%w: input module %d not found", ErrInvalidArgument, *input.InputModuleID)
		}
		if m.FluentType != "shared" && m.FluentType != ft {
			return fmt.Errorf("%w: input module %q is for %s, not %s", ErrInvalidArgument, m.Name, m.FluentType, ft)
		}
	}
	if len(input.FilterModuleIDs) > 0 {
		var mods []models.ConfigModule
		if err := s.db.Where("id IN ?", input.FilterModuleIDs).Find(&mods).Error; err != nil {
			return err
		}
		if len(mods) != len(input.FilterModuleIDs) {
			return fmt.Errorf("%w: one or more filter module IDs not found", ErrInvalidArgument)
		}
		for _, m := range mods {
			if m.FluentType != "shared" && m.FluentType != ft {
				return fmt.Errorf("%w: filter module %q is for %s, not %s", ErrInvalidArgument, m.Name, m.FluentType, ft)
			}
		}
	}
	if len(input.OutputTargetIDs) > 0 {
		var targets []models.OutputTarget
		if err := s.db.Where("id IN ?", input.OutputTargetIDs).Find(&targets).Error; err != nil {
			return err
		}
		if len(targets) != len(input.OutputTargetIDs) {
			return fmt.Errorf("%w: one or more output target IDs not found", ErrInvalidArgument)
		}
		for _, tgt := range targets {
			if tgt.FluentType != "shared" && tgt.FluentType != ft {
				return fmt.Errorf("%w: output target %q is for %s, not %s", ErrInvalidArgument, tgt.Name, tgt.FluentType, ft)
			}
		}
	}
	return nil
}

func (s *configService) CreatePipeline(input *ConfigPipelineInput, createdBy uint) (*ConfigPipelineDetail, error) {
	name, err := validatePipelineInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.validatePipelineRefs(input); err != nil {
		return nil, err
	}
	p := models.ConfigPipeline{
		Name:            name,
		Description:     strings.TrimSpace(input.Description),
		FluentType:      input.FluentType,
		InputModuleID:   input.InputModuleID,
		FilterModuleIDs: marshalUintSlice(input.FilterModuleIDs),
		OutputTargetIDs: marshalUintSlice(input.OutputTargetIDs),
		CreatedBy:       createdBy,
	}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return s.resolvePipelineDetail(p), nil
}

func (s *configService) UpdatePipeline(id uint, input *ConfigPipelineInput) (*ConfigPipelineDetail, error) {
	var p models.ConfigPipeline
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	name, err := validatePipelineInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.validatePipelineRefs(input); err != nil {
		return nil, err
	}
	p.Name = name
	p.Description = strings.TrimSpace(input.Description)
	p.FluentType = input.FluentType
	p.InputModuleID = input.InputModuleID
	p.FilterModuleIDs = marshalUintSlice(input.FilterModuleIDs)
	p.OutputTargetIDs = marshalUintSlice(input.OutputTargetIDs)
	if err := s.db.Save(&p).Error; err != nil {
		return nil, err
	}
	return s.resolvePipelineDetail(p), nil
}

func (s *configService) DeletePipeline(id uint) error {
	return s.db.Delete(&models.ConfigPipeline{}, id).Error
}

func isConfigValidationError(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}
