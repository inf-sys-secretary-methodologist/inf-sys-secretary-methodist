import { act } from '@testing-library/react'
import { useAuthStore } from '../authStore'
import { authApi } from '@/lib/api/auth'
import { UserRole } from '@/types/auth'

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

// Track time for bypassing debounce in checkAuth
let testTime = 0

describe('authStore', () => {
  beforeEach(() => {
    // Mock Date.now to bypass checkAuth debounce
    testTime += 2000 // Advance 2 seconds between tests
    jest.spyOn(Date, 'now').mockReturnValue(testTime)

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

  afterEach(() => {
    jest.restoreAllMocks()
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
        user: { id: 1, name: 'Test', email: 'test@test.com', role: UserRole.STUDENT },
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

    it('handles login response without wrapper data', async () => {
      const mockUser = {
        id: 1,
        name: 'Test User',
        email: 'test@test.com',
        role: 'student' as const,
      }

      mockedAuthApi.login.mockResolvedValue({
        user: mockUser,
        token: 'direct-token',
        refreshToken: 'direct-refresh',
      } as never)

      await act(async () => {
        await useAuthStore.getState().login({ email: 'test@test.com', password: 'password' })
      })

      const state = useAuthStore.getState()
      expect(state.token).toBe('direct-token')
      expect(state.user).toEqual(mockUser)
    })

    it('uses nested error message when available', async () => {
      mockedAuthApi.login.mockRejectedValue({
        response: { data: { error: { message: 'Nested error message' } } },
      })

      await expect(
        act(async () => {
          await useAuthStore.getState().login({ email: 'test@test.com', password: 'wrong' })
        })
      ).rejects.toBeDefined()

      expect(useAuthStore.getState().error).toBe('Nested error message')
    })

    it('uses default error code when no message in response', async () => {
      mockedAuthApi.login.mockRejectedValue({
        response: { data: {} },
      })

      await expect(
        act(async () => {
          await useAuthStore.getState().login({ email: 'test@test.com', password: 'wrong' })
        })
      ).rejects.toBeDefined()

      expect(useAuthStore.getState().error).toBe('LOGIN_ERROR')
    })
  })

  describe('register', () => {
    it('sets authenticated state on successful register', async () => {
      const mockUser = {
        id: 2,
        name: 'New User',
        email: 'new@test.com',
        role: UserRole.STUDENT,
      }

      mockedAuthApi.register.mockResolvedValue({
        data: {
          user: mockUser,
          token: 'new-token',
          refreshToken: 'new-refresh',
        },
      } as never)

      await act(async () => {
        await useAuthStore.getState().register({
          name: 'New User',
          email: 'new@test.com',
          password: 'Password1!',
          role: UserRole.STUDENT,
        })
      })

      const state = useAuthStore.getState()
      expect(state.user).toEqual(mockUser)
      expect(state.token).toBe('new-token')
      expect(state.isAuthenticated).toBe(true)
    })

    it('sets error state on failed register', async () => {
      mockedAuthApi.register.mockRejectedValue({
        response: { data: { message: 'Email already exists' } },
      })

      await expect(
        act(async () => {
          await useAuthStore.getState().register({
            name: 'New User',
            email: 'existing@test.com',
            password: 'Password1!',
            role: UserRole.STUDENT,
          })
        })
      ).rejects.toBeDefined()

      expect(useAuthStore.getState().error).toBe('Email already exists')
    })

    it('uses default error code when no message', async () => {
      mockedAuthApi.register.mockRejectedValue({
        response: { data: {} },
      })

      await expect(
        act(async () => {
          await useAuthStore.getState().register({
            name: 'New User',
            email: 'test@test.com',
            password: 'Password1!',
            role: UserRole.STUDENT,
          })
        })
      ).rejects.toBeDefined()

      expect(useAuthStore.getState().error).toBe('REGISTER_ERROR')
    })
  })

  describe('refreshAccessToken', () => {
    it('updates tokens on successful refresh', async () => {
      useAuthStore.setState({
        refreshToken: 'old-refresh-token',
        token: 'old-access-token',
      })

      mockedAuthApi.refreshToken.mockResolvedValue({
        token: 'new-access-token',
        refreshToken: 'new-refresh-token',
        expiresIn: 3600,
      })

      await act(async () => {
        await useAuthStore.getState().refreshAccessToken()
      })

      const state = useAuthStore.getState()
      expect(state.token).toBe('new-access-token')
      expect(state.refreshToken).toBe('new-refresh-token')
    })

    it('throws error when no refresh token available', async () => {
      useAuthStore.setState({ refreshToken: null })

      await expect(
        act(async () => {
          await useAuthStore.getState().refreshAccessToken()
        })
      ).rejects.toThrow('No refresh token available')
    })

    it('logs out user when refresh fails', async () => {
      useAuthStore.setState({
        refreshToken: 'invalid-token',
        token: 'access-token',
        isAuthenticated: true,
      })

      mockedAuthApi.refreshToken.mockRejectedValue(new Error('Refresh failed'))

      await expect(
        act(async () => {
          await useAuthStore.getState().refreshAccessToken()
        })
      ).rejects.toThrow()

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.token).toBeNull()
    })
  })

  describe('checkAuth', () => {
    it('sets isAuthenticated false when no token', async () => {
      useAuthStore.setState({ token: null })

      await act(async () => {
        await useAuthStore.getState().checkAuth()
      })

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })

    it('verifies token and loads user when token exists', async () => {
      const mockUser = {
        id: 1,
        name: 'Existing User',
        email: 'existing@test.com',
        role: UserRole.STUDENT,
      }

      useAuthStore.setState({ token: 'valid-token' })
      mockedAuthApi.getCurrentUser.mockResolvedValue(mockUser)

      await act(async () => {
        await useAuthStore.getState().checkAuth()
      })

      const state = useAuthStore.getState()
      expect(state.user).toEqual(mockUser)
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it('clears auth state when token verification fails', async () => {
      useAuthStore.setState({
        token: 'invalid-token',
        isAuthenticated: true,
        user: { id: 1, name: 'Test', email: 'test@test.com', role: UserRole.STUDENT },
      })

      mockedAuthApi.getCurrentUser.mockRejectedValue(new Error('Token invalid'))

      await act(async () => {
        await useAuthStore.getState().checkAuth()
      })

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
      expect(state.user).toBeNull()
      expect(state.token).toBeNull()
    })
  })
})
