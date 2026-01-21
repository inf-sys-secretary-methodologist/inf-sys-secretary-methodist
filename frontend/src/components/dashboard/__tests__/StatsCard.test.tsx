import { render, screen } from '@testing-library/react'
import { StatsCard } from '../StatsCard'
import { Users, FileText, Calendar } from 'lucide-react'

// Mock the GlowingEffect component
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => <div data-testid="glowing-effect" />,
}))

// Mock NumberTicker component
jest.mock('@/components/ui/number-ticker', () => ({
  NumberTicker: ({ value }: { value: number }) => <span data-testid="number-ticker">{value}</span>,
}))

describe('StatsCard', () => {
  const defaultProps = {
    icon: Users,
    title: 'Total Users',
    value: 1234,
    change: 12.5,
    period: 'месяц',
  }

  it('renders card with title', () => {
    render(<StatsCard {...defaultProps} />)
    expect(screen.getByText('Total Users')).toBeInTheDocument()
  })

  it('renders value with NumberTicker', () => {
    render(<StatsCard {...defaultProps} />)
    expect(screen.getByTestId('number-ticker')).toHaveTextContent('1234')
  })

  it('renders period text', () => {
    render(<StatsCard {...defaultProps} />)
    expect(screen.getByText('за месяц')).toBeInTheDocument()
  })

  it('renders positive change with plus sign', () => {
    render(<StatsCard {...defaultProps} change={12.5} />)
    expect(screen.getByText('+12.5%')).toBeInTheDocument()
  })

  it('renders negative change without plus sign', () => {
    render(<StatsCard {...defaultProps} change={-5.3} />)
    expect(screen.getByText('-5.3%')).toBeInTheDocument()
  })

  it('renders zero change with plus sign', () => {
    render(<StatsCard {...defaultProps} change={0} />)
    expect(screen.getByText('+0.0%')).toBeInTheDocument()
  })

  it('applies green styling for positive change', () => {
    render(<StatsCard {...defaultProps} change={10} />)
    const changeElement = screen.getByText('+10.0%')
    expect(changeElement).toHaveClass('text-green-600')
  })

  it('applies red styling for negative change', () => {
    render(<StatsCard {...defaultProps} change={-10} />)
    const changeElement = screen.getByText('-10.0%')
    expect(changeElement).toHaveClass('text-red-600')
  })

  it('renders icon', () => {
    render(<StatsCard {...defaultProps} icon={FileText} />)
    // Icon is rendered inside a div with specific classes
    const iconContainer = document.querySelector('.flex.h-12.w-12')
    expect(iconContainer).toBeInTheDocument()
  })

  it('renders GlowingEffect', () => {
    render(<StatsCard {...defaultProps} />)
    expect(screen.getByTestId('glowing-effect')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<StatsCard {...defaultProps} className="custom-class" />)
    const card = container.firstChild
    expect(card).toHaveClass('custom-class')
  })

  it('renders with different icons', () => {
    const { rerender } = render(<StatsCard {...defaultProps} icon={Users} />)
    expect(document.querySelector('.flex.h-12.w-12')).toBeInTheDocument()

    rerender(<StatsCard {...defaultProps} icon={Calendar} />)
    expect(document.querySelector('.flex.h-12.w-12')).toBeInTheDocument()
  })
})
