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
  // Secure flag only over HTTPS — checking process.env.NODE_ENV at runtime is
  // unreliable for Next.js apps (NODE_ENV=production baked at `next build`
  // time regardless of where the container later runs). Without HTTPS the
  // browser silently drops cookies marked Secure, breaking auth on
  // http://localhost. window.location.protocol gives the true runtime answer.
  const secureFlag =
    typeof window !== 'undefined' && window.location.protocol === 'https:' ? ';Secure' : ''
  document.cookie = `${name}=${encodedValue};expires=${expires.toUTCString()};path=/;SameSite=Lax${secureFlag}`
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

  // Ephemeral MFA challenge state (v0.125.2). When the backend Login
  // response signals data.mfa_required=true the user has cleared the
  // password gate but still owes the second factor. We hold the
  // 5-min intermediate token + pending user in memory and explicitly
  // exclude them from `partialize` so they never reach cookie / disk.
  mfaIntermediateToken: string | null
  mfaPendingUser: User | null

  // Actions
  login: (credentials: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => void
  refreshAccessToken: () => Promise<void>
  checkAuth: () => Promise<void>
  clearError: () => void
  setLoading: (loading: boolean) => void
  clearMFAChallenge: () => void
  verifyLoginMFA: (code: string) => Promise<void>
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
      mfaIntermediateToken: null,
      mfaPendingUser: null,

      // Login action
      login: async (credentials) => {
        set({ isLoading: true, error: null })
        try {
          const response = await authApi.login(credentials)

          // Extract data from response wrapper
          const authData =
            (
              response as {
                data?: {
                  user: User
                  token?: string
                  refreshToken?: string
                  mfa_required?: boolean
                  intermediate_token?: string
                }
              }
            ).data ||
            (response as {
              user: User
              token?: string
              refreshToken?: string
              mfa_required?: boolean
              intermediate_token?: string
            })

          // MFA gate: backend withheld access/refresh because the user has
          // mfa_enabled=true. Stash the 5-min intermediate token + pending
          // user in ephemeral state — the UI will collect the 6-digit code
          // and call verifyLoginMFA() to finish the handshake.
          if (authData.mfa_required && authData.intermediate_token) {
            set({
              user: null,
              token: null,
              refreshToken: null,
              isAuthenticated: false,
              mfaIntermediateToken: authData.intermediate_token,
              mfaPendingUser: authData.user,
              isLoading: false,
              error: null,
            })
            return
          }

          // Set token in API client
          apiClient.setAuthToken(authData.token as string)

          set({
            user: authData.user,
            token: authData.token as string,
            refreshToken: authData.refreshToken as string,
            isAuthenticated: true,
            mfaIntermediateToken: null,
            mfaPendingUser: null,
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
          mfaIntermediateToken: null,
          mfaPendingUser: null,
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

      // Reset ephemeral MFA challenge state — used when the user
      // cancels the second-factor step or after a successful exchange.
      clearMFAChallenge: () => {
        set({ mfaIntermediateToken: null, mfaPendingUser: null })
      },

      // Exchange the in-memory intermediate token + the 6-digit TOTP
      // code the user typed for the real access+refresh pair.
      // - On success: populates auth state and drops the challenge.
      // - On failure: writes the backend message to state.error and
      //   re-throws so the UI can branch on status. The challenge is
      //   intentionally preserved so the user can retry the code (the
      //   UI is responsible for clearing on a 401, where the
      //   intermediate is dead).
      verifyLoginMFA: async (code: string) => {
        const intermediate = get().mfaIntermediateToken
        if (!intermediate) {
          throw new Error('No MFA challenge in progress')
        }

        set({ isLoading: true, error: null })
        try {
          const response = await authApi.verifyLoginMFA(intermediate, code)

          /* c8 ignore start - response wrapper variants */
          const authData =
            (response as { data?: { user: User; token: string; refreshToken: string } }).data ||
            (response as { user: User; token: string; refreshToken: string })
          /* c8 ignore stop */

          apiClient.setAuthToken(authData.token)

          set({
            user: authData.user,
            token: authData.token,
            refreshToken: authData.refreshToken,
            isAuthenticated: true,
            mfaIntermediateToken: null,
            mfaPendingUser: null,
            isLoading: false,
            error: null,
          })
        } catch (error: unknown) {
          const errorMessage =
            (error as { response?: { data?: { error?: { message?: string }; message?: string } } })
              .response?.data?.error?.message ||
            (error as { response?: { data?: { message?: string } } }).response?.data?.message ||
            'MFA_VERIFY_ERROR'
          set({
            isLoading: false,
            error: errorMessage,
          })
          throw error
        }
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

// Imperative read of "is the user mid-MFA-verify?" used by callers
// that need a sync answer right after `await login()` (the post-await
// guards in LoginForm.onSubmit and useLogin.handleLogin). Subscriptions
// to mfaIntermediateToken are frozen at render-time and miss the value
// the store action wrote in the same microtask, so the helper goes
// through getState() rather than a hook.
export const isMFAChallengeActive = (): boolean =>
  useAuthStore.getState().mfaIntermediateToken !== null
