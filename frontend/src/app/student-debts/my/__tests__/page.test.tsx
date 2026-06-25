import { render, screen } from '@/test-utils'

jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseMyStudentDebts = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  useMyStudentDebts: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseMyStudentDebts(filter, opts),
}))

import MyStudentDebtsPage from '../page'
import type { StudentDebtListItem } from '@/types/studentDebts'

const sample = (overrides: Partial<StudentDebtListItem> = {}): StudentDebtListItem => ({
  id: 3,
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
  mockUseAuthCheck.mockReturnValue({
    user: { id: 7, role: 'student' as const },
    isAuthenticated: true,
    isLoading: false,
  })
  mockUseMyStudentDebts.mockReturnValue({ items: [], total: 0, isLoading: false, error: undefined })
})

describe('MyStudentDebtsPage', () => {
  it('renders the page title', () => {
    render(<MyStudentDebtsPage />)
    expect(screen.getByText('my.title')).toBeInTheDocument()
  })

  it('shows the empty state when the student has no debts', () => {
    render(<MyStudentDebtsPage />)
    expect(screen.getByText('empty.title')).toBeInTheDocument()
  })

  it('renders a card per own debt', () => {
    mockUseMyStudentDebts.mockReturnValue({
      items: [sample({ id: 1, discipline_name: 'Матанализ' })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<MyStudentDebtsPage />)
    expect(screen.getByText('Матанализ')).toBeInTheDocument()
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<MyStudentDebtsPage />)
    const lastCall = mockUseMyStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('fetches once authenticated', () => {
    render(<MyStudentDebtsPage />)
    const lastCall = mockUseMyStudentDebts.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: true })
  })
})
