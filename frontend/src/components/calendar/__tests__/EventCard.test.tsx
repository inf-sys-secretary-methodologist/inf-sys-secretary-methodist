import { render, screen, fireEvent } from '@testing-library/react'
import { EventCard } from '../EventCard'
import type { CalendarEvent } from '@/types/calendar'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { count?: number }) => {
    const translations: Record<string, string> = {
      'eventTypes.meeting': 'Meeting',
      'eventTypes.deadline': 'Deadline',
      'eventTypes.task': 'Task',
      'eventTypes.reminder': 'Reminder',
      'eventTypes.holiday': 'Holiday',
      'eventTypes.personal': 'Personal',
      recurring: 'Recurring',
      allDay: 'All day',
      menu: 'Menu',
      edit: 'Edit',
      delete: 'Delete',
      participants: `${params?.count} participants`,
    }
    return translations[key] || key
  },
}))

describe('EventCard', () => {
  const baseEvent: CalendarEvent = {
    id: 1,
    title: 'Team Meeting',
    event_type: 'meeting',
    status: 'scheduled',
    start_time: new Date(2024, 0, 15, 10, 0).toISOString(),
    end_time: new Date(2024, 0, 15, 11, 0).toISOString(),
    all_day: false,
    timezone: 'Europe/Moscow',
    is_recurring: false,
    organizer_id: 1,
    priority: 1,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }

  describe('Compact variant', () => {
    it('renders event title', () => {
      render(<EventCard event={baseEvent} variant="compact" />)
      expect(screen.getByText('Team Meeting')).toBeInTheDocument()
    })

    it('renders time for non-all-day events', () => {
      render(<EventCard event={baseEvent} variant="compact" />)
      expect(screen.getByText('10:00')).toBeInTheDocument()
    })

    it('does not show time for all-day events', () => {
      render(<EventCard event={{ ...baseEvent, all_day: true }} variant="compact" />)
      expect(screen.queryByText('10:00')).not.toBeInTheDocument()
    })

    it('calls onClick when clicked', () => {
      const onClick = jest.fn()
      render(<EventCard event={baseEvent} variant="compact" onClick={onClick} />)
      fireEvent.click(screen.getByText('Team Meeting'))
      expect(onClick).toHaveBeenCalled()
    })

    it('applies custom className', () => {
      const { container } = render(
        <EventCard event={baseEvent} variant="compact" className="custom-event" />
      )
      expect(container.querySelector('.custom-event')).toBeInTheDocument()
    })
  })

  describe('Full variant', () => {
    it('renders event title', () => {
      render(<EventCard event={baseEvent} variant="full" />)
      expect(screen.getByText('Team Meeting')).toBeInTheDocument()
    })

    it('renders event type badge', () => {
      render(<EventCard event={baseEvent} variant="full" />)
      expect(screen.getByText('Meeting')).toBeInTheDocument()
    })

    it('shows recurring badge for recurring events', () => {
      render(<EventCard event={{ ...baseEvent, is_recurring: true }} variant="full" />)
      expect(screen.getByText('Recurring')).toBeInTheDocument()
    })

    it('shows all day text for all-day events', () => {
      render(<EventCard event={{ ...baseEvent, all_day: true }} variant="full" />)
      expect(screen.getByText('All day')).toBeInTheDocument()
    })

    it('shows time range for timed events', () => {
      render(<EventCard event={baseEvent} variant="full" />)
      expect(screen.getByText(/10:00.*11:00/)).toBeInTheDocument()
    })

    it('renders description when provided', () => {
      render(<EventCard event={{ ...baseEvent, description: 'Weekly team sync' }} variant="full" />)
      expect(screen.getByText('Weekly team sync')).toBeInTheDocument()
    })

    it('renders location when provided', () => {
      render(<EventCard event={{ ...baseEvent, location: 'Conference Room A' }} variant="full" />)
      expect(screen.getByText('Conference Room A')).toBeInTheDocument()
    })

    it('renders participants count when provided', () => {
      render(
        <EventCard
          event={{
            ...baseEvent,
            participants: [
              { user_id: 1, user_name: 'John', response_status: 'accepted', role: 'required' },
              { user_id: 2, user_name: 'Jane', response_status: 'pending', role: 'optional' },
            ],
          }}
          variant="full"
        />
      )
      expect(screen.getByText('2 participants')).toBeInTheDocument()
    })

    it('calls onClick when title is clicked', () => {
      const onClick = jest.fn()
      render(<EventCard event={baseEvent} variant="full" onClick={onClick} />)
      fireEvent.click(screen.getByText('Team Meeting'))
      expect(onClick).toHaveBeenCalled()
    })

    it('shows menu button when onEdit or onDelete provided', () => {
      render(<EventCard event={baseEvent} variant="full" onEdit={jest.fn()} />)
      expect(screen.getByRole('button', { name: 'Menu' })).toBeInTheDocument()
    })

    it('does not show menu button when no handlers provided', () => {
      render(<EventCard event={baseEvent} variant="full" />)
      expect(screen.queryByRole('button', { name: 'Menu' })).not.toBeInTheDocument()
    })
  })

  describe('Event types', () => {
    const eventTypes: CalendarEvent['event_type'][] = [
      'meeting',
      'deadline',
      'task',
      'reminder',
      'holiday',
      'personal',
    ]

    eventTypes.forEach((eventType) => {
      it(`renders ${eventType} event type correctly`, () => {
        render(<EventCard event={{ ...baseEvent, event_type: eventType }} variant="full" />)
        expect(screen.getByText('Team Meeting')).toBeInTheDocument()
      })
    })
  })

  describe('Custom colors', () => {
    it('applies custom color when provided', () => {
      const { container } = render(
        <EventCard event={{ ...baseEvent, color: '#ff5500' }} variant="full" />
      )
      // Check that the color indicator has custom style
      const colorIndicator = container.querySelector('[style*="background-color"]')
      expect(colorIndicator).toBeInTheDocument()
    })
  })

  describe('Default variant', () => {
    it('defaults to compact variant', () => {
      render(<EventCard event={baseEvent} />)
      // Compact variant shows time in a specific format
      expect(screen.getByText('10:00')).toBeInTheDocument()
      // Full variant would show event type badge
      expect(screen.queryByText('Meeting')).not.toBeInTheDocument()
    })
  })
})
