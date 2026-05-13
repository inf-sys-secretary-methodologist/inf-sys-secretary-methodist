import { render, screen, waitFor } from '@testing-library/react'
import AdminComposioPage from '../page'
import type { ComposioConfig } from '@/types/composio'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseComposioConfig = jest.fn()
jest.mock('@/hooks/useComposioConfig', () => ({
  useComposioConfig: (opts?: unknown) => mockUseComposioConfig(opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

const configuredSnapshot: ComposioConfig = {
  configured: true,
  api_key_configured: true,
  entity_id_set: true,
  mcp_config_id_set: true,
}

const partialSnapshot: ComposioConfig = {
  configured: false,
  api_key_configured: true,
  entity_id_set: true,
  mcp_config_id_set: false,
}

const unconfiguredSnapshot: ComposioConfig = {
  configured: false,
  api_key_configured: false,
  entity_id_set: false,
  mcp_config_id_set: false,
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseComposioConfig.mockReturnValue({
    config: configuredSnapshot,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminComposioPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminComposioPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders for system_admin', () => {
    render(<AdminComposioPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-composio-page')).toBeInTheDocument()
  })
})

describe('AdminComposioPage — status card', () => {
  it('renders the configured badge when all three fields are set', () => {
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-status-card')).toBeInTheDocument()
    expect(screen.getByTestId('composio-status-configured')).toBeInTheDocument()
  })

  it('renders the unconfigured badge on partial config', () => {
    mockUseComposioConfig.mockReturnValue({
      config: partialSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-status-unconfigured')).toBeInTheDocument()
  })

  it('renders the unconfigured badge when all fields empty', () => {
    mockUseComposioConfig.mockReturnValue({
      config: unconfiguredSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-status-unconfigured')).toBeInTheDocument()
  })
})

describe('AdminComposioPage — per-field rows', () => {
  it('renders set markers for all three fields when fully configured', () => {
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-field-api-key')).toHaveTextContent('fields.set')
    expect(screen.getByTestId('composio-field-entity-id')).toHaveTextContent('fields.set')
    expect(screen.getByTestId('composio-field-mcp-config-id')).toHaveTextContent('fields.set')
  })

  it('renders unset marker only for the missing field on partial config', () => {
    mockUseComposioConfig.mockReturnValue({
      config: partialSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-field-api-key')).toHaveTextContent('fields.set')
    expect(screen.getByTestId('composio-field-entity-id')).toHaveTextContent('fields.set')
    expect(screen.getByTestId('composio-field-mcp-config-id')).toHaveTextContent('fields.unset')
  })

  it('renders unset markers for all fields when nothing is configured', () => {
    mockUseComposioConfig.mockReturnValue({
      config: unconfiguredSnapshot,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-field-api-key')).toHaveTextContent('fields.unset')
    expect(screen.getByTestId('composio-field-entity-id')).toHaveTextContent('fields.unset')
    expect(screen.getByTestId('composio-field-mcp-config-id')).toHaveTextContent('fields.unset')
  })
})

describe('AdminComposioPage — loading / error', () => {
  it('renders the loading spinner when isLoading=true', () => {
    mockUseComposioConfig.mockReturnValue({
      config: undefined,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseComposioConfig.mockReturnValue({
      config: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminComposioPage />)
    expect(screen.getByTestId('composio-error')).toBeInTheDocument()
  })
})
