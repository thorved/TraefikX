import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import type { User, LoginRequest } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  // Actions
  async function login(credentials: LoginRequest) {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.login(credentials)
      user.value = response.user
      return response
    } catch (err: any) {
      error.value = err.response?.data?.error || 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    try {
      await api.logout()
    } finally {
      user.value = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    }
  }

  async function fetchUser() {
    const token = localStorage.getItem('access_token')
    if (!token) {
      user.value = null
      return
    }

    try {
      user.value = await api.getCurrentUser()
    } catch (err) {
      user.value = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
    }
  }

  async function changePassword(currentPassword: string, newPassword: string) {
    try {
      await api.changePassword({ current_password: currentPassword, new_password: newPassword })
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Failed to change password')
    }
  }

  async function togglePasswordLogin(enabled: boolean) {
    try {
      await api.togglePasswordLogin(enabled)
      if (user.value) {
        user.value.password_enabled = enabled
      }
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Failed to toggle password login')
    }
  }

  async function removePassword() {
    try {
      await api.removePassword()
      if (user.value) {
        user.value.password_enabled = false
      }
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Failed to remove password')
    }
  }

  async function unlinkOIDC() {
    try {
      await api.unlinkOIDC()
      if (user.value) {
        user.value.oidc_enabled = false
        user.value.oidc_provider = undefined
        user.value.is_linked_to_oidc = false
      }
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Failed to unlink OIDC')
    }
  }

  return {
    user,
    isLoading,
    error,
    isAuthenticated,
    isAdmin,
    login,
    logout,
    fetchUser,
    changePassword,
    togglePasswordLogin,
    removePassword,
    unlinkOIDC,
  }
})