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

const mockUseCurricula = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  useCurricula: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseCurricula(filter, opts),
  useCurriculum: jest.fn(),
}))

import CurriculumPage from '../page'
import type { Curriculum } from '@/types/curriculum'

const methodistAuth = {
  user: { id: 5, role: 'methodist' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<Curriculum> = {}): Curriculum => ({
  id: 11,
  title: 'ИВТ-2026 / 4 года',
  code: '09.03.04-2026',
  specialty: 'Информатика и вычислительная техника',
  year: 2026,
  description: 'Учебный план',
  status: 'draft',
  created_by: 5,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-01T08:00:00Z',
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockUseAuthCheck.mockReturnValue(methodistAuth)
  mockUseCurricula.mockReturnValue({
    items: [],
    total: 0,
    isLoading: false,
    error: undefined,
  })
})

describe('CurriculumPage', () => {
  it('redirects student → /forbidden', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<CurriculumPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('does not redirect while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: true,
    })
    render(<CurriculumPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('renders page title and empty state when no items', () => {
    render(<CurriculumPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
    expect(screen.getByText('empty.title')).toBeInTheDocument()
  })

  it('renders cards for items', () => {
    mockUseCurricula.mockReturnValue({
      items: [
        sample({ id: 11, title: 'Lab 1' }),
        sample({ id: 12, title: 'Lab 2', status: 'approved' }),
      ],
      total: 2,
      isLoading: false,
      error: undefined,
    })
    render(<CurriculumPage />)

    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Lab 2')).toBeInTheDocument()
  })

  it('forwards specialty filter to hook on input change', () => {
    render(<CurriculumPage />)

    fireEvent.change(screen.getByLabelText('filters.specialty'), {
      target: { value: 'Информатика' },
    })

    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ specialty: 'Информатика' })
  })

  it('forwards year filter to hook on numeric input', () => {
    render(<CurriculumPage />)

    fireEvent.change(screen.getByLabelText('filters.year'), {
      target: { value: '2026' },
    })

    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ year: 2026 })
  })

  it('forwards status filter to hook on select change', () => {
    render(<CurriculumPage />)

    fireEvent.change(screen.getByLabelText('filters.status'), {
      target: { value: 'pending_approval' },
    })

    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ status: 'pending_approval' })
  })

  it('does NOT fetch when role is student (skip 401 round-trip)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<CurriculumPage />)
    // Hook is still called (rules of hooks), but enabled must be false.
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: true,
    })
    render(<CurriculumPage />)
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('shows error block when hook surfaces an error', () => {
    mockUseCurricula.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
    })
    render(<CurriculumPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it('hides countLabel when items list is empty', () => {
    // The countLabel ("Показано N из M") is only useful when there
    // are items to count — rendering it for an empty list reads as
    // "Showing 0 of 0" which is noisy. Pin the conditional in the
    // page so a future refactor doesn't regress it.
    render(<CurriculumPage />)
    expect(screen.queryByText('countLabel')).not.toBeInTheDocument()
  })

  it.each([
    ['methodist', true],
    ['system_admin', true],
    ['academic_secretary', false],
    ['teacher', false],
  ])('Create button visibility for role %s = %s', (role, visible) => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: role as 'methodist' | 'system_admin' | 'academic_secretary' | 'teacher' },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<CurriculumPage />)
    const btn = screen.queryByRole('button', { name: 'createButton' })
    if (visible) {
      expect(btn).toBeInTheDocument()
    } else {
      expect(btn).not.toBeInTheDocument()
    }
  })

  it('opens CreateCurriculumDialog when Create button is clicked (methodist)', () => {
    render(<CurriculumPage />)
    // The dialog is closed initially — title not rendered.
    expect(screen.queryByText('createDialog.title')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'createButton' }))
    // Once open, the dialog title appears.
    expect(screen.getByText('createDialog.title')).toBeInTheDocument()
  })

  it('starts pagination at offset=0 with default limit', () => {
    render(<CurriculumPage />)
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 0 })
  })

  it('renders Next button enabled when total > limit', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 50,
      isLoading: false,
      error: undefined,
    })
    render(<CurriculumPage />)
    expect(screen.getByRole('button', { name: 'pagination.next' })).not.toBeDisabled()
  })

  it('disables Next button when on last page', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<CurriculumPage />)
    expect(screen.getByRole('button', { name: 'pagination.next' })).toBeDisabled()
  })

  it('disables Prev button on first page', () => {
    render(<CurriculumPage />)
    expect(screen.getByRole('button', { name: 'pagination.prev' })).toBeDisabled()
  })

  it('advances offset by limit on Next click', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 100,
      isLoading: false,
      error: undefined,
    })
    render(<CurriculumPage />)
    fireEvent.click(screen.getByRole('button', { name: 'pagination.next' }))
    const lastCall = mockUseCurricula.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 20 })
  })

  it('shows countLabel when items list is non-empty', () => {
    mockUseCurricula.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<CurriculumPage />)
    expect(screen.getByText('countLabel')).toBeInTheDocument()
  })
})
