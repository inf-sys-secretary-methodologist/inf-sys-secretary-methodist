import { act } from '@testing-library/react'
import { useAuthStore } from '../authStore'
import { authApi } from '@/lib/api/auth'

// Mock the API
jest.mock('@/lib/api/auth', () => ({
  authApi: {
    login: jest.fn(),
    register: jest.fn(),
    refreshToken: jest.fn(),
    getCurrentUser: jest.fn(),
  },
}))

jest.mock('@/lib/api', () => ({
  apiClient: {
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockedAuthApi = jest.mocked(authApi)

// Mock document.cookie
const mockCookies: Record<string, string> = {}
Object.defineProperty(document, 'cookie', {
  get: () =>
    Object.entries(mockCookies)
      .map(([k, v]) => `${k}=${v}`)
      .join('; '),
  set: (value: string) => {
    const [cookiePart] = value.split(';')
    const [key, val] = cookiePart.split('=')
    if (val === '' || value.includes('expires=Thu, 01 Jan 1970')) {
      delete mockCookies[key]
    } else {
      mockCookies[key] = val
    }
  },
  configurable: true,
})

describe('authStore', () => {
  beforeEach(() => {
    // Reset store state
    useAuthStore.setState({
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    })

    // Clear mocks
    jest.clearAllMocks()

    // Clear mock cookies
    Object.keys(mockCookies).forEach((key) => delete mockCookies[key])
  })

  describe('initial state', () => {
    it('has correct initial values', () => {
      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.token).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.error).toBeNull()
    })
  })

  describe('logout', () => {
    it('clears auth state', () => {
      // Set up authenticated state
      useAuthStore.setState({
        user: { id: 1, name: 'Test', email: 'test@test.com', role: 'student' as const },
        token: 'test-token',
        refreshToken: 'test-refresh',
        isAuthenticated: true,
      })

      // Call logout
      act(() => {
        useAuthStore.getState().logout()
      })

      // Check state is cleared
      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.token).toBeNull()
      expect(state.refreshToken).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })
  })

  describe('clearError', () => {
    it('clears error state', () => {
      useAuthStore.setState({ error: 'Some error' })

      act(() => {
        useAuthStore.getState().clearError()
      })

      expect(useAuthStore.getState().error).toBeNull()
    })
  })

  describe('setLoading', () => {
    it('sets loading state', () => {
      expect(useAuthStore.getState().isLoading).toBe(false)

      act(() => {
        useAuthStore.getState().setLoading(true)
      })

      expect(useAuthStore.getState().isLoading).toBe(true)

      act(() => {
        useAuthStore.getState().setLoading(false)
      })

      expect(useAuthStore.getState().isLoading).toBe(false)
    })
  })

  describe('login', () => {
    it('sets authenticated state on successful login', async () => {
      const mockUser = {
        id: 1,
        name: 'Test User',
        email: 'test@test.com',
        role: 'student' as const,
      }

      mockedAuthApi.login.mockResolvedValue({
        data: {
          user: mockUser,
          token: 'access-token',
          refreshToken: 'refresh-token',
        },
      } as never)

      await act(async () => {
        await useAuthStore.getState().login({ email: 'test@test.com', password: 'password' })
      })

      const state = useAuthStore.getState()
      expect(state.user).toEqual(mockUser)
      expect(state.token).toBe('access-token')
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
    })

    it('sets error state on failed login', async () => {
      mockedAuthApi.login.mockRejectedValue({
        response: { data: { message: 'Invalid credentials' } },
      })

      await expect(
        act(async () => {
          await useAuthStore.getState().login({ email: 'test@test.com', password: 'wrong' })
        })
      ).rejects.toBeDefined()

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.isLoading).toBe(false)
      expect(state.error).toBe('Invalid credentials')
    })
  })
})
