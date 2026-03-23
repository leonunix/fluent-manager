package services

import (
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type ConfigService interface {
	ListTemplates(fluentType, search string, page, pageSize int) ([]models.ConfigTemplate, int64, error)
	GetTemplate(id uint) (*models.ConfigTemplate, error)
	CreateTemplate(name, description, fluentType, content, variables string, createdBy uint) (*models.ConfigTemplate, error)
	UpdateTemplate(id uint, name, description, fluentType, content, variables string) (*models.ConfigTemplate, error)
	DeleteTemplate(id uint) error
	ListVersions(templateID uint) ([]models.ConfigVersion, error)
	CreateVersion(templateID, createdBy uint, content, comment string) (*models.ConfigVersion, error)
	GetVersion(versionID uint) (*models.ConfigVersion, error)
}

type configService struct {
	db *gorm.DB
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

func (s *configService) CreateTemplate(name, description, fluentType, content, variables string, createdBy uint) (*models.ConfigTemplate, error) {
	tpl := models.ConfigTemplate{
		Name:        name,
		Description: description,
		FluentType:  fluentType,
		Content:     content,
		Variables:   variables,
		CreatedBy:   createdBy,
	}
	if err := s.db.Create(&tpl).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (s *configService) UpdateTemplate(id uint, name, description, fluentType, content, variables string) (*models.ConfigTemplate, error) {
	var tpl models.ConfigTemplate
	if err := s.db.First(&tpl, id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&tpl).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
		"fluent_type": fluentType,
		"content":     content,
		"variables":   variables,
	})
	return &tpl, nil
}

func (s *configService) DeleteTemplate(id uint) error {
	var tpl models.ConfigTemplate
	if err := s.db.First(&tpl, id).Error; err != nil {
		return err
	}
	s.db.Where("template_id = ?", id).Delete(&models.ConfigVersion{})
	return s.db.Delete(&tpl).Error
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
		TemplateID: templateID,
		Version:    maxVersion + 1,
		Content:    content,
		Hash:       models.HashConfig(content),
		Comment:    comment,
		CreatedBy:  createdBy,
	}
	s.db.Create(&version)
	return &version, nil
}

func (s *configService) GetVersion(versionID uint) (*models.ConfigVersion, error) {
	var version models.ConfigVersion
	if err := s.db.Preload("Template").Preload("Creator").First(&version, versionID).Error; err != nil {
		return nil, err
	}
	return &version, nil
}
