import { render, screen, fireEvent } from '@testing-library/react'

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const mockUseEvents = jest.fn()
jest.mock('@/hooks/useExtracurricularEvents', () => ({
  useExtracurricularEvents: (filter?: unknown) => mockUseEvents(filter),
}))

import { UpcomingEventsWidget } from '../UpcomingEventsWidget'

const futureEvent = (id: number, title: string, daysFromNow: number) => ({
  id,
  title,
  category: 'cultural' as const,
  target_audience: 'all' as const,
  status: 'published' as const,
  location: 'Hall',
  start_at: new Date(Date.now() + daysFromNow * 86_400_000).toISOString(),
  end_at: new Date(Date.now() + (daysFromNow + 1) * 86_400_000).toISOString(),
  max_capacity: 100,
  organizer_id: 1,
  participant_count: 0,
  version: 1,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
})

beforeEach(() => {
  mockPush.mockClear()
  mockUseEvents.mockReturnValue({
    events: [],
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('UpcomingEventsWidget', () => {
  it('renders the widget heading', () => {
    render(<UpcomingEventsWidget />)
    expect(screen.getByText('upcoming')).toBeInTheDocument()
  })

  it('renders up to 5 upcoming event titles', () => {
    mockUseEvents.mockReturnValue({
      events: [
        futureEvent(1, 'Hackathon', 3),
        futureEvent(2, 'Concert', 7),
        futureEvent(3, 'Sports day', 14),
      ],
      total: 3,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<UpcomingEventsWidget />)
    expect(screen.getByText('Hackathon')).toBeInTheDocument()
    expect(screen.getByText('Concert')).toBeInTheDocument()
    expect(screen.getByText('Sports day')).toBeInTheDocument()
  })

  it('skips past events (start_at before now)', () => {
    mockUseEvents.mockReturnValue({
      events: [futureEvent(1, 'Past event', -3), futureEvent(2, 'Future event', 5)],
      total: 2,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<UpcomingEventsWidget />)
    expect(screen.queryByText('Past event')).not.toBeInTheDocument()
    expect(screen.getByText('Future event')).toBeInTheDocument()
  })

  it('shows empty state when no upcoming events', () => {
    render(<UpcomingEventsWidget />)
    expect(screen.getByText('empty')).toBeInTheDocument()
  })

  it('shows loading skeleton when isLoading', () => {
    mockUseEvents.mockReturnValue({
      events: [],
      total: 0,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    const { container } = render(<UpcomingEventsWidget />)
    expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
  })

  it('passes filter status=published to the hook', () => {
    render(<UpcomingEventsWidget />)
    const filter = mockUseEvents.mock.calls[0]?.[0]
    expect(filter).toEqual(expect.objectContaining({ status: 'published' }))
  })

  it('navigates to event detail on click', () => {
    mockUseEvents.mockReturnValue({
      events: [futureEvent(42, 'Workshop', 5)],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<UpcomingEventsWidget />)
    fireEvent.click(screen.getByText('Workshop'))
    expect(mockPush).toHaveBeenCalledWith('/extracurricular/42')
  })
})
