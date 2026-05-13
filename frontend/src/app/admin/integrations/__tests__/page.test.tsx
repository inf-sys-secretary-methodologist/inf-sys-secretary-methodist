import { render, screen, waitFor } from '@testing-library/react'
import AdminIntegrationsPage from '../page'
import type { IntegrationsConfig } from '@/types/integrations'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseIntegrationsConfig = jest.fn()
jest.mock('@/hooks/useIntegrationsConfig', () => ({
  useIntegrationsConfig: (opts?: unknown) => mockUseIntegrationsConfig(opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

const configuredSnapshot: IntegrationsConfig = {
  vapid: {
    configured: true,
    public_key: 'BPublicKey123',
    subject: 'mailto:admin@example.com',
  },
  n8n: {
    enabled: true,
    webhook_url: 'https://n8n.example.com',
  },
}

const unconfiguredSnapshot: IntegrationsConfig = {
  vapid: { configured: false, public_key: '', subject: '' },
  n8n: { enabled: false, webhook_url: 'http://localhost:5678' },
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseIntegrationsConfig.mockReturnValue({
    config: configuredSnapshot,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminIntegrationsPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminIntegrationsPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders for system_admin', () => {
    render(<AdminIntegrationsPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-integrations-page')).toBeInTheDocument()
  })
})

describe('AdminIntegrationsPage — VAPID card', () => {
  it('renders the VAPID status card with configured badge', () => {
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('vapid-status-card')).toBeInTheDocument()
    expect(screen.getByTestId('vapid-status-configured')).toBeInTheDocument()
  })

  it('renders the unconfigured badge when VAPID is missing', () => {
    mockUseIntegrationsConfig.mockReturnValue({
      config: unconfiguredSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('vapid-status-unconfigured')).toBeInTheDocument()
  })

  it('renders the public key and subject', () => {
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('vapid-public-key')).toHaveTextContent('BPublicKey123')
    expect(screen.getByTestId('vapid-subject')).toHaveTextContent('mailto:admin@example.com')
  })
})

describe('AdminIntegrationsPage — n8n card', () => {
  it('renders the n8n status card with enabled badge', () => {
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('n8n-status-card')).toBeInTheDocument()
    expect(screen.getByTestId('n8n-status-enabled')).toBeInTheDocument()
  })

  it('renders the disabled badge when n8n disabled', () => {
    mockUseIntegrationsConfig.mockReturnValue({
      config: unconfiguredSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('n8n-status-disabled')).toBeInTheDocument()
  })

  it('renders the webhook URL', () => {
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('n8n-webhook-url')).toHaveTextContent('https://n8n.example.com')
  })
})

describe('AdminIntegrationsPage — loading / error', () => {
  it('renders the loading spinner when isLoading=true', () => {
    mockUseIntegrationsConfig.mockReturnValue({
      config: undefined,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('integrations-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseIntegrationsConfig.mockReturnValue({
      config: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminIntegrationsPage />)
    expect(screen.getByTestId('integrations-error')).toBeInTheDocument()
  })
})
