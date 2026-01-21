/* eslint-disable @typescript-eslint/no-explicit-any */
import { renderHook, act } from '@testing-library/react'
import { useAuth, useLogout, useLogin, useRegister, useAuthCheck, useRequireAuth } from '../useAuth'
import { useAuthStore } from '@/stores/authStore'
import { UserRole } from '@/types/auth'

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
  }),
}))

// Mock the auth store
jest.mock('@/stores/authStore')

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

describe('useAuth', () => {
  const mockUser = { id: 1, name: 'Test', email: 'test@test.com', role: 'student' }
  const mockLogin = jest.fn()
  const mockRegister = jest.fn()
  const mockLogout = jest.fn()
  const mockCheckAuth = jest.fn()
  const mockClearError = jest.fn()

  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      user: mockUser,
      isAuthenticated: true,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogout,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    } as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('returns user state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('returns auth functions', () => {
    const { result } = renderHook(() => useAuth())

    expect(typeof result.current.login).toBe('function')
    expect(typeof result.current.register).toBe('function')
    expect(typeof result.current.logout).toBe('function')
    expect(typeof result.current.checkAuth).toBe('function')
    expect(typeof result.current.clearError).toBe('function')
  })

  it('returns not authenticated state when user is null', () => {
    mockUseAuthStore.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogout,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useAuth())

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })

  it('returns loading state', () => {
    mockUseAuthStore.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: true,
      error: null,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogout,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useAuth())

    expect(result.current.isLoading).toBe(true)
  })

  it('returns error state', () => {
    const errorMessage = 'Authentication failed'
    mockUseAuthStore.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: errorMessage,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogout,
      checkAuth: mockCheckAuth,
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useAuth())

    expect(result.current.error).toBe(errorMessage)
  })
})

describe('useLogout', () => {
  const mockLogout = jest.fn()

  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      logout: mockLogout,
    } as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('calls logout from store', () => {
    const { result } = renderHook(() => useLogout())

    act(() => {
      result.current.logout('/login')
    })

    expect(mockLogout).toHaveBeenCalled()
  })

  it('returns isLoading as false', () => {
    const { result } = renderHook(() => useLogout())

    expect(result.current.isLoading).toBe(false)
  })
})

describe('useLogin', () => {
  const mockLogin = jest.fn()
  const mockClearError = jest.fn()

  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: null,
      clearError: mockClearError,
    } as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('returns login function and state', () => {
    const { result } = renderHook(() => useLogin())

    expect(typeof result.current.login).toBe('function')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(typeof result.current.clearError).toBe('function')
  })

  it('calls login with credentials', async () => {
    mockLogin.mockResolvedValue(undefined)

    const { result } = renderHook(() => useLogin())

    await act(async () => {
      await result.current.login({ email: 'test@test.com', password: 'password' })
    })

    expect(mockLogin).toHaveBeenCalledWith({ email: 'test@test.com', password: 'password' })
  })

  it('handles login error', async () => {
    const error = new Error('Login failed')
    mockLogin.mockRejectedValue(error)

    const { result } = renderHook(() => useLogin())

    await expect(
      act(async () => {
        await result.current.login({ email: 'test@test.com', password: 'wrong' })
      })
    ).rejects.toThrow('Login failed')
  })

  it('shows loading state', () => {
    mockUseAuthStore.mockReturnValue({
      login: mockLogin,
      isLoading: true,
      error: null,
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useLogin())

    expect(result.current.isLoading).toBe(true)
  })

  it('shows error state', () => {
    mockUseAuthStore.mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: 'Invalid credentials',
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useLogin())

    expect(result.current.error).toBe('Invalid credentials')
  })
})

describe('useRegister', () => {
  const mockRegister = jest.fn()
  const mockClearError = jest.fn()

  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      register: mockRegister,
      isLoading: false,
      error: null,
      clearError: mockClearError,
    } as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('returns register function and state', () => {
    const { result } = renderHook(() => useRegister())

    expect(typeof result.current.register).toBe('function')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(typeof result.current.clearError).toBe('function')
  })

  it('calls register with data', async () => {
    mockRegister.mockResolvedValue(undefined)

    const { result } = renderHook(() => useRegister())

    const registerData = {
      email: 'new@test.com',
      password: 'password123',
      name: 'New User',
      role: UserRole.STUDENT,
    }

    await act(async () => {
      await result.current.register(registerData)
    })

    expect(mockRegister).toHaveBeenCalledWith(registerData)
  })

  it('handles registration error', async () => {
    const error = new Error('Email already exists')
    mockRegister.mockRejectedValue(error)

    const { result } = renderHook(() => useRegister())

    await expect(
      act(async () => {
        await result.current.register({
          email: 'existing@test.com',
          password: 'password',
          name: 'User',
          role: UserRole.STUDENT,
        })
      })
    ).rejects.toThrow('Email already exists')
  })

  it('shows loading state', () => {
    mockUseAuthStore.mockReturnValue({
      register: mockRegister,
      isLoading: true,
      error: null,
      clearError: mockClearError,
    } as any)

    const { result } = renderHook(() => useRegister())

    expect(result.current.isLoading).toBe(true)
  })
})

describe('useAuthCheck', () => {
  const mockCheckAuth = jest.fn()
  const mockUser = { id: 1, name: 'Test', email: 'test@test.com', role: 'student' }

  beforeEach(() => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: mockUser,
      isAuthenticated: true,
      isLoading: false,
    } as any)
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('calls checkAuth on mount', () => {
    renderHook(() => useAuthCheck())

    expect(mockCheckAuth).toHaveBeenCalledTimes(1)
  })

  it('returns auth state', () => {
    const { result } = renderHook(() => useAuthCheck())

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.isLoading).toBe(false)
  })

  it('returns loading state when checking', () => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: null,
      isAuthenticated: false,
      isLoading: true,
    } as any)

    const { result } = renderHook(() => useAuthCheck())

    expect(result.current.isLoading).toBe(true)
    expect(result.current.user).toBeNull()
  })

  it('returns unauthenticated state', () => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: null,
      isAuthenticated: false,
      isLoading: false,
    } as any)

    const { result } = renderHook(() => useAuthCheck())

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
  })
})

describe('useRequireAuth', () => {
  const mockCheckAuth = jest.fn()
  const mockPush = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('does not redirect when authenticated', () => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: { id: 1 },
      isAuthenticated: true,
      isLoading: false,
    } as any)

    const { result } = renderHook(() => useRequireAuth())

    expect(mockPush).not.toHaveBeenCalled()
    expect(result.current.isAuthenticated).toBe(true)
  })

  it('does not redirect while loading', () => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: null,
      isAuthenticated: false,
      isLoading: true,
    } as any)

    const { result } = renderHook(() => useRequireAuth())

    expect(mockPush).not.toHaveBeenCalled()
    expect(result.current.isLoading).toBe(true)
  })

  it('returns authentication state', () => {
    mockUseAuthStore.mockReturnValue({
      checkAuth: mockCheckAuth,
      user: { id: 1 },
      isAuthenticated: true,
      isLoading: false,
    } as any)

    const { result } = renderHook(() => useRequireAuth('/custom-login'))

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.isLoading).toBe(false)
  })
})
