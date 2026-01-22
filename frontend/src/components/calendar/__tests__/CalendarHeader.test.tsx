import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CalendarHeader } from '../CalendarHeader'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      month: 'Month',
      week: 'Week',
      day: 'Day',
      today: 'Today',
      previous: 'Previous',
      next: 'Next',
      newEvent: 'New Event',
      weekNumber: 'Week',
    }
    return translations[key] || key
  },
  useLocale: () => 'en',
}))

describe('CalendarHeader', () => {
  const defaultProps = {
    currentDate: new Date(2024, 5, 15), // June 15, 2024
    view: 'month' as const,
    onDateChange: jest.fn(),
    onViewChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders header with current month in month view', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.getByText(/June 2024/i)).toBeInTheDocument()
  })

  it('renders header with current date in day view', () => {
    render(<CalendarHeader {...defaultProps} view="day" />)
    expect(screen.getByText(/15 June 2024/i)).toBeInTheDocument()
  })

  it('renders navigation buttons', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.getByRole('button', { name: /previous/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument()
    expect(screen.getByText('Today')).toBeInTheDocument()
  })

  it('calls onDateChange with previous month when previous button is clicked in month view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /previous/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getMonth()).toBe(4) // May (0-indexed)
  })

  it('calls onDateChange with next month when next button is clicked in month view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getMonth()).toBe(6) // July (0-indexed)
  })

  it('calls onDateChange with previous week when previous button is clicked in week view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} view="week" onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /previous/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getDate()).toBe(8) // Previous week
  })

  it('calls onDateChange with next week when next button is clicked in week view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} view="week" onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getDate()).toBe(22) // Next week
  })

  it('calls onDateChange with previous day when previous button is clicked in day view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} view="day" onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /previous/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getDate()).toBe(14) // Previous day
  })

  it('calls onDateChange with next day when next button is clicked in day view', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} view="day" onDateChange={onDateChange} />)

    await user.click(screen.getByRole('button', { name: /next/i }))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    expect(newDate.getDate()).toBe(16) // Next day
  })

  it('calls onDateChange with today when today button is clicked', async () => {
    const user = userEvent.setup()
    const onDateChange = jest.fn()
    render(<CalendarHeader {...defaultProps} onDateChange={onDateChange} />)

    await user.click(screen.getByText('Today'))
    expect(onDateChange).toHaveBeenCalled()
    const newDate = onDateChange.mock.calls[0][0]
    const today = new Date()
    expect(newDate.getDate()).toBe(today.getDate())
  })

  it('renders view tabs', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.getByRole('tab', { name: /month/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /week/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /day/i })).toBeInTheDocument()
  })

  it('calls onViewChange when view tab is clicked', async () => {
    const user = userEvent.setup()
    const onViewChange = jest.fn()
    render(<CalendarHeader {...defaultProps} onViewChange={onViewChange} />)

    await user.click(screen.getByRole('tab', { name: /week/i }))
    expect(onViewChange).toHaveBeenCalledWith('week')
  })

  it('renders add event button when onAddEvent is provided', () => {
    const onAddEvent = jest.fn()
    render(<CalendarHeader {...defaultProps} onAddEvent={onAddEvent} />)
    expect(screen.getByText('New Event')).toBeInTheDocument()
  })

  it('does not render add event button when onAddEvent is not provided', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.queryByText('New Event')).not.toBeInTheDocument()
  })

  it('calls onAddEvent when add event button is clicked', async () => {
    const user = userEvent.setup()
    const onAddEvent = jest.fn()
    render(<CalendarHeader {...defaultProps} onAddEvent={onAddEvent} />)

    await user.click(screen.getByText('New Event'))
    expect(onAddEvent).toHaveBeenCalled()
  })

  it('applies custom className', () => {
    const { container } = render(<CalendarHeader {...defaultProps} className="custom-class" />)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('displays day number from current date', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('displays abbreviated month', () => {
    render(<CalendarHeader {...defaultProps} />)
    expect(screen.getByText('Jun')).toBeInTheDocument()
  })
})
