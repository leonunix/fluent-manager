package services

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

var (
	validPipelineFluentTypes = map[string]bool{
		"fluentbit": true,
		"fluentd":   true,
		"shared":    true,
	}
	validPipelineProtocols = map[string]bool{
		"forward": true,
		"http":    true,
		"kafka":   true,
		"loki":    true,
		"custom":  true,
	}
	validOutputTargetTypes = map[string]bool{
		"opensearch": true,
		"loki":       true,
		"kafka":      true,
		"http":       true,
		"s3":         true,
		"stdout":     true,
		"custom":     true,
	}
)

type LogPipelineInput struct {
	Name                          string `json:"name"`
	Description                   string `json:"description"`
	FluentType                    string `json:"fluent_type"`
	Protocol                      string `json:"protocol"`
	SourceClusterID               *uint  `json:"source_cluster_id"`
	SourceAggregationGroupID      *uint  `json:"source_aggregation_group_id"`
	SourceLabelSelector           string `json:"source_label_selector"`
	UpstreamRole                  string `json:"upstream_role"`
	DestinationAggregationGroupID *uint  `json:"destination_aggregation_group_id"`
	DestinationOutputTargetID     *uint  `json:"destination_output_target_id"`
	DestinationOutputName         string `json:"destination_output_name"`
	DestinationOutputType         string `json:"destination_output_type"`
	TagStrategy                   string `json:"tag_strategy"`
	Enabled                       bool   `json:"enabled"`
}

type OutputTargetInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FluentType  string `json:"fluent_type"`
	TargetType  string `json:"target_type"`
	Endpoint    string `json:"endpoint"`
	Settings    string `json:"settings"`
}

type ConfigLintInput struct {
	FluentType     string `json:"fluent_type"`
	RuntimeVersion string `json:"runtime_version"`
	Content        string `json:"content"`
}

type ConfigImportInput struct {
	FluentType string `json:"fluent_type"`
	NamePrefix string `json:"name_prefix"`
	Content    string `json:"content"`
}

type ImportedConfigModule struct {
	Order                 int      `json:"order"`
	Name                  string   `json:"name"`
	Summary               string   `json:"summary"`
	ModuleType            string   `json:"module_type"`
	FluentType            string   `json:"fluent_type"`
	DetectedPlugin        string   `json:"detected_plugin"`
	Content               string   `json:"content"`
	Variables             string   `json:"variables"`
	VariableKeys          []string `json:"variable_keys"`
	ImportAction          string   `json:"import_action"`
	ExistingModuleID      *uint    `json:"existing_module_id,omitempty"`
	ExistingModuleName    string   `json:"existing_module_name,omitempty"`
	OutputTargetID        *uint    `json:"output_target_id,omitempty"`
	OutputTargetName      string   `json:"output_target_name,omitempty"`
	OutputTargetType      string   `json:"output_target_type,omitempty"`
	OutputTargetEndpoint  string   `json:"output_target_endpoint,omitempty"`
	OutputTargetMatchType string   `json:"output_target_match_type,omitempty"`
}

type ImportedConfigDestination struct {
	OutputModuleName  string `json:"output_module_name"`
	OutputModuleOrder int    `json:"output_module_order"`
	OutputTargetID    uint   `json:"output_target_id"`
	Name              string `json:"name"`
	TargetType        string `json:"target_type"`
	Endpoint          string `json:"endpoint"`
	MatchType         string `json:"match_type"`
}

type ConfigImportValidation struct {
	Equivalent      bool                           `json:"equivalent"`
	Verdict         string                         `json:"verdict"`
	Summary         string                         `json:"summary"`
	SemanticDiff    *ConfigSemanticDiffResult      `json:"semantic_diff"`
	LintSummary     string                         `json:"lint_summary"`
	LintFindings    []models.ConfigAnalysisFinding `json:"lint_findings"`
	RenderedContent string                         `json:"rendered_content"`
}

type ConfigImportResult struct {
	FluentType            string                      `json:"fluent_type"`
	NamePrefix            string                      `json:"name_prefix"`
	Summary               string                      `json:"summary"`
	SuggestedTemplateName string                      `json:"suggested_template_name"`
	Warnings              []string                    `json:"warnings"`
	Modules               []ImportedConfigModule      `json:"modules"`
	Destinations          []ImportedConfigDestination `json:"destinations"`
	FlowPath              []string                    `json:"flow_path"`
	FlowLayout            map[string]interface{}      `json:"flow_layout"`
	TemplateDraftContent  string                      `json:"template_draft_content"`
	Validation            ConfigImportValidation      `json:"validation"`
}

type PipelineGraphNode struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	NodeType    string `json:"node_type"`
	Health      string `json:"health"`
	Description string `json:"description"`
}

type PipelineGraphEdge struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	Label      string `json:"label"`
	Protocol   string `json:"protocol"`
	EdgeType   string `json:"edge_type"`
	Health     string `json:"health"`
	PipelineID uint   `json:"pipeline_id"`
}

