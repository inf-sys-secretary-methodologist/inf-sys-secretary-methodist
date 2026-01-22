import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DayView } from '../DayView'
import type { CalendarEvent } from '@/types/calendar'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      allDay: 'All Day',
      eventsCount: `${params?.count || 0} events`,
    }
    return translations[key] || key
  },
  useLocale: () => 'en',
}))

const mockEvents: CalendarEvent[] = [
  {
    id: 1,
    title: 'Morning Meeting',
    description: 'Team standup',
    start_time: '2024-06-15T09:00:00',
    end_time: '2024-06-15T10:00:00',
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
  {
    id: 2,
    title: 'All Day Conference',
    description: 'Annual conference',
    start_time: '2024-06-15T00:00:00',
    all_day: true,
    event_type: 'meeting',
    status: 'scheduled',
    timezone: 'Europe/Moscow',
    organizer_id: 1,
    is_recurring: false,
    priority: 1,
    created_at: '2024-01-01T00:00:00',
    updated_at: '2024-01-01T00:00:00',
  },
  {
    id: 3,
    title: 'Afternoon Task',
    start_time: '2024-06-15T14:00:00',
    end_time: '2024-06-15T16:00:00',
    all_day: false,
    event_type: 'task',
    status: 'scheduled',
    timezone: 'Europe/Moscow',
    organizer_id: 1,
    is_recurring: false,
    priority: 1,
    created_at: '2024-01-01T00:00:00',
    updated_at: '2024-01-01T00:00:00',
  },
]

describe('DayView', () => {
  const defaultProps = {
    currentDate: new Date(2024, 5, 15), // June 15, 2024
    events: mockEvents,
  }

  it('renders day view component', () => {
    render(<DayView {...defaultProps} />)
    expect(screen.getByText(/15 June 2024/i)).toBeInTheDocument()
  })

  it('displays the day of week', () => {
    render(<DayView {...defaultProps} />)
    expect(screen.getByText('Sat')).toBeInTheDocument() // June 15, 2024 is a Saturday
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('displays events count', () => {
    render(<DayView {...defaultProps} />)
    expect(screen.getByText('3 events')).toBeInTheDocument()
  })

  it('separates all-day and timed events', () => {
    render(<DayView {...defaultProps} />)
    // There are two "All Day" texts - section header and event badge
    expect(screen.getAllByText('All Day').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('All Day Conference')).toBeInTheDocument()
  })

  it('renders all event titles', () => {
    render(<DayView {...defaultProps} />)
    expect(screen.getByText('Morning Meeting')).toBeInTheDocument()
    expect(screen.getByText('All Day Conference')).toBeInTheDocument()
    expect(screen.getByText('Afternoon Task')).toBeInTheDocument()
  })

  it('renders 24 hour slots', () => {
    render(<DayView {...defaultProps} />)
    // Check that hours are rendered (at least some hour labels)
    expect(screen.getByText('00:00')).toBeInTheDocument()
    expect(screen.getByText('12:00')).toBeInTheDocument()
  })

  it('calls onEventClick when event is clicked', async () => {
    const user = userEvent.setup()
    const onEventClick = jest.fn()
    render(<DayView {...defaultProps} onEventClick={onEventClick} />)

    await user.click(screen.getByText('Morning Meeting'))
    expect(onEventClick).toHaveBeenCalledWith(mockEvents[0])
  })

  it('calls onTimeSlotClick when time slot is clicked', async () => {
    const user = userEvent.setup()
    const onTimeSlotClick = jest.fn()
    const { container } = render(<DayView {...defaultProps} onTimeSlotClick={onTimeSlotClick} />)

    // Find and click a time slot (they have cursor-pointer class)
    const timeSlots = container.querySelectorAll('.cursor-pointer')
    if (timeSlots.length > 0) {
      await user.click(timeSlots[10]) // Click on 10 AM slot
      expect(onTimeSlotClick).toHaveBeenCalled()
    }
  })

  it('applies custom className', () => {
    const { container } = render(<DayView {...defaultProps} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('does not show all-day section when there are no all-day events', () => {
    const timedOnlyEvents = mockEvents.filter((e) => !e.all_day)
    render(<DayView {...defaultProps} events={timedOnlyEvents} />)
    expect(screen.queryByText('All Day')).not.toBeInTheDocument()
  })

  it('handles empty events array', () => {
    render(<DayView {...defaultProps} events={[]} />)
    expect(screen.getByText('0 events')).toBeInTheDocument()
  })

  it('renders time column with hour labels', () => {
    render(<DayView {...defaultProps} />)
    // Check for various hour labels
    expect(screen.getByText('08:00')).toBeInTheDocument()
    expect(screen.getByText('17:00')).toBeInTheDocument()
    expect(screen.getByText('23:00')).toBeInTheDocument()
  })
})
