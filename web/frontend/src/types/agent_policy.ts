import type { Cluster, Environment } from './topology'
import type { User } from './auth'

export interface AgentSettings {
  heartbeat_interval: number
  metrics_interval: number
  log_upload_interval: number
  log_buffer_lines: number
  health_port: number
  max_retries: number
  retry_base_delay: number
  fluent_type: string
  fluent_config_path: string
  fluent_config_dir: string
  fluent_binary: string
  fluent_service_unit: string
  fluent_restart_cmd: string
  fluent_reload_cmd: string
  fluent_dry_run_cmd: string
  fluent_log_path: string
  fluent_extra_files: string[]
  fluent_metrics_url: string
  fluent_metrics_format: string
  backup_dir: string
  max_backups: number
}

export interface AgentSettingsPatch {
  heartbeat_interval?: number
  metrics_interval?: number
  log_upload_interval?: number
  log_buffer_lines?: number
  health_port?: number
  max_retries?: number
  retry_base_delay?: number
  fluent_type?: string
  fluent_config_path?: string
  fluent_config_dir?: string
  fluent_binary?: string
  fluent_service_unit?: string
  fluent_restart_cmd?: string
  fluent_reload_cmd?: string
  fluent_dry_run_cmd?: string
  fluent_log_path?: string
  fluent_extra_files?: string[]
  fluent_metrics_url?: string
  fluent_metrics_format?: string
  backup_dir?: string
  max_backups?: number
}

export interface AgentPolicyInput {
  name: string
  description: string
  scope_type: 'global' | 'environment' | 'cluster' | 'label_selector'
  environment_id?: number | null
  cluster_id?: number | null
  label_selector: string
  priority: number
  is_enabled: boolean
  settings: AgentSettingsPatch
}

export interface AgentPolicy {
  id: number
  name: string
  description: string
  scope_type: 'global' | 'environment' | 'cluster' | 'label_selector'
  environment_id: number | null
  environment?: Environment
  cluster_id: number | null
  cluster?: Cluster
  label_selector: string
  priority: number
  is_enabled: boolean
  settings: AgentSettingsPatch
  created_by: number
  creator?: User
  created_at: string
  updated_at: string
}

export interface AgentPolicyListResponse {
  data: AgentPolicy[]
  defaults: AgentSettings
}

export interface ResolvedAgentPolicy {
  settings: AgentSettings
  matched_policies: AgentPolicy[]
}
