package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type AuthenticatedAgentKey struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	ClusterID  *uint  `json:"cluster_id,omitempty"`
	KeyPreview string `json:"key_preview,omitempty"`
	Legacy     bool   `json:"legacy"`
}

type AgentAccessKeyInput struct {
	Name        string `json:"name"`
	ClusterID   *uint  `json:"cluster_id"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

type AgentAccessKeyCreateResult struct {
	Key          *models.AgentAccessKey `json:"key"`
	PlaintextKey string                 `json:"plaintext_key"`
}

type AgentAccessKeyService interface {
	List(allowedClusters []uint) ([]models.AgentAccessKey, error)
	Create(input AgentAccessKeyInput, createdBy uint, allowedClusters []uint) (*AgentAccessKeyCreateResult, error)
	Update(id uint, input AgentAccessKeyInput, allowedClusters []uint) (*models.AgentAccessKey, error)
	Delete(id uint, allowedClusters []uint) error
	Authenticate(rawKey, legacyKey string) (*AuthenticatedAgentKey, error)
}

type agentAccessKeyService struct {
	db *gorm.DB
}

func NewAgentAccessKeyService(db *gorm.DB) AgentAccessKeyService {
	return &agentAccessKeyService{db: db}
}

func (s *agentAccessKeyService) List(allowedClusters []uint) ([]models.AgentAccessKey, error) {
	var keys []models.AgentAccessKey
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	if allowedClusters != nil {
		query = query.Where("cluster_id IN ?", allowedClusters)
	}
	if err := query.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *agentAccessKeyService) Create(input AgentAccessKeyInput, createdBy uint, allowedClusters []uint) (*AgentAccessKeyCreateResult, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}

	clusterID, err := s.resolveScopedClusterID(input.ClusterID, allowedClusters)
	if err != nil {
		return nil, err
	}

	rawKey, err := generateAgentAccessKey()
	if err != nil {
		return nil, err
	}

	key := &models.AgentAccessKey{
		Name:        name,
		KeyHash:     hashAgentAccessKey(rawKey),
		KeyPreview:  previewAgentAccessKey(rawKey),
		ClusterID:   clusterID,
		Description: strings.TrimSpace(input.Description),
		IsActive:    true,
		CreatedBy:   createdBy,
	}
	if err := s.db.Create(key).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator").First(key, key.ID).Error; err != nil {
		return nil, err
	}
	return &AgentAccessKeyCreateResult{
		Key:          key,
		PlaintextKey: rawKey,
	}, nil
}

func (s *agentAccessKeyService) Update(id uint, input AgentAccessKeyInput, allowedClusters []uint) (*models.AgentAccessKey, error) {
	key, err := s.getScopedKey(id, allowedClusters)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	clusterID, err := s.resolveScopedClusterID(input.ClusterID, allowedClusters)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name":        name,
		"cluster_id":  clusterID,
		"description": strings.TrimSpace(input.Description),
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if err := s.db.Model(key).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator").First(key, key.ID).Error; err != nil {
		return nil, err
	}
	return key, nil
}

func (s *agentAccessKeyService) Delete(id uint, allowedClusters []uint) error {
	key, err := s.getScopedKey(id, allowedClusters)
	if err != nil {
		return err
	}
	return s.db.Delete(&models.AgentAccessKey{}, key.ID).Error
}

func (s *agentAccessKeyService) Authenticate(rawKey, legacyKey string) (*AuthenticatedAgentKey, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var key models.AgentAccessKey
	if err := s.db.Where("key_hash = ? AND is_active = ?", hashAgentAccessKey(rawKey), true).First(&key).Error; err == nil {
		now := time.Now()
		_ = s.db.Model(&key).Update("last_used_at", &now).Error
		return &AuthenticatedAgentKey{
			ID:         key.ID,
			Name:       key.Name,
			ClusterID:  key.ClusterID,
			KeyPreview: key.KeyPreview,
		}, nil
	}

	if legacyKey != "" && rawKey == legacyKey {
		return &AuthenticatedAgentKey{
			Name:   "legacy-config-key",
			Legacy: true,
		}, nil
	}

	return nil, gorm.ErrRecordNotFound
}

func (s *agentAccessKeyService) getScopedKey(id uint, allowedClusters []uint) (*models.AgentAccessKey, error) {
	var key models.AgentAccessKey
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	if allowedClusters != nil {
		query = query.Where("cluster_id IN ?", allowedClusters)
	}
	if err := query.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *agentAccessKeyService) resolveScopedClusterID(clusterID *uint, allowedClusters []uint) (*uint, error) {
	if clusterID == nil {
		if allowedClusters != nil {
			return nil, fmt.Errorf("%w: scoped users must bind agent keys to an allowed cluster", ErrForbidden)
		}
		return nil, nil
	}

	if allowedClusters != nil && !containsUint(allowedClusters, *clusterID) {
		return nil, fmt.Errorf("%w: cluster is outside current scope", ErrForbidden)
	}

	var cluster models.Cluster
	if err := s.db.First(&cluster, *clusterID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: cluster not found", ErrInvalidArgument)
		}
		return nil, err
	}
	return clusterID, nil
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func generateAgentAccessKey() (string, error) {
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate agent access key: %w", err)
	}
	return "fmak_" + base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func hashAgentAccessKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func previewAgentAccessKey(raw string) string {
	if len(raw) <= 14 {
		return raw
	}
	return raw[:10] + "..." + raw[len(raw)-4:]
}
