export interface User {
  id: number
  username: string
  email: string
  display_name: string
  auth_source: string
  is_active: boolean
  last_login_at: string | null
  roles: Role[]
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

export interface LoginRequest {
  username: string
  password: string
  auth_source?: string
}

export interface LoginResponse {
  token: string
  user: Pick<User, 'id' | 'username' | 'email' | 'display_name' | 'auth_source'>
}
