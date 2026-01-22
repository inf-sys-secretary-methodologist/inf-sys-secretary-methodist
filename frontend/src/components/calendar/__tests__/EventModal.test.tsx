import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EventModal } from '../EventModal'
import type { CalendarEvent } from '@/types/calendar'

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      'modal.newTitle': 'New Event',
      'modal.editTitle': 'Edit Event',
      'modal.viewTitle': 'View Event',
      'modal.newDescription': 'Create a new event',
      'modal.editDescription': 'Edit event details',
      'modal.viewDescription': 'Event details',
      'labels.title': 'Title',
      'labels.titleRequired': 'Title *',
      'labels.eventType': 'Event Type',
      'labels.allDay': 'All Day',
      'labels.startDate': 'Start Date',
      'labels.startDateRequired': 'Start Date *',
      'labels.startTime': 'Start Time',
      'labels.startTimeRequired': 'Start Time *',
      'labels.endDate': 'End Date',
      'labels.endTime': 'End Time',
      'labels.location': 'Location',
      'labels.color': 'Color',
      'labels.description': 'Description',
      'form.eventNamePlaceholder': 'Enter event name',
      'form.selectTypePlaceholder': 'Select type',
      'form.locationPlaceholder': 'Add location',
      'form.eventDescriptionPlaceholder': 'Add description',
      'placeholders.selectDate': 'Select date',
      'buttons.delete': 'Delete',
      'buttons.close': 'Close',
      'buttons.cancel': 'Cancel',
      'buttons.saving': 'Saving...',
      'buttons.save': 'Save',
      'buttons.create': 'Create',
      'colors.default': 'Default',
      'eventTypes.meeting': 'Meeting',
      'eventTypes.deadline': 'Deadline',
      'eventTypes.task': 'Task',
      'eventTypes.reminder': 'Reminder',
      'eventTypes.holiday': 'Holiday',
      'eventTypes.personal': 'Personal',
    }
    return translations[key] || key
  },
  useLocale: () => 'en',
}))

// Mock next/dynamic to return Calendar synchronously
jest.mock('next/dynamic', () => () => {
  const MockCalendar = () => <div data-testid="calendar">Calendar</div>
  return MockCalendar
})

const mockEvent: CalendarEvent = {
  id: 1,
  title: 'Existing Meeting',
  description: 'Team sync',
  start_time: '2024-06-15T10:00:00',
  end_time: '2024-06-15T11:00:00',
  all_day: false,
  event_type: 'meeting',
  status: 'scheduled',
  timezone: 'Europe/Moscow',
  organizer_id: 1,
  is_recurring: false,
  priority: 1,
  created_at: '2024-01-01T00:00:00',
  updated_at: '2024-01-01T00:00:00',
  location: 'Conference Room A',
}

describe('EventModal', () => {
  const defaultProps = {
    open: true,
    onOpenChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders new event modal', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('New Event')).toBeInTheDocument()
  })

  it('renders edit event modal with event data', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} />)
    expect(screen.getByText('Edit Event')).toBeInTheDocument()
  })

  it('renders view-only modal when onSubmit is not provided', () => {
    render(<EventModal {...defaultProps} event={mockEvent} />)
    expect(screen.getByText('View Event')).toBeInTheDocument()
  })

  it('renders title label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Title *')).toBeInTheDocument()
  })

  it('renders description label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Description')).toBeInTheDocument()
  })

  it('renders all day checkbox label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('All Day')).toBeInTheDocument()
  })

  it('renders create button for new event', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Create')).toBeInTheDocument()
  })

  it('renders cancel button', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })

  it('renders save button when editing existing event', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} />)
    expect(screen.getByText('Save')).toBeInTheDocument()
  })

  it('renders delete button when editing existing event', () => {
    const onSubmit = jest.fn()
    const onDelete = jest.fn()
    render(
      <EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} onDelete={onDelete} />
    )
    expect(screen.getByText('Delete')).toBeInTheDocument()
  })

  it('does not render delete button for new event', () => {
    const onSubmit = jest.fn()
    const onDelete = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} onDelete={onDelete} />)
    expect(screen.queryByText('Delete')).not.toBeInTheDocument()
  })

  it('calls onOpenChange when cancel is clicked', async () => {
    const user = userEvent.setup()
    const onOpenChange = jest.fn()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onOpenChange={onOpenChange} onSubmit={onSubmit} />)

    await user.click(screen.getByText('Cancel'))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('pre-fills form with event title when editing', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} />)

    const titleInput = screen.getByDisplayValue('Existing Meeting')
    expect(titleInput).toBeInTheDocument()
  })

  it('pre-fills description when editing', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} />)

    const descInput = screen.getByDisplayValue('Team sync')
    expect(descInput).toBeInTheDocument()
  })

  it('pre-fills location when editing', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} />)

    const locationInput = screen.getByDisplayValue('Conference Room A')
    expect(locationInput).toBeInTheDocument()
  })

  it('shows event type label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Event Type')).toBeInTheDocument()
  })

  it('shows location label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Location')).toBeInTheDocument()
  })

  it('shows color label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Color')).toBeInTheDocument()
  })

  it('shows start date label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Start Date *')).toBeInTheDocument()
  })

  it('shows start time label when not all day', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('Start Time *')).toBeInTheDocument()
  })

  it('calls onDelete when delete button is clicked', async () => {
    const user = userEvent.setup()
    const onDelete = jest.fn().mockResolvedValue(undefined)
    const onSubmit = jest.fn()
    render(
      <EventModal {...defaultProps} event={mockEvent} onSubmit={onSubmit} onDelete={onDelete} />
    )

    await user.click(screen.getByText('Delete'))
    await waitFor(() => {
      expect(onDelete).toHaveBeenCalledWith(1)
    })
  })

  it('renders close button in view-only mode', () => {
    render(<EventModal {...defaultProps} event={mockEvent} />)
    // Close text appears in dialog close button and custom close button
    expect(screen.getAllByText('Close').length).toBeGreaterThan(0)
  })

  it('uses initialDate when provided for new event', () => {
    const onSubmit = jest.fn()
    const initialDate = new Date(2024, 5, 20, 14, 30)
    render(<EventModal {...defaultProps} initialDate={initialDate} onSubmit={onSubmit} />)

    const timeInput = screen.getByDisplayValue('14:30')
    expect(timeInput).toBeInTheDocument()
  })

  it('renders end date label', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('End Date')).toBeInTheDocument()
  })

  it('renders end time label when not all day', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)
    expect(screen.getByText('End Time')).toBeInTheDocument()
  })
})
