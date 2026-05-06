import { render, screen } from '@/test-utils'
import { CurriculumCard } from '../CurriculumCard'
import type { Curriculum, CurriculumStatus } from '@/types/curriculum'

const sample: Curriculum = {
  id: 11,
  title: 'ИВТ-2026 / 4 года',
  code: '09.03.04-2026',
  specialty: 'Информатика и вычислительная техника',
  year: 2026,
  description: 'Учебный план направления подготовки',
  status: 'draft',
  created_by: 5,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-01T08:00:00Z',
}

describe('CurriculumCard', () => {
  it('renders title, code, specialty, year and links to the detail page', () => {
    render(<CurriculumCard curriculum={sample} />)

    expect(screen.getByText('ИВТ-2026 / 4 года')).toBeInTheDocument()
    expect(screen.getByText('09.03.04-2026')).toBeInTheDocument()
    expect(screen.getByText('Информатика и вычислительная техника')).toBeInTheDocument()
    expect(screen.getByText('2026')).toBeInTheDocument()

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/curriculum/11')
  })

  it('renders description when present', () => {
    render(<CurriculumCard curriculum={sample} />)
    expect(screen.getByText('Учебный план направления подготовки')).toBeInTheDocument()
  })

  it('omits description chip when description is empty', () => {
    const { container } = render(
      <CurriculumCard curriculum={{ ...sample, description: '' }} />
    )
    expect(container.textContent).not.toMatch('Учебный план направления подготовки')
  })

  it('renders the status pill via translated key', () => {
    // jest.setup.ts mocks useTranslations to return the key verbatim
    // (no interpolation), so we assert on the key, not the value.
    render(<CurriculumCard curriculum={sample} />)
    expect(screen.getByText('card.status.draft')).toBeInTheDocument()
  })

  it.each<CurriculumStatus>([
    'draft',
    'pending_approval',
    'approved',
    'archived',
  ])('uses a distinct translation key for status=%s', (status) => {
    render(<CurriculumCard curriculum={{ ...sample, status }} />)
    const expectedKey =
      status === 'pending_approval' ? 'card.status.pending' : `card.status.${status}`
    expect(screen.getByText(expectedKey)).toBeInTheDocument()
  })

  it('renders the openAria translation key on the link', () => {
    render(<CurriculumCard curriculum={sample} />)
    expect(screen.getByLabelText('card.openAria')).toBeInTheDocument()
  })
})
