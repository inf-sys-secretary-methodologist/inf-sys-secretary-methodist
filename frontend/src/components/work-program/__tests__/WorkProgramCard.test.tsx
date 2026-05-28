import { render, screen } from '@testing-library/react'
import { WorkProgramCard } from '../WorkProgramCard'
import type { WorkProgramSummary } from '@/types/workProgram'

// next-intl is auto-mocked in jest.setup.ts: useTranslations returns the
// key verbatim (no interpolation). Real-string parity is covered by
// hooks/__tests__/workProgram.i18n.test.ts.

const baseWP: WorkProgramSummary = {
  id: 7,
  discipline_id: 10,
  specialty_code: '09.03.01',
  applicable_from_year: 2026,
  title: 'Базы данных',
  status: 'approved',
  author_id: 5,
  version: 2,
}

describe('WorkProgramCard', () => {
  it('renders the work program title', () => {
    render(<WorkProgramCard workProgram={baseWP} />)
    expect(screen.getByText('Базы данных')).toBeInTheDocument()
  })

  it('renders the status pill key for approved (t-mock verbatim)', () => {
    render(<WorkProgramCard workProgram={baseWP} />)
    expect(screen.getByText('card.status.approved')).toBeInTheDocument()
  })

  it('maps pending_approval to the short pill key "pending"', () => {
    render(<WorkProgramCard workProgram={{ ...baseWP, status: 'pending_approval' }} />)
    expect(screen.getByText('card.status.pending')).toBeInTheDocument()
  })

  it('maps needs_revision to the camelCase pill key "needsRevision"', () => {
    render(<WorkProgramCard workProgram={{ ...baseWP, status: 'needs_revision' }} />)
    expect(screen.getByText('card.status.needsRevision')).toBeInTheDocument()
  })

  it('renders the specialty code', () => {
    render(<WorkProgramCard workProgram={baseWP} />)
    expect(screen.getByText('09.03.01')).toBeInTheDocument()
  })

  it('renders the cohort year', () => {
    render(<WorkProgramCard workProgram={baseWP} />)
    expect(screen.getByText('2026')).toBeInTheDocument()
  })

  it('links to the detail page', () => {
    render(<WorkProgramCard workProgram={baseWP} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/work-programs/7')
  })
})
