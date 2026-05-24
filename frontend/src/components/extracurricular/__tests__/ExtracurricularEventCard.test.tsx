import { render, screen, fireEvent } from '@testing-library/react'
import { ExtracurricularEventCard } from '../ExtracurricularEventCard'
import type { ExtracurricularEventSummary } from '@/types/extracurricular'

// next-intl is auto-mocked in jest.setup.ts: useTranslations returns
// the key verbatim. Parity test loads real JSON for translations —
// see hooks/__tests__/extracurricular.i18n.test.ts.

const baseEvent: ExtracurricularEventSummary = {
  id: 7,
  title: 'Spring concert',
  category: 'cultural',
  target_audience: 'all',
  status: 'published',
  location: 'Main hall',
  start_at: '2026-06-15T18:00:00Z',
  end_at: '2026-06-15T21:00:00Z',
  max_capacity: 200,
  organizer_id: 5,
  participant_count: 42,
  version: 3,
  created_at: '2026-05-20T10:00:00Z',
  updated_at: '2026-05-25T12:00:00Z',
}

describe('ExtracurricularEventCard', () => {
  it('renders event title', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('Spring concert')).toBeInTheDocument()
  })

  it('renders status badge (key — t-mock verbatim)', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('status.published')).toBeInTheDocument()
  })

  it('renders category badge', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('category.cultural')).toBeInTheDocument()
  })

  it('renders audience badge', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('audience.all')).toBeInTheDocument()
  })

  it('renders location text', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText(/Main hall/i)).toBeInTheDocument()
  })

  it('renders participant count and capacity', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('42 / 200')).toBeInTheDocument()
  })

  it('omits capacity when max_capacity is null/undefined', () => {
    render(
      <ExtracurricularEventCard
        event={{ ...baseEvent, max_capacity: null, participant_count: 5 }}
      />
    )
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.queryByText(/\/ 200/)).not.toBeInTheDocument()
  })

  it('calls onClick when the title is clicked', () => {
    const onClick = jest.fn()
    render(<ExtracurricularEventCard event={baseEvent} onClick={onClick} />)
    fireEvent.click(screen.getByText('Spring concert'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders Register action when onRegister provided and not registered', () => {
    const onRegister = jest.fn()
    render(<ExtracurricularEventCard event={baseEvent} onRegister={onRegister} />)
    const btn = screen.getByText('register')
    fireEvent.click(btn)
    expect(onRegister).toHaveBeenCalledTimes(1)
  })

  it('renders Unregister action instead of Register when isRegistered=true', () => {
    const onUnregister = jest.fn()
    render(
      <ExtracurricularEventCard
        event={baseEvent}
        onRegister={jest.fn()}
        onUnregister={onUnregister}
        isRegistered
      />
    )
    expect(screen.queryByText('register')).not.toBeInTheDocument()
    const btn = screen.getByText('unregister')
    fireEvent.click(btn)
    expect(onUnregister).toHaveBeenCalledTimes(1)
  })

  it('shows dropdown menu trigger when edit/delete handlers supplied', () => {
    render(<ExtracurricularEventCard event={baseEvent} onEdit={jest.fn()} onDelete={jest.fn()} />)
    expect(screen.getByTestId('event-card-menu-trigger')).toBeInTheDocument()
  })

  it('omits dropdown menu trigger when no actions supplied', () => {
    render(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.queryByTestId('event-card-menu-trigger')).not.toBeInTheDocument()
  })
})
