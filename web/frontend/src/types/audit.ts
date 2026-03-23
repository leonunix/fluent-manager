export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource: string
  resource_type: string
  resource_id: number
  detail: string
  ip: string
  created_at: string
}
