import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../input'

describe('Input', () => {
  it('renders an input element', () => {
    render(<Input />)
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()
  })

  it('forwards ref correctly', () => {
    const ref = { current: null }
    render(<Input ref={ref} />)
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })

  it('applies custom className', () => {
    render(<Input className="custom-class" />)
    const input = screen.getByRole('textbox')
    expect(input).toHaveClass('custom-class')
  })

  it('handles different input types', () => {
    const { rerender } = render(<Input type="text" />)
    let input = screen.getByRole('textbox')
    expect(input).toHaveAttribute('type', 'text')

    rerender(<Input type="email" />)
    input = screen.getByRole('textbox')
    expect(input).toHaveAttribute('type', 'email')

    rerender(<Input type="password" />)
    input = screen.getByRole('textbox', { hidden: true }) || document.querySelector('input[type="password"]')
    expect(input).toHaveAttribute('type', 'password')
  })

  it('handles search input type with special styling', () => {
    const { container } = render(<Input type="search" />)
    const input = container.querySelector('input[type="search"]')
    expect(input).toHaveClass('[&::-webkit-search-cancel-button]:appearance-none')
  })

  it('handles disabled state', () => {
    render(<Input disabled />)
    const input = screen.getByRole('textbox')
    expect(input).toBeDisabled()
    expect(input).toHaveClass('disabled:cursor-not-allowed')
    expect(input).toHaveClass('disabled:opacity-50')
  })

  it('handles user input', async () => {
    const user = userEvent.setup()
    render(<Input />)
    const input = screen.getByRole('textbox')

    await user.type(input, 'Hello World')
    expect(input).toHaveValue('Hello World')
  })

  it('handles placeholder', () => {
    render(<Input placeholder="Enter text..." />)
    const input = screen.getByRole('textbox')
    expect(input).toHaveAttribute('placeholder', 'Enter text...')
  })

  it('handles onChange callback', async () => {
    const handleChange = jest.fn()
    const user = userEvent.setup()
    render(<Input onChange={handleChange} />)
    const input = screen.getByRole('textbox')

    await user.type(input, 'a')
    expect(handleChange).toHaveBeenCalledTimes(1)
  })

  it('applies focus-visible styles on focus', async () => {
    const user = userEvent.setup()
    render(<Input />)
    const input = screen.getByRole('textbox')

    await user.click(input)
    expect(input).toHaveFocus()
    expect(input).toHaveClass('focus-visible:border-ring')
    expect(input).toHaveClass('focus-visible:ring-[3px]')
  })

  it('has proper default styling classes', () => {
    render(<Input />)
    const input = screen.getByRole('textbox')

    expect(input).toHaveClass('flex')
    expect(input).toHaveClass('h-10')
    expect(input).toHaveClass('w-full')
    expect(input).toHaveClass('rounded-lg')
    expect(input).toHaveClass('border')
    expect(input).toHaveClass('bg-background')
    expect(input).toHaveClass('text-foreground')
  })

  it('passes through additional props', () => {
    render(<Input data-testid="custom-input" aria-label="Custom input" />)
    const input = screen.getByTestId('custom-input')
    expect(input).toHaveAttribute('aria-label', 'Custom input')
  })
})
