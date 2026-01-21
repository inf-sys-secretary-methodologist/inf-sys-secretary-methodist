import { authApi } from '../auth'
import { apiClient } from '../../api'
import { UserRole } from '@/types/auth'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('authApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('login', () => {
    it('calls login endpoint with credentials', async () => {
      const mockResponse = {
        user: { id: 1, email: 'test@test.com', name: 'Test', role: UserRole.STUDENT },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
        expiresIn: 3600,
      }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const credentials = { email: 'test@test.com', password: 'password123' }
      const result = await authApi.login(credentials)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/login', credentials)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('register', () => {
    it('calls register endpoint with user data', async () => {
      const mockResponse = {
        user: { id: 1, email: 'new@test.com', name: 'New User', role: UserRole.STUDENT },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
        expiresIn: 3600,
      }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const userData = {
        email: 'new@test.com',
        password: 'password123',
        name: 'New User',
        role: UserRole.STUDENT,
      }
      const result = await authApi.register(userData)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/register', userData)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('logout', () => {
    it('calls logout endpoint', async () => {
      mockedApiClient.post.mockResolvedValue(undefined)

      await authApi.logout()

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/logout')
    })
  })

  describe('refreshToken', () => {
    it('calls refresh endpoint with token', async () => {
      const mockResponse = {
        token: 'new-jwt-token',
        refreshToken: 'new-refresh-token',
        expiresIn: 3600,
      }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const result = await authApi.refreshToken({ refreshToken: 'old-refresh-token' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/refresh', {
        refreshToken: 'old-refresh-token',
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getCurrentUser', () => {
    it('calls get current user endpoint', async () => {
      const mockUser = {
        id: 1,
        email: 'test@test.com',
        name: 'Test User',
        role: UserRole.STUDENT,
      }
      mockedApiClient.get.mockResolvedValue(mockUser)

      const result = await authApi.getCurrentUser()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/me')
      expect(result).toEqual(mockUser)
    })
  })

  describe('verifyToken', () => {
    it('calls verify endpoint with token', async () => {
      mockedApiClient.post.mockResolvedValue({ valid: true })

      const result = await authApi.verifyToken('test-token')

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/auth/verify', { token: 'test-token' })
      expect(result).toEqual({ valid: true })
    })

    it('returns invalid for expired token', async () => {
      mockedApiClient.post.mockResolvedValue({ valid: false })

      const result = await authApi.verifyToken('expired-token')

      expect(result).toEqual({ valid: false })
    })
  })
})
