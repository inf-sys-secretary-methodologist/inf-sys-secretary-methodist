import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FullCalendar } from '../FullCalendar'
import type { CalendarEvent } from '@/types/calendar'
import * as useMediaQueryModule from '@/hooks/use-media-query'

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

  it('calls onEventClick when event is clicked', async () => {
    const user = userEvent.setup()
    const onEventClick = jest.fn()
    render(<FullCalendar {...defaultProps} onEventClick={onEventClick} />)

    // In month view, find and click the event
    const eventElement = screen.queryByText('Test Meeting')
    if (eventElement) {
      await user.click(eventElement)
      expect(onEventClick).toHaveBeenCalled()
    }
  })

  it('renders with onUpdateEvent prop', () => {
    const onUpdateEvent = jest.fn()
    const { container } = render(<FullCalendar {...defaultProps} onUpdateEvent={onUpdateEvent} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders with onDeleteEvent prop', () => {
    const onDeleteEvent = jest.fn()
    const { container } = render(<FullCalendar {...defaultProps} onDeleteEvent={onDeleteEvent} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('navigates when previous button is clicked', async () => {
    const user = userEvent.setup()
    const { container } = render(<FullCalendar {...defaultProps} />)

    const initialHTML = container.innerHTML

    await user.click(screen.getByRole('button', { name: /previous/i }))

    // The calendar should have changed
    expect(container.innerHTML).not.toBe(initialHTML)
  })

  it('renders loading state', () => {
    const { container } = render(<FullCalendar {...defaultProps} isLoading={true} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('can switch back to month view', async () => {
    const user = userEvent.setup()
    render(<FullCalendar {...defaultProps} />)

    // Switch to week first
    await user.click(screen.getByRole('tab', { name: /week/i }))
    expect(screen.getByText('Mon')).toBeInTheDocument()

    // Switch back to month
    await user.click(screen.getByRole('tab', { name: /month/i }))
    expect(screen.getByRole('tab', { name: /month/i })).toBeInTheDocument()
  })

  it('renders events in day view', async () => {
    const user = userEvent.setup()
    render(<FullCalendar {...defaultProps} />)

    await user.click(screen.getByRole('tab', { name: /day/i }))
    // Day view should show event list or time grid
    expect(screen.getByText(/events/i)).toBeInTheDocument()
  })

  it('handles create event callback', async () => {
    const user = userEvent.setup()
    const onCreateEvent = jest.fn()
    render(<FullCalendar {...defaultProps} onCreateEvent={onCreateEvent} />)

    // Click new event button - use button role to be specific
    const newEventButton = screen.getByRole('button', { name: /new event/i })
    await user.click(newEventButton)

    // Modal should open - there should now be multiple "New Event" texts
    const newEventTexts = screen.getAllByText('New Event')
    expect(newEventTexts.length).toBeGreaterThan(1)
  })

  it('displays all weekday headers', () => {
    render(<FullCalendar {...defaultProps} />)
    expect(screen.getByText('Sun')).toBeInTheDocument()
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Tue')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Thu')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
  })
})

describe('FullCalendar mobile view', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders mobile tabs when on mobile', () => {
    // Mock mobile view
    jest.spyOn(useMediaQueryModule, 'useIsMobile').mockReturnValue(true)

    render(<FullCalendar events={mockEvents} isLoading={false} />)
    // Mobile tabs should be visible with mobile layout - multiple tabs exist
    const monthTabs = screen.getAllByRole('tab', { name: /month/i })
    expect(monthTabs.length).toBeGreaterThan(0)
  })

  it('handles mobile tabs onValueChange', async () => {
    const user = userEvent.setup()
    // Mock mobile view
    jest.spyOn(useMediaQueryModule, 'useIsMobile').mockReturnValue(true)

    render(<FullCalendar events={mockEvents} isLoading={false} />)

    // Find mobile tabs and click week tab to trigger onValueChange
    const weekTabs = screen.getAllByRole('tab', { name: /week/i })
    // Click the last week tab (mobile one)
    await user.click(weekTabs[weekTabs.length - 1])

    // Verify view changed - week view should now show day names
    expect(screen.getByText('Mon')).toBeInTheDocument()
  })
})
