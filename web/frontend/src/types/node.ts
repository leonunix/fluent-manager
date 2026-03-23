export interface Node {
  id: number
  node_uid: string
  hostname: string
  ip_address: string
  os: string
  agent_version: string
  fluent_type: string
  fluent_version: string
  status: string
  cluster_id: number | null
  cluster?: import('./topology').Cluster
  environment_id: number | null
  environment?: import('./topology').Environment
  labels: string
  config_id: number | null
  config?: import('./config').ConfigVersion
  last_heartbeat: string | null
  created_at: string
  updated_at: string
}

export interface NodeMetrics {
  id: number
  node_id: number
  cpu_usage_percent: number
  mem_total_mb: number
  mem_used_mb: number
  mem_usage_percent: number
  disk_total_gb: number
  disk_used_gb: number
  disk_usage_percent: number
  load_avg_1: number
  load_avg_5: number
  load_avg_15: number
  fluent_running: boolean
  fluent_pid: number
  fluent_cpu_percent: number
  fluent_mem_mb: number
  fluent_open_fds: number
  updated_at: string
}

export interface RemoteCommand {
  id: number
  node_id: number
  node?: Node
  action: string
  args: string
  status: string
  output: string
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface NodeLog {
  id: number
  node_id: number
  lines: string
  line_count: number
  created_at: string
}

export interface NodeStats {
  total: number
  statuses: { status: string; count: number }[]
}
