import { render, screen, waitFor } from '@testing-library/react'
import AdminSentryPage from '../page'
import type { SentryConfig } from '@/types/sentry'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseSentryConfig = jest.fn()
jest.mock('@/hooks/useSentryConfig', () => ({
  useSentryConfig: (opts?: unknown) => mockUseSentryConfig(opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

const configuredSnapshot: SentryConfig = {
  dsn_configured: true,
  environment: 'production',
  release: '0.133.0',
  traces_sample_rate: 0.1,
  tracing_enabled: true,
}

const unconfiguredSnapshot: SentryConfig = {
  dsn_configured: false,
  environment: 'development',
  release: '0.133.0',
  traces_sample_rate: 0.1,
  tracing_enabled: true,
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseSentryConfig.mockReturnValue({
    config: configuredSnapshot,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminSentryPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminSentryPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders for system_admin', () => {
    render(<AdminSentryPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-sentry-page')).toBeInTheDocument()
  })
})

describe('AdminSentryPage — content', () => {
  it('renders the status card with DSN configured badge', () => {
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-status-card')).toBeInTheDocument()
    expect(screen.getByTestId('sentry-status-configured')).toBeInTheDocument()
  })

  it('renders the unconfigured badge when DSN is missing', () => {
    mockUseSentryConfig.mockReturnValue({
      config: unconfiguredSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-status-unconfigured')).toBeInTheDocument()
  })

  it('renders environment and release metadata', () => {
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-meta-environment')).toHaveTextContent('production')
    expect(screen.getByTestId('sentry-meta-release')).toHaveTextContent('0.133.0')
  })

  it('renders traces sample rate and tracing enabled', () => {
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-meta-traces')).toHaveTextContent('0.1')
    expect(screen.getByTestId('sentry-meta-tracing')).toBeInTheDocument()
  })

  it('renders the loading spinner when isLoading=true', () => {
    mockUseSentryConfig.mockReturnValue({
      config: undefined,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseSentryConfig.mockReturnValue({
      config: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminSentryPage />)
    expect(screen.getByTestId('sentry-error')).toBeInTheDocument()
  })
})
