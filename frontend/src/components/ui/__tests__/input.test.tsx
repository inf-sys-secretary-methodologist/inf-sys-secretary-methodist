import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../input'

describe('Input', () => {
  it('renders input element', () => {
    render(<Input placeholder="Enter text" />)
    const input = screen.getByPlaceholderText('Enter text')
    expect(input).toBeInTheDocument()
    expect(input.tagName).toBe('INPUT')
  })

  it('renders with specified type', () => {
    render(<Input type="email" placeholder="Enter email" />)
    const input = screen.getByPlaceholderText('Enter email')
    expect(input).toHaveAttribute('type', 'email')
  })

  it('renders password input', () => {
    render(<Input type="password" placeholder="Enter password" />)
    const input = screen.getByPlaceholderText('Enter password')
    expect(input).toHaveAttribute('type', 'password')
  })

  it('applies custom className', () => {
    render(<Input className="custom-class" placeholder="Input" />)
    const input = screen.getByPlaceholderText('Input')
    expect(input).toHaveClass('custom-class')
  })

  it('applies default classes', () => {
    render(<Input placeholder="Input" />)
    const input = screen.getByPlaceholderText('Input')
    expect(input).toHaveClass('flex', 'h-9', 'w-full', 'rounded-md', 'border')
  })

  it('handles value changes', async () => {
    const handleChange = jest.fn()
    const user = userEvent.setup()

    render(<Input placeholder="Input" onChange={handleChange} />)
    const input = screen.getByPlaceholderText('Input')

    await user.type(input, 'test')

    expect(handleChange).toHaveBeenCalled()
    expect(input).toHaveValue('test')
  })

  it('handles disabled state', () => {
    render(<Input disabled placeholder="Disabled input" />)
    const input = screen.getByPlaceholderText('Disabled input')
    expect(input).toBeDisabled()
  })

  it('handles readonly state', () => {
    render(<Input readOnly value="readonly value" placeholder="Readonly" />)
    const input = screen.getByPlaceholderText('Readonly')
    expect(input).toHaveAttribute('readonly')
  })

  it('forwards ref', () => {
    const ref = { current: null }
    render(<Input ref={ref} placeholder="Input" />)
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })

  it('handles focus and blur', async () => {
    const handleFocus = jest.fn()
    const handleBlur = jest.fn()
    const user = userEvent.setup()

    render(<Input placeholder="Input" onFocus={handleFocus} onBlur={handleBlur} />)
    const input = screen.getByPlaceholderText('Input')

    await user.click(input)
    expect(handleFocus).toHaveBeenCalledTimes(1)

    await user.tab()
    expect(handleBlur).toHaveBeenCalledTimes(1)
  })

  it('renders with aria-label', () => {
    render(<Input aria-label="Search input" />)
    expect(screen.getByLabelText('Search input')).toBeInTheDocument()
  })
})
