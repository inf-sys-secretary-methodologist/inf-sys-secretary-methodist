import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { SubmissionRow } from '../SubmissionRow'
import type { SubmissionView } from '@/types/assignments'

jest.mock('@/hooks/useAssignments', () => ({
  saveGrade: jest.fn(),
}))

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const base: SubmissionView = {
  id: 1,
  assignment_id: 10,
  student_id: 7,
  student_name: 'Иван Петров',
  status: 'pending',
  created_at: '2026-05-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
}

describe('SubmissionRow', () => {
  it('renders student name and id label', () => {
    // jest.setup.ts mocks useTranslations to return the key verbatim
    // (no interpolation), so we assert on the key + the dynamic name.
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={base} />)
    expect(screen.getByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('submissionRow.studentIdLabel')).toBeInTheDocument()
  })

  it('shows the pending status label', () => {
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={base} />)
    expect(screen.getByText(/status\.pending/)).toBeInTheDocument()
  })

  it('shows graded score and feedback for graded submissions', () => {
    const graded: SubmissionView = {
      ...base,
      id: 2,
      status: 'graded',
      grade_value: 85,
      feedback: 'great',
      graded_at: '2026-05-02T10:00:00Z',
    }
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={graded} />)
    expect(screen.getByText(/85 \/ 100/)).toBeInTheDocument()
    expect(screen.getByText(/great/)).toBeInTheDocument()
    expect(screen.getByText(/status\.graded/)).toBeInTheDocument()
  })

  it('shows the returned status label', () => {
    const returned: SubmissionView = { ...base, status: 'returned' }
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={returned} />)
    expect(screen.getByText(/status\.returned/)).toBeInTheDocument()
  })
})

describe('SubmissionRow Return button integration', () => {
  it('renders Return button when status is pending', () => {
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={base} />)
    expect(
      screen.getByRole('button', { name: /returnButton/ })
    ).toBeInTheDocument()
  })

  it('renders Return button when status is graded', () => {
    const graded: SubmissionView = {
      ...base,
      id: 2,
      status: 'graded',
      grade_value: 85,
      feedback: 'great',
      graded_at: '2026-05-02T10:00:00Z',
    }
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={graded} />)
    expect(
      screen.getByRole('button', { name: /returnButton/ })
    ).toBeInTheDocument()
  })

  it('does NOT render Return button when status is returned', () => {
    const returned: SubmissionView = { ...base, status: 'returned' }
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={returned} />)
    expect(
      screen.queryByRole('button', { name: /returnButton/ })
    ).not.toBeInTheDocument()
  })

  it('opens ReturnDialog on click', async () => {
    const user = userEvent.setup()
    render(<SubmissionRow assignmentId={10} maxScore={100} submission={base} />)

    // Dialog title is hidden by default (Radix)
    expect(screen.queryByText(/returnDialog\.title/)).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /returnButton/ }))
    expect(screen.getByText(/returnDialog\.title/)).toBeInTheDocument()
  })
})
