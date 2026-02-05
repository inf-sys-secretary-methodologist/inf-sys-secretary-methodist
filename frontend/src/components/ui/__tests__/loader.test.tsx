import { render, screen } from '@testing-library/react'
import {
  Loader,
  CircularLoader,
  DotsLoader,
  WaveLoader,
  TextShimmerLoader,
  TextDotsLoader,
} from '../loader'

describe('CircularLoader', () => {
  it('renders without crashing', () => {
    render(<CircularLoader />)
    expect(screen.getByText('Loading')).toBeInTheDocument()
  })

  it('applies default size (md)', () => {
    const { container } = render(<CircularLoader />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('size-5')
  })

  it('applies small size', () => {
    const { container } = render(<CircularLoader size="sm" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('size-4')
  })

  it('applies large size', () => {
    const { container } = render(<CircularLoader size="lg" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('size-6')
  })

  it('accepts custom className', () => {
    const { container } = render(<CircularLoader className="custom-class" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('custom-class')
  })

  it('has sr-only text for accessibility', () => {
    render(<CircularLoader />)
    const srText = screen.getByText('Loading')
    expect(srText).toHaveClass('sr-only')
  })
})

describe('DotsLoader', () => {
  it('renders three dots', () => {
    const { container } = render(<DotsLoader />)
    const dots = container.querySelectorAll('.rounded-full.bg-primary')
    expect(dots.length).toBe(3)
  })

  it('applies default size (md)', () => {
    const { container } = render(<DotsLoader />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('h-5')
  })

  it('applies small size', () => {
    const { container } = render(<DotsLoader size="sm" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('h-4')
  })

  it('applies large size', () => {
    const { container } = render(<DotsLoader size="lg" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('h-6')
  })

  it('has animation delays for each dot', () => {
    const { container } = render(<DotsLoader />)
    const dots = container.querySelectorAll('.rounded-full.bg-primary')
    expect(dots[0]).toHaveStyle({ animationDelay: '0ms' })
    expect(dots[1]).toHaveStyle({ animationDelay: '160ms' })
    expect(dots[2]).toHaveStyle({ animationDelay: '320ms' })
  })
})

describe('WaveLoader', () => {
  it('renders five bars', () => {
    const { container } = render(<WaveLoader />)
    const bars = container.querySelectorAll('.rounded-full.bg-primary')
    expect(bars.length).toBe(5)
  })

  it('applies default size (md)', () => {
    const { container } = render(<WaveLoader />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('h-5')
  })

  it('has varying heights for wave effect', () => {
    const { container } = render(<WaveLoader size="md" />)
    const bars = container.querySelectorAll('.rounded-full.bg-primary')
    expect(bars[0]).toHaveStyle({ height: '8px' })
    expect(bars[1]).toHaveStyle({ height: '12px' })
    expect(bars[2]).toHaveStyle({ height: '16px' })
    expect(bars[3]).toHaveStyle({ height: '12px' })
    expect(bars[4]).toHaveStyle({ height: '8px' })
  })
})

describe('TextShimmerLoader', () => {
  it('renders default text', () => {
    render(<TextShimmerLoader />)
    expect(screen.getByText('Loading')).toBeInTheDocument()
  })

  it('renders custom text', () => {
    render(<TextShimmerLoader text="Processing..." />)
    expect(screen.getByText('Processing...')).toBeInTheDocument()
  })

  it('applies default size (md)', () => {
    const { container } = render(<TextShimmerLoader />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('text-sm')
  })

  it('applies small size', () => {
    const { container } = render(<TextShimmerLoader size="sm" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('text-xs')
  })

  it('applies large size', () => {
    const { container } = render(<TextShimmerLoader size="lg" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('text-base')
  })
})

describe('TextDotsLoader', () => {
  it('renders default text with dots', () => {
    render(<TextDotsLoader />)
    expect(screen.getByText('Loading')).toBeInTheDocument()
    const dots = screen.getAllByText('.')
    expect(dots.length).toBe(3)
  })

  it('renders custom text', () => {
    render(<TextDotsLoader text="Saving" />)
    expect(screen.getByText('Saving')).toBeInTheDocument()
  })

  it('applies default size (md)', () => {
    render(<TextDotsLoader />)
    const textElement = screen.getByText('Loading')
    expect(textElement).toHaveClass('text-sm')
  })
})

describe('Loader', () => {
  it('renders circular variant by default', () => {
    const { container } = render(<Loader />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('animate-spin')
  })

  it('renders circular variant', () => {
    const { container } = render(<Loader variant="circular" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('animate-spin')
  })

  it('renders dots variant', () => {
    const { container } = render(<Loader variant="dots" />)
    const dots = container.querySelectorAll('.rounded-full.bg-primary')
    expect(dots.length).toBe(3)
  })

  it('renders wave variant', () => {
    const { container } = render(<Loader variant="wave" />)
    const bars = container.querySelectorAll('.rounded-full.bg-primary')
    expect(bars.length).toBe(5)
  })

  it('renders text-shimmer variant', () => {
    render(<Loader variant="text-shimmer" text="Loading..." />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders loading-dots variant', () => {
    render(<Loader variant="loading-dots" text="Processing" />)
    expect(screen.getByText('Processing')).toBeInTheDocument()
  })

  it('passes size to child components', () => {
    const { container } = render(<Loader variant="circular" size="lg" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('size-6')
  })

  it('passes className to child components', () => {
    const { container } = render(<Loader variant="circular" className="custom-loader" />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('custom-loader')
  })

  it('falls back to circular for unknown variants', () => {
    const { container } = render(<Loader variant={'unknown' as 'circular'} />)
    const element = container.firstChild as HTMLElement
    expect(element).toHaveClass('animate-spin')
  })
})