type PipelineGraph struct {
	Nodes []PipelineGraphNode `json:"nodes"`
	Edges []PipelineGraphEdge `json:"edges"`
}

type RuntimeDriftItem struct {
	NodeID              uint    `json:"node_id"`
	Hostname            string  `json:"hostname"`
	ClusterName         string  `json:"cluster_name"`
	AggregationGroup    string  `json:"aggregation_group"`
	DesiredConfigHash   string  `json:"desired_config_hash"`
	EffectiveConfigHash string  `json:"effective_config_hash"`
	Status              string  `json:"status"`
	LastSyncAt          *string `json:"last_sync_at,omitempty"`
	LastReloadAt        *string `json:"last_reload_at,omitempty"`
	LastError           string  `json:"last_error"`
}

type AggregationGroupRuntimeMetric struct {
	AggregationGroupID uint    `json:"aggregation_group_id"`
	Name               string  `json:"name"`
	AssignedNodes      int64   `json:"assigned_nodes"`
	OnlineNodes        int64   `json:"online_nodes"`
	DestinationPipes   int64   `json:"destination_pipelines"`
	SourcePipes        int64   `json:"source_pipelines"`
	AvgCPU             float64 `json:"avg_cpu"`
	AvgMem             float64 `json:"avg_mem"`
	TLSCoverageRate    float64 `json:"tls_coverage_rate"`
}

type FluentOpsService interface {
	ListOutputTargets() ([]models.OutputTarget, error)
	GetOutputTarget(id uint) (*models.OutputTarget, error)
	CreateOutputTarget(input *OutputTargetInput, createdBy uint) (*models.OutputTarget, error)
	UpdateOutputTarget(id uint, input *OutputTargetInput) (*models.OutputTarget, error)
	DeleteOutputTarget(id uint) error
	ListPipelines(allowedClusters []uint) ([]models.LogPipeline, error)
	GetPipeline(id uint, allowedClusters []uint) (*models.LogPipeline, error)
	CreatePipeline(input *LogPipelineInput, createdBy uint, allowedClusters []uint) (*models.LogPipeline, error)
	UpdatePipeline(id uint, input *LogPipelineInput, allowedClusters []uint) (*models.LogPipeline, error)
	DeletePipeline(id uint, allowedClusters []uint) error
	PipelineGraph(allowedClusters []uint) (*PipelineGraph, error)
	RuntimeHealthGraph(allowedClusters []uint) (*PipelineGraph, error)
	LintConfig(input *ConfigLintInput, createdBy uint) (*models.ConfigAnalysisResult, error)
	ImportExistingConfig(input *ConfigImportInput) (*ConfigImportResult, error)
	ReplayConfig(input *ConfigReplayInput) (*ConfigReplayResult, error)
	SemanticDiff(input *ConfigSemanticDiffInput) (*ConfigSemanticDiffResult, error)
	CheckCompatibility(input *CompatibilityCheckInput, allowedClusters []uint) (*CompatibilityCheckResult, error)
	GetAnalysisResult(id uint) (*models.ConfigAnalysisResult, error)
	ListRuntimeDrift(allowedClusters []uint) ([]RuntimeDriftItem, error)
	RuntimeRecommendations(allowedClusters []uint) ([]RuntimeRecommendation, error)
	AggregationGroupMetrics(id uint, allowedClusters []uint) (*AggregationGroupRuntimeMetric, error)
}

type fluentOpsService struct {
	db *gorm.DB
}

func NewFluentOpsService(db *gorm.DB) FluentOpsService {
	return &fluentOpsService{db: db}
}

func (s *fluentOpsService) ListOutputTargets() ([]models.OutputTarget, error) {
	var targets []models.OutputTarget
	if err := s.db.Preload("Creator").Order("target_type, name").Find(&targets).Error; err != nil {
		return nil, err
	}
	return targets, nil
}

func (s *fluentOpsService) GetOutputTarget(id uint) (*models.OutputTarget, error) {
	var target models.OutputTarget
	if err := s.db.Preload("Creator").First(&target, id).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *fluentOpsService) CreateOutputTarget(input *OutputTargetInput, createdBy uint) (*models.OutputTarget, error) {
	target, err := s.buildOutputTargetModel(input, nil, createdBy)
	if err != nil {
		return nil, err
	}
	if err := s.db.Create(target).Error; err != nil {
		return nil, err
	}
	return s.GetOutputTarget(target.ID)
}

func (s *fluentOpsService) UpdateOutputTarget(id uint, input *OutputTargetInput) (*models.OutputTarget, error) {
	current, err := s.GetOutputTarget(id)
	if err != nil {
		return nil, err
	}
	target, err := s.buildOutputTargetModel(input, current, current.CreatedBy)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(current).Updates(map[string]interface{}{
		"name":        target.Name,
		"description": target.Description,
		"fluent_type": target.FluentType,
		"target_type": target.TargetType,
		"endpoint":    target.Endpoint,
		"settings":    target.Settings,
	}).Error; err != nil {
		return nil, err
	}
	return s.GetOutputTarget(id)
}

