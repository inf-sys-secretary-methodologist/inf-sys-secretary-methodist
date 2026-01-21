import { render, screen, waitFor } from '@testing-library/react'
import { withAuth } from '../withAuth'
import { UserRole } from '@/types/auth'

// Mock next/navigation
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      loading: 'Loading...',
    }
    return translations[key] || key
  },
}))

// Mock auth store
const mockAuthStore = {
  isAuthenticated: false,
  isLoading: false,
  user: null as null | { role: UserRole; id: number; email: string; name: string },
  checkAuth: jest.fn().mockResolvedValue(undefined),
}

jest.mock('@/stores/authStore', () => ({
  useAuthStore: () => mockAuthStore,
}))

// Test component
const TestComponent = () => <div data-testid="protected-content">Protected Content</div>

describe('withAuth HOC', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset mock state
    mockAuthStore.isAuthenticated = false
    mockAuthStore.isLoading = false
    mockAuthStore.user = null
    mockAuthStore.checkAuth = jest.fn().mockResolvedValue(undefined)

    // Mock window.location
    Object.defineProperty(window, 'location', {
      value: { pathname: '/dashboard' },
      writable: true,
    })
  })

  it('shows loading state while checking auth', () => {
    mockAuthStore.isLoading = true

    const ProtectedComponent = withAuth(TestComponent)
    render(<ProtectedComponent />)

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
  })

  it('shows custom loading component when provided', () => {
    mockAuthStore.isLoading = true

    const CustomLoading = () => <div data-testid="custom-loading">Custom Loading</div>
    const ProtectedComponent = withAuth(TestComponent, { LoadingComponent: CustomLoading })
    render(<ProtectedComponent />)

    expect(screen.getByTestId('custom-loading')).toBeInTheDocument()
  })

  it('redirects to login when not authenticated', async () => {
    mockAuthStore.isAuthenticated = false

    const ProtectedComponent = withAuth(TestComponent)
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/login?redirect=%2Fdashboard')
    })
  })

  it('redirects to custom path when specified', async () => {
    mockAuthStore.isAuthenticated = false

    const ProtectedComponent = withAuth(TestComponent, { redirectTo: '/custom-login' })
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/custom-login?redirect=%2Fdashboard')
    })
  })

  it('renders component when authenticated', async () => {
    mockAuthStore.isAuthenticated = true
    mockAuthStore.user = {
      id: 1,
      email: 'test@test.com',
      name: 'Test User',
      role: UserRole.SYSTEM_ADMIN,
    }

    const ProtectedComponent = withAuth(TestComponent)
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(screen.getByTestId('protected-content')).toBeInTheDocument()
    })
  })

  it('redirects to forbidden when user lacks required role', async () => {
    mockAuthStore.isAuthenticated = true
    mockAuthStore.user = {
      id: 1,
      email: 'test@test.com',
      name: 'Test User',
      role: UserRole.ACADEMIC_SECRETARY,
    }

    const ProtectedComponent = withAuth(TestComponent, { roles: [UserRole.SYSTEM_ADMIN] })
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/forbidden')
    })
  })

  it('renders component when user has required role', async () => {
    mockAuthStore.isAuthenticated = true
    mockAuthStore.user = {
      id: 1,
      email: 'test@test.com',
      name: 'Test User',
      role: UserRole.SYSTEM_ADMIN,
    }

    const ProtectedComponent = withAuth(TestComponent, { roles: [UserRole.SYSTEM_ADMIN] })
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(screen.getByTestId('protected-content')).toBeInTheDocument()
    })
  })

  it('renders component when user has one of multiple allowed roles', async () => {
    mockAuthStore.isAuthenticated = true
    mockAuthStore.user = {
      id: 1,
      email: 'test@test.com',
      name: 'Test User',
      role: UserRole.METHODIST,
    }

    const ProtectedComponent = withAuth(TestComponent, {
      roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
    })
    render(<ProtectedComponent />)

    await waitFor(() => {
      expect(screen.getByTestId('protected-content')).toBeInTheDocument()
    })
  })

  it('calls checkAuth on mount', () => {
    const ProtectedComponent = withAuth(TestComponent)
    render(<ProtectedComponent />)

    expect(mockAuthStore.checkAuth).toHaveBeenCalled()
  })

  it('sets correct displayName', () => {
    const ProtectedComponent = withAuth(TestComponent)
    expect(ProtectedComponent.displayName).toBe('withAuth(TestComponent)')
  })

  it('passes props to wrapped component', async () => {
    mockAuthStore.isAuthenticated = true
    mockAuthStore.user = {
      id: 1,
      email: 'test@test.com',
      name: 'Test User',
      role: UserRole.SYSTEM_ADMIN,
    }

    const ComponentWithProps = ({ title }: { title: string }) => (
      <div data-testid="component-with-props">{title}</div>
    )

    const ProtectedComponent = withAuth(ComponentWithProps)
    render(<ProtectedComponent title="Test Title" />)

    await waitFor(() => {
      expect(screen.getByText('Test Title')).toBeInTheDocument()
    })
  })
})
