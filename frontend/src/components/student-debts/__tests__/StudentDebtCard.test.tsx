import { render, screen } from '@/test-utils'
import { StudentDebtCard } from '../StudentDebtCard'
import type { StudentDebtListItem } from '@/types/studentDebts'

const sample = (overrides: Partial<StudentDebtListItem> = {}): StudentDebtListItem => ({
  id: 7,
  student_full_name: 'Иванов Иван Иванович',
  group_name: 'ИС-21',
  discipline_name: 'Базы данных',
  semester: 4,
  control_form: 'exam',
  status: 'open',
  version: 1,
  ...overrides,
})

describe('StudentDebtCard', () => {
  it('renders the student full name as the card heading', () => {
    render(<StudentDebtCard debt={sample()} />)
    expect(screen.getByText('Иванов Иван Иванович')).toBeInTheDocument()
  })

  it('renders the group and discipline', () => {
    render(
      <StudentDebtCard debt={sample({ group_name: 'ИС-21', discipline_name: 'Базы данных' })} />
    )
    expect(screen.getByText('ИС-21')).toBeInTheDocument()
    expect(screen.getByText('Базы данных')).toBeInTheDocument()
  })

  it('links to the debt detail page', () => {
    render(<StudentDebtCard debt={sample({ id: 42 })} />)
    expect(screen.getByRole('link')).toHaveAttribute('href', '/student-debts/42')
  })

  it('shows the status pill via the mapped camelCase key', () => {
    render(<StudentDebtCard debt={sample({ status: 'resit_scheduled' })} />)
    expect(screen.getByText('card.status.resitScheduled')).toBeInTheDocument()
  })

  it('shows a distinct status label for closed_passed', () => {
    render(<StudentDebtCard debt={sample({ status: 'closed_passed' })} />)
    expect(screen.getByText('card.status.closedPassed')).toBeInTheDocument()
  })
})
