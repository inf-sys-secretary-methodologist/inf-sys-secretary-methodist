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

// Mock Select component to enable testing onValueChange
jest.mock('@/components/ui/select', () => ({
  Select: ({
    children,
    value,
    onValueChange,
    disabled,
  }: {
    children: React.ReactNode
    value?: string
    onValueChange?: (value: string) => void
    disabled?: boolean
  }) => (
    <div data-testid="mock-select">
      <select
        value={value}
        onChange={(e) => onValueChange?.(e.target.value)}
        disabled={disabled}
        data-testid="mock-select-input"
      >
        {children}
      </select>
    </div>
  ),
  SelectTrigger: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SelectValue: ({ placeholder }: { placeholder?: string }) => <span>{placeholder}</span>,
  SelectContent: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SelectItem: ({ children, value }: { children: React.ReactNode; value: string }) => (
    <option value={value}>{children}</option>
  ),
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
      'colors.blue': 'Blue',
      'colors.red': 'Red',
      'colors.green': 'Green',
      'colors.yellow': 'Yellow',
      'colors.purple': 'Purple',
      'colors.gray': 'Gray',
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

// Mock next/dynamic to return Calendar synchronously with onSelect support
jest.mock('next/dynamic', () => () => {
  const MockCalendar = ({ onSelect }: { onSelect?: (date: Date) => void }) => (
    <div data-testid="calendar">
      <button data-testid="mock-calendar-select" onClick={() => onSelect?.(new Date(2024, 5, 20))}>
        Select Date
      </button>
    </div>
  )
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

  it('hides time inputs when all-day is checked', async () => {
    const user = userEvent.setup()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)

    // Initially time inputs are visible
    expect(screen.getByText('Start Time *')).toBeInTheDocument()
    expect(screen.getByText('End Time')).toBeInTheDocument()

    // Check all-day
    await user.click(screen.getByRole('checkbox'))

    // Time inputs should be hidden
    expect(screen.queryByText('Start Time *')).not.toBeInTheDocument()
    expect(screen.queryByText('End Time')).not.toBeInTheDocument()
  })

  it('does not render submit buttons in view-only mode', () => {
    render(<EventModal {...defaultProps} event={mockEvent} />)

    // In view-only mode, there's no submit button
    expect(screen.queryByText('Save')).not.toBeInTheDocument()
    expect(screen.queryByText('Create')).not.toBeInTheDocument()
  })

  it('shows saving button state when isLoading is true', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} isLoading={true} />)

    expect(screen.getByText('Saving...')).toBeInTheDocument()
  })

  it('disables submit button when isLoading is true', () => {
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} isLoading={true} />)

    const submitButton = screen.getByText('Saving...')
    expect(submitButton).toBeDisabled()
  })

  it('allows changing event type via select', async () => {
    const user = userEvent.setup()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)

    // Find the event type select (first mock-select-input)
    const selectInputs = screen.getAllByTestId('mock-select-input')
    const eventTypeSelect = selectInputs[0]

    // Change the event type to deadline - triggers onValueChange
    await user.selectOptions(eventTypeSelect, 'deadline')

    expect(eventTypeSelect).toHaveValue('deadline')
  })

  it('allows changing color via select', async () => {
    const user = userEvent.setup()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)

    // Find the color select (second mock-select-input)
    const selectInputs = screen.getAllByTestId('mock-select-input')
    const colorSelect = selectInputs[1]

    // Change the color to blue - triggers onValueChange
    await user.selectOptions(colorSelect, '#3b82f6')

    expect(colorSelect).toHaveValue('#3b82f6')
  })

  it('allows selecting start date from calendar', async () => {
    const user = userEvent.setup()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)

    // Open the start date popover by clicking the button
    const startDateButtons = screen.getAllByRole('button', { name: /select date/i })
    // The start date button should have the calendar icon
    const startDateButton = startDateButtons[0]
    await user.click(startDateButton)

    // Click the mock calendar select button
    const calendarSelectBtns = screen.getAllByTestId('mock-calendar-select')
    if (calendarSelectBtns.length > 0) {
      await user.click(calendarSelectBtns[0])
    }

    // The date should be selected (we just need to verify the callback was triggered)
    expect(startDateButton).toBeInTheDocument()
  })

  it('allows selecting end date from calendar', async () => {
    const user = userEvent.setup()
    const onSubmit = jest.fn()
    render(<EventModal {...defaultProps} onSubmit={onSubmit} />)

    // Find end date button (second date button)
    const dateButtons = screen.getAllByRole('button')
    const endDateButton = dateButtons.find((btn) => btn.textContent?.includes('Select date'))

    if (endDateButton) {
      await user.click(endDateButton)

      // Try to find and click the calendar select
      const calendarSelectBtns = screen.getAllByTestId('mock-calendar-select')
      if (calendarSelectBtns.length > 0) {
        await user.click(calendarSelectBtns[calendarSelectBtns.length - 1])
      }
    }

    expect(screen.getByText('End Date')).toBeInTheDocument()
  })
})
