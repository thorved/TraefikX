export type UserRole = 'admin' | 'user';

export interface User {
  id: number;
  email: string;
  role: UserRole;
  is_active: boolean;
  password_enabled: boolean;
  oidc_provider?: string;
  oidc_enabled: boolean;
  is_linked_to_oidc: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface UserResponse {
  id: number;
  email: string;
  role: UserRole;
  is_active: boolean;
  password_enabled: boolean;
  oidc_provider?: string;
  oidc_enabled: boolean;
  is_linked_to_oidc: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

export interface CreateUserRequest {
  email: string;
  password?: string;
  role: UserRole;
  oidc_enabled: boolean;
}

export interface UpdateUserRequest {
  email?: string;
  role?: UserRole;
  is_active?: boolean;
  oidc_enabled?: boolean;
}

export interface ChangePasswordRequest {
  current_password?: string;
  new_password: string;
}

export interface OIDCStatus {
  enabled: boolean;
  provider_name?: string;
}

export interface OIDCAuthResponse {
  auth_url: string;
  state: string;
}

export interface ApiError {
  error: string;
}

export interface UsersListResponse {
  users: UserResponse[];
}
