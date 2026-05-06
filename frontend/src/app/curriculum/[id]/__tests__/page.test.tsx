import { render, screen, fireEvent } from '@/test-utils'

const mockReplace = jest.fn()
const mockUseParams = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => mockUseParams(),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  useCurricula: jest.fn(),
  useCurriculum: (id: number | null, opts?: { enabled?: boolean }) => mockUseCurriculum(id, opts),
  updateCurriculum: jest.fn(),
  submitCurriculum: jest.fn(),
}))

// Dialog children are stubbed — page tests assert on which dialog
// mounts open=true and that the wired callbacks fire mutate(). Each
// dialog has its own dedicated test suite covering form / submit /
// error semantics.
jest.mock('@/components/curriculum/EditCurriculumDialog', () => ({
  EditCurriculumDialog: ({
    open,
    onClose,
    onSaved,
  }: {
    open: boolean
    onClose: () => void
    onSaved?: () => void
  }) => (
    <div data-testid="edit-dialog-stub" data-open={String(open)}>
      <button type="button" onClick={onClose}>
        edit-close-stub
      </button>
      <button type="button" onClick={() => onSaved?.()}>
        edit-saved-stub
      </button>
    </div>
  ),
}))

jest.mock('@/components/curriculum/SubmitCurriculumDialog', () => ({
  SubmitCurriculumDialog: ({
    open,
    onClose,
    onSubmitted,
  }: {
    open: boolean
    onClose: () => void
    onSubmitted?: () => void
  }) => (
    <div data-testid="submit-dialog-stub" data-open={String(open)}>
      <button type="button" onClick={onClose}>
        submit-close-stub
      </button>
      <button type="button" onClick={() => onSubmitted?.()}>
        submit-confirmed-stub
      </button>
    </div>
  ),
}))

import CurriculumDetailPage from '../page'
import type { Curriculum, CurriculumStatus } from '@/types/curriculum'

const methodistAuth = {
  user: { id: 5, role: 'methodist' as const },
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
  status: 'draft',
  created_by: 5,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-01T08:00:00Z',
  ...overrides,
})

