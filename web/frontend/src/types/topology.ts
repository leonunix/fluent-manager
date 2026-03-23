export interface Environment {
  id: number
  name: string
  alias: string
  color: string
  sort_order: number
  description: string
  created_at: string
  updated_at: string
}

export interface DataCenter {
  id: number
  name: string
  alias: string
  provider: string
  location: string
  description: string
  regions?: Region[]
  tags: string
  created_at: string
  updated_at: string
}

export interface Region {
  id: number
  name: string
  alias: string
  datacenter_id: number
  datacenter?: DataCenter
  clusters?: Cluster[]
  description: string
  tags: string
  created_at: string
  updated_at: string
}

export interface Cluster {
  id: number
  name: string
  alias: string
  region_id: number
  region?: Region
  environment_id: number | null
  environment?: Environment
  is_default: boolean
  nodes?: import('./node').Node[]
  match_rules?: ClusterMatchRule[]
  config_id: number | null
  config?: import('./config').ConfigVersion
  description: string
  tags: string
  created_at: string
  updated_at: string
}

export interface ClusterMatchRule {
  id: number
  cluster_id: number
  name: string
  priority: number
  hostname_pattern: string
  ip_pattern: string
  fluent_type: string
  label_selector: string
  os_pattern: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface TopologyTreeNode {
  id: number
  name: string
  alias: string
  provider?: string
  regions?: TopologyTreeRegion[]
}

export interface TopologyTreeRegion {
  id: number
  name: string
  alias: string
  clusters?: TopologyTreeCluster[]
}

export interface TopologyTreeCluster {
  id: number
  name: string
  alias: string
  is_default: boolean
  environment: string
  env_color: string
  node_count: number
  online_count: number
}
