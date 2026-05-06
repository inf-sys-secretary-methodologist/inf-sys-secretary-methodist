import { render, screen, fireEvent, waitFor } from '@/test-utils'

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
const mockSubmitCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  useCurricula: jest.fn(),
  useCurriculum: (id: number | null, opts?: { enabled?: boolean }) =>
    mockUseCurriculum(id, opts),
  updateCurriculum: jest.fn(),
  submitCurriculum: (...args: unknown[]) => mockSubmitCurriculum(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

// EditCurriculumDialog is heavy (Radix dialog). Stub it so the page
// tests don't have to render the full form — they assert that the
// page wires the props (curriculum / open / onClose / onSaved).
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
    <div data-testid="edit-dialog-stub">
      <span>open={String(open)}</span>
      <button type="button" onClick={onClose}>close-stub</button>
      <button type="button" onClick={() => onSaved?.()}>saved-stub</button>
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
      expect(
        screen.queryByRole('button', { name: 'detail.actions.edit' })
      ).not.toBeInTheDocument()
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

  it('clicking Submit calls submitCurriculum + mutate + toast.success', async () => {
    const mutate = jest.fn()
    mockUseCurriculum.mockReturnValue({
      curriculum: sample({ status: 'draft' }),
      isLoading: false,
      error: undefined,
      mutate,
    })
    mockSubmitCurriculum.mockResolvedValueOnce(sample({ status: 'pending_approval' }))

    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.submit' }))

    await waitFor(() => expect(mockSubmitCurriculum).toHaveBeenCalledWith(11))
    await waitFor(() => expect(mutate).toHaveBeenCalled())
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [422, 'submitToast.errors.notDraft'],
    [403, 'submitToast.errors.forbidden'],
    [500, 'submitToast.errors.generic'],
  ])('clicking Submit maps HTTP %i error to toast key', async (status, expectedKey) => {
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status },
    })
    mockSubmitCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(<CurriculumDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.submit' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
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
