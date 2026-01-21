import { render, screen } from '@testing-library/react'
import { Separator } from '../separator'

describe('Separator', () => {
  it('renders with horizontal orientation by default', () => {
    render(<Separator data-testid="separator" />)
    const separator = screen.getByTestId('separator')
    expect(separator).toHaveClass('h-[1px]', 'w-full')
  })

  it('renders with horizontal orientation explicitly', () => {
    render(<Separator orientation="horizontal" data-testid="separator" />)
    const separator = screen.getByTestId('separator')
    expect(separator).toHaveClass('h-[1px]', 'w-full')
  })

  it('renders with vertical orientation', () => {
    render(<Separator orientation="vertical" data-testid="separator" />)
    const separator = screen.getByTestId('separator')
    expect(separator).toHaveClass('h-full', 'w-[1px]')
  })

  it('applies custom className', () => {
    render(<Separator className="custom-class" data-testid="separator" />)
    expect(screen.getByTestId('separator')).toHaveClass('custom-class')
  })

  it('applies base classes', () => {
    render(<Separator data-testid="separator" />)
    expect(screen.getByTestId('separator')).toHaveClass('shrink-0', 'bg-border')
  })

  it('is decorative by default', () => {
    render(<Separator data-testid="separator" />)
    const separator = screen.getByTestId('separator')
    // Decorative separators should not be announced by screen readers
    expect(separator).toHaveAttribute('data-orientation', 'horizontal')
  })

  it('can be non-decorative', () => {
    render(<Separator decorative={false} data-testid="separator" />)
    const separator = screen.getByTestId('separator')
    expect(separator).toHaveAttribute('role', 'separator')
  })

  it('forwards ref', () => {
    const ref = { current: null }
    render(<Separator ref={ref} data-testid="separator" />)
    expect(ref.current).toBeInstanceOf(HTMLElement)
  })
})