func (s *fluentOpsService) DeleteOutputTarget(id uint) error {
	target, err := s.GetOutputTarget(id)
	if err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&models.LogPipeline{}).
		Where("destination_output_target_id = ?", target.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: output target is referenced by %d pipeline(s)", ErrConflict, count)
	}
	return s.db.Delete(&models.OutputTarget{}, target.ID).Error
}

func (s *fluentOpsService) ListPipelines(allowedClusters []uint) ([]models.LogPipeline, error) {
	var pipelines []models.LogPipeline
	query := s.applyPipelineScope(s.basePipelineQuery().Order("name"), allowedClusters)
	if err := query.Find(&pipelines).Error; err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (s *fluentOpsService) GetPipeline(id uint, allowedClusters []uint) (*models.LogPipeline, error) {
	var pipeline models.LogPipeline
	if err := s.basePipelineQuery().First(&pipeline, id).Error; err != nil {
		return nil, err
	}
	if !pipelineInScope(&pipeline, allowedClusters) {
		return nil, ErrForbidden
	}
	return &pipeline, nil
}

func (s *fluentOpsService) CreatePipeline(input *LogPipelineInput, createdBy uint, allowedClusters []uint) (*models.LogPipeline, error) {
	pipeline, err := s.buildPipelineModel(input, nil, createdBy, allowedClusters)
	if err != nil {
		return nil, err
	}
	if err := s.db.Create(pipeline).Error; err != nil {
		return nil, err
	}
	return s.GetPipeline(pipeline.ID, allowedClusters)
}

func (s *fluentOpsService) UpdatePipeline(id uint, input *LogPipelineInput, allowedClusters []uint) (*models.LogPipeline, error) {
	current, err := s.GetPipeline(id, allowedClusters)
	if err != nil {
		return nil, err
	}
	pipeline, err := s.buildPipelineModel(input, current, current.CreatedBy, allowedClusters)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(current).Updates(map[string]interface{}{
		"name":                             pipeline.Name,
		"description":                      pipeline.Description,
		"fluent_type":                      pipeline.FluentType,
		"protocol":                         pipeline.Protocol,
		"source_cluster_id":                pipeline.SourceClusterID,
		"source_aggregation_group_id":      pipeline.SourceAggregationGroupID,
		"source_label_selector":            pipeline.SourceLabelSelector,
		"upstream_role":                    pipeline.UpstreamRole,
		"destination_aggregation_group_id": pipeline.DestinationAggregationGroupID,
		"destination_output_target_id":     pipeline.DestinationOutputTargetID,
		"destination_output_name":          pipeline.DestinationOutputName,
		"destination_output_type":          pipeline.DestinationOutputType,
		"tag_strategy":                     pipeline.TagStrategy,
		"enabled":                          pipeline.Enabled,
	}).Error; err != nil {
		return nil, err
	}
	return s.GetPipeline(id, allowedClusters)
}

func (s *fluentOpsService) DeletePipeline(id uint, allowedClusters []uint) error {
	pipeline, err := s.GetPipeline(id, allowedClusters)
	if err != nil {
		return err
	}
	return s.db.Delete(&models.LogPipeline{}, pipeline.ID).Error
}

func (s *fluentOpsService) PipelineGraph(allowedClusters []uint) (*PipelineGraph, error) {
	pipelines, err := s.ListPipelines(allowedClusters)
	if err != nil {
		return nil, err
	}
	return buildPipelineGraph(pipelines, false), nil
}

func (s *fluentOpsService) RuntimeHealthGraph(allowedClusters []uint) (*PipelineGraph, error) {
	pipelines, err := s.ListPipelines(allowedClusters)
	if err != nil {
		return nil, err
	}
	return buildPipelineGraph(pipelines, true), nil
}

func (s *fluentOpsService) LintConfig(input *ConfigLintInput, createdBy uint) (*models.ConfigAnalysisResult, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: lint payload is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validRenderedConfigTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: content is required", ErrInvalidArgument)
	}

	findings := lintConfigContent(fluentType, content)
	summary := buildLintSummary(findings)
	result := &models.ConfigAnalysisResult{
		FluentType:     fluentType,
		RuntimeVersion: strings.TrimSpace(input.RuntimeVersion),
		Content:        content,
		Summary:        summary,
		Status:         "completed",
		CreatedBy:      createdBy,
	}
	if err := s.db.Create(result).Error; err != nil {
		return nil, err
	}
	for i := range findings {
		findings[i].AnalysisResultID = result.ID
	}
	if len(findings) > 0 {
		if err := s.db.Create(&findings).Error; err != nil {
			return nil, err
		}
	}
	return s.GetAnalysisResult(result.ID)
}

