import { render, screen } from '@/test-utils'
import { fireEvent } from '@testing-library/react'

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const mockUseMyStudentDebts = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  useMyStudentDebts: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseMyStudentDebts(filter, opts),
}))

import { MyStudentDebtsWidget } from '../MyStudentDebtsWidget'
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
  mockPush.mockClear()
  mockUseMyStudentDebts.mockReturnValue({ items: [], total: 0, isLoading: false, error: undefined })
})

describe('MyStudentDebtsWidget', () => {
  it('renders the widget title', () => {
    render(<MyStudentDebtsWidget />)
    expect(screen.getByText('widget.title')).toBeInTheDocument()
  })

  it('shows the empty message when there are no debts', () => {
    render(<MyStudentDebtsWidget />)
    expect(screen.getByText('widget.empty')).toBeInTheDocument()
  })

  it('lists active debts and navigates to detail on click', () => {
    mockUseMyStudentDebts.mockReturnValue({
      items: [sample({ id: 11, discipline_name: 'Физика' })],
      total: 1,
      isLoading: false,
      error: undefined,
    })
    render(<MyStudentDebtsWidget />)
    const row = screen.getByText('Физика')
    fireEvent.click(row)
    expect(mockPush).toHaveBeenCalledWith('/student-debts/11')
  })
})
