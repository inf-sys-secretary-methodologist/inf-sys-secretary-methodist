import { render, screen } from '@/test-utils'
import { AssignmentCard } from '../AssignmentCard'
import type { Assignment } from '@/types/assignments'

const sample: Assignment = {
  id: 10,
  title: 'Lab 1',
  description: 'doubly-linked list',
  teacher_id: 42,
  group_name: 'ИС-21',
  subject: 'Algorithms',
  max_score: 100,
  due_date: '2026-05-15T00:00:00Z',
  created_at: '2026-05-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
}

describe('AssignmentCard', () => {
  it('renders title, subject, group_name and links to the detail page', () => {
    render(<AssignmentCard assignment={sample} />)

    expect(screen.getByText('Lab 1')).toBeInTheDocument()
    expect(screen.getByText('Algorithms')).toBeInTheDocument()
    expect(screen.getByText('ИС-21')).toBeInTheDocument()

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/assignments/10')
  })

  it('renders description when present', () => {
    render(<AssignmentCard assignment={sample} />)
    expect(screen.getByText('doubly-linked list')).toBeInTheDocument()
  })

  it('renders the max score chip via translated key', () => {
    // jest.setup.ts mocks useTranslations to return the key verbatim
    // (no interpolation), so we assert on the key, not the value.
    render(<AssignmentCard assignment={sample} />)
    expect(screen.getByText('card.maxScoreLabel')).toBeInTheDocument()
  })

  it('does not render a due-date label when due_date is missing', () => {
    const { container } = render(<AssignmentCard assignment={{ ...sample, due_date: undefined }} />)
    expect(container.textContent).not.toMatch(/2026/)
  })

  it('parses due_date as local midnight (no UTC day shift)', () => {
    // CLAUDE.md rule #9: a 2026-12-31T00:00:00Z due-date must render
    // as 31 Dec, not 30 Dec, regardless of browser timezone.
    render(
      <AssignmentCard assignment={{ ...sample, due_date: '2026-12-31T00:00:00Z' }} />
    )
    expect(screen.getByText(/31/)).toBeInTheDocument()
  })
})
