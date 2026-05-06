import { render, screen } from '@/test-utils'
import { fireEvent } from '@testing-library/react'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => ({}),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseMyAssignments = jest.fn()
jest.mock('@/hooks/useMyAssignments', () => ({
  useMyAssignments: (status?: string, opts?: { enabled?: boolean }) =>
    mockUseMyAssignments(status, opts),
  useMyAssignment: jest.fn(),
}))

import MyAssignmentsPage from '../page'

const studentAuth = {
  user: { id: 7, role: 'student' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sampleView = (overrides: Record<string, unknown> = {}) => ({
  assignment_id: 10,
  title: 'Lab 1',
  description: 'Solve A',
  subject: 'Math',
  group_name: 'БСБО-01-22',
  max_score: 100,
  due_date: '2026-05-15T00:00:00Z',
  assignment_created_at: '2026-05-01T00:00:00Z',
  assignment_updated_at: '2026-05-01T00:00:00Z',
  submission_id: 1,
  student_id: 7,
  status: 'pending' as const,
  feedback: '',
  return_reason: '',
  submission_created_at: '2026-05-01T00:00:00Z',
  submission_updated_at: '2026-05-01T00:00:00Z',
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockUseAuthCheck.mockReturnValue(studentAuth)
  mockUseMyAssignments.mockReturnValue({ items: [], total: 0, isLoading: false, error: undefined })
})

describe('MyAssignmentsPage', () => {
  it('redirects non-student to /forbidden', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MyAssignmentsPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('does not redirect while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<MyAssignmentsPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('renders page title and empty state when no items', () => {
    render(<MyAssignmentsPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
    expect(screen.getByText('empty.title')).toBeInTheDocument()
  })

  it('renders cards for returned items', () => {
    mockUseMyAssignments.mockReturnValue({
      items: [
        sampleView({ assignment_id: 10, title: 'Lab 1' }),
        sampleView({ assignment_id: 11, title: 'Lab 2', status: 'graded', grade_value: 85 }),
      ],
      total: 2,
      isLoading: false,
      error: undefined,
    })
    render(<MyAssignmentsPage />)

    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Lab 2')).toBeInTheDocument()
  })

  it('renders status filter tabs and forwards selection to the hook', () => {
    render(<MyAssignmentsPage />)
    expect(mockUseMyAssignments).toHaveBeenLastCalledWith(undefined, { enabled: true })

    fireEvent.click(screen.getByRole('tab', { name: /returned/i }))
    expect(mockUseMyAssignments).toHaveBeenLastCalledWith('returned', { enabled: true })

    fireEvent.click(screen.getByRole('tab', { name: /graded/i }))
    expect(mockUseMyAssignments).toHaveBeenLastCalledWith('graded', { enabled: true })

    fireEvent.click(screen.getByRole('tab', { name: /all/i }))
    expect(mockUseMyAssignments).toHaveBeenLastCalledWith(undefined, { enabled: true })
  })

  it('does NOT fetch when the role is not student (skip 401 round-trip)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MyAssignmentsPage />)
    // Hook is still called (rules of hooks), but the second arg
    // must carry enabled: false so the SWR key short-circuits.
    expect(mockUseMyAssignments).toHaveBeenCalledWith(undefined, { enabled: false })
  })

  it('shows error block when the hook surfaces an error', () => {
    mockUseMyAssignments.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
    })
    render(<MyAssignmentsPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })
})
