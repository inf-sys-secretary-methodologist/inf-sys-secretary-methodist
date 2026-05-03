import { apiClient } from '../api'
import type {
  LoginRequest,
  RegisterRequest,
  RefreshTokenRequest,
  AuthResponse,
  RefreshTokenResponse,
  User,
} from '@/types/auth'

/**
 * Auth API endpoints
 */
export const authApi = {
  /**
   * Login user
   */
  login: async (credentials: LoginRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/api/auth/login', credentials)
  },

  /**
   * Register new user
   */
  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/api/auth/register', data)
  },

  /**
   * Logout user
   */
  logout: async (): Promise<void> => {
    return apiClient.post<void>('/api/auth/logout')
  },

  /**
   * Refresh access token
   */
  refreshToken: async (request: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
    return apiClient.post<RefreshTokenResponse>('/api/auth/refresh', request)
  },

  /**
   * Get current user profile
   */
  getCurrentUser: async (): Promise<User> => {
    return apiClient.get<User>('/api/me')
  },

  /**
   * Verify token validity
   */
  verifyToken: async (token: string): Promise<{ valid: boolean }> => {
    return apiClient.post<{ valid: boolean }>('/api/auth/verify', { token })
  },

  // --- Password recovery (v0.108.0) ---

  /**
   * Request a password reset email. Always resolves to undefined for
   * any well-formed email (anti-enumeration honored by the backend);
   * the form renders a generic success regardless.
   */
  requestPasswordReset: async (email: string): Promise<void> => {
    return apiClient.post<void>('/api/auth/password-reset/request', { email })
  },

  /**
   * Probe a reset token's validity without consuming it. Resolves on
   * 204; the caller catches a 410 Gone to render "link expired".
   */
  verifyPasswordResetToken: async (token: string): Promise<void> => {
    return apiClient.get<void>(`/api/auth/password-reset/verify/${encodeURIComponent(token)}`)
  },

  /**
   * Consume a reset token and set a new password. 410 Gone on dead
   * link, 400 on weak password — caller maps to UI message.
   */
  confirmPasswordReset: async (token: string, password: string): Promise<void> => {
    return apiClient.post<void>('/api/auth/password-reset/confirm', { token, password })
  },
}
