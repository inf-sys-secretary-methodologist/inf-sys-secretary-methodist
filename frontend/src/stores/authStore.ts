import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { authApi } from '@/lib/api/auth'
import { apiClient } from '@/lib/api'
import type { User, LoginRequest, RegisterRequest } from '@/types/auth'

interface AuthState {
  // State
  user: User | null
  token: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  login: (credentials: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => void
  refreshAccessToken: () => Promise<void>
  checkAuth: () => Promise<void>
  clearError: () => void
  setLoading: (loading: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Login action
      login: async (credentials) => {
        set({ isLoading: true, error: null })
        try {
          const response = await authApi.login(credentials)

          // Set token in API client
          apiClient.setAuthToken(response.token)

          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || 'Ошибка входа'
          set({
            isLoading: false,
            error: errorMessage,
          })
          throw error
        }
      },

      // Register action
      register: async (data) => {
        set({ isLoading: true, error: null })
        try {
          const response = await authApi.register(data)

          // Set token in API client
          apiClient.setAuthToken(response.token)

          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })
        } catch (error: any) {
          const errorMessage = error.response?.data?.message || 'Ошибка регистрации'
          set({
            isLoading: false,
            error: errorMessage,
          })
          throw error
        }
      },

      // Logout action
      logout: () => {
        try {
          authApi.logout().catch(() => {
            // Ignore logout API errors
          })
        } finally {
          // Clear API client token
          apiClient.clearAuthToken()

          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          })
        }
      },

      // Refresh token action
      refreshAccessToken: async () => {
        const { refreshToken } = get()
        if (!refreshToken) {
          throw new Error('No refresh token available')
        }

        try {
          const response = await authApi.refreshToken({ refreshToken })

          // Update token in API client
          apiClient.setAuthToken(response.token)

          set({
            token: response.token,
            refreshToken: response.refreshToken,
          })
        } catch (error) {
          // If refresh fails, logout user
          get().logout()
          throw error
        }
      },

      // Check auth status on app load
      checkAuth: async () => {
        const { token } = get()
        if (!token) {
          set({ isAuthenticated: false, isLoading: false })
          return
        }

        set({ isLoading: true })
        try {
          // Set token in API client first
          apiClient.setAuthToken(token)

          // Verify token and get user
          const user = await authApi.getCurrentUser()

          set({
            user,
            isAuthenticated: true,
            isLoading: false,
          })
        } catch (error) {
          // If check fails, clear auth state
          get().logout()
          set({ isLoading: false })
        }
      },

      // Clear error
      clearError: () => {
        set({ error: null })
      },

      // Set loading
      setLoading: (loading: boolean) => {
        set({ isLoading: loading })
      },
    }),
    {
      name: 'auth-storage', // localStorage key
      storage: createJSONStorage(() => localStorage),
      // Only persist necessary fields
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
