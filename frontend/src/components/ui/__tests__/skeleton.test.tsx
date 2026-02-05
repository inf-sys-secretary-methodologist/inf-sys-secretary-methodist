import { render } from '@testing-library/react'
import { Skeleton } from '../skeleton'

describe('Skeleton', () => {
  it('renders without crashing', () => {
    const { container } = render(<Skeleton />)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('applies default classes', () => {
    const { container } = render(<Skeleton />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('animate-pulse')
    expect(element).toHaveClass('rounded-md')
    expect(element).toHaveClass('bg-muted')
  })

  it('accepts custom className', () => {
    const { container } = render(<Skeleton className="h-10 w-20" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('h-10')
    expect(element).toHaveClass('w-20')
    expect(element).toHaveClass('animate-pulse')
  })

  it('passes through additional props', () => {
    const { container } = render(<Skeleton data-testid="test-skeleton" aria-label="Loading" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveAttribute('data-testid', 'test-skeleton')
    expect(element).toHaveAttribute('aria-label', 'Loading')
  })

  it('renders as a div element', () => {
    const { container } = render(<Skeleton />)
    expect(container.firstChild?.nodeName).toBe('DIV')
  })
})
