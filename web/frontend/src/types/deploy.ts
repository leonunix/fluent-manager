export interface DeployTask {
  id: number
  config_id: number
  config?: import('./config').ConfigVersion
  scope: string
  scope_id: number
  status: string
  total_nodes: number
  success_count: number
  fail_count: number
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface DeployRecord {
  id: number
  deploy_task_id: number
  node_id: number
  node?: import('./node').Node
  status: string
  message: string
  created_at: string
  updated_at: string
}
