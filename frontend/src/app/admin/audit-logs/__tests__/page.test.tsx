import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminAuditLogsPage from '../page'
import type { AuditLog, AuditLogPagination } from '@/types/audit'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseAuditLogs = jest.fn()
jest.mock('@/hooks/useAuditLogs', () => ({
  useAuditLogs: (filter?: unknown, opts?: unknown) => mockUseAuditLogs(filter, opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const sampleLog: AuditLog = {
  id: 11,
  created_at: '2026-05-10T12:30:00Z',
  action: 'curriculum.approved',
  resource: 'curriculum',
  actor_user_id: 42,
  actor_ip: '10.0.0.5',
  correlation_id: 'req-7c4f',
  fields: { curriculum_id: 7 },
}

const samplePagination: AuditLogPagination = {
  page: 1,
  per_page: 50,
  total: 1,
  total_pages: 1,
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseAuditLogs.mockReturnValue({
    items: [],
    pagination: samplePagination,
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminAuditLogsPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminAuditLogsPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('does not redirect when role is system_admin', () => {
    render(<AdminAuditLogsPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('does not fire the hook while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({
      user: undefined,
      isAuthenticated: false,
      isLoading: true,
    })
    render(<AdminAuditLogsPage />)
    const lastCall = mockUseAuditLogs.mock.calls.at(-1)
    expect(lastCall?.[1]).toMatchObject({ enabled: false })
  })
})

describe('AdminAuditLogsPage — render states', () => {
  it('renders header with title', () => {
    render(<AdminAuditLogsPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
  })

  it('renders empty state when no logs', () => {
    render(<AdminAuditLogsPage />)
    expect(screen.getByTestId('audit-logs-empty')).toBeInTheDocument()
  })

  it('renders loading spinner while loading', () => {
    mockUseAuditLogs.mockReturnValue({
      items: [],
      pagination: undefined,
      total: 0,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminAuditLogsPage />)
    expect(screen.getByTestId('audit-logs-loading')).toBeInTheDocument()
  })

  it('renders error message when API fails', () => {
    mockUseAuditLogs.mockReturnValue({
      items: [],
      pagination: undefined,
      total: 0,
      isLoading: false,
      error: new Error('500'),
      mutate: jest.fn(),
    })
    render(<AdminAuditLogsPage />)
    expect(screen.getByTestId('audit-logs-error')).toBeInTheDocument()
  })

  it('renders table rows when logs present', () => {
    mockUseAuditLogs.mockReturnValue({
      items: [sampleLog],
      pagination: samplePagination,
      total: 1,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminAuditLogsPage />)
    expect(screen.getByText('curriculum.approved')).toBeInTheDocument()
    expect(screen.getByText('curriculum')).toBeInTheDocument()
    expect(screen.getByText('10.0.0.5')).toBeInTheDocument()
  })
})

describe('AdminAuditLogsPage — filters', () => {
  it('forwards action filter to useAuditLogs after typing', async () => {
    const user = userEvent.setup()
    render(<AdminAuditLogsPage />)

    const input = screen.getByLabelText('filters.action')
    await user.type(input, 'auth.login')

    await waitFor(() => {
      const lastCall = mockUseAuditLogs.mock.calls.at(-1)
      expect(lastCall?.[0]).toMatchObject({ action: 'auth.login' })
    })
  })

  it('resets all filters when reset button clicked', async () => {
    const user = userEvent.setup()
    render(<AdminAuditLogsPage />)

    const input = screen.getByLabelText('filters.action')
    await user.type(input, 'auth.login')

    const reset = screen.getByRole('button', { name: 'filters.reset' })
    await user.click(reset)

    await waitFor(() => {
      const lastCall = mockUseAuditLogs.mock.calls.at(-1)
      expect(lastCall?.[0]?.action).toBeFalsy()
    })
  })
})

describe('AdminAuditLogsPage — pagination', () => {
  it('disables prev button on first page', () => {
    mockUseAuditLogs.mockReturnValue({
      items: [sampleLog],
      pagination: { page: 1, per_page: 50, total: 100, total_pages: 2 },
      total: 100,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminAuditLogsPage />)
    const prev = screen.getByRole('button', { name: 'pagination.prev' })
    expect(prev).toBeDisabled()
  })

  it('advances offset when next clicked', async () => {
    mockUseAuditLogs.mockReturnValue({
      items: [sampleLog],
      pagination: { page: 1, per_page: 50, total: 100, total_pages: 2 },
      total: 100,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    const user = userEvent.setup()
    render(<AdminAuditLogsPage />)

    const next = screen.getByRole('button', { name: 'pagination.next' })
    await user.click(next)

    await waitFor(() => {
      const lastCall = mockUseAuditLogs.mock.calls.at(-1)
      expect(lastCall?.[0]?.offset).toBe(50)
    })
  })
})
