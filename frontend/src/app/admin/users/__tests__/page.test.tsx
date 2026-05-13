import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdminUsersPage from '../page'
import type { User } from '@/types/user'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUseUsers = jest.fn()
jest.mock('@/hooks/useUsers', () => ({
  useUsers: (filter: unknown, opts?: unknown) => mockUseUsers(filter, opts),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

const sampleUsers: User[] = [
  {
    id: 1,
    email: 'admin@example.com',
    name: 'Системный администратор',
    role: 'system_admin',
    status: 'active',
    department_id: null,
    position_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
  },
  {
    id: 2,
    email: 'teacher@example.com',
    name: 'Иванова И.И.',
    role: 'teacher',
    status: 'inactive',
    department_id: 3,
    department_name: 'Кафедра ИТ',
    position_id: 5,
    position_name: 'Доцент',
    created_at: '2026-02-15T00:00:00Z',
    updated_at: '2026-04-10T00:00:00Z',
  },
]

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue({
    user: { id: 1, role: 'system_admin' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseUsers.mockReturnValue({
    users: sampleUsers,
    total: 2,
    page: 1,
    limit: 20,
    totalPages: 1,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminUsersPage — role guard', () => {
  it('redirects non-admin to /forbidden', async () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminUsersPage />)
    await waitFor(() => expect(mockReplace).toHaveBeenCalledWith('/forbidden'))
  })

  it('renders for system_admin', () => {
    render(<AdminUsersPage />)
    expect(mockReplace).not.toHaveBeenCalled()
    expect(screen.getByTestId('admin-users-page')).toBeInTheDocument()
  })
})

describe('AdminUsersPage — list content', () => {
  it('renders the users table with rows for each user', () => {
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-table')).toBeInTheDocument()
    expect(screen.getByTestId('user-row-1')).toBeInTheDocument()
    expect(screen.getByTestId('user-row-2')).toBeInTheDocument()
  })

  it('renders the user name and email in each row', () => {
    render(<AdminUsersPage />)
    expect(screen.getByText('Системный администратор')).toBeInTheDocument()
    expect(screen.getByText('admin@example.com')).toBeInTheDocument()
    expect(screen.getByText('Иванова И.И.')).toBeInTheDocument()
    expect(screen.getByText('teacher@example.com')).toBeInTheDocument()
  })

  it('renders the role and status badges', () => {
    render(<AdminUsersPage />)
    expect(screen.getByTestId('user-role-1')).toBeInTheDocument()
    expect(screen.getByTestId('user-status-1')).toBeInTheDocument()
    expect(screen.getByTestId('user-role-2')).toBeInTheDocument()
    expect(screen.getByTestId('user-status-2')).toBeInTheDocument()
  })

  it('renders the department and position when present', () => {
    render(<AdminUsersPage />)
    expect(screen.getByText('Кафедра ИТ')).toBeInTheDocument()
    expect(screen.getByText('Доцент')).toBeInTheDocument()
  })

  it('renders the loading spinner when isLoading=true', () => {
    mockUseUsers.mockReturnValue({
      users: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-loading')).toBeInTheDocument()
  })

  it('renders the error state on fetch failure', () => {
    mockUseUsers.mockReturnValue({
      users: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-error')).toBeInTheDocument()
  })

  it('renders the empty state when there are no users', () => {
    mockUseUsers.mockReturnValue({
      users: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-empty')).toBeInTheDocument()
  })
})

describe('AdminUsersPage — filters', () => {
  it('renders the search and role/status filter inputs', () => {
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-search')).toBeInTheDocument()
    expect(screen.getByTestId('users-role-filter')).toBeInTheDocument()
    expect(screen.getByTestId('users-status-filter')).toBeInTheDocument()
  })

  it('updates the search filter and forwards it to useUsers', async () => {
    render(<AdminUsersPage />)
    const search = screen.getByTestId('users-search') as HTMLInputElement
    await userEvent.type(search, 'ivan')
    await waitFor(() => {
      const lastCall = mockUseUsers.mock.calls.at(-1)
      expect(lastCall?.[0]).toEqual(expect.objectContaining({ search: 'ivan' }))
    })
  })

  it('clears all filters via the reset button', async () => {
    render(<AdminUsersPage />)
    const search = screen.getByTestId('users-search') as HTMLInputElement
    await userEvent.type(search, 'something')
    const reset = screen.getByTestId('users-reset')
    await userEvent.click(reset)
    expect(search.value).toBe('')
  })
})

describe('AdminUsersPage — pagination', () => {
  it('renders prev/next controls with the page indicator', () => {
    mockUseUsers.mockReturnValue({
      users: sampleUsers,
      total: 50,
      page: 2,
      limit: 20,
      totalPages: 3,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminUsersPage />)
    expect(screen.getByTestId('users-pagination-prev')).toBeInTheDocument()
    expect(screen.getByTestId('users-pagination-next')).toBeInTheDocument()
    expect(screen.getByTestId('users-pagination-indicator')).toBeInTheDocument()
  })
})
