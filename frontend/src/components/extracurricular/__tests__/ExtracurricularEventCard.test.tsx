import { render, screen, fireEvent } from '@testing-library/react'
import { NextIntlClientProvider } from 'next-intl'
import { ExtracurricularEventCard } from '../ExtracurricularEventCard'
import ruMessages from '../../../../messages/ru.json'
import type { ExtracurricularEventSummary } from '@/types/extracurricular'

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

function renderWith(node: React.ReactNode) {
  return render(
    <NextIntlClientProvider locale="ru" messages={ruMessages}>
      {node}
    </NextIntlClientProvider>
  )
}

describe('ExtracurricularEventCard', () => {
  it('renders event title', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText('Spring concert')).toBeInTheDocument()
  })

  it('renders status badge with localized label', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    // ru.json: extracurricular.status.published = "Опубликовано"
    expect(screen.getByText('Опубликовано')).toBeInTheDocument()
  })

  it('renders category badge with localized label', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    // ru.json: extracurricular.category.cultural = "Культурное"
    expect(screen.getByText('Культурное')).toBeInTheDocument()
  })

  it('renders audience badge with localized label', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    // ru.json: extracurricular.audience.all = "Все"
    expect(screen.getByText('Все')).toBeInTheDocument()
  })

  it('renders location text', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText(/Main hall/i)).toBeInTheDocument()
  })

  it('renders participant count and capacity', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.getByText(/42/)).toBeInTheDocument()
    expect(screen.getByText(/200/)).toBeInTheDocument()
  })

  it('omits capacity when max_capacity is null/undefined', () => {
    renderWith(
      <ExtracurricularEventCard
        event={{ ...baseEvent, max_capacity: null, participant_count: 5 }}
      />
    )
    expect(screen.getByText(/5/)).toBeInTheDocument()
    expect(screen.queryByText(/200/)).not.toBeInTheDocument()
  })

  it('calls onClick when the title is clicked', () => {
    const onClick = jest.fn()
    renderWith(<ExtracurricularEventCard event={baseEvent} onClick={onClick} />)
    fireEvent.click(screen.getByText('Spring concert'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders Register action when onRegister provided and not registered', () => {
    const onRegister = jest.fn()
    renderWith(<ExtracurricularEventCard event={baseEvent} onRegister={onRegister} />)
    const btn = screen.getByText('Записаться')
    fireEvent.click(btn)
    expect(onRegister).toHaveBeenCalledTimes(1)
  })

  it('renders Unregister action instead of Register when isRegistered=true', () => {
    const onUnregister = jest.fn()
    renderWith(
      <ExtracurricularEventCard
        event={baseEvent}
        onRegister={jest.fn()}
        onUnregister={onUnregister}
        isRegistered
      />
    )
    expect(screen.queryByText('Записаться')).not.toBeInTheDocument()
    const btn = screen.getByText('Отменить запись')
    fireEvent.click(btn)
    expect(onUnregister).toHaveBeenCalledTimes(1)
  })

  it('shows dropdown menu trigger when edit/delete handlers supplied', () => {
    renderWith(
      <ExtracurricularEventCard event={baseEvent} onEdit={jest.fn()} onDelete={jest.fn()} />
    )
    expect(screen.getByTestId('event-card-menu-trigger')).toBeInTheDocument()
  })

  it('omits dropdown menu trigger when no actions supplied', () => {
    renderWith(<ExtracurricularEventCard event={baseEvent} />)
    expect(screen.queryByTestId('event-card-menu-trigger')).not.toBeInTheDocument()
  })
})
