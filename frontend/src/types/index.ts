export interface User {
  id: number
  email: string
  role: 'admin' | 'user'
  is_active: boolean
  password_enabled: boolean
  oidc_provider?: string
  oidc_enabled: boolean
  is_linked_to_oidc: boolean
  created_at: string
  updated_at: string
  last_login_at?: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface ChangePasswordRequest {
  current_password?: string
  new_password: string
}

export interface CreateUserRequest {
  email: string
  password?: string
  role: 'admin' | 'user'
  oidc_enabled: boolean
}

export interface UpdateUserRequest {
  email?: string
  role?: 'admin' | 'user'
  is_active?: boolean
  oidc_enabled?: boolean
}