import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FullCalendar } from '../FullCalendar'
import type { CalendarEvent } from '@/types/calendar'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      month: 'Month',
      week: 'Week',
      day: 'Day',
      today: 'Today',
      previous: 'Previous',
      next: 'Next',
      newEvent: 'New Event',
      weekNumber: 'Week',
      allDay: 'All Day',
      eventsCount: `${params?.count || 0} events`,
      'modal.newTitle': 'New Event',
      'modal.editTitle': 'Edit Event',
      dayEventsTitle: 'Events',
      noEvents: 'No events',
      'weekdays.sun': 'Sun',
      'weekdays.mon': 'Mon',
      'weekdays.tue': 'Tue',
      'weekdays.wed': 'Wed',
      'weekdays.thu': 'Thu',
      'weekdays.fri': 'Fri',
      'weekdays.sat': 'Sat',
      moreEvents: `+${params?.count || 0} more`,
    }
    return translations[key] || key
  },
  useLocale: () => 'en',
}))

// Mock hooks
jest.mock('@/hooks/use-media-query', () => ({
  useIsMobile: () => false,
}))

// Mock next/dynamic
jest.mock('next/dynamic', () => () => {
  const MockCalendar = () => <div data-testid="calendar">Calendar</div>
  return MockCalendar
})

const mockEvents: CalendarEvent[] = [
  {
    id: 1,
    title: 'Test Meeting',
    description: 'Team standup',
    start_time: new Date().toISOString(),
    end_time: new Date(Date.now() + 3600000).toISOString(),
    all_day: false,
    event_type: 'meeting',
    status: 'scheduled',
    timezone: 'Europe/Moscow',
    organizer_id: 1,
    is_recurring: false,
    priority: 1,
    created_at: '2024-01-01T00:00:00',
    updated_at: '2024-01-01T00:00:00',
  },
]

describe('FullCalendar', () => {
  const defaultProps = {
    events: mockEvents,
    isLoading: false,
  }

  it('renders the full calendar component', () => {
    const { container } = render(<FullCalendar {...defaultProps} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders calendar header with today button', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.getByText('Today')).toBeInTheDocument()
  })

  it('renders view tabs', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.getByRole('tab', { name: /month/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /week/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /day/i })).toBeInTheDocument()
  })

  it('renders navigation buttons', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.getByRole('button', { name: /previous/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument()
  })

  it('renders new event button when onCreateEvent is provided', () => {
    const onCreateEvent = jest.fn()
    render(<FullCalendar {...defaultProps} onCreateEvent={onCreateEvent} />)
    expect(screen.getByText('New Event')).toBeInTheDocument()
  })

  it('does not render new event button when onCreateEvent is not provided', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.queryByText('New Event')).not.toBeInTheDocument()
  })

  it('switches to week view when week tab is clicked', async () => {
    const user = userEvent.setup()
    render(<FullCalendar {...defaultProps} />)

    await user.click(screen.getByRole('tab', { name: /week/i }))
    // Week view should now be active - check for weekday headers
    expect(screen.getByText('Mon')).toBeInTheDocument()
  })

  it('switches to day view when day tab is clicked', async () => {
    const user = userEvent.setup()
    render(<FullCalendar {...defaultProps} />)

    await user.click(screen.getByRole('tab', { name: /day/i }))
    // Day view should now be active - check for events count text
    expect(screen.getByText(/events/i)).toBeInTheDocument()
  })

  it('navigates when next button is clicked', async () => {
    const user = userEvent.setup()
    const { container } = render(<FullCalendar {...defaultProps} />)

    const initialHTML = container.innerHTML

    await user.click(screen.getByRole('button', { name: /next/i }))

    // The calendar should have changed
    expect(container.innerHTML).not.toBe(initialHTML)
  })

  it('applies custom className', () => {
    const { container } = render(<FullCalendar {...defaultProps} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('handles empty events array', () => {
    const { container } = render(<FullCalendar events={[]} isLoading={false} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders weekday headers in month view', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.getByText('Sun')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
  })

  it('can click on today button', async () => {
    const user = userEvent.setup()
    render(<FullCalendar {...defaultProps} />)

    // Navigate away first
    await user.click(screen.getByRole('button', { name: /next/i }))

    // Click today
    await user.click(screen.getByText('Today'))

    // Component should still be rendered
    expect(screen.getByText('Today')).toBeInTheDocument()
  })
})
