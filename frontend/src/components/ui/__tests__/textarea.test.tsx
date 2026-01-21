import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Textarea } from '../textarea'

describe('Textarea', () => {
  it('renders textarea element', () => {
    render(<Textarea placeholder="Enter text" />)
    expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<Textarea className="custom-class" placeholder="Text" />)
    expect(screen.getByPlaceholderText('Text')).toHaveClass('custom-class')
  })

  it('applies default classes', () => {
    render(<Textarea placeholder="Text" />)
    const textarea = screen.getByPlaceholderText('Text')
    expect(textarea).toHaveClass('flex', 'min-h-[60px]', 'w-full', 'rounded-md', 'border')
  })

  it('handles value changes', async () => {
    const handleChange = jest.fn()
    const user = userEvent.setup()

    render(<Textarea placeholder="Text" onChange={handleChange} />)
    const textarea = screen.getByPlaceholderText('Text')

    await user.type(textarea, 'test content')

    expect(handleChange).toHaveBeenCalled()
    expect(textarea).toHaveValue('test content')
  })

  it('handles disabled state', () => {
    render(<Textarea disabled placeholder="Disabled" />)
    expect(screen.getByPlaceholderText('Disabled')).toBeDisabled()
  })

  it('handles readonly state', () => {
    render(<Textarea readOnly value="readonly value" placeholder="Readonly" />)
    expect(screen.getByPlaceholderText('Readonly')).toHaveAttribute('readonly')
  })

  it('forwards ref', () => {
    const ref = { current: null }
    render(<Textarea ref={ref} placeholder="Text" />)
    expect(ref.current).toBeInstanceOf(HTMLTextAreaElement)
  })

  it('supports rows attribute', () => {
    render(<Textarea rows={5} placeholder="Text" />)
    expect(screen.getByPlaceholderText('Text')).toHaveAttribute('rows', '5')
  })

  it('renders with aria-label', () => {
    render(<Textarea aria-label="Description input" />)
    expect(screen.getByLabelText('Description input')).toBeInTheDocument()
  })
})
