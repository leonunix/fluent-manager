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
	Name        string `json:"name"`
	Description string `json:"description"`
	ModuleType  string `json:"module_type"`
	FluentType  string `json:"fluent_type"`
	Content     string `json:"content"`
	Variables   string `json:"variables"`
	IsBuiltin   bool   `json:"is_builtin"`
	PresetKind  string `json:"preset_kind"`
	PresetKey   string `json:"preset_key"`
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

	var maxVersion int
	s.db.Model(&models.ConfigVersion{}).
		Where("template_id = ?", templateID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	version := models.ConfigVersion{
		TemplateID:    templateID,
		Version:       maxVersion + 1,
		Content:       content,
		Hash:          models.HashConfig(content),
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
			"name":        module.Name,
			"description": module.Description,
			"module_type": module.ModuleType,
			"fluent_type": module.FluentType,
			"content":     module.Content,
			"variables":   module.Variables,
			"is_builtin":  module.IsBuiltin,
			"preset_kind": module.PresetKind,
			"preset_key":  module.PresetKey,
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
				return fmt.Errorf("%w: modules are still referenced by templates or versions: %s", ErrForbidden, strings.Join(usedNames, ", "))
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

	variables, err := parseRenderVariables(input.Variables)
	if err != nil {
		return nil, err
	}

	type renderPart struct {
		module  models.ConfigModule
		version *models.ConfigModuleVersion
		content string
		ref     RenderModuleRef
	}

	parts := make([]renderPart, 0, len(input.Modules))
	sourceRefs := make([]map[string]interface{}, 0, len(input.Modules))

	for _, ref := range input.Modules {
		module, version, content, err := s.resolveRenderModule(ref, fluentType)
		if err != nil {
			return nil, err
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

	var sections []string
	for _, part := range parts {
		moduleVariables := cloneRenderVariables(variables)
		overrideVariables, err := parseRenderVariables(part.ref.Variables)
		if err != nil {
			return nil, err
		}
		for key, value := range overrideVariables {
			moduleVariables[key] = value
		}

		rendered, err := renderModuleTemplate(part.content, moduleVariables)
		if err != nil {
			return nil, fmt.Errorf("%w: render module %q failed: %v", ErrInvalidArgument, part.module.Name, err)
		}
		header := fmt.Sprintf("# module:%s name:%s runtime:%s", part.module.ModuleType, part.module.Name, fluentType)
		sections = append(sections, header+"\n"+strings.TrimSpace(rendered))
	}

	sourcePayload, _ := json.Marshal(sourceRefs)
	content := strings.Join(sections, "\n\n")
	rendered := &models.RenderedConfig{
		Name:           strings.TrimSpace(input.Name),
		FluentType:     fluentType,
		RuntimeVersion: strings.TrimSpace(input.RuntimeVersion),
		SourceModules:  string(sourcePayload),
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
	if _, err := parseRenderVariables(input.Variables); err != nil {
		return nil, err
	}
	presetKind := strings.TrimSpace(input.PresetKind)
	if presetKind != "" && presetKind != "input" && presetKind != "output" {
		return nil, fmt.Errorf("%w: unsupported preset_kind %q", ErrInvalidArgument, presetKind)
	}

	return &models.ConfigModule{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		ModuleType:  moduleType,
		FluentType:  fluentType,
		Content:     content,
		Variables:   normalizeJSONString(input.Variables),
		IsBuiltin:   input.IsBuiltin,
		PresetKind:  presetKind,
		PresetKey:   strings.TrimSpace(input.PresetKey),
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

func isConfigValidationError(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}
