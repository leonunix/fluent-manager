package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

var (
	validAggregationFluentTypes = map[string]bool{
		"fluentbit": true,
		"fluentd":   true,
	}
	validAggregationModes = map[string]bool{
		"forward": true,
		"http":    true,
		"custom":  true,
	}
)

type AggregationGroupInput struct {
	Name         string  `json:"name"`
	Alias        string  `json:"alias"`
	Description  string  `json:"description"`
	FluentType   string  `json:"fluent_type"`
	Mode         string  `json:"mode"`
	EndpointHost string  `json:"endpoint_host"`
	EndpointPort int     `json:"endpoint_port"`
	EnableTLS    bool    `json:"enable_tls"`
	SharedKey    *string `json:"shared_key"`
	ClusterID    *uint   `json:"cluster_id"`
}

type NodeFluentProfileInput struct {
	NodeRole             string `json:"node_role"`
	AggregationGroupID   *uint  `json:"aggregation_group_id"`
	LoadedPlugins        string `json:"loaded_plugins"`
	SupportsHotReload    bool   `json:"supports_hot_reload"`
	SupportsMultiline    bool   `json:"supports_multiline"`
	SupportsStorageLayer bool   `json:"supports_storage_layer"`
	SupportsForwardTLS   bool   `json:"supports_forward_tls"`
	SupportsMetricsAPI   bool   `json:"supports_metrics_api"`
	Metadata             string `json:"metadata"`
}

type FluentService interface {
	ListAggregationGroups(allowedClusters []uint) ([]models.AggregationGroup, error)
	ListDeletedAggregationGroups(allowedClusters []uint) ([]models.AggregationGroup, error)
	GetAggregationGroup(id uint, allowedClusters []uint) (*models.AggregationGroup, error)
	CreateAggregationGroup(input *AggregationGroupInput) (*models.AggregationGroup, error)
	UpdateAggregationGroup(id uint, input *AggregationGroupInput) (*models.AggregationGroup, error)
	DeleteAggregationGroup(id uint) error
	RestoreAggregationGroup(id uint, allowedClusters []uint) (*models.AggregationGroup, error)
	GetNodeProfile(nodeID uint) (*models.NodeFluentProfile, error)
	UpsertNodeProfile(nodeID uint, input *NodeFluentProfileInput) (*models.NodeFluentProfile, error)
}

type fluentService struct {
	db     *gorm.DB
	cipher *sharedKeyCipher
}

func NewFluentService(db *gorm.DB, secret string) FluentService {
	cipher, err := newSharedKeyCipher(secret)
	if err != nil {
		panic(err)
	}

	svc := &fluentService{
		db:     db,
		cipher: cipher,
	}
	svc.migrateLegacySharedKeys()
	return svc
}

func (s *fluentService) ListAggregationGroups(allowedClusters []uint) ([]models.AggregationGroup, error) {
	var groups []models.AggregationGroup
	query := s.db.Preload("Cluster").Order("name")
	query = applyAggregationGroupScope(query, allowedClusters)
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	sanitizeAggregationGroups(groups)
	return groups, nil
}

func (s *fluentService) ListDeletedAggregationGroups(allowedClusters []uint) ([]models.AggregationGroup, error) {
	var groups []models.AggregationGroup
	query := s.db.Unscoped().Preload("Cluster").Where("deleted_at IS NOT NULL").Order("name")
	query = applyAggregationGroupScope(query, allowedClusters)
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	sanitizeAggregationGroups(groups)
	return groups, nil
}

func (s *fluentService) GetAggregationGroup(id uint, allowedClusters []uint) (*models.AggregationGroup, error) {
	var group models.AggregationGroup
	if err := s.db.Preload("Cluster").Preload("Nodes").First(&group, id).Error; err != nil {
		return nil, err
	}
	if !aggregationGroupInScope(&group, allowedClusters) {
		return nil, ErrForbidden
	}
	sanitizeAggregationGroup(&group)
	return &group, nil
}

func (s *fluentService) CreateAggregationGroup(input *AggregationGroupInput) (*models.AggregationGroup, error) {
	group, err := s.buildAggregationGroupModel(input, nil)
	if err != nil {
		return nil, err
	}
	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}
	sanitizeAggregationGroup(group)
	return group, nil
}

