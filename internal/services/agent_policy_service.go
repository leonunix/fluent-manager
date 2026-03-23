package services

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type AgentSettingsPatch struct {
	HeartbeatInterval   *int     `json:"heartbeat_interval,omitempty"`
	MetricsInterval     *int     `json:"metrics_interval,omitempty"`
	LogUploadInterval   *int     `json:"log_upload_interval,omitempty"`
	LogBufferLines      *int     `json:"log_buffer_lines,omitempty"`
	HealthPort          *int     `json:"health_port,omitempty"`
	MaxRetries          *int     `json:"max_retries,omitempty"`
	RetryBaseDelay      *int     `json:"retry_base_delay,omitempty"`
	FluentType          *string  `json:"fluent_type,omitempty"`
	FluentConfigPath    *string  `json:"fluent_config_path,omitempty"`
	FluentConfigDir     *string  `json:"fluent_config_dir,omitempty"`
	FluentBinary        *string  `json:"fluent_binary,omitempty"`
	FluentServiceUnit   *string  `json:"fluent_service_unit,omitempty"`
	FluentRestartCmd    *string  `json:"fluent_restart_cmd,omitempty"`
	FluentReloadCmd     *string  `json:"fluent_reload_cmd,omitempty"`
	FluentDryRunCmd     *string  `json:"fluent_dry_run_cmd,omitempty"`
	FluentLogPath       *string  `json:"fluent_log_path,omitempty"`
	FluentExtraFiles    *[]string `json:"fluent_extra_files,omitempty"`
	FluentMetricsURL    *string  `json:"fluent_metrics_url,omitempty"`
	FluentMetricsFormat *string  `json:"fluent_metrics_format,omitempty"`
	BackupDir           *string  `json:"backup_dir,omitempty"`
	MaxBackups          *int     `json:"max_backups,omitempty"`
}

type AgentPolicyInput struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	ScopeType     string             `json:"scope_type"`
	EnvironmentID *uint              `json:"environment_id"`
	ClusterID     *uint              `json:"cluster_id"`
	LabelSelector string             `json:"label_selector"`
	Priority      int                `json:"priority"`
	IsEnabled     bool               `json:"is_enabled"`
	Settings      AgentSettingsPatch `json:"settings"`
}

type AgentPolicyView struct {
	ID            uint               `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	ScopeType     string             `json:"scope_type"`
	EnvironmentID *uint              `json:"environment_id"`
	Environment   *models.Environment `json:"environment,omitempty"`
	ClusterID     *uint              `json:"cluster_id"`
	Cluster       *models.Cluster    `json:"cluster,omitempty"`
	LabelSelector string             `json:"label_selector"`
	Priority      int                `json:"priority"`
	IsEnabled     bool               `json:"is_enabled"`
	Settings      AgentSettingsPatch `json:"settings"`
	CreatedBy     uint               `json:"created_by"`
	Creator       *models.User       `json:"creator,omitempty"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}

type ResolvedAgentPolicy struct {
	Settings        AgentSettings     `json:"settings"`
	MatchedPolicies []AgentPolicyView `json:"matched_policies"`
}

type AgentPolicyService interface {
	ListPolicies() ([]AgentPolicyView, error)
	GetPolicy(id uint) (*AgentPolicyView, error)
	CreatePolicy(input *AgentPolicyInput, createdBy uint) (*AgentPolicyView, error)
	UpdatePolicy(id uint, input *AgentPolicyInput) (*AgentPolicyView, error)
	DeletePolicy(id uint) error
	GetDefaultSettings() AgentSettings
	ResolveForNode(node *models.Node) (*ResolvedAgentPolicy, error)
	ResolveForNodeID(nodeID uint) (*ResolvedAgentPolicy, error)
}

type agentPolicyService struct {
	db       *gorm.DB
	defaults AgentSettings
}

