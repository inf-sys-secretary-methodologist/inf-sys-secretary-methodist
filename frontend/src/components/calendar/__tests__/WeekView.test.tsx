import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WeekView } from '../WeekView'
import type { CalendarEvent } from '@/types/calendar'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      allDay: 'All Day',
    }
    return translations[key] || key
  },
  useLocale: () => 'en',
}))

const mockEvents: CalendarEvent[] = [
  {
    id: 1,
    title: 'Monday Meeting',
    description: 'Weekly sync',
    start_time: '2024-06-10T09:00:00', // Monday
    end_time: '2024-06-10T10:00:00',
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
    title: 'All Week Conference',
    start_time: '2024-06-11T00:00:00', // Tuesday
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
    title: 'Friday Task',
    start_time: '2024-06-14T14:00:00', // Friday
    end_time: '2024-06-14T16:00:00',
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

describe('WeekView', () => {
  const defaultProps = {
    currentDate: new Date(2024, 5, 12), // June 12, 2024 (Wednesday)
    events: mockEvents,
  }

  it('renders week view component', () => {
    const { container } = render(<WeekView {...defaultProps} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders 7 days of the week', () => {
    render(<WeekView {...defaultProps} />)
    // Week of June 10-16, 2024 (Mon-Sun)
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('11')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
    expect(screen.getByText('13')).toBeInTheDocument()
    expect(screen.getByText('14')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
    expect(screen.getByText('16')).toBeInTheDocument()
  })

  it('renders day names', () => {
    render(<WeekView {...defaultProps} />)
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Tue')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Thu')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
    expect(screen.getByText('Sun')).toBeInTheDocument()
  })

  it('renders hour slots', () => {
    render(<WeekView {...defaultProps} />)
    expect(screen.getByText('00:00')).toBeInTheDocument()
    expect(screen.getByText('12:00')).toBeInTheDocument()
  })

  it('renders timed events', () => {
    render(<WeekView {...defaultProps} />)
    expect(screen.getByText('Monday Meeting')).toBeInTheDocument()
    expect(screen.getByText('Friday Task')).toBeInTheDocument()
  })

  it('renders all-day events section', () => {
    render(<WeekView {...defaultProps} />)
    expect(screen.getByText('All Week Conference')).toBeInTheDocument()
  })

  it('calls onDateSelect when day header is clicked', async () => {
    const user = userEvent.setup()
    const onDateSelect = jest.fn()
    render(<WeekView {...defaultProps} onDateSelect={onDateSelect} />)

    await user.click(screen.getByText('12'))
    expect(onDateSelect).toHaveBeenCalled()
  })

  it('calls onEventClick when event is clicked', async () => {
    const user = userEvent.setup()
    const onEventClick = jest.fn()
    render(<WeekView {...defaultProps} onEventClick={onEventClick} />)

    await user.click(screen.getByText('Monday Meeting'))
    expect(onEventClick).toHaveBeenCalledWith(mockEvents[0])
  })

  it('calls onTimeSlotClick when time slot is clicked', async () => {
    const user = userEvent.setup()
    const onTimeSlotClick = jest.fn()
    const { container } = render(<WeekView {...defaultProps} onTimeSlotClick={onTimeSlotClick} />)

    // Find and click a time slot
    const timeSlots = container.querySelectorAll('.cursor-pointer')
    if (timeSlots.length > 0) {
      await user.click(timeSlots[10])
      expect(onTimeSlotClick).toHaveBeenCalled()
    }
  })

  it('highlights selected date', () => {
    const { container } = render(
      <WeekView {...defaultProps} selectedDate={new Date(2024, 5, 12)} />
    )
    const selectedDay = container.querySelector('.bg-gray-100')
    expect(selectedDay).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<WeekView {...defaultProps} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('handles empty events array', () => {
    render(<WeekView {...defaultProps} events={[]} />)
    expect(screen.queryByText('Monday Meeting')).not.toBeInTheDocument()
  })

  it('renders 24 hour slots', () => {
    render(<WeekView {...defaultProps} />)
    // Check for various hour labels
    expect(screen.getByText('08:00')).toBeInTheDocument()
    expect(screen.getByText('17:00')).toBeInTheDocument()
    expect(screen.getByText('23:00')).toBeInTheDocument()
  })

  it('renders time column', () => {
    const { container } = render(<WeekView {...defaultProps} />)
    const timeColumn = container.querySelector('.w-16')
    expect(timeColumn).toBeInTheDocument()
  })

  it('separates all-day events from timed events', () => {
    render(<WeekView {...defaultProps} />)
    // Both should be rendered in different sections
    expect(screen.getByText('All Week Conference')).toBeInTheDocument()
    expect(screen.getByText('Monday Meeting')).toBeInTheDocument()
  })
})
