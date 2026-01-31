import axios, { AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import type { AuthResponse, LoginRequest, ChangePasswordRequest, User, CreateUserRequest, UpdateUserRequest } from '@/types'

class ApiClient {
  private client: AxiosInstance
  private refreshPromise: Promise<string> | null = null

  constructor() {
    this.client = axios.create({
      baseURL: '/api',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = localStorage.getItem('access_token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor to handle token refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true

          try {
            const newToken = await this.refreshToken()
            originalRequest.headers.Authorization = `Bearer ${newToken}`
            return this.client(originalRequest)
          } catch (refreshError) {
            // Refresh failed, logout user
            localStorage.removeItem('access_token')
            localStorage.removeItem('refresh_token')
            window.location.href = '/login'
            return Promise.reject(refreshError)
          }
        }

        return Promise.reject(error)
      }
    )
  }

  private async refreshToken(): Promise<string> {
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.client
      .post<AuthResponse>('/auth/refresh', {
        refresh_token: localStorage.getItem('refresh_token'),
      })
      .then((response) => {
        localStorage.setItem('access_token', response.data.access_token)
        localStorage.setItem('refresh_token', response.data.refresh_token)
        return response.data.access_token
      })
      .finally(() => {
        this.refreshPromise = null
      })

    return this.refreshPromise
  }

  // Auth endpoints
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<AuthResponse>('/auth/login', credentials)
    this.setTokens(response.data.access_token, response.data.refresh_token)
    return response.data
  }

  async logout(): Promise<void> {
    await this.client.post('/auth/logout')
    this.clearTokens()
  }

  async getCurrentUser(): Promise<User> {
    const response = await this.client.get<User>('/auth/me')
    return response.data
  }

  async changePassword(data: ChangePasswordRequest): Promise<void> {
    await this.client.put('/auth/password', data)
  }

  async togglePasswordLogin(enabled: boolean): Promise<void> {
    await this.client.post('/auth/password/toggle', { enabled })
  }

  async removePassword(): Promise<void> {
    await this.client.delete('/auth/password')
  }

  // OIDC endpoints
  async getOIDCStatus(): Promise<{ enabled: boolean; provider_name: string }> {
    const response = await this.client.get('/auth/oidc/status')
    return response.data
  }

  async initiateOIDCLogin(): Promise<{ auth_url: string; state: string }> {
    const response = await this.client.get('/auth/oidc')
    return response.data
  }

  async handleOIDCCallback(code: string, state: string): Promise<AuthResponse> {
    const response = await this.client.get<AuthResponse>(`/auth/oidc/callback?code=${code}&state=${state}`)
    this.setTokens(response.data.access_token, response.data.refresh_token)
    return response.data
  }

  async initiateOIDCLink(): Promise<{ auth_url: string; state: string }> {
    const response = await this.client.post('/auth/oidc/link')
    return response.data
  }

  async unlinkOIDC(): Promise<void> {
    await this.client.delete('/auth/oidc/link')
  }

  // User management endpoints (admin only)
  async listUsers(): Promise<User[]> {
    const response = await this.client.get<{ users: User[] }>('/users')
    return response.data.users
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    const response = await this.client.post<User>('/users', data)
    return response.data
  }

  async getUser(id: number | 'me'): Promise<User> {
    const response = await this.client.get<User>(`/users/${id}`)
    return response.data
  }

  async updateUser(id: number, data: UpdateUserRequest): Promise<User> {
    const response = await this.client.put<User>(`/users/${id}`, data)
    return response.data
  }

  async deleteUser(id: number): Promise<void> {
    await this.client.delete(`/users/${id}`)
  }

  async resetPassword(id: number, newPassword: string): Promise<void> {
    await this.client.post(`/users/${id}/reset-password`, { new_password: newPassword })
  }

  private setTokens(accessToken: string, refreshToken: string) {
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
  }

  private clearTokens() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }
}

export const api = new ApiClient()