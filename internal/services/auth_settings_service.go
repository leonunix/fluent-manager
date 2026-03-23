package services

import (
	"encoding/json"
	"fmt"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

// LDAPSettingsDTO mirrors the LDAP config for API read/write.
type LDAPSettingsDTO struct {
	Enabled      bool   `json:"enabled"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	UseTLS       bool   `json:"use_tls"`
	BindDN       string `json:"bind_dn"`
	BindPassword string `json:"bind_password"`
	BaseDN       string `json:"base_dn"`
	UserFilter   string `json:"user_filter"`
	GroupFilter  string `json:"group_filter"`
	Attributes   struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Name     string `json:"name"`
	} `json:"attributes"`
	GroupSyncStrategy string `json:"group_sync_strategy"`
}

// SAMLSettingsDTO mirrors the SAML config for API read/write.
type SAMLSettingsDTO struct {
	Enabled           bool   `json:"enabled"`
	IDPMetadata       string `json:"idp_metadata"`
	EntityID          string `json:"entity_id"`
	ACSURL            string `json:"acs_url"`
	CertFile          string `json:"cert_file"`
	KeyFile           string `json:"key_file"`
	GroupAttribute    string `json:"group_attribute"`
	GroupSyncStrategy string `json:"group_sync_strategy"`
}

// GroupMappingInput is used for setting external group mappings.
type GroupMappingInput struct {
	ExternalGroupName string `json:"external_group_name"`
	GroupID           uint   `json:"group_id"`
}

// AuthSettingsService manages LDAP/SAML configuration stored in the database.
type AuthSettingsService interface {
	GetLDAPSettings() (*LDAPSettingsDTO, error)
	UpdateLDAPSettings(settings LDAPSettingsDTO) error
	GetSAMLSettings() (*SAMLSettingsDTO, error)
	UpdateSAMLSettings(settings SAMLSettingsDTO) error
	ListGroupMappings(source string) ([]models.ExternalGroupMapping, error)
	SetGroupMappings(source string, mappings []GroupMappingInput) error
	ResolveExternalGroups(source string, externalGroupNames []string) ([]models.Group, error)
	GetGroupSyncStrategy(source string) string
	SeedFromConfig(authCfg config.AuthConfig)
}

type authSettingsService struct {
	db *gorm.DB
}

func NewAuthSettingsService(db *gorm.DB) AuthSettingsService {
	return &authSettingsService{db: db}
}

func (s *authSettingsService) getOrCreate(provider string) (*models.AuthSettings, error) {
	var settings models.AuthSettings
	result := s.db.Where("provider = ?", provider).First(&settings)
	if result.RowsAffected == 0 {
		settings = models.AuthSettings{
			Provider:          provider,
			Config:            "{}",
			GroupSyncStrategy: "always",
		}
		if err := s.db.Create(&settings).Error; err != nil {
			return nil, err
		}
	}
	return &settings, nil
}

func (s *authSettingsService) GetLDAPSettings() (*LDAPSettingsDTO, error) {
	settings, err := s.getOrCreate("ldap")
	if err != nil {
		return nil, err
	}

	var dto LDAPSettingsDTO
	if err := json.Unmarshal([]byte(settings.Config), &dto); err != nil {
		return &LDAPSettingsDTO{}, nil
	}
	dto.GroupSyncStrategy = settings.GroupSyncStrategy
	return &dto, nil
}

func (s *authSettingsService) UpdateLDAPSettings(dto LDAPSettingsDTO) error {
	settings, err := s.getOrCreate("ldap")
	if err != nil {
		return err
	}

	syncStrategy := dto.GroupSyncStrategy
	if syncStrategy == "" {
		syncStrategy = "always"
	}
	dto.GroupSyncStrategy = "" // don't store in JSON blob

	configJSON, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	return s.db.Model(settings).Updates(map[string]interface{}{
		"config":              string(configJSON),
		"group_sync_strategy": syncStrategy,
	}).Error
}

func (s *authSettingsService) GetSAMLSettings() (*SAMLSettingsDTO, error) {
	settings, err := s.getOrCreate("saml")
	if err != nil {
		return nil, err
	}

	var dto SAMLSettingsDTO
	if err := json.Unmarshal([]byte(settings.Config), &dto); err != nil {
		return &SAMLSettingsDTO{}, nil
	}
	dto.GroupSyncStrategy = settings.GroupSyncStrategy
	return &dto, nil
}

func (s *authSettingsService) UpdateSAMLSettings(dto SAMLSettingsDTO) error {
	settings, err := s.getOrCreate("saml")
	if err != nil {
		return err
	}

	syncStrategy := dto.GroupSyncStrategy
	if syncStrategy == "" {
		syncStrategy = "always"
	}
	dto.GroupSyncStrategy = ""

	configJSON, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	return s.db.Model(settings).Updates(map[string]interface{}{
		"config":              string(configJSON),
		"group_sync_strategy": syncStrategy,
	}).Error
}

func (s *authSettingsService) ListGroupMappings(source string) ([]models.ExternalGroupMapping, error) {
	var mappings []models.ExternalGroupMapping
	if err := s.db.Preload("Group").Where("source = ?", source).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (s *authSettingsService) SetGroupMappings(source string, inputs []GroupMappingInput) error {
	// Delete existing mappings for this source
	s.db.Unscoped().Where("source = ?", source).Delete(&models.ExternalGroupMapping{})

	// Create new mappings
	for _, input := range inputs {
		if input.ExternalGroupName == "" || input.GroupID == 0 {
			continue
		}
		mapping := models.ExternalGroupMapping{
			Source:            source,
			ExternalGroupName: input.ExternalGroupName,
			GroupID:           input.GroupID,
		}
		if err := s.db.Create(&mapping).Error; err != nil {
			return fmt.Errorf("failed to create mapping for %q: %w", input.ExternalGroupName, err)
		}
	}
	return nil
}

func (s *authSettingsService) ResolveExternalGroups(source string, externalGroupNames []string) ([]models.Group, error) {
	if len(externalGroupNames) == 0 {
		return nil, nil
	}

	var mappings []models.ExternalGroupMapping
	s.db.Where("source = ? AND external_group_name IN ?", source, externalGroupNames).Find(&mappings)

	if len(mappings) == 0 {
		return nil, nil
	}

	groupIDs := make([]uint, 0, len(mappings))
	seen := map[uint]bool{}
	for _, m := range mappings {
		if !seen[m.GroupID] {
			groupIDs = append(groupIDs, m.GroupID)
			seen[m.GroupID] = true
		}
	}

	var groups []models.Group
	s.db.Where("id IN ?", groupIDs).Find(&groups)
	return groups, nil
}

func (s *authSettingsService) GetGroupSyncStrategy(source string) string {
	settings, err := s.getOrCreate(source)
	if err != nil {
		return "always"
	}
	if settings.GroupSyncStrategy == "" {
		return "always"
	}
	return settings.GroupSyncStrategy
}

// SeedFromConfig seeds DB auth settings from config.yaml values on startup.
// Only populates if no settings exist yet for the provider (does not overwrite user edits).
func (s *authSettingsService) SeedFromConfig(authCfg config.AuthConfig) {
	// Seed LDAP
	if authCfg.LDAP.Host != "" {
		var count int64
		s.db.Model(&models.AuthSettings{}).Where("provider = ?", "ldap").Count(&count)
		if count == 0 {
			dto := LDAPSettingsDTO{
				Enabled:      authCfg.LDAP.Enabled,
				Host:         authCfg.LDAP.Host,
				Port:         authCfg.LDAP.Port,
				UseTLS:       authCfg.LDAP.UseTLS,
				BindDN:       authCfg.LDAP.BindDN,
				BindPassword: authCfg.LDAP.BindPassword,
				BaseDN:       authCfg.LDAP.BaseDN,
				UserFilter:   authCfg.LDAP.UserFilter,
				GroupFilter:  authCfg.LDAP.GroupFilter,
				GroupSyncStrategy: authCfg.LDAP.GroupSyncStrategy,
			}
			dto.Attributes.Username = authCfg.LDAP.Attributes.Username
			dto.Attributes.Email = authCfg.LDAP.Attributes.Email
			dto.Attributes.Name = authCfg.LDAP.Attributes.Name
			s.UpdateLDAPSettings(dto)
		}
	}

	// Seed SAML
	if authCfg.SAML.IDPMetadata != "" {
		var count int64
		s.db.Model(&models.AuthSettings{}).Where("provider = ?", "saml").Count(&count)
		if count == 0 {
			dto := SAMLSettingsDTO{
				Enabled:           authCfg.SAML.Enabled,
				IDPMetadata:       authCfg.SAML.IDPMetadata,
				EntityID:          authCfg.SAML.EntityID,
				ACSURL:            authCfg.SAML.ACSURL,
				CertFile:          authCfg.SAML.CertFile,
				KeyFile:           authCfg.SAML.KeyFile,
				GroupAttribute:    authCfg.SAML.GroupAttribute,
				GroupSyncStrategy: authCfg.SAML.GroupSyncStrategy,
			}
			s.UpdateSAMLSettings(dto)
		}
	}
}
