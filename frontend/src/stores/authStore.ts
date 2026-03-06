import { create } from 'zustand'
import { persist, PersistStorage, StorageValue } from 'zustand/middleware'
import { authApi } from '@/lib/api/auth'
import { apiClient } from '@/lib/api'
import { getStoredToken } from '@/lib/auth/token'
import type { User, LoginRequest, RegisterRequest } from '@/types/auth'

/* c8 ignore start - Cookie helper functions */
// Helper functions for cookie operations
const setCookie = (name: string, value: string, days: number = 7) => {
  const expires = new Date()
  expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000)
  // Encode the JSON value properly for cookie storage
  const encodedValue = encodeURIComponent(value)
  document.cookie = `${name}=${encodedValue};expires=${expires.toUTCString()};path=/;SameSite=Lax${process.env.NODE_ENV === 'production' ? ';Secure' : ''}`
}

const getCookie = (name: string): string | null => {
  const nameEQ = name + '='
  const ca = document.cookie.split(';')
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i]
    while (c.charAt(0) === ' ') c = c.substring(1, c.length)
    if (c.indexOf(nameEQ) === 0) {
      const value = c.substring(nameEQ.length, c.length)
      // Decode the cookie value
      return decodeURIComponent(value)
    }
  }
  return null
}
/* c8 ignore stop */

const deleteCookie = (name: string) => {
  document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/`
}

/* c8 ignore start - Cookie storage for Zustand persist */
// Custom cookie storage for Zustand persist
const cookieStorage = {
  getItem: (name: string): StorageValue<unknown> | null => {
    if (typeof window === 'undefined') return null
    const value = getCookie(name)
    if (!value) return null
    try {
      return JSON.parse(value) as StorageValue<unknown>
    } catch {
      return null
    }
  },
  setItem: (name: string, value: StorageValue<unknown>): void => {
    if (typeof window === 'undefined') return
    const jsonString = JSON.stringify(value)
    setCookie(name, jsonString, 7)
  },
  removeItem: (name: string): void => {
    if (typeof window === 'undefined') return
    deleteCookie(name)
  },
} satisfies PersistStorage<unknown>
/* c8 ignore stop */

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

/* c8 ignore start - Broken cookie cleanup */
// Initialize by cleaning up broken cookies
if (typeof window !== 'undefined') {
  const brokenCookie = getCookie('auth-storage')
  if (
    brokenCookie &&
    (brokenCookie === '[object Object]' || brokenCookie.includes('[object Object]'))
  ) {
    deleteCookie('auth-storage')
  }
}
/* c8 ignore stop */

// Flag to prevent multiple simultaneous checkAuth calls
let isCheckingAuth = false
let lastCheckAuthTime = 0
const CHECK_AUTH_DEBOUNCE_MS = 1000 // Don't call checkAuth more than once per second

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // Initial state - isLoading: true until Zustand hydrates from cookie
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: true, // Start with true, will be set to false after hydration
      error: null,

      // Login action
      login: async (credentials) => {
        set({ isLoading: true, error: null })
        try {
          const response = await authApi.login(credentials)

          // Extract data from response wrapper
          const authData =
            (response as { data?: { user: User; token: string; refreshToken: string } }).data ||
            (response as { user: User; token: string; refreshToken: string })

          // Set token in API client
          apiClient.setAuthToken(authData.token)

          set({
            user: authData.user,
            token: authData.token,
            refreshToken: authData.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })
        } catch (error: unknown) {
          const errorMessage =
            (error as { response?: { data?: { error?: { message?: string }; message?: string } } })
              .response?.data?.error?.message ||
            (error as { response?: { data?: { message?: string } } }).response?.data?.message ||
            'LOGIN_ERROR' // Error code, will be translated by UI components
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

          /* c8 ignore start - Extract data from response wrapper */
          // Extract data from response wrapper
          const authData =
            (response as { data?: { user: User; token: string; refreshToken: string } }).data ||
            (response as { user: User; token: string; refreshToken: string })
          /* c8 ignore stop */

          // Set token in API client
          apiClient.setAuthToken(authData.token)

          set({
            user: authData.user,
            token: authData.token,
            refreshToken: authData.refreshToken,
            isAuthenticated: true,
            isLoading: false,
            error: null,
          })
        } catch (error: unknown) {
          /* c8 ignore start - Error message extraction */
          const errorMessage =
            (error as { response?: { data?: { error?: { message?: string }; message?: string } } })
              .response?.data?.error?.message ||
            (error as { response?: { data?: { message?: string } } }).response?.data?.message ||
            'REGISTER_ERROR' // Error code, will be translated by UI components
          /* c8 ignore stop */
          set({
            isLoading: false,
            error: errorMessage,
          })
          throw error
        }
      },

      // Logout action
      logout: () => {
        // Clear API client token
        apiClient.clearAuthToken()

        // Explicitly delete the auth cookie
        deleteCookie('auth-storage')

        set({
          user: null,
          token: null,
          refreshToken: null,
          isAuthenticated: false,
          isLoading: false,
          error: null,
        })
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
        // Prevent multiple simultaneous calls and debounce
        const now = Date.now()
        /* c8 ignore start - Debounce logic */
        if (isCheckingAuth || now - lastCheckAuthTime < CHECK_AUTH_DEBOUNCE_MS) {
          return
        }
        /* c8 ignore stop */
        isCheckingAuth = true
        lastCheckAuthTime = now

        try {
          const state = get()

          // Try to get token from state or storage
          const token = state.token || getStoredToken()

          if (!token) {
            set({ isAuthenticated: false, isLoading: false })
            return
          }

          set({ isLoading: true })
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
          console.error('❌ checkAuth failed:', error)
          // If check fails, clear auth state directly (don't call logout to avoid cascade)
          apiClient.clearAuthToken()
          deleteCookie('auth-storage')
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          })
        } finally {
          isCheckingAuth = false
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
      name: 'auth-storage', // cookie name
      storage: cookieStorage,
      // Only persist necessary fields
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
      // Called after hydration completes
      onRehydrateStorage: () => {
        return () => {
          // Set isLoading to false after hydration completes
          useAuthStore.setState({ isLoading: false })
        }
      },
    }
  )
)
