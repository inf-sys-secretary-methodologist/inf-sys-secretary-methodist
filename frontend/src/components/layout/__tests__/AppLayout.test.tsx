import { render, screen } from '@testing-library/react'
import { AppLayout } from '../AppLayout'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      loading: 'Loading...',
    }
    return translations[key] || key
  },
}))

// Mock useAuthCheck
const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

// Mock getAvailableNavItems
jest.mock('@/config/navigation', () => ({
  getAvailableNavItems: jest.fn(() => [
    { href: '/dashboard', label: 'Dashboard' },
    { href: '/documents', label: 'Documents' },
  ]),
}))

// Mock useRouteAnnouncer
jest.mock('@/hooks/useRouteAnnouncer', () => ({
  useRouteAnnouncer: jest.fn(),
}))

// Mock AppHeader
jest.mock('../AppHeader', () => ({
  AppHeader: ({ items }: { items: unknown[] }) => (
    <header data-testid="app-header">AppHeader with {items.length} items</header>
  ),
}))

// Mock InstallPrompt
jest.mock('@/components/pwa/install-prompt', () => ({
  InstallPrompt: () => <div data-testid="install-prompt">InstallPrompt</div>,
}))

// Mock SkipToContent
jest.mock('@/components/ui/skip-to-content', () => ({
  SkipToContent: () => <a data-testid="skip-to-content">Skip to content</a>,
}))

describe('AppLayout', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders loading state when isLoading is true', () => {
    mockUseAuthCheck.mockReturnValue({
      user: null,
      isLoading: true,
    })

    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>
    )

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Content')).not.toBeInTheDocument()
  })

  it('renders children when not loading', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'admin' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Main Content</div>
      </AppLayout>
    )

    expect(screen.getByText('Main Content')).toBeInTheDocument()
    expect(screen.queryByText('Loading...')).not.toBeInTheDocument()
  })

  it('renders AppHeader with navigation items', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'admin' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>
    )

    expect(screen.getByTestId('app-header')).toBeInTheDocument()
    expect(screen.getByText('AppHeader with 2 items')).toBeInTheDocument()
  })

  it('renders SkipToContent component', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'admin' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>
    )

    expect(screen.getByTestId('skip-to-content')).toBeInTheDocument()
  })

  it('renders InstallPrompt component', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'admin' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>
    )

    expect(screen.getByTestId('install-prompt')).toBeInTheDocument()
  })

  it('renders main content with correct id and tabIndex', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'admin' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Content</div>
      </AppLayout>
    )

    const main = screen.getByRole('main')
    expect(main).toHaveAttribute('id', 'main-content')
    expect(main).toHaveAttribute('tabindex', '-1')
  })

  it('renders with user having different roles', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: '1', role: 'methodist' },
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Methodist Content</div>
      </AppLayout>
    )

    expect(screen.getByText('Methodist Content')).toBeInTheDocument()
  })

  it('renders when user is null (not authenticated)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: null,
      isLoading: false,
    })

    render(
      <AppLayout>
        <div>Public Content</div>
      </AppLayout>
    )

    expect(screen.getByText('Public Content')).toBeInTheDocument()
  })
})
