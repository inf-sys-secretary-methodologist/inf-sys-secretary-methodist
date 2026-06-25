import { render, screen } from '@/test-utils'
import { fireEvent } from '@testing-library/react'

jest.mock('next/navigation', () => ({
  useParams: () => ({ id: '9' }),
  useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('sonner', () => ({ toast: { success: jest.fn(), error: jest.fn() } }))

const mockUseStudentDebt = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  useStudentDebt: (id: number | null, opts?: { enabled?: boolean }) => mockUseStudentDebt(id, opts),
  scheduleResit: jest.fn(),
  recordResitResult: jest.fn(),
  pickStudentDebtErrorKey: () => 'generic',
}))

import StudentDebtDetailPage from '../page'
import type { StudentDebt, ResitAttempt } from '@/types/studentDebts'

const attempt = (overrides: Partial<ResitAttempt> = {}): ResitAttempt => ({
  id: 1,
  attempt_no: 1,
  is_commission: false,
  scheduled_date: '2026-07-01T00:00:00Z',
  examiner: 'Петров П.П.',
  result: 'pending',
  ...overrides,
})

const debt = (overrides: Partial<StudentDebt> = {}): StudentDebt => ({
  id: 9,
  student_full_name: 'Иванов Иван',
  group_name: 'ИС-21',
  discipline_name: 'Базы данных',
  semester: 4,
  control_form: 'exam',
  status: 'open',
  version: 1,
  created_at: '2026-06-01T00:00:00Z',
  updated_at: '2026-06-01T00:00:00Z',
  attempts: [],
  ...overrides,
})

const staffAuth = {
  user: { id: 1, role: 'academic_secretary' as const },
  isAuthenticated: true,
  isLoading: false,
}

beforeEach(() => {
  mockUseAuthCheck.mockReturnValue(staffAuth)
  mockUseStudentDebt.mockReturnValue({
    debt: debt(),
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('StudentDebtDetailPage', () => {
  it('renders the student name and discipline', () => {
    render(<StudentDebtDetailPage />)
    expect(screen.getByText('Иванов Иван')).toBeInTheDocument()
    expect(screen.getByText('Базы данных')).toBeInTheDocument()
  })

  it('parses the route id and fetches the debt', () => {
    render(<StudentDebtDetailPage />)
    const lastCall = mockUseStudentDebt.mock.calls.at(-1)
    expect(lastCall?.[0]).toBe(9)
    expect(lastCall?.[1]).toEqual({ enabled: true })
  })

  it('shows the empty-attempts message when there are no resits', () => {
    render(<StudentDebtDetailPage />)
    expect(screen.getByText('detail.attempts.empty')).toBeInTheDocument()
  })

  it('renders attempts in the timeline', () => {
    mockUseStudentDebt.mockReturnValue({
      debt: debt({ status: 'resit_scheduled', attempts: [attempt({ examiner: 'Сидоров С.С.' })] }),
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<StudentDebtDetailPage />)
    // The examiner value shares a label-prefixed line, so match a substring.
    expect(screen.getByText(/Сидоров С\.С\./)).toBeInTheDocument()
  })

  // Schedule resit is available from open / commission for a manager.
  it('shows Schedule resit for a manager when status is open', () => {
    render(<StudentDebtDetailPage />)
    expect(screen.getByRole('button', { name: 'detail.actions.scheduleResit' })).toBeInTheDocument()
  })

  it('hides Schedule resit when status is resit_scheduled', () => {
    mockUseStudentDebt.mockReturnValue({
      debt: debt({ status: 'resit_scheduled', attempts: [attempt()] }),
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<StudentDebtDetailPage />)
    expect(
      screen.queryByRole('button', { name: 'detail.actions.scheduleResit' })
    ).not.toBeInTheDocument()
  })

  // Record result is available only from resit_scheduled for a manager.
  it('shows Record result when status is resit_scheduled', () => {
    mockUseStudentDebt.mockReturnValue({
      debt: debt({ status: 'resit_scheduled', attempts: [attempt()] }),
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<StudentDebtDetailPage />)
    expect(screen.getByRole('button', { name: 'detail.actions.recordResult' })).toBeInTheDocument()
  })

  it('hides all actions for a teacher (read-only)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<StudentDebtDetailPage />)
    expect(
      screen.queryByRole('button', { name: 'detail.actions.scheduleResit' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'detail.actions.recordResult' })
    ).not.toBeInTheDocument()
  })

  it('opens the schedule dialog on Schedule resit click', () => {
    render(<StudentDebtDetailPage />)
    expect(screen.queryByText('scheduleDialog.title')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.scheduleResit' }))
    expect(screen.getByText('scheduleDialog.title')).toBeInTheDocument()
  })

  it('shows the not-found block when the debt is missing', () => {
    mockUseStudentDebt.mockReturnValue({
      debt: undefined,
      isLoading: false,
      error: { response: { status: 404 } },
      mutate: jest.fn(),
    })
    render(<StudentDebtDetailPage />)
    expect(screen.getByText('detail.notFound')).toBeInTheDocument()
  })
})