func (s *fluentService) UpdateAggregationGroup(id uint, input *AggregationGroupInput) (*models.AggregationGroup, error) {
	var group models.AggregationGroup
	if err := s.db.Preload("Nodes").First(&group, id).Error; err != nil {
		return nil, err
	}

	updated, err := s.buildAggregationGroupModel(input, &group)
	if err != nil {
		return nil, err
	}

	if err := s.validateClusterReassignment(&group, updated.ClusterID); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name":          updated.Name,
		"alias":         updated.Alias,
		"description":   updated.Description,
		"fluent_type":   updated.FluentType,
		"mode":          updated.Mode,
		"endpoint_host": updated.EndpointHost,
		"endpoint_port": updated.EndpointPort,
		"enable_tls":    updated.EnableTLS,
		"shared_key":    updated.SharedKey,
		"cluster_id":    updated.ClusterID,
	}
	if err := s.db.Model(&group).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.GetAggregationGroup(id, nil)
}

func (s *fluentService) DeleteAggregationGroup(id uint) error {
	var nodeCount int64
	s.db.Model(&models.Node{}).Where("aggregation_group_id = ?", id).Count(&nodeCount)
	if nodeCount > 0 {
		return ErrHasChildren
	}
	return s.db.Delete(&models.AggregationGroup{}, id).Error
}

func (s *fluentService) RestoreAggregationGroup(id uint, allowedClusters []uint) (*models.AggregationGroup, error) {
	var group models.AggregationGroup
	if err := s.db.Unscoped().Preload("Cluster").Where("deleted_at IS NOT NULL").First(&group, id).Error; err != nil {
		return nil, err
	}
	if !aggregationGroupInScope(&group, allowedClusters) {
		return nil, ErrForbidden
	}
	if err := s.db.Unscoped().Model(&models.AggregationGroup{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error; err != nil {
		return nil, err
	}
	return s.GetAggregationGroup(id, allowedClusters)
}

func (s *fluentService) GetNodeProfile(nodeID uint) (*models.NodeFluentProfile, error) {
	var profile models.NodeFluentProfile
	if err := s.db.Preload("Node.AggregationGroup").Where("node_id = ?", nodeID).First(&profile).Error; err != nil {
		return nil, err
	}
	sanitizeNodeAggregationGroup(profile.Node)
	return &profile, nil
}

func (s *fluentService) UpsertNodeProfile(nodeID uint, input *NodeFluentProfileInput) (*models.NodeFluentProfile, error) {
	var node models.Node
	if err := s.db.First(&node, nodeID).Error; err != nil {
		return nil, err
	}

	role := input.NodeRole
	if role == "" {
		role = node.NodeRole
	}
	if role == "" {
		role = models.NodeRoleStandalone
	}
	if !models.IsValidNodeRole(role) {
		return nil, fmt.Errorf("%w: invalid node role", ErrInvalidArgument)
	}

	if input.AggregationGroupID != nil {
		var group models.AggregationGroup
		if err := s.db.First(&group, *input.AggregationGroupID).Error; err != nil {
			return nil, err
		}
	}

	if err := s.db.Model(&node).Updates(map[string]interface{}{
		"node_role":            role,
		"aggregation_group_id": input.AggregationGroupID,
	}).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	var profile models.NodeFluentProfile
	err := s.db.Where("node_id = ?", nodeID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = models.NodeFluentProfile{
			NodeID: nodeID,
		}
		if err := s.db.Create(&profile).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"loaded_plugins":         input.LoadedPlugins,
		"supports_hot_reload":    input.SupportsHotReload,
		"supports_multiline":     input.SupportsMultiline,
		"supports_storage_layer": input.SupportsStorageLayer,
		"supports_forward_tls":   input.SupportsForwardTLS,
		"supports_metrics_api":   input.SupportsMetricsAPI,
		"metadata":               input.Metadata,
		"last_reported_at":       &now,
	}
	if err := s.db.Model(&profile).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := s.db.Preload("Node.AggregationGroup").Where("node_id = ?", nodeID).First(&profile).Error; err != nil {
		return nil, err
	}
	sanitizeNodeAggregationGroup(profile.Node)
	return &profile, nil
}

func (s *fluentService) buildAggregationGroupModel(input *AggregationGroupInput, existing *models.AggregationGroup) (*models.AggregationGroup, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: aggregation group payload is required", ErrInvalidArgument)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}

	fluentType := strings.TrimSpace(input.FluentType)
	if fluentType == "" {
		fluentType = "fluentd"
	}
	if !validAggregationFluentTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}

	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = "forward"
	}
	if !validAggregationModes[mode] {
		return nil, fmt.Errorf("%w: unsupported mode %q", ErrInvalidArgument, mode)
	}

	if input.EndpointPort < 0 || input.EndpointPort > 65535 {
		return nil, fmt.Errorf("%w: endpoint_port must be between 0 and 65535", ErrInvalidArgument)
	}

	var dup models.AggregationGroup
	err := s.db.
		Where("name = ?", name).
		First(&dup).Error
	if err == nil && (existing == nil || dup.ID != existing.ID) {
		return nil, fmt.Errorf("%w: aggregation group name already exists", ErrConflict)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if input.ClusterID != nil {
		var cluster models.Cluster
		if err := s.db.First(&cluster, *input.ClusterID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: cluster not found", ErrInvalidArgument)
			}
			return nil, err
		}
	}

	sharedKey := ""
	if existing != nil {
		sharedKey = existing.SharedKey
	}
	if input.SharedKey != nil {
		encrypted, err := s.cipher.Encrypt(*input.SharedKey)
		if err != nil {
			return nil, err
		}
		sharedKey = encrypted
	}

	group := &models.AggregationGroup{
		Name:         name,
		Alias:        strings.TrimSpace(input.Alias),
		Description:  strings.TrimSpace(input.Description),
		FluentType:   fluentType,
		Mode:         mode,
		EndpointHost: strings.TrimSpace(input.EndpointHost),
		EndpointPort: input.EndpointPort,
		EnableTLS:    input.EnableTLS,
		SharedKey:    sharedKey,
		ClusterID:    input.ClusterID,
	}
	return group, nil
}