beforeEach(() => {
  jest.clearAllMocks()
  mockUseAuthCheck.mockReturnValue(methodistAuth)
  mockUseParams.mockReturnValue({ id: '11' })
  mockUseCurriculum.mockReturnValue({
    curriculum: sample(),
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('CurriculumDetailPage', () => {
  it('redirects student → /forbidden', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<CurriculumDetailPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('does not redirect while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<CurriculumDetailPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('renders curriculum metadata when loaded', () => {
    render(<CurriculumDetailPage />)
    expect(screen.getByText('ИВТ-2026 / 4 года')).toBeInTheDocument()
    expect(screen.getByText('09.03.04-2026')).toBeInTheDocument()
    expect(screen.getByText('Информатика')).toBeInTheDocument()
    expect(screen.getByText('2026')).toBeInTheDocument()
    expect(screen.getByText('Описание')).toBeInTheDocument()
  })

  it('renders the status pill via card.status.draft key', () => {
    render(<CurriculumDetailPage />)
    expect(screen.getByText('card.status.draft')).toBeInTheDocument()
  })

  it('shows loadFailed block when hook surfaces error', () => {
    mockUseCurriculum.mockReturnValue({
      curriculum: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<CurriculumDetailPage />)
    expect(screen.getByText('detail.loadFailed')).toBeInTheDocument()
  })

  it('shows notFound block when path id is invalid', () => {
    mockUseParams.mockReturnValue({ id: 'abc' })
    render(<CurriculumDetailPage />)
    expect(screen.getByText('detail.notFound')).toBeInTheDocument()
    // Hook must NOT fetch — invalid id short-circuits.
    const lastCall = mockUseCurriculum.mock.calls.at(-1)
    expect(lastCall?.[0]).toBeNull()
  })

  it.each<CurriculumStatus>(['pending_approval', 'approved', 'archived'])(
    'hides Edit button when status=%s',
    (status) => {
      mockUseCurriculum.mockReturnValue({
        curriculum: sample({ status }),
        isLoading: false,
        error: undefined,
        mutate: jest.fn(),
      })
      render(<CurriculumDetailPage />)
      expect(screen.queryByRole('button', { name: 'detail.actions.edit' })).not.toBeInTheDocument()
    }
  )

  it.each<CurriculumStatus>(['pending_approval', 'approved', 'archived'])(
    'hides Submit button when status=%s',
    (status) => {
      mockUseCurriculum.mockReturnValue({
        curriculum: sample({ status }),
        isLoading: false,
        error: undefined,
        mutate: jest.fn(),
      })
      render(<CurriculumDetailPage />)
      expect(
        screen.queryByRole('button', { name: 'detail.actions.submit' })
      ).not.toBeInTheDocument()
    }
  )

  it('shows Edit + Submit buttons when status=draft', () => {
    render(<CurriculumDetailPage />)
    expect(screen.getByRole('button', { name: 'detail.actions.edit' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'detail.actions.submit' })).toBeInTheDocument()
  })

  it.each<[CurriculumStatus, string]>([
    ['pending_approval', 'detail.statusHint.pending'],
    ['approved', 'detail.statusHint.approved'],
    ['archived', 'detail.statusHint.archived'],
  ])('renders status hint for status=%s', (status, expectedKey) => {
    mockUseCurriculum.mockReturnValue({
      curriculum: sample({ status }),
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<CurriculumDetailPage />)
    expect(screen.getByText(expectedKey)).toBeInTheDocument()
  })

  it('does NOT render status hint for status=draft', () => {
    render(<CurriculumDetailPage />)
    expect(screen.queryByText('detail.statusHint.pending')).not.toBeInTheDocument()
    expect(screen.queryByText('detail.statusHint.approved')).not.toBeInTheDocument()
    expect(screen.queryByText('detail.statusHint.archived')).not.toBeInTheDocument()
  })

  it('Submit dialog mounts open=false initially', () => {
    render(<CurriculumDetailPage />)
    expect(screen.getByTestId('submit-dialog-stub')).toHaveAttribute('data-open', 'false')
  })

  it('clicking Submit button opens the SubmitCurriculumDialog', () => {
    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.submit' }))
    expect(screen.getByTestId('submit-dialog-stub')).toHaveAttribute('data-open', 'true')
  })

  it('SubmitCurriculumDialog onSubmitted fires mutate', () => {
    const mutate = jest.fn()
    mockUseCurriculum.mockReturnValue({
      curriculum: sample({ status: 'draft' }),
      isLoading: false,
      error: undefined,
      mutate,
    })
    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'submit-confirmed-stub' }))
    expect(mutate).toHaveBeenCalled()
  })

  it('EditCurriculumDialog onSaved fires mutate', () => {
    const mutate = jest.fn()
    mockUseCurriculum.mockReturnValue({
      curriculum: sample({ status: 'draft' }),
      isLoading: false,
      error: undefined,
      mutate,
    })
    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'edit-saved-stub' }))
    expect(mutate).toHaveBeenCalled()
  })

  it('clicking Edit button opens the EditCurriculumDialog', () => {
    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.edit' }))
    expect(screen.getByTestId('edit-dialog-stub')).toHaveAttribute('data-open', 'true')
  })

  it('passes enabled=false to useCurriculum when role is student', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<CurriculumDetailPage />)
    const lastCall = mockUseCurriculum.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('passes enabled=false to useCurriculum while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<CurriculumDetailPage />)
    const lastCall = mockUseCurriculum.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('renders backToList link to /curriculum', () => {
    render(<CurriculumDetailPage />)
    const link = screen.getByText('detail.backToList').closest('a')
    expect(link).toHaveAttribute('href', '/curriculum')
  })
})
