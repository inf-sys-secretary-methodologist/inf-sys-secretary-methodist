import { render, screen } from '@testing-library/react'
import { Label } from '../label'

describe('Label', () => {
  it('renders with children', () => {
    render(<Label>Label text</Label>)
    expect(screen.getByText('Label text')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<Label className="custom-class">Label</Label>)
    expect(screen.getByText('Label')).toHaveClass('custom-class')
  })

  it('applies default classes', () => {
    render(<Label>Label</Label>)
    const label = screen.getByText('Label')
    expect(label).toHaveClass('text-sm', 'font-medium', 'leading-none')
  })

  it('forwards ref', () => {
    const ref = { current: null }
    render(<Label ref={ref}>Label</Label>)
    expect(ref.current).toBeInstanceOf(HTMLLabelElement)
  })

  it('associates with form element via htmlFor', () => {
    render(
      <>
        <Label htmlFor="test-input">Test Label</Label>
        <input id="test-input" />
      </>
    )

    const label = screen.getByText('Test Label')
    expect(label).toHaveAttribute('for', 'test-input')
  })

  it('forwards additional props', () => {
    render(<Label data-testid="test-label">Test</Label>)
    expect(screen.getByTestId('test-label')).toBeInTheDocument()
  })
})
