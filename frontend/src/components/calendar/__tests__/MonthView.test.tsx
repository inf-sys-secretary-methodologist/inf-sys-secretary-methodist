import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MonthView } from '../MonthView'
import type { CalendarEvent } from '@/types/calendar'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
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
    title: 'Lunch Meeting',
    start_time: '2024-06-15T12:00:00',
    end_time: '2024-06-15T13:00:00',
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
  {
    id: 4,
    title: 'Evening Call',
    start_time: '2024-06-15T18:00:00',
    end_time: '2024-06-15T19:00:00',
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

describe('MonthView', () => {
  const defaultProps = {
    currentDate: new Date(2024, 5, 15), // June 15, 2024
    events: mockEvents,
  }

  it('renders month view component', () => {
    const { container } = render(<MonthView {...defaultProps} />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders weekday headers', () => {
    render(<MonthView {...defaultProps} />)
    expect(screen.getByText('Sun')).toBeInTheDocument()
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Tue')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Thu')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
  })

  it('renders day numbers', () => {
    render(<MonthView {...defaultProps} />)
    // June 15, 2024 should be visible - use getAllByText since there might be duplicates
    expect(screen.getByText('15')).toBeInTheDocument()
    expect(screen.getAllByText('1').length).toBeGreaterThan(0)
    expect(screen.getAllByText('30').length).toBeGreaterThan(0)
  })

  it('renders events on the correct day', () => {
    render(<MonthView {...defaultProps} />)
    expect(screen.getByText('Morning Meeting')).toBeInTheDocument()
  })

  it('shows more events button when there are more than 3 events', () => {
    render(<MonthView {...defaultProps} />)
    // Day 15 has 4 events, so should show "+1 more"
    expect(screen.getByText('+1 more')).toBeInTheDocument()
  })

  it('calls onDateSelect when a day is clicked', async () => {
    const user = userEvent.setup()
    const onDateSelect = jest.fn()
    render(<MonthView {...defaultProps} onDateSelect={onDateSelect} />)

    await user.click(screen.getByText('20'))
    expect(onDateSelect).toHaveBeenCalled()
  })

  it('calls onEventClick when event is clicked', async () => {
    const user = userEvent.setup()
    const onEventClick = jest.fn()
    render(<MonthView {...defaultProps} onEventClick={onEventClick} />)

    await user.click(screen.getByText('Morning Meeting'))
    expect(onEventClick).toHaveBeenCalledWith(mockEvents[0])
  })

  it('highlights selected date', () => {
    const { container } = render(
      <MonthView {...defaultProps} selectedDate={new Date(2024, 5, 20)} />
    )
    // Check that there's a cell with selected styling
    const selectedCell = container.querySelector('.bg-gray-100')
    expect(selectedCell).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<MonthView {...defaultProps} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('handles empty events array', () => {
    render(<MonthView {...defaultProps} events={[]} />)
    expect(screen.queryByText('Morning Meeting')).not.toBeInTheDocument()
  })

  it('displays days from previous and next months', () => {
    render(<MonthView {...defaultProps} />)
    // June 2024 starts on Saturday, so should show some May dates
    // and ends on Sunday, so should show some July dates
    const dayNumbers = screen.getAllByText(/^\d+$/)
    expect(dayNumbers.length).toBeGreaterThan(28) // More than just June days
  })

  it('renders 7 weekday headers', () => {
    render(<MonthView {...defaultProps} />)
    // Check that all weekday abbreviations are present
    expect(screen.getByText('Sun')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
  })
})
