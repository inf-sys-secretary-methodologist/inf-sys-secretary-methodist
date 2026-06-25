import { render, screen } from '@/test-utils'
import { fireEvent } from '@testing-library/react'

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

const mockUseStudentDebts = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  useStudentDebts: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseStudentDebts(filter, opts),
}))

const mockExport = jest.fn()
const mockImport = jest.fn()
jest.mock('@/lib/api/studentDebts', () => ({
  studentDebtsApi: {
    export: (...args: unknown[]) => mockExport(...args),
    import: (...args: unknown[]) => mockImport(...args),
  },
}))

import StudentDebtsPage from '../page'
import type { StudentDebtListItem } from '@/types/studentDebts'

const staffAuth = {
  user: { id: 1, role: 'academic_secretary' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<StudentDebtListItem> = {}): StudentDebtListItem => ({
  id: 11,
  student_full_name: 'Иванов Иван',
  group_name: 'ИС-21',
  discipline_name: 'Базы данных',
  semester: 4,
  control_form: 'exam',
  status: 'open',
  version: 1,
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockPush.mockClear()
  mockExport.mockReset()
  mockImport.mockReset()
  mockUseAuthCheck.mockReturnValue(staffAuth)
  mockUseStudentDebts.mockReturnValue({
    items: [],
    total: 0,
    isLoading: false,
    error: undefined,
  })
})

describe('StudentDebtsPage', () => {
  it('renders title and empty state when no items', () => {
    render(<StudentDebtsPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
    expect(screen.getByText('empty.title')).toBeInTheDocument()
  })

  it('renders cards for items', () => {
    mockUseStudentDebts.mockReturnValue({
      items: [
        sample({ id: 11, student_full_name: 'Студент А' }),
        sample({ id: 12, student_full_name: 'Студент Б' }),
      ],
      total: 2,
      isLoading: false,
      error: undefined,
    })
    render(<StudentDebtsPage />)
    expect(screen.getByText('Студент А')).toBeInTheDocument()
    expect(screen.getByText('Студент Б')).toBeInTheDocument()
  })

  // A student is denied the registry endpoint; the page redirects them to
  // their own-debts view rather than fire a guaranteed-403 fetch.
  it('redirects students to /student-debts/my', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<StudentDebtsPage />)
    expect(mockReplace).toHaveBeenCalledWith('/student-debts/my')
  })

  it('does NOT fetch the registry for a student (enabled=false)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<StudentDebtsPage />)
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('fetches for staff (enabled=true)', () => {
    render(<StudentDebtsPage />)
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: true })
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<StudentDebtsPage />)
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('forwards group_name filter to hook on input change', () => {
    render(<StudentDebtsPage />)
    fireEvent.change(screen.getByLabelText('filters.group'), { target: { value: 'ИС-21' } })
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ group_name: 'ИС-21' })
  })

  it('forwards semester filter to hook on numeric input', () => {
    render(<StudentDebtsPage />)
    fireEvent.change(screen.getByLabelText('filters.semester'), { target: { value: '4' } })
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ semester: 4 })
  })

  it('forwards status filter to hook on select change', () => {
    render(<StudentDebtsPage />)
    fireEvent.change(screen.getByLabelText('filters.status'), {
      target: { value: 'resit_scheduled' },
    })
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ status: 'resit_scheduled' })
  })

  it('shows error block when hook surfaces an error', () => {
    mockUseStudentDebts.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
    })
    render(<StudentDebtsPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it('starts pagination at offset=0 with default limit', () => {
    render(<StudentDebtsPage />)
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 0 })
  })

  it('advances offset by limit on Next click', () => {
    mockUseStudentDebts.mockReturnValue({
      items: [sample()],
      total: 100,
      isLoading: false,
      error: undefined,
    })
    render(<StudentDebtsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'pagination.next' }))
    const lastCall = mockUseStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[0]).toMatchObject({ limit: 20, offset: 20 })
  })

  it('resets offset to 0 when a filter changes', () => {
    mockUseStudentDebts.mockReturnValue({
      items: [sample()],
      total: 100,
      isLoading: false,
      error: undefined,
    })
    render(<StudentDebtsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'pagination.next' }))
    expect(mockUseStudentDebts.mock.calls.at(-1)?.[0]).toMatchObject({ offset: 20 })
    fireEvent.change(screen.getByLabelText('filters.status'), { target: { value: 'open' } })
    expect(mockUseStudentDebts.mock.calls.at(-1)?.[0]).toMatchObject({ status: 'open', offset: 0 })
  })

  // Import/Export are EDIT_ROLES (isDebtManager) affordances: admin /
  // methodist / secretary. A teacher reads the registry but cannot manage it.
  it('shows Import and Export for a manager (academic_secretary)', () => {
    render(<StudentDebtsPage />)
    expect(screen.getByRole('button', { name: 'importButton' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'exportButton' })).toBeInTheDocument()
  })

  it('hides Import and Export for a teacher (read-only)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<StudentDebtsPage />)
    expect(screen.queryByRole('button', { name: 'importButton' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'exportButton' })).not.toBeInTheDocument()
  })

  it('opens the import dialog when Import is clicked', () => {
    render(<StudentDebtsPage />)
    expect(screen.queryByText('importDialog.title')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'importButton' }))
    expect(screen.getByText('importDialog.title')).toBeInTheDocument()
  })

  it('calls the export api when Export is clicked', () => {
    mockExport.mockResolvedValueOnce(new Blob(['x']))
    render(<StudentDebtsPage />)
    fireEvent.click(screen.getByRole('button', { name: 'exportButton' }))
    expect(mockExport).toHaveBeenCalled()
  })
})
