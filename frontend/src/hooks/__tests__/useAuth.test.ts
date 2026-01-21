/* eslint-disable @typescript-eslint/no-explicit-any */
import { renderHook, act } from '@testing-library/react'
import { useAuth, useLogout } from '../useAuth'
import { useAuthStore } from '@/stores/authStore'

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