func NewAgentPolicyService(db *gorm.DB, defaults AgentSettings) AgentPolicyService {
	return &agentPolicyService{db: db, defaults: defaults}
}

func (s *agentPolicyService) ListPolicies() ([]AgentPolicyView, error) {
	var policies []models.AgentPolicy
	if err := s.db.Preload("Environment").Preload("Cluster.Region.DataCenter").Preload("Creator").
		Order("scope_type ASC, priority ASC, id ASC").
		Find(&policies).Error; err != nil {
		return nil, err
	}

	result := make([]AgentPolicyView, 0, len(policies))
	for _, policy := range policies {
		view, err := toAgentPolicyView(&policy)
		if err != nil {
			return nil, err
		}
		result = append(result, *view)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if agentPolicyScopeOrder(result[i].ScopeType) != agentPolicyScopeOrder(result[j].ScopeType) {
			return agentPolicyScopeOrder(result[i].ScopeType) < agentPolicyScopeOrder(result[j].ScopeType)
		}
		if result[i].Priority != result[j].Priority {
			return result[i].Priority < result[j].Priority
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (s *agentPolicyService) GetPolicy(id uint) (*AgentPolicyView, error) {
	var policy models.AgentPolicy
	if err := s.db.Preload("Environment").Preload("Cluster.Region.DataCenter").Preload("Creator").First(&policy, id).Error; err != nil {
		return nil, err
	}
	return toAgentPolicyView(&policy)
}

func (s *agentPolicyService) CreatePolicy(input *AgentPolicyInput, createdBy uint) (*AgentPolicyView, error) {
	policy, err := s.buildPolicyModel(input, nil)
	if err != nil {
		return nil, err
	}
	policy.CreatedBy = createdBy
	if err := s.db.Create(policy).Error; err != nil {
		return nil, err
	}
	return s.GetPolicy(policy.ID)
}

func (s *agentPolicyService) UpdatePolicy(id uint, input *AgentPolicyInput) (*AgentPolicyView, error) {
	var existing models.AgentPolicy
	if err := s.db.First(&existing, id).Error; err != nil {
		return nil, err
	}

	policy, err := s.buildPolicyModel(input, &existing)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"name":           policy.Name,
		"description":    policy.Description,
		"scope_type":     policy.ScopeType,
		"environment_id": policy.EnvironmentID,
		"cluster_id":     policy.ClusterID,
		"label_selector": policy.LabelSelector,
		"priority":       policy.Priority,
		"is_enabled":     policy.IsEnabled,
		"settings":       policy.Settings,
	}
	if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetPolicy(id)
}

func (s *agentPolicyService) DeletePolicy(id uint) error {
	return s.db.Delete(&models.AgentPolicy{}, id).Error
}

func (s *agentPolicyService) GetDefaultSettings() AgentSettings {
	return s.defaults
}

func (s *agentPolicyService) ResolveForNodeID(nodeID uint) (*ResolvedAgentPolicy, error) {
	var node models.Node
	if err := s.db.Preload("Cluster.Environment").Preload("Environment").First(&node, nodeID).Error; err != nil {
		return nil, err
	}
	return s.ResolveForNode(&node)
}

func (s *agentPolicyService) ResolveForNode(node *models.Node) (*ResolvedAgentPolicy, error) {
	settings := s.defaults
	if node == nil {
		return &ResolvedAgentPolicy{Settings: settings}, nil
	}

	var policies []models.AgentPolicy
	if err := s.db.Preload("Environment").Preload("Cluster.Region.DataCenter").Preload("Creator").
		Where("is_enabled = ?", true).
		Find(&policies).Error; err != nil {
		return nil, err
	}

	matches := make([]models.AgentPolicy, 0, len(policies))
	for _, policy := range policies {
		ok, err := matchesAgentPolicy(&policy, node)
		if err != nil {
			return nil, err
		}
		if ok {
			matches = append(matches, policy)
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority < matches[j].Priority
		}
		if agentPolicyScopeOrder(matches[i].ScopeType) != agentPolicyScopeOrder(matches[j].ScopeType) {
			return agentPolicyScopeOrder(matches[i].ScopeType) < agentPolicyScopeOrder(matches[j].ScopeType)
		}
		return matches[i].ID < matches[j].ID
	})

	resolvedMatches := make([]AgentPolicyView, 0, len(matches))
	for _, policy := range matches {
		patch, err := decodeAgentSettingsPatch(policy.Settings)
		if err != nil {
			return nil, err
		}
		applyPatch(&settings, patch)

		view, err := toAgentPolicyView(&policy)
		if err != nil {
			return nil, err
		}
		resolvedMatches = append(resolvedMatches, *view)
	}

	return &ResolvedAgentPolicy{
		Settings:        settings,
		MatchedPolicies: resolvedMatches,
	}, nil
}

func (s *agentPolicyService) buildPolicyModel(input *AgentPolicyInput, existing *models.AgentPolicy) (*models.AgentPolicy, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: policy payload is required", ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}

	scopeType := strings.TrimSpace(input.ScopeType)
	if !isValidAgentPolicyScope(scopeType) {
		return nil, fmt.Errorf("%w: invalid scope_type", ErrInvalidArgument)
	}

	if err := s.validatePolicyScope(scopeType, input.EnvironmentID, input.ClusterID, input.LabelSelector); err != nil {
		return nil, err
	}
	if err := s.ensureScopeTargetsExist(input.EnvironmentID, input.ClusterID); err != nil {
		return nil, err
	}
	if err := validateLabelSelector(input.LabelSelector, scopeType); err != nil {
		return nil, err
	}
	if err := s.ensureSingleGlobal(scopeType, existing); err != nil {
		return nil, err
	}

	settingsJSON, err := json.Marshal(input.Settings)
	if err != nil {
		return nil, fmt.Errorf("marshal policy settings: %w", err)
	}

	policy := &models.AgentPolicy{
		Name:          name,
		Description:   strings.TrimSpace(input.Description),
		ScopeType:     scopeType,
		EnvironmentID: input.EnvironmentID,
		ClusterID:     input.ClusterID,
		LabelSelector: strings.TrimSpace(input.LabelSelector),
		Priority:      input.Priority,
		IsEnabled:     input.IsEnabled,
		Settings:      string(settingsJSON),
	}
	if !input.IsEnabled && existing == nil {
		policy.IsEnabled = false
	} else if existing == nil && input.IsEnabled == false {
		policy.IsEnabled = false
	}
	if existing != nil {
		policy.ID = existing.ID
		policy.CreatedBy = existing.CreatedBy
		policy.CreatedAt = existing.CreatedAt
	}
	return policy, nil
}

func (s *agentPolicyService) validatePolicyScope(scopeType string, environmentID, clusterID *uint, labelSelector string) error {
	switch scopeType {
	case models.AgentPolicyScopeGlobal:
		if environmentID != nil || clusterID != nil || strings.TrimSpace(labelSelector) != "" {
			return fmt.Errorf("%w: global policy cannot target environment, cluster, or labels", ErrInvalidArgument)
		}
	case models.AgentPolicyScopeEnvironment:
		if environmentID == nil || clusterID != nil {
			return fmt.Errorf("%w: environment policy requires environment_id only", ErrInvalidArgument)
		}
	case models.AgentPolicyScopeCluster:
		if clusterID == nil || environmentID != nil {
			return fmt.Errorf("%w: cluster policy requires cluster_id only", ErrInvalidArgument)
		}
	case models.AgentPolicyScopeLabelSelector:
		if strings.TrimSpace(labelSelector) == "" {
			return fmt.Errorf("%w: label selector policy requires label_selector", ErrInvalidArgument)
		}
	}
	return nil
}

func (s *agentPolicyService) ensureScopeTargetsExist(environmentID, clusterID *uint) error {
	if environmentID != nil {
		var env models.Environment
		if err := s.db.First(&env, *environmentID).Error; err != nil {
			return fmt.Errorf("%w: environment not found", ErrInvalidArgument)
		}
	}
	if clusterID != nil {
		var cluster models.Cluster
		if err := s.db.First(&cluster, *clusterID).Error; err != nil {
			return fmt.Errorf("%w: cluster not found", ErrInvalidArgument)
		}
	}
	return nil
}

func (s *agentPolicyService) ensureSingleGlobal(scopeType string, existing *models.AgentPolicy) error {
	if scopeType != models.AgentPolicyScopeGlobal {
		return nil
	}
	query := s.db.Model(&models.AgentPolicy{}).Where("scope_type = ?", models.AgentPolicyScopeGlobal)
	if existing != nil {
		query = query.Where("id <> ?", existing.ID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: only one global agent policy is allowed", ErrConflict)
	}
	return nil
}

func toAgentPolicyView(policy *models.AgentPolicy) (*AgentPolicyView, error) {
	if policy == nil {
		return nil, gorm.ErrRecordNotFound
	}
	patch, err := decodeAgentSettingsPatch(policy.Settings)
	if err != nil {
		return nil, err
	}
	return &AgentPolicyView{
		ID:            policy.ID,
		Name:          policy.Name,
		Description:   policy.Description,
		ScopeType:     policy.ScopeType,
		EnvironmentID: policy.EnvironmentID,
		Environment:   policy.Environment,
		ClusterID:     policy.ClusterID,
		Cluster:       policy.Cluster,
		LabelSelector: policy.LabelSelector,
		Priority:      policy.Priority,
		IsEnabled:     policy.IsEnabled,
		Settings:      patch,
		CreatedBy:     policy.CreatedBy,
		Creator:       policy.Creator,
		CreatedAt:     policy.CreatedAt.Format(timeRFC3339),
		UpdatedAt:     policy.UpdatedAt.Format(timeRFC3339),
	}, nil
}

func decodeAgentSettingsPatch(raw string) (AgentSettingsPatch, error) {
	var patch AgentSettingsPatch
	if strings.TrimSpace(raw) == "" {
		return patch, nil
	}
	if err := json.Unmarshal([]byte(raw), &patch); err != nil {
		return AgentSettingsPatch{}, fmt.Errorf("decode policy settings: %w", err)
	}
	return patch, nil
}

func applyPatch(settings *AgentSettings, patch AgentSettingsPatch) {
	if settings == nil {
		return
	}
	if patch.HeartbeatInterval != nil {
		settings.HeartbeatInterval = *patch.HeartbeatInterval
	}
	if patch.MetricsInterval != nil {
		settings.MetricsInterval = *patch.MetricsInterval
	}
	if patch.LogUploadInterval != nil {
		settings.LogUploadInterval = *patch.LogUploadInterval
	}
	if patch.LogBufferLines != nil {
		settings.LogBufferLines = *patch.LogBufferLines
	}
	if patch.HealthPort != nil {
		settings.HealthPort = *patch.HealthPort
	}
	if patch.MaxRetries != nil {
		settings.MaxRetries = *patch.MaxRetries
	}
	if patch.RetryBaseDelay != nil {
		settings.RetryBaseDelay = *patch.RetryBaseDelay
	}
	if patch.FluentType != nil {
		settings.FluentType = *patch.FluentType
	}
	if patch.FluentConfigPath != nil {
		settings.FluentConfigPath = *patch.FluentConfigPath
	}
	if patch.FluentConfigDir != nil {
		settings.FluentConfigDir = *patch.FluentConfigDir
	}
	if patch.FluentBinary != nil {
		settings.FluentBinary = *patch.FluentBinary
	}
	if patch.FluentServiceUnit != nil {
		settings.FluentServiceUnit = *patch.FluentServiceUnit
	}
	if patch.FluentRestartCmd != nil {
		settings.FluentRestartCmd = *patch.FluentRestartCmd
	}
	if patch.FluentReloadCmd != nil {
		settings.FluentReloadCmd = *patch.FluentReloadCmd
	}
	if patch.FluentDryRunCmd != nil {
		settings.FluentDryRunCmd = *patch.FluentDryRunCmd
	}
	if patch.FluentLogPath != nil {
		settings.FluentLogPath = *patch.FluentLogPath
	}
	if patch.FluentExtraFiles != nil {
		settings.FluentExtraFiles = append([]string(nil), (*patch.FluentExtraFiles)...)
	}
	if patch.FluentMetricsURL != nil {
		settings.FluentMetricsURL = *patch.FluentMetricsURL
	}
	if patch.FluentMetricsFormat != nil {
		settings.FluentMetricsFormat = *patch.FluentMetricsFormat
	}
	if patch.BackupDir != nil {
		settings.BackupDir = *patch.BackupDir
	}
	if patch.MaxBackups != nil {
		settings.MaxBackups = *patch.MaxBackups
	}
}

func matchesAgentPolicy(policy *models.AgentPolicy, node *models.Node) (bool, error) {
	if policy == nil || node == nil || !policy.IsEnabled {
		return false, nil
	}

	switch policy.ScopeType {
	case models.AgentPolicyScopeGlobal:
		return true, nil
	case models.AgentPolicyScopeEnvironment:
		effectiveEnvID := node.EffectiveEnvironmentID()
		return effectiveEnvID != nil && policy.EnvironmentID != nil && *effectiveEnvID == *policy.EnvironmentID, nil
	case models.AgentPolicyScopeCluster:
		return node.ClusterID != nil && policy.ClusterID != nil && *node.ClusterID == *policy.ClusterID, nil
	case models.AgentPolicyScopeLabelSelector:
		return labelSelectorMatches(policy.LabelSelector, node.Labels)
	default:
		return false, nil
	}
}

func labelSelectorMatches(selectorJSON, labelsJSON string) (bool, error) {
	if strings.TrimSpace(selectorJSON) == "" || strings.TrimSpace(labelsJSON) == "" {
		return false, nil
	}
	var selector map[string]string
	if err := json.Unmarshal([]byte(selectorJSON), &selector); err != nil {
		return false, fmt.Errorf("%w: invalid label_selector JSON", ErrInvalidArgument)
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
		return false, nil
	}
	for key, value := range selector {
		if labels[key] != value {
			return false, nil
		}
	}
	return true, nil
}

func validateLabelSelector(labelSelector, scopeType string) error {
	if scopeType != models.AgentPolicyScopeLabelSelector {
		return nil
	}
	if strings.TrimSpace(labelSelector) == "" {
		return fmt.Errorf("%w: label_selector is required", ErrInvalidArgument)
	}
	var selector map[string]string
	if err := json.Unmarshal([]byte(labelSelector), &selector); err != nil {
		return fmt.Errorf("%w: label_selector must be a JSON object of string pairs", ErrInvalidArgument)
	}
	if len(selector) == 0 {
		return fmt.Errorf("%w: label_selector cannot be empty", ErrInvalidArgument)
	}
	return nil
}

func isValidAgentPolicyScope(scopeType string) bool {
	switch scopeType {
	case models.AgentPolicyScopeGlobal, models.AgentPolicyScopeEnvironment, models.AgentPolicyScopeCluster, models.AgentPolicyScopeLabelSelector:
		return true
	default:
		return false
	}
}

func agentPolicyScopeOrder(scopeType string) int {
	switch scopeType {
	case models.AgentPolicyScopeGlobal:
		return 0
	case models.AgentPolicyScopeEnvironment:
		return 1
	case models.AgentPolicyScopeCluster:
		return 2
	case models.AgentPolicyScopeLabelSelector:
		return 3
	default:
		return 99
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
