export interface AggregationGroupInput {
  name: string
  alias?: string
  description?: string
  fluent_type?: string
  mode?: string
  endpoint_host?: string
  endpoint_port?: number
  enable_tls?: boolean
  shared_key?: string
  cluster_id?: number | null
}

export interface NodeFluentProfileInput {
  node_role?: string
  aggregation_group_id?: number | null
  loaded_plugins?: string
  supports_hot_reload?: boolean
  supports_multiline?: boolean
  supports_storage_layer?: boolean
  supports_forward_tls?: boolean
  supports_metrics_api?: boolean
  metadata?: string
}

export interface LogPipeline {
  id: number
  name: string
  description: string
  fluent_type: string
  protocol: string
  source_cluster_id: number | null
  source_cluster?: import('./topology').Cluster
  source_aggregation_group_id: number | null
  source_aggregation_group?: import('./node').AggregationGroup
  source_label_selector: string
  upstream_role: string
  destination_aggregation_group_id: number | null
  destination_aggregation_group?: import('./node').AggregationGroup
  destination_output_target_id: number | null
  destination_output_target?: OutputTarget
  destination_output_name: string
  destination_output_type: string
  tag_strategy: string
  enabled: boolean
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface LogPipelineInput {
  name: string
  description?: string
  fluent_type: string
  protocol: string
  source_cluster_id?: number | null
  source_aggregation_group_id?: number | null
  source_label_selector?: string
  upstream_role?: string
  destination_aggregation_group_id?: number | null
  destination_output_target_id?: number | null
  destination_output_name?: string
  destination_output_type?: string
  tag_strategy?: string
  enabled?: boolean
}

export interface OutputTarget {
  id: number
  name: string
  description: string
  fluent_type: string
  target_type: string
  endpoint: string
  settings: string
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface OutputTargetInput {
  name: string
  description?: string
  fluent_type: string
  target_type: string
  endpoint?: string
  settings?: string
}

export interface PipelineGraphNode {
  id: string
  label: string
  node_type: string
  health: string
  description: string
}

export interface PipelineGraphEdge {
  id: string
  source: string
  target: string
  label: string
  protocol: string
  edge_type: string
  health: string
  pipeline_id: number
}

export interface PipelineGraph {
  nodes: PipelineGraphNode[]
  edges: PipelineGraphEdge[]
}

export interface ConfigAnalysisFinding {
  id: number
  analysis_result_id: number
  severity: string
  rule_code: string
  message: string
  suggestion: string
  line: number
}

export interface ConfigAnalysisResult {
  id: number
  fluent_type: string
  runtime_version: string
  content: string
  summary: string
  status: string
  findings?: ConfigAnalysisFinding[]
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface ConfigLintInput {
  fluent_type: string
  runtime_version?: string
  content: string
}

export interface ConfigReplayStep {
  stage: string
  name: string
  status: string
  detail: string
}

export interface ConfigReplayResult {
  fluent_type: string
  runtime_version: string
  sample_tag: string
  detected_parser: string
  parsed_record: Record<string, any>
  matched_filters: string[]
  route_matched: boolean
  final_output: string
  final_output_type: string
  warnings: string[]
  steps: ConfigReplayStep[]
}

export interface ConfigReplayInput {
  fluent_type: string
  runtime_version?: string
  content: string
  sample_log: string
  sample_tag?: string
}

export interface SemanticChange {
  category: string
  change_type: string
  item: string
  detail: string
}

export interface ConfigSemanticDiffResult {
  fluent_type: string
  summary: string
  changes: SemanticChange[]
}

export interface ConfigSemanticDiffInput {
  fluent_type: string
  before_content: string
  after_content: string
}

export interface CompatibilityCheckResult {
  fluent_type: string
  runtime_version: string
  compatible: boolean
  hot_reload_supported: boolean
  checked_node_id?: number
  missing_plugins: string[]
  findings: ConfigAnalysisFinding[]
}

export interface CompatibilityCheckInput {
  fluent_type: string
  runtime_version?: string
  content: string
  node_id?: number
}

export interface RuntimeDriftItem {
  node_id: number
  hostname: string
  cluster_name: string
  aggregation_group: string
  desired_config_hash: string
  effective_config_hash: string
  status: string
  last_sync_at?: string
  last_reload_at?: string
  last_error: string
}

export interface AggregationGroupRuntimeMetric {
  aggregation_group_id: number
  name: string
  assigned_nodes: number
  online_nodes: number
  destination_pipelines: number
  source_pipelines: number
  avg_cpu: number
  avg_mem: number
  tls_coverage_rate: number
}

export interface RuntimeRecommendation {
  severity: string
  scope_type: string
  scope_id: number
  title: string
  detail: string
  suggestion: string
}
