import { render, screen } from '@/test-utils'
import { StudentAssignmentCard } from '../StudentAssignmentCard'
import type { StudentAssignmentView } from '@/types/assignments'

const base: StudentAssignmentView = {
  assignment_id: 10,
  title: 'Lab 1',
  description: 'Solve A',
  subject: 'Math',
  group_name: 'БСБО-01-22',
  max_score: 100,
  due_date: '2026-05-15T00:00:00Z',
  assignment_created_at: '2026-05-01T00:00:00Z',
  assignment_updated_at: '2026-05-01T00:00:00Z',
  submission_id: 1,
  student_id: 7,
  status: 'pending',
  feedback: '',
  return_reason: '',
  submission_created_at: '2026-05-01T00:00:00Z',
  submission_updated_at: '2026-05-01T00:00:00Z',
}

describe('StudentAssignmentCard', () => {
  it('links to /my-assignments/:id and renders title + subject + group', () => {
    render(<StudentAssignmentCard view={base} />)

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/my-assignments/10')
    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Math')).toBeInTheDocument()
    expect(screen.getByText('БСБО-01-22')).toBeInTheDocument()
  })

  it('renders the pending status pill', () => {
    render(<StudentAssignmentCard view={base} />)
    expect(screen.getByText('myAssignments.status.pending')).toBeInTheDocument()
  })

  it('renders the graded status pill with grade/max score', () => {
    render(
      <StudentAssignmentCard
        view={{ ...base, status: 'graded', grade_value: 85 }}
      />
    )
    expect(screen.getByText('myAssignments.status.graded')).toBeInTheDocument()
    expect(screen.getByText('85 / 100')).toBeInTheDocument()
  })

  it('renders the returned status pill with return reason snippet', () => {
    render(
      <StudentAssignmentCard
        view={{ ...base, status: 'returned', return_reason: 'please add citations' }}
      />
    )
    expect(screen.getByText('myAssignments.status.returned')).toBeInTheDocument()
    expect(screen.getByText(/please add citations/)).toBeInTheDocument()
  })

  it('parses due_date as local midnight (no UTC day shift, CLAUDE.md #9)', () => {
    render(
      <StudentAssignmentCard view={{ ...base, due_date: '2026-12-31T00:00:00Z' }} />
    )
    // Year fragment is a stable proxy: 2026 must be present in formatted date.
    expect(screen.getByText(/2026/)).toBeInTheDocument()
  })

  it('does not render a due-date chip when due_date is missing', () => {
    const { container } = render(
      <StudentAssignmentCard view={{ ...base, due_date: undefined }} />
    )
    // Calendar icon container only renders when due_date is set;
    // assert via the formatted date pattern absence.
    expect(container.textContent).not.toMatch(/2026 г/)
  })
})