func (s *fluentOpsService) GetAnalysisResult(id uint) (*models.ConfigAnalysisResult, error) {
	var result models.ConfigAnalysisResult
	if err := s.db.Preload("Creator").Preload("Findings").First(&result, id).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(result.Findings, func(i, j int) bool {
		if result.Findings[i].Severity != result.Findings[j].Severity {
			return severityRank(result.Findings[i].Severity) < severityRank(result.Findings[j].Severity)
		}
		return result.Findings[i].Line < result.Findings[j].Line
	})
	return &result, nil
}

func (s *fluentOpsService) ListRuntimeDrift(allowedClusters []uint) ([]RuntimeDriftItem, error) {
	var nodes []models.Node
	query := s.db.Preload("Config").
		Preload("Cluster.Config").
		Preload("Cluster.Region.DataCenter").
		Preload("AggregationGroup")
	if allowedClusters != nil {
		if len(allowedClusters) == 0 {
			return []RuntimeDriftItem{}, nil
		}
		query = query.Where("cluster_id IN ?", allowedClusters)
	}
	if err := query.Order("hostname").Find(&nodes).Error; err != nil {
		return nil, err
	}

	nodeIDs := make([]uint, 0, len(nodes))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}

	stateMap := map[uint]models.NodeRuntimeState{}
	if len(nodeIDs) > 0 {
		var states []models.NodeRuntimeState
		if err := s.db.Where("node_id IN ?", nodeIDs).Find(&states).Error; err != nil {
			return nil, err
		}
		for _, state := range states {
			stateMap[state.NodeID] = state
		}
	}

	items := make([]RuntimeDriftItem, 0, len(nodes))
	for _, node := range nodes {
		state := stateMap[node.ID]
		desiredHash := ""
		if node.Config != nil {
			desiredHash = node.Config.Hash
		} else if node.Cluster != nil && node.Cluster.Config != nil {
			desiredHash = node.Cluster.Config.Hash
		}

		status := "not_managed"
		if desiredHash != "" {
			status = "unknown"
			if state.EffectiveConfigHash != "" {
				if state.EffectiveConfigHash == desiredHash {
					status = "in_sync"
				} else {
					status = "drifted"
				}
			}
			if state.LastError != "" {
				status = "apply_failed"
			}
		}

		item := RuntimeDriftItem{
			NodeID:              node.ID,
			Hostname:            node.Hostname,
			ClusterName:         resolvedNodeClusterName(&node),
			AggregationGroup:    resolvedNodeGroupName(&node),
			DesiredConfigHash:   desiredHash,
			EffectiveConfigHash: state.EffectiveConfigHash,
			Status:              status,
			LastError:           state.LastError,
		}
		if state.LastSyncAt != nil {
			v := state.LastSyncAt.Format(timeLayoutRFC3339)
			item.LastSyncAt = &v
		}
		if state.LastReloadAt != nil {
			v := state.LastReloadAt.Format(timeLayoutRFC3339)
			item.LastReloadAt = &v
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *fluentOpsService) AggregationGroupMetrics(id uint, allowedClusters []uint) (*AggregationGroupRuntimeMetric, error) {
	var group models.AggregationGroup
	if err := s.db.Preload("Cluster").First(&group, id).Error; err != nil {
		return nil, err
	}
	if !aggregationGroupInScope(&group, allowedClusters) {
		return nil, ErrForbidden
	}

	metrics, err := s.aggregationGroupMetricsForGroups([]models.AggregationGroup{group})
	if err != nil {
		return nil, err
	}
	metric := metrics[id]
	if metric == nil {
		return &AggregationGroupRuntimeMetric{
			AggregationGroupID: group.ID,
			Name:               firstNonEmpty(group.Alias, group.Name),
		}, nil
	}
	return metric, nil
}

func (s *fluentOpsService) basePipelineQuery() *gorm.DB {
	return s.db.Preload("Creator").
		Preload("SourceCluster.Region.DataCenter").
		Preload("SourceAggregationGroup.Cluster.Region.DataCenter").
		Preload("DestinationAggregationGroup.Cluster.Region.DataCenter").
		Preload("DestinationOutputTarget")
}

func (s *fluentOpsService) buildOutputTargetModel(input *OutputTargetInput, existing *models.OutputTarget, createdBy uint) (*models.OutputTarget, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: output target payload is required", ErrInvalidArgument)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validPipelineFluentTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	targetType := strings.TrimSpace(input.TargetType)
	if !validOutputTargetTypes[targetType] {
		return nil, fmt.Errorf("%w: unsupported target_type %q", ErrInvalidArgument, targetType)
	}
	settings := strings.TrimSpace(input.Settings)
	if settings == "" {
		settings = "{}"
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &parsed); err != nil {
		return nil, fmt.Errorf("%w: settings must be a JSON object", ErrInvalidArgument)
	}
	if parsed == nil {
		settings = "{}"
	}

	var duplicate models.OutputTarget
	err := s.db.Where("name = ?", name).First(&duplicate).Error
	if err == nil && (existing == nil || duplicate.ID != existing.ID) {
		return nil, fmt.Errorf("%w: output target name already exists", ErrConflict)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &models.OutputTarget{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		FluentType:  fluentType,
		TargetType:  targetType,
		Endpoint:    strings.TrimSpace(input.Endpoint),
		Settings:    settings,
		CreatedBy:   createdBy,
	}, nil
}

func (s *fluentOpsService) buildPipelineModel(input *LogPipelineInput, existing *models.LogPipeline, createdBy uint, allowedClusters []uint) (*models.LogPipeline, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: pipeline payload is required", ErrInvalidArgument)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	fluentType := strings.TrimSpace(input.FluentType)
	if !validPipelineFluentTypes[fluentType] {
		return nil, fmt.Errorf("%w: unsupported fluent_type %q", ErrInvalidArgument, fluentType)
	}
	protocol := strings.TrimSpace(input.Protocol)
	if !validPipelineProtocols[protocol] {
		return nil, fmt.Errorf("%w: unsupported protocol %q", ErrInvalidArgument, protocol)
	}

	sourceCount := 0
	if input.SourceClusterID != nil {
		sourceCount++
	}
	if input.SourceAggregationGroupID != nil {
		sourceCount++
	}
	if strings.TrimSpace(input.SourceLabelSelector) != "" {
		sourceCount++
	}
	if sourceCount != 1 {
		return nil, fmt.Errorf("%w: exactly one source selector is required", ErrInvalidArgument)
	}

	destinationCount := 0
	if input.DestinationAggregationGroupID != nil {
		destinationCount++
	}
	if input.DestinationOutputTargetID != nil {
		destinationCount++
	}
	if strings.TrimSpace(input.DestinationOutputName) != "" {
		destinationCount++
	}
	if destinationCount != 1 {
		return nil, fmt.Errorf("%w: exactly one destination selector is required", ErrInvalidArgument)
	}

	if input.SourceClusterID != nil {
		var cluster models.Cluster
		if err := s.db.First(&cluster, *input.SourceClusterID).Error; err != nil {
			return nil, fmt.Errorf("%w: source cluster not found", ErrInvalidArgument)
		}
		if !clusterAllowed(*input.SourceClusterID, allowedClusters) {
			return nil, ErrForbidden
		}
	}
	if input.SourceAggregationGroupID != nil {
		var group models.AggregationGroup
		if err := s.db.Preload("Cluster").First(&group, *input.SourceAggregationGroupID).Error; err != nil {
			return nil, fmt.Errorf("%w: source aggregation group not found", ErrInvalidArgument)
		}
		if !aggregationGroupInScope(&group, allowedClusters) {
			return nil, ErrForbidden
		}
	}
	if input.DestinationAggregationGroupID != nil {
		var group models.AggregationGroup
		if err := s.db.Preload("Cluster").First(&group, *input.DestinationAggregationGroupID).Error; err != nil {
			return nil, fmt.Errorf("%w: destination aggregation group not found", ErrInvalidArgument)
		}
		if !aggregationGroupInScope(&group, allowedClusters) {
			return nil, ErrForbidden
		}
	}
	if input.DestinationOutputTargetID != nil {
		if err := s.db.First(&models.OutputTarget{}, *input.DestinationOutputTargetID).Error; err != nil {
			return nil, fmt.Errorf("%w: destination output target not found", ErrInvalidArgument)
		}
	}
	if allowedClusters != nil && strings.TrimSpace(input.SourceLabelSelector) != "" {
		return nil, fmt.Errorf("%w: scoped users cannot create global label-selector pipelines", ErrForbidden)
	}

	var existingDup models.LogPipeline
	err := s.db.Where("name = ?", name).First(&existingDup).Error
	if err == nil && (existing == nil || existingDup.ID != existing.ID) {
		return nil, fmt.Errorf("%w: pipeline name already exists", ErrConflict)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &models.LogPipeline{
		Name:                          name,
		Description:                   strings.TrimSpace(input.Description),
		FluentType:                    fluentType,
		Protocol:                      protocol,
		SourceClusterID:               input.SourceClusterID,
		SourceAggregationGroupID:      input.SourceAggregationGroupID,
		SourceLabelSelector:           strings.TrimSpace(input.SourceLabelSelector),
		UpstreamRole:                  strings.TrimSpace(input.UpstreamRole),
		DestinationAggregationGroupID: input.DestinationAggregationGroupID,
		DestinationOutputTargetID:     input.DestinationOutputTargetID,
		DestinationOutputName:         strings.TrimSpace(input.DestinationOutputName),
		DestinationOutputType:         strings.TrimSpace(input.DestinationOutputType),
		TagStrategy:                   strings.TrimSpace(input.TagStrategy),
		Enabled:                       input.Enabled,
		CreatedBy:                     createdBy,
	}, nil
}

func (s *fluentOpsService) applyPipelineScope(query *gorm.DB, allowedClusters []uint) *gorm.DB {
	if allowedClusters == nil {
		return query
	}

	sourceGroupScope := applyAggregationGroupScope(s.db.Model(&models.AggregationGroup{}).Select("id"), allowedClusters)
	destinationGroupScope := applyAggregationGroupScope(s.db.Model(&models.AggregationGroup{}).Select("id"), allowedClusters)

	conditions := make([]string, 0, 3)
	args := make([]interface{}, 0, 3)
	if len(allowedClusters) > 0 {
		conditions = append(conditions, "source_cluster_id IN ?")
		args = append(args, allowedClusters)
	}
	conditions = append(conditions, "source_aggregation_group_id IN (?)")
	args = append(args, sourceGroupScope)
	conditions = append(conditions, "destination_aggregation_group_id IN (?)")
	args = append(args, destinationGroupScope)

	return query.Where(strings.Join(conditions, " OR "), args...)
}

func pipelineInScope(pipeline *models.LogPipeline, allowedClusters []uint) bool {
	if allowedClusters == nil {
		return true
	}
	if pipeline == nil {
		return false
	}
	if pipeline.SourceClusterID != nil && clusterAllowed(*pipeline.SourceClusterID, allowedClusters) {
		return true
	}
	if pipeline.SourceAggregationGroup != nil && aggregationGroupInScope(pipeline.SourceAggregationGroup, allowedClusters) {
		return true
	}
	if pipeline.DestinationAggregationGroup != nil && aggregationGroupInScope(pipeline.DestinationAggregationGroup, allowedClusters) {
		return true
	}
	return false
}

func clusterAllowed(clusterID uint, allowedClusters []uint) bool {
	if allowedClusters == nil {
		return true
	}
	for _, allowed := range allowedClusters {
		if allowed == clusterID {
			return true
		}
	}
	return false
}

func (s *fluentOpsService) aggregationGroupMetricsForGroups(groups []models.AggregationGroup) (map[uint]*AggregationGroupRuntimeMetric, error) {
	metrics := make(map[uint]*AggregationGroupRuntimeMetric, len(groups))
	if len(groups) == 0 {
		return metrics, nil
	}

	groupIDs := make([]uint, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
		metrics[group.ID] = &AggregationGroupRuntimeMetric{
			AggregationGroupID: group.ID,
			Name:               firstNonEmpty(group.Alias, group.Name),
		}
	}

	type countRow struct {
		GroupID uint
		Count   int64
	}
	type avgRow struct {
		GroupID uint
		AvgCPU  float64
		AvgMem  float64
	}

	var assignedRows []countRow
	if err := s.db.Model(&models.Node{}).
		Select("aggregation_group_id as group_id, COUNT(*) as count").
		Where("aggregation_group_id IN ?", groupIDs).
		Group("aggregation_group_id").
		Scan(&assignedRows).Error; err != nil {
		return nil, err
	}
	for _, row := range assignedRows {
		if metric := metrics[row.GroupID]; metric != nil {
			metric.AssignedNodes = row.Count
		}
	}

	var onlineRows []countRow
	if err := s.db.Model(&models.Node{}).
		Select("aggregation_group_id as group_id, COUNT(*) as count").
		Where("aggregation_group_id IN ? AND status = ?", groupIDs, "online").
		Group("aggregation_group_id").
		Scan(&onlineRows).Error; err != nil {
		return nil, err
	}
	for _, row := range onlineRows {
		if metric := metrics[row.GroupID]; metric != nil {
			metric.OnlineNodes = row.Count
		}
	}

	var destinationRows []countRow
	if err := s.db.Model(&models.LogPipeline{}).
		Select("destination_aggregation_group_id as group_id, COUNT(*) as count").
		Where("destination_aggregation_group_id IN ?", groupIDs).
		Group("destination_aggregation_group_id").
		Scan(&destinationRows).Error; err != nil {
		return nil, err
	}
	for _, row := range destinationRows {
		if metric := metrics[row.GroupID]; metric != nil {
			metric.DestinationPipes = row.Count
		}
	}

	var sourceRows []countRow
	if err := s.db.Model(&models.LogPipeline{}).
		Select("source_aggregation_group_id as group_id, COUNT(*) as count").
		Where("source_aggregation_group_id IN ?", groupIDs).
		Group("source_aggregation_group_id").
		Scan(&sourceRows).Error; err != nil {
		return nil, err
	}
	for _, row := range sourceRows {
		if metric := metrics[row.GroupID]; metric != nil {
			metric.SourcePipes = row.Count
		}
	}

	var avgRows []avgRow
	if err := s.db.Model(&models.NodeMetrics{}).
		Select("nodes.aggregation_group_id as group_id, AVG(node_metrics.cpu_usage_percent) as avg_cpu, AVG(node_metrics.mem_usage_percent) as avg_mem").
		Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.deleted_at IS NULL").
		Where("nodes.aggregation_group_id IN ?", groupIDs).
		Group("nodes.aggregation_group_id").
		Scan(&avgRows).Error; err != nil {
		return nil, err
	}
	for _, row := range avgRows {
		if metric := metrics[row.GroupID]; metric != nil {
			metric.AvgCPU = row.AvgCPU
			metric.AvgMem = row.AvgMem
		}
	}

	var tlsRows []countRow
	if err := s.db.Model(&models.Node{}).
		Select("nodes.aggregation_group_id as group_id, COUNT(*) as count").
		Joins("JOIN node_fluent_profiles ON node_fluent_profiles.node_id = nodes.id").
		Where("nodes.aggregation_group_id IN ? AND node_fluent_profiles.supports_forward_tls = ?", groupIDs, true).
		Group("nodes.aggregation_group_id").
		Scan(&tlsRows).Error; err != nil {
		return nil, err
	}
	for _, row := range tlsRows {
		if metric := metrics[row.GroupID]; metric != nil && metric.AssignedNodes > 0 {
			metric.TLSCoverageRate = float64(row.Count) * 100 / float64(metric.AssignedNodes)
		}
	}

	return metrics, nil
}

func buildPipelineGraph(pipelines []models.LogPipeline, includeHealth bool) *PipelineGraph {
	nodeMap := map[string]PipelineGraphNode{}
	edges := make([]PipelineGraphEdge, 0, len(pipelines))

	for _, pipeline := range pipelines {
		sourceID, sourceNode := pipelineSourceNode(&pipeline, includeHealth)
		targetID, targetNode := pipelineTargetNode(&pipeline, includeHealth)
		nodeMap[sourceID] = sourceNode
		nodeMap[targetID] = targetNode

		edges = append(edges, PipelineGraphEdge{
			ID:         fmt.Sprintf("pipeline:%d", pipeline.ID),
			Source:     sourceID,
			Target:     targetID,
			Label:      pipeline.Name,
			Protocol:   pipeline.Protocol,
			EdgeType:   "pipeline",
			Health:     boolHealth(pipeline.Enabled),
			PipelineID: pipeline.ID,
		})
	}

	nodes := make([]PipelineGraphNode, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Label < nodes[j].Label })
	sort.Slice(edges, func(i, j int) bool { return edges[i].Label < edges[j].Label })

	return &PipelineGraph{Nodes: nodes, Edges: edges}
}

func pipelineSourceNode(pipeline *models.LogPipeline, includeHealth bool) (string, PipelineGraphNode) {
	if pipeline.SourceCluster != nil {
		label := firstNonEmpty(pipeline.SourceCluster.Alias, pipeline.SourceCluster.Name)
		return fmt.Sprintf("cluster:%d", pipeline.SourceCluster.ID), PipelineGraphNode{
			ID:          fmt.Sprintf("cluster:%d", pipeline.SourceCluster.ID),
			Label:       label,
			NodeType:    "cluster",
			Health:      boolHealth(includeHealth),
			Description: "cluster source",
		}
	}
	if pipeline.SourceAggregationGroup != nil {
		label := firstNonEmpty(pipeline.SourceAggregationGroup.Alias, pipeline.SourceAggregationGroup.Name)
		return fmt.Sprintf("group:%d", pipeline.SourceAggregationGroup.ID), PipelineGraphNode{
			ID:          fmt.Sprintf("group:%d", pipeline.SourceAggregationGroup.ID),
			Label:       label,
			NodeType:    "aggregation_group",
			Health:      boolHealth(includeHealth),
			Description: "aggregation source",
		}
	}
	label := firstNonEmpty(pipeline.SourceLabelSelector, "label selector")
	return fmt.Sprintf("selector:%d", pipeline.ID), PipelineGraphNode{
		ID:          fmt.Sprintf("selector:%d", pipeline.ID),
		Label:       label,
		NodeType:    "selector",
		Health:      boolHealth(includeHealth),
		Description: "label selector source",
	}
}

func pipelineTargetNode(pipeline *models.LogPipeline, includeHealth bool) (string, PipelineGraphNode) {
	if pipeline.DestinationAggregationGroup != nil {
		label := firstNonEmpty(pipeline.DestinationAggregationGroup.Alias, pipeline.DestinationAggregationGroup.Name)
		return fmt.Sprintf("group:%d", pipeline.DestinationAggregationGroup.ID), PipelineGraphNode{
			ID:          fmt.Sprintf("group:%d", pipeline.DestinationAggregationGroup.ID),
			Label:       label,
			NodeType:    "aggregation_group",
			Health:      boolHealth(includeHealth),
			Description: "aggregation destination",
		}
	}
	if pipeline.DestinationOutputTarget != nil {
		label := firstNonEmpty(pipeline.DestinationOutputTarget.Name, "output")
		desc := firstNonEmpty(pipeline.DestinationOutputTarget.TargetType, pipeline.DestinationOutputTarget.Endpoint, "terminal output")
		return fmt.Sprintf("output-target:%d", pipeline.DestinationOutputTarget.ID), PipelineGraphNode{
			ID:          fmt.Sprintf("output-target:%d", pipeline.DestinationOutputTarget.ID),
			Label:       label,
			NodeType:    "output",
			Health:      boolHealth(includeHealth),
			Description: desc,
		}
	}
	label := firstNonEmpty(pipeline.DestinationOutputName, "output")
	desc := firstNonEmpty(pipeline.DestinationOutputType, "terminal output")
	return fmt.Sprintf("output:%d", pipeline.ID), PipelineGraphNode{
		ID:          fmt.Sprintf("output:%d", pipeline.ID),
		Label:       label,
		NodeType:    "output",
		Health:      boolHealth(includeHealth),
		Description: desc,
	}
}

func lintConfigContent(fluentType, content string) []models.ConfigAnalysisFinding {
	lines := strings.Split(content, "\n")
	findings := make([]models.ConfigAnalysisFinding, 0, 8)

	if strings.Contains(content, "{{.") {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "error",
			RuleCode:   "UNRESOLVED_TEMPLATE",
			Message:    "config still contains unresolved template variables",
			Suggestion: "render the config with variables before deployment",
			Line:       firstLineContaining(lines, "{{."),
		})
	}

	if fluentType == "fluentbit" && strings.Contains(content, "[INPUT]") && strings.Contains(content, "Name tail") {
		if !strings.Contains(content, "storage.type") && !strings.Contains(content, "storage.path") {
			findings = append(findings, models.ConfigAnalysisFinding{
				Severity:   "warning",
				RuleCode:   "FB_STORAGE_MISSING",
				Message:    "tail input is configured without persistent storage settings",
				Suggestion: "add storage.path or storage.type filesystem for safer buffering",
				Line:       firstLineContaining(lines, "Name tail"),
			})
		}
	}

	if strings.Contains(strings.ToLower(content), "forward") && !strings.Contains(strings.ToLower(content), "tls") && !strings.Contains(strings.ToLower(content), "shared_key") {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "warning",
			RuleCode:   "FORWARD_SECURITY_WEAK",
			Message:    "forward transport appears to be configured without TLS or shared_key",
			Suggestion: "enable TLS and configure shared_key or mutual auth where possible",
			Line:       firstLineContaining(lines, "forward"),
		})
	}

	if strings.Contains(content, "[OUTPUT]") && !strings.Contains(strings.ToLower(content), "retry") {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "info",
			RuleCode:   "OUTPUT_RETRY_UNSET",
			Message:    "output section does not declare retry behavior",
			Suggestion: "set explicit retry/backoff options to avoid runtime ambiguity",
			Line:       firstLineContaining(lines, "[OUTPUT]"),
		})
	}

	if fluentType == "fluentd" && strings.Contains(content, "<source") && !strings.Contains(content, "<match") {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "warning",
			RuleCode:   "FD_ROUTE_MISSING",
			Message:    "config defines source blocks but no match routes",
			Suggestion: "add at least one <match> block to route incoming events",
			Line:       firstLineContaining(lines, "<source"),
		})
	}

	if len(findings) == 0 {
		findings = append(findings, models.ConfigAnalysisFinding{
			Severity:   "info",
			RuleCode:   "NO_FINDINGS",
			Message:    "no obvious lint findings detected in the baseline rule set",
			Suggestion: "continue validating against real plugins and sample traffic",
			Line:       1,
		})
	}

	return findings
}

func buildLintSummary(findings []models.ConfigAnalysisFinding) string {
	counts := map[string]int{}
	for _, finding := range findings {
		counts[finding.Severity]++
	}
	parts := make([]string, 0, len(counts))
	for _, severity := range []string{"error", "warning", "info"} {
		if counts[severity] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", counts[severity], severity))
		}
	}
	if len(parts) == 0 {
		return "no findings"
	}
	return strings.Join(parts, ", ")
}

func firstLineContaining(lines []string, needle string) int {
	for i, line := range lines {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	return 1
}

func severityRank(severity string) int {
	switch severity {
	case "error":
		return 0
	case "warning":
		return 1
	default:
		return 2
	}
}

func resolvedNodeClusterName(node *models.Node) string {
	if node == nil || node.Cluster == nil {
		return "-"
	}
	return firstNonEmpty(node.Cluster.Alias, node.Cluster.Name)
}

func resolvedNodeGroupName(node *models.Node) string {
	if node == nil || node.AggregationGroup == nil {
		return "-"
	}
	return firstNonEmpty(node.AggregationGroup.Alias, node.AggregationGroup.Name)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolHealth(ok bool) string {
	if ok {
		return "healthy"
	}
	return "unknown"
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
