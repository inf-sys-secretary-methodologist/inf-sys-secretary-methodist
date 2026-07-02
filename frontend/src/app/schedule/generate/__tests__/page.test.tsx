import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { UserRole } from '@/types/auth'

jest.mock('@/hooks/useAuth', () => ({ useAuthCheck: jest.fn() }))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const activeSemester = { id: 7, name: 'Осень 2026', number: 1, is_active: true }
jest.mock('@/hooks/useSchedule', () => ({
  useSemesters: () => ({ semesters: [activeSemester], isLoading: false, error: null }),
}))

let mockRole: UserRole = UserRole.METHODIST
jest.mock('@/stores/authStore', () => ({
  useAuthStore: (selector: (s: { user: { role: UserRole } }) => unknown) =>
    selector({ user: { role: mockRole } }),
}))

const toast = { success: jest.fn(), error: jest.fn() }
jest.mock('sonner', () => ({
  toast: { success: (m: string) => toast.success(m), error: (m: string) => toast.error(m) },
}))

const preview = jest.fn()
const apply = jest.fn()
jest.mock('@/lib/api/schedule', () => ({
  scheduleGenerateApi: {
    preview: (...args: unknown[]) => preview(...args),
    apply: (...args: unknown[]) => apply(...args),
  },
}))

const draft = {
  lessons: [
    {
      load_id: 1,
      group_id: 1,
      group_name: 'ИС-21',
      teacher_id: 1,
      teacher_name: 'Иванов И.',
      discipline_id: 1,
      discipline_name: 'Математика',
      lesson_type_id: 1,
      lesson_type_name: 'Лекция',
      week_type: 'all',
      day_of_week: 1,
      slot_number: 1,
      time_start: '09:00',
      time_end: '10:30',
      classroom_id: 1,
      classroom_name: 'А-101',
    },
  ],
  unplaced: [
    {
      load_id: 2,
      group_name: 'ИС-22',
      discipline_name: 'Физика',
      lesson_type_name: 'Практика',
      week_type: 'odd',
    },
  ],
  total_requested: 2,
  placed_count: 1,
  unplaced_count: 1,
}

import GenerateSchedulePage from '../page'

describe('GenerateSchedulePage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockRole = UserRole.METHODIST
  })

  it('renders the title and generate button for a methodist', () => {
    render(<GenerateSchedulePage />)
    expect(screen.getByText('title')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'generate' })).toBeInTheDocument()
  })

  it('denies access to a student', () => {
    mockRole = UserRole.STUDENT
    render(<GenerateSchedulePage />)
    expect(screen.getByText('accessDenied')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'generate' })).not.toBeInTheDocument()
  })

  it('generates a preview and renders placed + unplaced lessons', async () => {
    preview.mockResolvedValue(draft)
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    await user.click(screen.getByRole('button', { name: 'generate' }))

    await waitFor(() => expect(screen.getByText('ИС-21')).toBeInTheDocument())
    expect(preview).toHaveBeenCalledWith({ semester_id: 7 })
    expect(screen.getByText('Математика')).toBeInTheDocument()
    expect(screen.getByText('ИС-22')).toBeInTheDocument()
    expect(screen.getByTestId('summary-placed')).toHaveTextContent('1')
    expect(screen.getByTestId('summary-unplaced')).toHaveTextContent('1')
    expect(screen.getByTestId('summary-total')).toHaveTextContent('2')
  })

  it('applies the draft and toasts the created count', async () => {
    preview.mockResolvedValue(draft)
    apply.mockResolvedValue({ created: 5, unplaced: 1 })
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    await user.click(screen.getByRole('button', { name: 'generate' }))
    await waitFor(() => expect(screen.getByText('ИС-21')).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: 'apply' }))

    await waitFor(() => expect(apply).toHaveBeenCalledWith({ semester_id: 7 }))
    expect(toast.success).toHaveBeenCalled()
  })

  it('shows a conflict message when a schedule already exists (409)', async () => {
    preview.mockResolvedValue(draft)
    apply.mockRejectedValue({ response: { status: 409 } })
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    await user.click(screen.getByRole('button', { name: 'generate' }))
    await waitFor(() => expect(screen.getByText('ИС-21')).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: 'apply' }))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('errors.alreadyExists'))
  })

  it('invalidates a stale preview when generation params change', async () => {
    preview.mockResolvedValue(draft)
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    await user.click(screen.getByRole('button', { name: 'generate' }))
    await waitFor(() => expect(screen.getByText('ИС-21')).toBeInTheDocument())

    // Changing the day selection must drop the now-stale preview so a later
    // apply can never persist params different from what was shown.
    await user.click(screen.getByRole('button', { name: 'days.monday' }))

    await waitFor(() => expect(screen.queryByText('ИС-21')).not.toBeInTheDocument())
    expect(screen.queryByRole('button', { name: 'apply' })).not.toBeInTheDocument()
  })

  it('marks selected day toggles as pressed for assistive tech', () => {
    render(<GenerateSchedulePage />)
    // All six days are selected by default → aria-pressed=true.
    expect(screen.getByRole('button', { name: 'days.monday', pressed: true })).toBeInTheDocument()
  })

  it('toasts a generic error when preview generation fails', async () => {
    preview.mockRejectedValue(new Error('boom'))
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    await user.click(screen.getByRole('button', { name: 'generate' }))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('errors.generateFailed'))
    expect(screen.queryByText('ИС-21')).not.toBeInTheDocument()
  })

  it('includes an explicit days list when the full week is not selected', async () => {
    preview.mockResolvedValue(draft)
    const user = userEvent.setup()
    render(<GenerateSchedulePage />)

    // Drop Monday, then generate: the request must carry the remaining days.
    await user.click(screen.getByRole('button', { name: 'days.monday', pressed: true }))
    await user.click(screen.getByRole('button', { name: 'generate' }))

    await waitFor(() =>
      expect(preview).toHaveBeenCalledWith({ semester_id: 7, days: [2, 3, 4, 5, 6] })
    )
  })
})
