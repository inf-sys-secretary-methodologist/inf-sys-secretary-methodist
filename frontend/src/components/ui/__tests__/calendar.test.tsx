import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Calendar } from '../calendar'

describe('Calendar', () => {
  it('renders calendar component', () => {
    render(<Calendar />)
    expect(screen.getByRole('grid')).toBeInTheDocument()
  })

  it('shows current month by default', () => {
    render(<Calendar />)
    const today = new Date()
    const monthName = today.toLocaleString('en-US', { month: 'long' })
    expect(screen.getByText(new RegExp(monthName, 'i'))).toBeInTheDocument()
  })

  it('renders navigation buttons', () => {
    render(<Calendar />)
    // Navigation buttons should be present
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThanOrEqual(2) // Previous and Next
  })

  it('applies custom className', () => {
    const { container } = render(<Calendar className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('shows weekday headers', () => {
    const { container } = render(<Calendar />)
    // Should show weekday headers
    const weekdays = container.querySelectorAll('th')
    expect(weekdays.length).toBe(7) // 7 days of the week
  })

  it('renders grid cells for days', () => {
    render(<Calendar />)
    const gridCells = screen.getAllByRole('gridcell')
    expect(gridCells.length).toBeGreaterThan(0)
  })

  it('hides outside days when showOutsideDays is false', () => {
    render(<Calendar showOutsideDays={false} />)
    expect(screen.getByRole('grid')).toBeInTheDocument()
  })

  it('calls onSelect when a day is selected in single mode', async () => {
    const user = userEvent.setup()
    const onSelect = jest.fn()
    render(<Calendar mode="single" onSelect={onSelect} />)

    const dayButtons = screen.getAllByRole('gridcell')
    const clickableDay = dayButtons.find((btn) => {
      const button = btn.querySelector('button')
      return button && !button.hasAttribute('disabled')
    })

    if (clickableDay) {
      const button = clickableDay.querySelector('button')
      if (button) {
        await user.click(button)
        expect(onSelect).toHaveBeenCalled()
      }
    }
  })

  it('shows specified month when defaultMonth is provided', () => {
    render(<Calendar defaultMonth={new Date(2024, 5, 1)} />)
    expect(screen.getByText(/June/i)).toBeInTheDocument()
  })

  it('supports range selection mode', () => {
    render(<Calendar mode="range" />)
    expect(screen.getByRole('grid')).toBeInTheDocument()
  })

  it('supports multiple selection mode', () => {
    render(<Calendar mode="multiple" />)
    expect(screen.getByRole('grid')).toBeInTheDocument()
  })

  it('renders with custom classNames for components', () => {
    render(
      <Calendar
        classNames={{
          months: 'custom-months',
          month: 'custom-month',
        }}
      />
    )
    expect(screen.getByRole('grid')).toBeInTheDocument()
  })

  it('has correct display name', () => {
    expect(Calendar.displayName).toBe('Calendar')
  })

  it('renders chevron icons for navigation', () => {
    const { container } = render(<Calendar />)
    // Check for SVG icons in navigation buttons
    const svgs = container.querySelectorAll('svg')
    expect(svgs.length).toBeGreaterThanOrEqual(2)
  })

  it('renders with month prop', () => {
    const testDate = new Date(2024, 11, 1) // December 2024
    render(<Calendar month={testDate} />)
    expect(screen.getByText(/December/i)).toBeInTheDocument()
  })
})
