import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

const mockReplace = jest.fn()
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush }),
  useParams: () => ({}),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockUseWorkPrograms = jest.fn()
const mockCreateWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  useWorkPrograms: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseWorkPrograms(filter, opts),
  useWorkProgram: jest.fn(),
  createWorkProgram: (...args: unknown[]) => mockCreateWorkProgram(...args),
  pickWorkProgramErrorKey: () => 'generic',
}))

import WorkProgramsPage from '../page'
import type { WorkProgramSummary } from '@/types/workProgram'

const teacherAuth = {
  user: { id: 5, role: 'teacher' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<WorkProgramSummary> = {}): WorkProgramSummary => ({
  id: 11,
  discipline_id: 10,
  specialty_code: '09.03.01',
  applicable_from_year: 2026,
  title: 'Базы данных',
  status: 'approved',
  author_id: 5,
  version: 1,
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockPush.mockClear()
  mockCreateWorkProgram.mockReset()
  mockUseAuthCheck.mockReturnValue(teacherAuth)
  mockUseWorkPrograms.mockReturnValue({
    items: [],
    total: 0,
    isLoading: false,
    error: undefined,
  })
})

describe('WorkProgramsPage', () => {
  it('renders page title and empty state when no items', () => {
    render(<WorkProgramsPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
    expect(screen.getByText('empty.title')).toBeInTheDocument()
  })

  it('renders cards for items', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11, title: 'РПД 1' }), sample({ id: 12, title: 'РПД 2' })],
      total: 2,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    expect(screen.getByText('РПД 1')).toBeInTheDocument()
    expect(screen.getByText('РПД 2')).toBeInTheDocument()
  })

  // 273-ФЗ ст. 29: students see approved РПД. Unlike /curriculum, the
  // page must NOT redirect students and MUST fetch (server forces
  // status=approved for them).
  it('does NOT redirect students', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramsPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('fetches for students (enabled=true) — server scopes to approved', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramsPage />)
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: true })
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: true,
    })
    render(<WorkProgramsPage />)
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('forwards specialty_code filter to hook on input change', () => {
    render(<WorkProgramsPage />)
    fireEvent.change(screen.getByLabelText('filters.specialty'), {
      target: { value: '09.03.01' },
    })
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ specialty_code: '09.03.01' })
  })

  it('forwards applicable_from_year filter to hook on numeric input', () => {
    render(<WorkProgramsPage />)
    fireEvent.change(screen.getByLabelText('filters.year'), {
      target: { value: '2026' },
    })
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ applicable_from_year: 2026 })
  })

  it('forwards status filter to hook on select change', () => {
    render(<WorkProgramsPage />)
    fireEvent.change(screen.getByLabelText('filters.status'), {
      target: { value: 'pending_approval' },
    })
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ status: 'pending_approval' })
  })

  it('shows error block when hook surfaces an error', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
    })
    render(<WorkProgramsPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it('hides countLabel when items list is empty', () => {
    render(<WorkProgramsPage />)
    expect(screen.queryByText('countLabel')).not.toBeInTheDocument()
  })

  it('shows countLabel when items list is non-empty', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    expect(screen.getByText('countLabel')).toBeInTheDocument()
  })

  it('starts pagination at offset=0 with default limit', () => {
    render(<WorkProgramsPage />)
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 0 })
  })

  it('disables Prev button on first page', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 50,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    expect(screen.getByRole('button', { name: 'pagination.prev' })).toBeDisabled()
  })

  it('disables Next button when on last page', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    expect(screen.getByRole('button', { name: 'pagination.next' })).toBeDisabled()
  })

  it('advances offset by limit on Next click', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 100,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'pagination.next' }))
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 20 })
  })

  it('resets offset to 0 when a filter changes (avoids out-of-range page)', () => {
    mockUseWorkPrograms.mockReturnValue({
      items: [sample({ id: 11 })],
      total: 100,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'pagination.next' }))
    expect(mockUseWorkPrograms.mock.calls.at(-1)?.[0]).toMatchObject({ offset: 20 })
    fireEvent.change(screen.getByLabelText('filters.status'), {
      target: { value: 'approved' },
    })
    const lastCall = mockUseWorkPrograms.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ status: 'approved', offset: 0 })
  })

  // Create button is gated by canCreateWorkProgram (teacher / methodist /
  // admin per ADR-5). The backend whitelist is the source of truth; this
  // hides the affordance for read-only roles.
  it('shows the Создать button for a create-capable role (teacher)', () => {
    render(<WorkProgramsPage />)
    expect(screen.getByRole('button', { name: 'createButton' })).toBeInTheDocument()
  })

  it('hides the Создать button for students', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramsPage />)
    expect(screen.queryByRole('button', { name: 'createButton' })).not.toBeInTheDocument()
  })

  it('hides the Создать button for academic_secretary (curriculum author, not РПД)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 8, role: 'academic_secretary' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramsPage />)
    expect(screen.queryByRole('button', { name: 'createButton' })).not.toBeInTheDocument()
  })

  it('opens the CreateWorkProgramDialog when Создать is clicked', () => {
    render(<WorkProgramsPage />)
    expect(screen.queryByText('createDialog.title')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'createButton' }))
    expect(screen.getByText('createDialog.title')).toBeInTheDocument()
  })

  it('navigates to the new РПД detail after a successful create', async () => {
    mockCreateWorkProgram.mockResolvedValueOnce({ id: 99 })
    render(<WorkProgramsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'createButton' }))
    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'Базы данных' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.disciplineId'), {
      target: { value: '42' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: '09.03.01' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'createDialog.create' }))
    await waitFor(() => expect(mockPush).toHaveBeenCalledWith('/work-programs/99'))
  })
})
