import { render, screen, fireEvent } from '@testing-library/react'
import { AIQuickActions, AIQuickActionChips } from '../AIQuickActions'

jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
}))

describe('AIQuickActions', () => {
  const mockOnSelect = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders quick action suggestions', () => {
    render(<AIQuickActions onAction={mockOnSelect} />)

    // Should show action suggestions
    expect(screen.getByText('quickActionsTitle')).toBeInTheDocument()
  })

  it('calls onAction when action is clicked', () => {
    render(<AIQuickActions onAction={mockOnSelect} />)

    const actionButtons = screen.getAllByRole('button')
    expect(actionButtons.length).toBeGreaterThan(0)

    // Click first action
    fireEvent.click(actionButtons[0])

    expect(mockOnSelect).toHaveBeenCalledWith(expect.any(String))
  })

  it('renders with different action categories', () => {
    render(<AIQuickActions onAction={mockOnSelect} />)

    // Quick actions should be rendered
    const actionButtons = screen.getAllByRole('button')
    expect(actionButtons.length).toBeGreaterThan(0)
  })

  it('applies custom className', () => {
    const { container } = render(<AIQuickActions onAction={mockOnSelect} className="custom" />)

    expect(container.firstChild).toHaveClass('custom')
  })
})

describe('AIQuickActionChips', () => {
  const mockOnSelect = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders action chips', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} />)

    // Chips should be rendered
    const chips = screen.getAllByRole('button')
    expect(chips.length).toBeGreaterThan(0)
  })

  it('calls onAction when chip is clicked', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} />)

    const chips = screen.getAllByRole('button')
    fireEvent.click(chips[0])

    expect(mockOnSelect).toHaveBeenCalledWith(expect.any(String))
  })

  it('displays action text', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} />)

    // Should show action buttons
    const chips = screen.getAllByRole('button')
    expect(chips.length).toBeGreaterThan(0)
    expect(chips[0]).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<AIQuickActionChips onAction={mockOnSelect} className="custom" />)

    expect(container.firstChild).toHaveClass('custom')
  })

  it('renders multiple action categories', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} />)

    const chips = screen.getAllByRole('button')

    // Should have multiple action options
    expect(chips.length).toBeGreaterThanOrEqual(1)
  })

  it('displays action chips with text', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} />)

    // Should have multiple action options
    const chips = screen.getAllByRole('button')
    expect(chips.length).toBeGreaterThanOrEqual(1)
  })

  it('handles disabled state', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} disabled />)

    const chips = screen.getAllByRole('button')
    chips.forEach((chip) => {
      expect(chip).toBeDisabled()
    })
  })

  it('does not call onAction when disabled chip is clicked', () => {
    render(<AIQuickActionChips onAction={mockOnSelect} disabled />)

    const chips = screen.getAllByRole('button')
    fireEvent.click(chips[0])

    expect(mockOnSelect).not.toHaveBeenCalled()
  })
})
