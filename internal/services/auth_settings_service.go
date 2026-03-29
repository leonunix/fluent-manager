package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
// CertData / KeyData accept either a file path (/etc/saml/cert.pem)
// or inline PEM content (-----BEGIN CERTIFICATE-----...).
type SAMLSettingsDTO struct {
	Enabled           bool   `json:"enabled"`
	IDPMetadata       string `json:"idp_metadata"`
	EntityID          string `json:"entity_id"`
	ACSURL            string `json:"acs_url"`
	CertData          string `json:"cert_data"`
	KeyData           string `json:"key_data"`
	GroupAttribute    string `json:"group_attribute"`
	GroupSyncStrategy string `json:"group_sync_strategy"`
}

// AIAccountDTO stores a single AI account under a provider.
type AIAccountDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	APIKey      string `json:"api_key"`
	BaseURL     string `json:"base_url"`
	Model       string `json:"model"`
}

// AISettingsDTO stores multi-account AI assistant settings editable from the GUI.
type AISettingsDTO struct {
	Enabled               bool           `json:"enabled"`
	ActiveProvider        string         `json:"active_provider"`
	ActiveAccountID       string         `json:"active_account_id"`
	RequestTimeoutSeconds int            `json:"request_timeout_seconds"`
	SystemPrompt          string         `json:"system_prompt"`
	Accounts              []AIAccountDTO `json:"accounts"`
}

type legacyAIProviderSettingsDTO struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

type legacyAISettingsDTO struct {
	Enabled               bool                        `json:"enabled"`
	ActiveProvider        string                      `json:"active_provider"`
	RequestTimeoutSeconds int                         `json:"request_timeout_seconds"`
	SystemPrompt          string                      `json:"system_prompt"`
	OpenAI                legacyAIProviderSettingsDTO `json:"openai"`
	Claude                legacyAIProviderSettingsDTO `json:"claude"`
	Gemini                legacyAIProviderSettingsDTO `json:"gemini"`
	DeepSeek              legacyAIProviderSettingsDTO `json:"deepseek"`
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
	GetAISettings() (*AISettingsDTO, error)
	UpdateAISettings(settings AISettingsDTO) error
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

func defaultAISettings() AISettingsDTO {
	return AISettingsDTO{
		RequestTimeoutSeconds: 60,
		SystemPrompt:          "", // empty = use built-in base prompt only; custom text here is appended on top
		Accounts:              []AIAccountDTO{},
	}
}

func normalizeAISettings(dto *AISettingsDTO) {
	defaults := defaultAISettings()
	if dto.RequestTimeoutSeconds <= 0 {
		dto.RequestTimeoutSeconds = defaults.RequestTimeoutSeconds
	}
	if strings.TrimSpace(dto.SystemPrompt) == "" {
		dto.SystemPrompt = defaults.SystemPrompt
	}

	normalized := make([]AIAccountDTO, 0, len(dto.Accounts))
	activeExists := false
	for idx, account := range dto.Accounts {
		switch account.Provider {
		case "openai", "claude", "gemini", "deepseek":
		default:
			continue
		}
		if account.ID == "" {
			account.ID = fmt.Sprintf("%s-%d-%d", account.Provider, time.Now().UnixNano(), idx)
		}
		if account.Name == "" {
			account.Name = fmt.Sprintf("%s Account %d", strings.ToUpper(account.Provider[:1])+account.Provider[1:], idx+1)
		}
		if account.ID == dto.ActiveAccountID {
			activeExists = true
			dto.ActiveProvider = account.Provider
		}
		normalized = append(normalized, account)
	}
	dto.Accounts = normalized

	if !activeExists {
		dto.ActiveAccountID = ""
		dto.ActiveProvider = ""
		for _, account := range dto.Accounts {
			if account.Enabled {
				dto.ActiveAccountID = account.ID
				dto.ActiveProvider = account.Provider
				break
			}
		}
		if dto.ActiveAccountID == "" && len(dto.Accounts) > 0 {
			dto.ActiveAccountID = dto.Accounts[0].ID
			dto.ActiveProvider = dto.Accounts[0].Provider
		}
	}
}

func convertLegacyAISettings(legacy legacyAISettingsDTO) AISettingsDTO {
	settings := AISettingsDTO{
		Enabled:               legacy.Enabled,
		RequestTimeoutSeconds: legacy.RequestTimeoutSeconds,
		SystemPrompt:          legacy.SystemPrompt,
	}
	appendAccount := func(provider string, item legacyAIProviderSettingsDTO) {
		if !item.Enabled && item.APIKey == "" && item.BaseURL == "" && item.Model == "" {
			return
		}
		settings.Accounts = append(settings.Accounts, AIAccountDTO{
			ID:       provider + "-legacy",
			Name:     strings.ToUpper(provider[:1]) + provider[1:] + " Account 1",
			Provider: provider,
			Enabled:  item.Enabled,
			APIKey:   item.APIKey,
			BaseURL:  item.BaseURL,
			Model:    item.Model,
		})
		if legacy.ActiveProvider == provider {
			settings.ActiveAccountID = provider + "-legacy"
		}
	}
	appendAccount("openai", legacy.OpenAI)
	appendAccount("claude", legacy.Claude)
	appendAccount("gemini", legacy.Gemini)
	appendAccount("deepseek", legacy.DeepSeek)
	normalizeAISettings(&settings)
	return settings
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

func (s *authSettingsService) GetAISettings() (*AISettingsDTO, error) {
	settings, err := s.getOrCreate("ai")
	if err != nil {
		return nil, err
	}

	dto := defaultAISettings()
	if err := json.Unmarshal([]byte(settings.Config), &dto); err != nil {
		return &dto, nil
	}
	if len(dto.Accounts) == 0 && (strings.Contains(settings.Config, "\"openai\"") || strings.Contains(settings.Config, "\"claude\"") || strings.Contains(settings.Config, "\"gemini\"") || strings.Contains(settings.Config, "\"deepseek\"")) {
		var legacy legacyAISettingsDTO
		if legacyErr := json.Unmarshal([]byte(settings.Config), &legacy); legacyErr == nil {
			converted := convertLegacyAISettings(legacy)
			return &converted, nil
		}
	}
	normalizeAISettings(&dto)
	return &dto, nil
}

func (s *authSettingsService) UpdateAISettings(dto AISettingsDTO) error {
	settings, err := s.getOrCreate("ai")
	if err != nil {
		return err
	}

	normalizeAISettings(&dto)
	configJSON, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	return s.db.Model(settings).Update("config", string(configJSON)).Error
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
				Enabled:           authCfg.LDAP.Enabled,
				Host:              authCfg.LDAP.Host,
				Port:              authCfg.LDAP.Port,
				UseTLS:            authCfg.LDAP.UseTLS,
				BindDN:            authCfg.LDAP.BindDN,
				BindPassword:      authCfg.LDAP.BindPassword,
				BaseDN:            authCfg.LDAP.BaseDN,
				UserFilter:        authCfg.LDAP.UserFilter,
				GroupFilter:       authCfg.LDAP.GroupFilter,
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
				CertData:          authCfg.SAML.CertFile,
				KeyData:           authCfg.SAML.KeyFile,
				GroupAttribute:    authCfg.SAML.GroupAttribute,
				GroupSyncStrategy: authCfg.SAML.GroupSyncStrategy,
			}
			s.UpdateSAMLSettings(dto)
		}
	}
}
