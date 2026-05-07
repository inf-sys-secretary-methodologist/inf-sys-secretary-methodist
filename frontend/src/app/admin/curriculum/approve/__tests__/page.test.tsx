import { render, screen, fireEvent } from '@/test-utils'

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

const mockUseCurricula = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  useCurricula: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseCurricula(filter, opts),
  useCurriculum: jest.fn(),
  approveCurriculum: jest.fn(),
  rejectCurriculum: jest.fn(),
  updateCurriculum: jest.fn(),
  submitCurriculum: jest.fn(),
}))

jest.mock('@/components/curriculum/ApproveCurriculumDialog', () => ({
  ApproveCurriculumDialog: ({
    curriculumId,
    open,
    onClose,
    onApproved,
  }: {
    curriculumId: number
    open: boolean
    onClose: () => void
    onApproved?: () => void
  }) => (
    <div data-testid="approve-dialog-stub" data-open={String(open)} data-id={curriculumId}>
      <button type="button" onClick={onClose}>
        approve-close-stub
      </button>
      <button type="button" onClick={() => onApproved?.()}>
        approve-confirmed-stub
      </button>
    </div>
  ),
}))

jest.mock('@/components/curriculum/RejectCurriculumDialog', () => ({
  RejectCurriculumDialog: ({
    curriculumId,
    open,
    onClose,
    onRejected,
  }: {
    curriculumId: number
    open: boolean
    onClose: () => void
    onRejected?: () => void
  }) => (
    <div data-testid="reject-dialog-stub" data-open={String(open)} data-id={curriculumId}>
      <button type="button" onClick={onClose}>
        reject-close-stub
      </button>
      <button type="button" onClick={() => onRejected?.()}>
        reject-confirmed-stub
      </button>
    </div>
  ),
}))

import AdminCurriculumApprovePage from '../page'
import type { Curriculum } from '@/types/curriculum'

const adminAuth = {
  user: { id: 1, role: 'system_admin' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<Curriculum> = {}): Curriculum => ({
  id: 11,
  title: 'ИВТ-2026 / 4 года',
  code: '09.03.04-2026',
  specialty: 'Информатика',
  year: 2026,
  description: 'Описание',
  status: 'pending_approval',
  created_by: 5,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-01T08:00:00Z',
  ...overrides,
})

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue(adminAuth)
  mockUseCurricula.mockReturnValue({
    items: [],
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('AdminCurriculumApprovePage', () => {
  it.each(['methodist', 'academic_secretary', 'teacher', 'student'] as const)(
    'redirects non-admin (%s) → /forbidden',
    (role) => {
      mockUseAuthCheck.mockReturnValue({
        user: { id: 7, role },
        isAuthenticated: true,
        isLoading: false,
      })
      render(<AdminCurriculumApprovePage />)
      expect(mockReplace).toHaveBeenCalledWith('/forbidden')
    }
  )

  it('does not redirect while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<AdminCurriculumApprovePage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('does NOT fetch when role is non-admin (skip 401 round-trip)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: 'methodist' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<AdminCurriculumApprovePage />)
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<AdminCurriculumApprovePage />)
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('passes status=pending_approval filter to useCurricula', () => {
    render(<AdminCurriculumApprovePage />)
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ status: 'pending_approval' })
  })

  it('renders title and description headers', () => {
    render(<AdminCurriculumApprovePage />)
    expect(screen.getByText('adminApprove.title')).toBeInTheDocument()
    expect(screen.getByText('adminApprove.description')).toBeInTheDocument()
  })

  it('renders empty state when items=[]', () => {
    render(<AdminCurriculumApprovePage />)
    expect(screen.getByText('adminApprove.empty.title')).toBeInTheDocument()
  })

  it('renders rows with metadata + Approve + Reject buttons for items', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11, title: 'Lab 1' }), sample({ id: 12, title: 'Lab 2' })],
      total: 2,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminCurriculumApprovePage />)

    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Lab 2')).toBeInTheDocument()
    expect(
      screen.getAllByRole('button', { name: 'adminApprove.actions.approve' })
    ).toHaveLength(2)
    expect(
      screen.getAllByRole('button', { name: 'adminApprove.actions.reject' })
    ).toHaveLength(2)
  })

  it('clicking Approve opens ApproveCurriculumDialog для конкретного curriculum', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminCurriculumApprovePage />)

    fireEvent.click(screen.getByRole('button', { name: 'adminApprove.actions.approve' }))
    const dialog = screen.getByTestId('approve-dialog-stub')
    expect(dialog).toHaveAttribute('data-open', 'true')
    expect(dialog).toHaveAttribute('data-id', '11')
  })

  it('clicking Reject opens RejectCurriculumDialog для конкретного curriculum', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 12 })],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<AdminCurriculumApprovePage />)

    fireEvent.click(screen.getByRole('button', { name: 'adminApprove.actions.reject' }))
    const dialog = screen.getByTestId('reject-dialog-stub')
    expect(dialog).toHaveAttribute('data-open', 'true')
    expect(dialog).toHaveAttribute('data-id', '12')
  })

  it('Approve dialog onApproved fires mutate', () => {
    const mutate = jest.fn()
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate,
    })
    render(<AdminCurriculumApprovePage />)

    // Open dialog first.
    fireEvent.click(screen.getByRole('button', { name: 'adminApprove.actions.approve' }))
    fireEvent.click(screen.getByRole('button', { name: 'approve-confirmed-stub' }))
    expect(mutate).toHaveBeenCalled()
  })

  it('Reject dialog onRejected fires mutate', () => {
    const mutate = jest.fn()
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate,
    })
    render(<AdminCurriculumApprovePage />)

    fireEvent.click(screen.getByRole('button', { name: 'adminApprove.actions.reject' }))
    fireEvent.click(screen.getByRole('button', { name: 'reject-confirmed-stub' }))
    expect(mutate).toHaveBeenCalled()
  })

  it('shows loadFailed block when hook surfaces error', () => {
    mockUseCurricula.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<AdminCurriculumApprovePage />)
    expect(screen.getByText('adminApprove.loadFailed')).toBeInTheDocument()
  })
})
