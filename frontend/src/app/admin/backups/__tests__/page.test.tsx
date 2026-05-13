import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminBackupsPage from '../page'
import type { BackupFile, BackupMetricsResponse } from '@/types/backup'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockGetStoredToken = jest.fn()
jest.mock('@/lib/auth/token', () => ({
  getStoredToken: () => mockGetStoredToken(),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseBackups = jest.fn()
jest.mock('@/hooks/useBackups', () => ({
  useBackups: (opts?: unknown) => mockUseBackups(opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const sampleFile: BackupFile = {
  name: 'postgres_20260510_020000.sql.gz.age',
  type: 'postgres',
  size: 1048576,
  modified_at: 1705708800,
  encryption: 'age',
}

const sampleMetrics: BackupMetricsResponse = {
  postgres: {
    last_run_at: 1705708800,
    last_success_at: 1705708800,
    last_run_success: true,
    duration_seconds: 120,
    size_bytes: 1048576,
    age_seconds: 3600,
    total_count: 100,
    success_count: 99,
    failure_count: 1,
  },
  minio: null,
  remote_sync: null,
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseBackups.mockReturnValue({
    files: [],
    metrics: null,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
  mockGetStoredToken.mockReturnValue('test-jwt-token')
})

describe('AdminBackupsPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminBackupsPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders for system_admin', () => {
    render(<AdminBackupsPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-backups-page')).toBeInTheDocument()
  })
})

describe('AdminBackupsPage — content', () => {
  it('renders the metrics tile for postgres when present', () => {
    mockUseBackups.mockReturnValue({
      files: [sampleFile],
      metrics: sampleMetrics,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminBackupsPage />)
    expect(screen.getByTestId('backup-metrics-postgres')).toBeInTheDocument()
  })

  it('renders an empty state when there are no files', () => {
    render(<AdminBackupsPage />)
    expect(screen.getByTestId('backups-empty')).toBeInTheDocument()
  })

  it('renders the file row with encryption badge', () => {
    mockUseBackups.mockReturnValue({
      files: [sampleFile],
      metrics: sampleMetrics,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminBackupsPage />)
    expect(screen.getByText(sampleFile.name)).toBeInTheDocument()
    expect(screen.getByTestId(`backup-encryption-${sampleFile.name}`)).toBeInTheDocument()
  })

  it('renders the loading spinner when isLoading=true', () => {
    mockUseBackups.mockReturnValue({
      files: [],
      metrics: null,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminBackupsPage />)
    expect(screen.getByTestId('backups-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseBackups.mockReturnValue({
      files: [],
      metrics: null,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminBackupsPage />)
    expect(screen.getByTestId('backups-error')).toBeInTheDocument()
  })

  it('opens download URL with auth token on click (no plain anchor)', async () => {
    const windowOpenSpy = jest.spyOn(window, 'open').mockImplementation(() => null)
    mockUseBackups.mockReturnValue({
      files: [sampleFile],
      metrics: sampleMetrics,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })

    render(<AdminBackupsPage />)
    const button = screen.getByTestId(`backup-download-${sampleFile.name}`)
    await userEvent.click(button)

    // The audit-gated route requires the JWT to be in the URL
    // because <a download> cannot carry headers; mirror documents
    // page pattern.
    expect(windowOpenSpy).toHaveBeenCalledWith(
      expect.stringContaining(
        `/api/admin/backups/postgres/${encodeURIComponent(sampleFile.name)}/download?token=`
      ),
      '_blank'
    )
    expect(windowOpenSpy).toHaveBeenCalledWith(
      expect.stringContaining(encodeURIComponent('test-jwt-token')),
      '_blank'
    )

    windowOpenSpy.mockRestore()
  })
})
