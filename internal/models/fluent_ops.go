package models

import (
	"time"

	"gorm.io/gorm"
)

// OutputTarget represents a reusable terminal destination such as OpenSearch,
// Loki, Kafka, or a custom HTTP sink.
type OutputTarget struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null;uniqueIndex:idx_output_target_name_active,priority:1" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	FluentType  string         `gorm:"size:32;not null" json:"fluent_type"`
	TargetType  string         `gorm:"size:64;not null" json:"target_type"`
	Endpoint    string         `gorm:"size:255" json:"endpoint"`
	Settings    string         `gorm:"type:text" json:"settings"`
	CreatedBy   uint           `json:"created_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;uniqueIndex:idx_output_target_name_active,priority:2" json:"-"`
}

// LogPipeline represents a logical log flow from a source cluster/group to an
// aggregation group or terminal output.
type LogPipeline struct {
	ID                            uint              `gorm:"primaryKey" json:"id"`
	Name                          string            `gorm:"size:128;not null;uniqueIndex:idx_log_pipeline_name_active,priority:1" json:"name"`
	Description                   string            `gorm:"size:512" json:"description"`
	FluentType                    string            `gorm:"size:32;not null" json:"fluent_type"`
	Protocol                      string            `gorm:"size:32;not null" json:"protocol"`
	SourceClusterID               *uint             `gorm:"index" json:"source_cluster_id"`
	SourceCluster                 *Cluster          `json:"source_cluster,omitempty"`
	SourceAggregationGroupID      *uint             `gorm:"index" json:"source_aggregation_group_id"`
	SourceAggregationGroup        *AggregationGroup `json:"source_aggregation_group,omitempty"`
	SourceLabelSelector           string            `gorm:"type:text" json:"source_label_selector"`
	UpstreamRole                  string            `gorm:"size:32" json:"upstream_role"`
	DestinationAggregationGroupID *uint             `gorm:"index" json:"destination_aggregation_group_id"`
	DestinationAggregationGroup   *AggregationGroup `json:"destination_aggregation_group,omitempty"`
	DestinationOutputTargetID     *uint             `gorm:"index" json:"destination_output_target_id"`
	DestinationOutputTarget       *OutputTarget     `json:"destination_output_target,omitempty"`
	DestinationOutputName         string            `gorm:"size:128" json:"destination_output_name"`
	DestinationOutputType         string            `gorm:"size:64" json:"destination_output_type"`
	TagStrategy                   string            `gorm:"size:128" json:"tag_strategy"`
	Enabled                       bool              `gorm:"default:true" json:"enabled"`
	CreatedBy                     uint              `json:"created_by"`
	Creator                       *User             `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt                     time.Time         `json:"created_at"`
	UpdatedAt                     time.Time         `json:"updated_at"`
	DeletedAt                     gorm.DeletedAt    `gorm:"index;uniqueIndex:idx_log_pipeline_name_active,priority:2" json:"-"`
}

// ConfigAnalysisResult stores a semantic analysis run and its findings.
type ConfigAnalysisResult struct {
	ID             uint                    `gorm:"primaryKey" json:"id"`
	FluentType     string                  `gorm:"size:32;not null" json:"fluent_type"`
	RuntimeVersion string                  `gorm:"size:64" json:"runtime_version"`
	Content        string                  `gorm:"type:text;not null" json:"content"`
	Summary        string                  `gorm:"size:512" json:"summary"`
	Status         string                  `gorm:"size:32;default:completed" json:"status"`
	Findings       []ConfigAnalysisFinding `gorm:"foreignKey:AnalysisResultID" json:"findings,omitempty"`
	CreatedBy      uint                    `json:"created_by"`
	Creator        *User                   `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	DeletedAt      gorm.DeletedAt          `gorm:"index" json:"-"`
}

// ConfigAnalysisFinding is a single lint / compatibility issue.
type ConfigAnalysisFinding struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	AnalysisResultID uint           `gorm:"index;not null" json:"analysis_result_id"`
	Severity         string         `gorm:"size:16;not null" json:"severity"`
	RuleCode         string         `gorm:"size:64;not null" json:"rule_code"`
	Message          string         `gorm:"size:512;not null" json:"message"`
	Suggestion       string         `gorm:"size:512" json:"suggestion"`
	Line             int            `json:"line"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// NodeRuntimeState stores runtime-reported state used for drift and health
// calculations. Heartbeat updates EffectiveConfigHash even if the agent does not
// report advanced queue metrics yet.
type NodeRuntimeState struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	NodeID              uint           `gorm:"uniqueIndex;not null" json:"node_id"`
	Node                *Node          `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	DesiredConfigHash   string         `gorm:"size:64" json:"desired_config_hash"`
	EffectiveConfigHash string         `gorm:"size:64" json:"effective_config_hash"`
	LastSyncAt          *time.Time     `json:"last_sync_at"`
	LastReloadAt        *time.Time     `json:"last_reload_at"`
	LastError           string         `gorm:"size:512" json:"last_error"`
	QueueDepth          int            `json:"queue_depth"`
	RetryCount          int            `json:"retry_count"`
	FlushLatencyMS      int            `json:"flush_latency_ms"`
	InputStatus         string         `gorm:"size:32" json:"input_status"`
	OutputStatus        string         `gorm:"size:32" json:"output_status"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}
