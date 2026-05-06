import { render, screen } from '@/test-utils'

const mockReplace = jest.fn()
let mockParamsValue: { id?: string } = { id: '10' }
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn(), back: jest.fn() }),
  useParams: () => mockParamsValue,
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseMyAssignment = jest.fn()
jest.mock('@/hooks/useMyAssignments', () => ({
  useMyAssignment: (id: number | null, opts?: { enabled?: boolean }) =>
    mockUseMyAssignment(id, opts),
  useMyAssignments: jest.fn(),
}))

import MyAssignmentDetailPage from '../page'

const studentAuth = {
  user: { id: 7, role: 'student' as const },
  isAuthenticated: true,
  isLoading: false,
}

const baseView = (overrides: Record<string, unknown> = {}) => ({
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
  mockParamsValue = { id: '10' }
  mockUseAuthCheck.mockReturnValue(studentAuth)
  mockUseMyAssignment.mockReturnValue({ view: undefined, isLoading: false, error: undefined })
})

describe('MyAssignmentDetailPage', () => {
  it('redirects non-student to /forbidden', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MyAssignmentDetailPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('passes a parsed numeric id to the hook with enabled:true for student', () => {
    render(<MyAssignmentDetailPage />)
    expect(mockUseMyAssignment).toHaveBeenCalledWith(10, { enabled: true })
  })

  it('passes null to the hook for invalid path id', () => {
    mockParamsValue = { id: 'abc' }
    render(<MyAssignmentDetailPage />)
    expect(mockUseMyAssignment).toHaveBeenCalledWith(null, { enabled: true })
  })

  it('passes null to the hook for fractional path id (no useless 4xx)', () => {
    mockParamsValue = { id: '1.5' }
    render(<MyAssignmentDetailPage />)
    expect(mockUseMyAssignment).toHaveBeenCalledWith(null, { enabled: true })
  })

  it('passes enabled:false when caller is not a student', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 1, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MyAssignmentDetailPage />)
    // Hook still called per rules of hooks, but second arg carries
    // enabled:false so the SWR key short-circuits to null.
    expect(mockUseMyAssignment).toHaveBeenCalledWith(10, { enabled: false })
  })

  it('renders error block when the hook reports an error', () => {
    mockUseMyAssignment.mockReturnValue({
      view: undefined,
      isLoading: false,
      error: new Error('404'),
    })
    render(<MyAssignmentDetailPage />)
    expect(screen.getByText('detail.loadFailed')).toBeInTheDocument()
  })

  it('renders title, description and assignment metadata for a pending submission', () => {
    mockUseMyAssignment.mockReturnValue({
      view: baseView(),
      isLoading: false,
      error: undefined,
    })
    render(<MyAssignmentDetailPage />)

    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Solve A')).toBeInTheDocument()
    expect(screen.getByText('Math')).toBeInTheDocument()
    expect(screen.getByText('БСБО-01-22')).toBeInTheDocument()
    // Pending status panel.
    expect(screen.getByText('detail.pendingTitle')).toBeInTheDocument()
  })

  it('renders the grade panel for a graded submission', () => {
    mockUseMyAssignment.mockReturnValue({
      view: baseView({ status: 'graded', grade_value: 85, feedback: 'great work' }),
      isLoading: false,
      error: undefined,
    })
    render(<MyAssignmentDetailPage />)

    expect(screen.getByText('detail.gradedTitle')).toBeInTheDocument()
    expect(screen.getByText('85 / 100')).toBeInTheDocument()
    expect(screen.getByText('great work')).toBeInTheDocument()
  })

  it('renders the return-reason panel for a returned submission', () => {
    mockUseMyAssignment.mockReturnValue({
      view: baseView({
        status: 'returned',
        return_reason: 'please add citations',
        returned_at: '2026-05-04T10:00:00Z',
      }),
      isLoading: false,
      error: undefined,
    })
    render(<MyAssignmentDetailPage />)

    expect(screen.getByText('detail.returnedTitle')).toBeInTheDocument()
    expect(screen.getByText('please add citations')).toBeInTheDocument()
  })

  it('shows the Resubmit button only when status is returned', () => {
    // pending: no button
    mockUseMyAssignment.mockReturnValueOnce({
      view: baseView({ status: 'pending' }),
      isLoading: false,
      error: undefined,
    })
    const { unmount } = render(<MyAssignmentDetailPage />)
    expect(
      screen.queryByRole('button', { name: /resubmitButton/ })
    ).not.toBeInTheDocument()
    unmount()

    // graded: no button
    mockUseMyAssignment.mockReturnValueOnce({
      view: baseView({ status: 'graded', grade_value: 85 }),
      isLoading: false,
      error: undefined,
    })
    const r2 = render(<MyAssignmentDetailPage />)
    expect(
      screen.queryByRole('button', { name: /resubmitButton/ })
    ).not.toBeInTheDocument()
    r2.unmount()

    // returned: button present
    mockUseMyAssignment.mockReturnValueOnce({
      view: baseView({ status: 'returned', return_reason: 'redo' }),
      isLoading: false,
      error: undefined,
    })
    render(<MyAssignmentDetailPage />)
    expect(
      screen.getByRole('button', { name: /resubmitButton/ })
    ).toBeInTheDocument()
  })
})
