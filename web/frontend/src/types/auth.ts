export interface User {
  id: number
  username: string
  email: string
  display_name: string
  auth_source: string
  is_active: boolean
  last_login_at: string | null
  roles: Role[]
  groups: Group[]
  scopes: UserScope[]
  created_at: string
  updated_at: string
}

export interface Role {
  id: number
  name: string
  description: string
  permissions: Permission[]
  created_at: string
  updated_at: string
}

export interface Permission {
  id: number
  resource: string
  action: string
  name: string
}

export interface UserScope {
  id: number
  user_id: number
  scope_type: string
  scope_id: number
  scope_name: string
  created_at: string
}

export interface Group {
  id: number
  name: string
  description: string
  roles: Role[]
  scopes: GroupScope[]
  users?: User[]
  created_at: string
  updated_at: string
}

export interface GroupScope {
  id: number
  group_id: number
  scope_type: string
  scope_id: number
  scope_name: string
  created_at: string
}

export interface ExternalGroupMapping {
  id: number
  source: string
  external_group_name: string
  group_id: number
  group?: Group
  created_at: string
  updated_at: string
}

export interface LDAPSettings {
  enabled: boolean
  host: string
  port: number
  use_tls: boolean
  bind_dn: string
  bind_password: string
  base_dn: string
  user_filter: string
  group_filter: string
  attributes: { username: string; email: string; name: string }
  group_sync_strategy: string
}

export interface SAMLSettings {
  enabled: boolean
  idp_metadata: string
  entity_id: string
  acs_url: string
  cert_data: string
  key_data: string
  group_attribute: string
  group_sync_strategy: string
}

export interface LoginRequest {
  username: string
  password: string
  auth_source?: string
}

export interface LoginResponse {
  token: string
  user: Pick<User, 'id' | 'username' | 'email' | 'display_name' | 'auth_source'>
}
