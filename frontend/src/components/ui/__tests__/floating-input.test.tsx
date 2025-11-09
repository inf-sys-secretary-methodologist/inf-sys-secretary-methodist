import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FloatingInput } from '../floating-input'

describe('FloatingInput', () => {
  it('renders with a label', () => {
    render(<FloatingInput label="Email" />)
    const label = screen.getByText('Email')
    expect(label).toBeInTheDocument()
  })

  it('renders an input element', () => {
    render(<FloatingInput label="Email" />)
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()
  })

  it('generates unique id for input and label', () => {
    render(<FloatingInput label="Email" />)
    const input = screen.getByRole('textbox')
    const label = screen.getByText('Email')

    expect(input).toHaveAttribute('id')
    expect(label).toHaveAttribute('for', input.getAttribute('id'))
  })

  it('has placeholder space by default', () => {
    render(<FloatingInput label="Email" />)
    const input = screen.getByRole('textbox')
    expect(input).toHaveAttribute('placeholder', ' ')
  })

  it('applies custom className to input', () => {
    render(<FloatingInput label="Email" className="custom-input" />)
    const input = screen.getByRole('textbox')
    expect(input).toHaveClass('custom-input')
  })

  it('applies containerClassName to wrapper', () => {
    const { container } = render(
      <FloatingInput label="Email" containerClassName="custom-container" />
    )
    const wrapper = container.firstChild
    expect(wrapper).toHaveClass('custom-container')
  })

  it('handles user input', async () => {
    const user = userEvent.setup()
    render(<FloatingInput label="Email" />)
    const input = screen.getByRole('textbox')

    await user.type(input, 'test@example.com')
    expect(input).toHaveValue('test@example.com')
  })

  it('label has proper animation classes', () => {
    render(<FloatingInput label="Email" />)
    const label = screen.getByText('Email').parentElement

    expect(label).toHaveClass('absolute')
    expect(label).toHaveClass('top-1/2')
    expect(label).toHaveClass('-translate-y-1/2')
    expect(label).toHaveClass('transition-all')
    expect(label).toHaveClass('duration-200')
  })

  it('label animates on focus using group-focus-within', () => {
    render(<FloatingInput label="Email" />)
    const label = screen.getByText('Email').parentElement

    expect(label).toHaveClass('group-focus-within:top-0')
    expect(label).toHaveClass('group-focus-within:text-xs')
    expect(label).toHaveClass('group-focus-within:font-medium')
  })

  it('label stays up when input has value using has selector', () => {
    render(<FloatingInput label="Email" />)
    const label = screen.getByText('Email').parentElement

    expect(label).toHaveClass('has-[+input:not(:placeholder-shown)]:top-0')
    expect(label).toHaveClass('has-[+input:not(:placeholder-shown)]:text-xs')
  })

  it('input has peer class for CSS animations', () => {
    render(<FloatingInput label="Email" />)
    const input = screen.getByRole('textbox')
    expect(input).toHaveClass('peer')
  })

  it('container has group class for focus-within', () => {
    const { container } = render(<FloatingInput label="Email" />)
    const wrapper = container.firstChild
    expect(wrapper).toHaveClass('group')
    expect(wrapper).toHaveClass('relative')
  })

  it('container has minimum width', () => {
    const { container } = render(<FloatingInput label="Email" />)
    const wrapper = container.firstChild
    expect(wrapper).toHaveClass('min-w-[200px]')
  })

  it('forwards input props correctly', async () => {
    const handleChange = jest.fn()
    const user = userEvent.setup()

    render(
      <FloatingInput
        label="Email"
        type="email"
        onChange={handleChange}
        disabled={false}
      />
    )

    const input = screen.getByRole('textbox')
    expect(input).toHaveAttribute('type', 'email')

    await user.type(input, 'a')
    expect(handleChange).toHaveBeenCalled()
  })

  it('label background prevents border overlap', () => {
    render(<FloatingInput label="Email" />)
    const labelSpan = screen.getByText('Email')

    expect(labelSpan).toHaveClass('bg-background')
    expect(labelSpan).toHaveClass('px-2')
  })

  it('handles disabled state', () => {
    render(<FloatingInput label="Email" disabled />)
    const input = screen.getByRole('textbox')
    expect(input).toBeDisabled()
  })

  it('renders multiple instances with unique ids', () => {
    render(
      <>
        <FloatingInput label="Email" />
        <FloatingInput label="Password" />
      </>
    )

    const inputs = screen.getAllByRole('textbox')
    expect(inputs).toHaveLength(2)
    expect(inputs[0].getAttribute('id')).not.toBe(inputs[1].getAttribute('id'))
  })
})
