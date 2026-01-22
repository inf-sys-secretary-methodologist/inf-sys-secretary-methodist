import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Switch } from '../switch'

describe('Switch', () => {
  it('renders switch component', () => {
    render(<Switch />)
    expect(screen.getByRole('switch')).toBeInTheDocument()
  })

  it('is unchecked by default', () => {
    render(<Switch />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')
  })

  it('can be checked by default', () => {
    render(<Switch defaultChecked />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  })

  it('toggles when clicked', async () => {
    const user = userEvent.setup()
    render(<Switch />)

    const switchEl = screen.getByRole('switch')
    expect(switchEl).toHaveAttribute('aria-checked', 'false')

    await user.click(switchEl)
    expect(switchEl).toHaveAttribute('aria-checked', 'true')

    await user.click(switchEl)
    expect(switchEl).toHaveAttribute('aria-checked', 'false')
  })

  it('calls onCheckedChange when toggled', async () => {
    const user = userEvent.setup()
    const onCheckedChange = jest.fn()
    render(<Switch onCheckedChange={onCheckedChange} />)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).toHaveBeenCalledWith(true)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).toHaveBeenCalledWith(false)
  })

  it('applies custom className', () => {
    render(<Switch className="custom-switch" />)
    expect(screen.getByRole('switch')).toHaveClass('custom-switch')
  })

  it('is disabled when disabled prop is true', () => {
    render(<Switch disabled />)
    expect(screen.getByRole('switch')).toBeDisabled()
  })

  it('does not toggle when disabled', async () => {
    const user = userEvent.setup()
    const onCheckedChange = jest.fn()
    render(<Switch disabled onCheckedChange={onCheckedChange} />)

    await user.click(screen.getByRole('switch'))
    expect(onCheckedChange).not.toHaveBeenCalled()
  })

  it('works with controlled checked state', () => {
    const { rerender } = render(<Switch checked={false} />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')

    rerender(<Switch checked={true} />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  })

  it('has correct display name', () => {
    expect(Switch.displayName).toBe('Switch')
  })

  it('supports value attribute', () => {
    render(<Switch value="on" />)
    expect(screen.getByRole('switch')).toHaveAttribute('value', 'on')
  })

  it('supports required attribute via aria-required', () => {
    render(<Switch required />)
    expect(screen.getByRole('switch')).toHaveAttribute('aria-required', 'true')
  })

  it('toggles with keyboard space', async () => {
    const user = userEvent.setup()
    render(<Switch />)

    const switchEl = screen.getByRole('switch')
    switchEl.focus()
    expect(switchEl).toHaveAttribute('aria-checked', 'false')

    await user.keyboard(' ')
    expect(switchEl).toHaveAttribute('aria-checked', 'true')
  })

  it('has accessible label when associated with label element', () => {
    render(
      <div>
        <label htmlFor="test-switch">Enable notifications</label>
        <Switch id="test-switch" />
      </div>
    )

    expect(screen.getByRole('switch', { name: 'Enable notifications' })).toBeInTheDocument()
  })

  it('renders thumb element', () => {
    const { container } = render(<Switch />)
    // Thumb is rendered with specific classes
    expect(container.querySelector('.pointer-events-none')).toBeInTheDocument()
  })
})