func (s *fluentService) validateClusterReassignment(group *models.AggregationGroup, targetClusterID *uint) error {
	if sameUintPointer(group.ClusterID, targetClusterID) {
		return nil
	}

	var mismatched int64
	query := s.db.Model(&models.Node{}).Where("aggregation_group_id = ?", group.ID)
	if targetClusterID == nil {
		query = query.Where("cluster_id IS NOT NULL")
	} else {
		query = query.Where("cluster_id IS NULL OR cluster_id <> ?", *targetClusterID)
	}
	if err := query.Count(&mismatched).Error; err != nil {
		return err
	}
	if mismatched > 0 {
		return fmt.Errorf("%w: cannot change cluster while assigned nodes belong to a different cluster", ErrInvalidArgument)
	}
	return nil
}

func (s *fluentService) migrateLegacySharedKeys() {
	var groups []models.AggregationGroup
	if err := s.db.Unscoped().Where("shared_key <> ''").Find(&groups).Error; err != nil {
		log.Printf("WARNING: failed to load aggregation group shared keys for migration: %v", err)
		return
	}

	for _, group := range groups {
		if isEncryptedSharedKey(group.SharedKey) {
			continue
		}
		encrypted, err := s.cipher.Encrypt(group.SharedKey)
		if err != nil {
			log.Printf("WARNING: failed to encrypt aggregation group shared key for group %d: %v", group.ID, err)
			continue
		}
		if err := s.db.Unscoped().
			Model(&models.AggregationGroup{}).
			Where("id = ?", group.ID).
			Update("shared_key", encrypted).Error; err != nil {
			log.Printf("WARNING: failed to persist encrypted shared key for group %d: %v", group.ID, err)
		}
	}
}

func applyAggregationGroupScope(query *gorm.DB, allowedClusters []uint) *gorm.DB {
	if allowedClusters == nil {
		return query
	}
	if len(allowedClusters) == 0 {
		return query.Where("cluster_id IS NULL")
	}
	return query.Where("cluster_id IS NULL OR cluster_id IN ?", allowedClusters)
}

func aggregationGroupInScope(group *models.AggregationGroup, allowedClusters []uint) bool {
	if allowedClusters == nil {
		return true
	}
	if group.ClusterID == nil {
		return true
	}
	for _, clusterID := range allowedClusters {
		if clusterID == *group.ClusterID {
			return true
		}
	}
	return false
}

func sanitizeAggregationGroups(groups []models.AggregationGroup) {
	for i := range groups {
		sanitizeAggregationGroup(&groups[i])
	}
}

func sanitizeAggregationGroup(group *models.AggregationGroup) {
	if group == nil {
		return
	}
	group.HasSharedKey = group.SharedKey != ""
	group.SharedKey = ""
}

func sanitizeNodeAggregationGroup(node *models.Node) {
	if node == nil || node.AggregationGroup == nil {
		return
	}
	sanitizeAggregationGroup(node.AggregationGroup)
}

func sameUintPointer(a, b *uint) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
