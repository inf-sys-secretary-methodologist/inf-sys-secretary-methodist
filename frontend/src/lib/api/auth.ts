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
}
