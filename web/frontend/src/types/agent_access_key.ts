import type { Cluster } from './topology'
import type { User } from './auth'

export interface AgentAccessKey {
  id: number
  name: string
  key_preview: string
  cluster_id: number | null
  cluster?: Cluster | null
  description: string
  is_active: boolean
  last_used_at: string | null
  created_by: number
  creator?: User | null
  created_at: string
  updated_at: string
}

export interface AgentAccessKeyInput {
  name: string
  cluster_id: number | null
  description: string
  is_active?: boolean
}

export interface AgentAccessKeyCreateResult {
  key: AgentAccessKey
  plaintext_key: string
}
