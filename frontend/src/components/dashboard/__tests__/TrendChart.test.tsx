import { render, screen } from '@testing-library/react'
import { TrendChart } from '../TrendChart'
import type { TrendPoint } from '@/types/dashboard'

// Track domain function calls - used in test assertions
// eslint-disable-next-line @typescript-eslint/no-unused-vars
let lastDomainFunction: ((dataMax: number) => number) | null = null

// Mock recharts
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  AreaChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="area-chart" data-points={data.length}>
      {children}
    </div>
  ),
  Area: ({ dataKey, stroke }: { dataKey: string; stroke: string }) => (
    <div data-testid={`area-${dataKey}`} data-stroke={stroke} />
  ),
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: ({ domain }: { domain?: [number, (dataMax: number) => number] }) => {
    // Call the domain function to cover the code
    if (domain && typeof domain[1] === 'function') {
      lastDomainFunction = domain[1]
    }
    return <div data-testid="y-axis" />
  },
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => <div data-testid="glowing-effect" />,
}))

describe('TrendChart', () => {
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const mockDatasets = [
    {
      name: 'Documents',
      data: [
        { date: yesterday.toISOString(), value: 5 },
        { date: today.toISOString(), value: 10 },
      ] as TrendPoint[],
      color: '#3b82f6',
    },
    {
      name: 'Tasks',
      data: [
        { date: yesterday.toISOString(), value: 3 },
        { date: today.toISOString(), value: 7 },
      ] as TrendPoint[],
      color: '#10b981',
    },
  ]

  it('renders chart title', () => {
    render(<TrendChart title="Activity Overview" datasets={mockDatasets} />)
    expect(screen.getByText('Activity Overview')).toBeInTheDocument()
  })

  it('renders chart components', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('area-chart')).toBeInTheDocument()
    expect(screen.getByTestId('x-axis')).toBeInTheDocument()
    expect(screen.getByTestId('y-axis')).toBeInTheDocument()
    expect(screen.getByTestId('cartesian-grid')).toBeInTheDocument()
    expect(screen.getByTestId('tooltip')).toBeInTheDocument()
    expect(screen.getByTestId('legend')).toBeInTheDocument()
  })

  it('renders areas for each dataset', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)
    expect(screen.getByTestId('area-Documents')).toBeInTheDocument()
    expect(screen.getByTestId('area-Tasks')).toBeInTheDocument()
  })

  it('applies dataset colors to areas', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)
    expect(screen.getByTestId('area-Documents')).toHaveAttribute('data-stroke', '#3b82f6')
    expect(screen.getByTestId('area-Tasks')).toHaveAttribute('data-stroke', '#10b981')
  })

  it('renders GlowingEffect', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)
    expect(screen.getByTestId('glowing-effect')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <TrendChart title="Test Chart" datasets={mockDatasets} className="custom-chart" />
    )
    expect(container.firstChild).toHaveClass('custom-chart')
  })

  it('defaults to month period (30 days)', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)
    const chart = screen.getByTestId('area-chart')
    // Should have 30 data points for month period
    expect(chart).toHaveAttribute('data-points', '30')
  })

  it('uses week period when specified (7 days)', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} period="week" />)
    const chart = screen.getByTestId('area-chart')
    expect(chart).toHaveAttribute('data-points', '7')
  })

  it('uses quarter period when specified (90 days)', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} period="quarter" />)
    const chart = screen.getByTestId('area-chart')
    expect(chart).toHaveAttribute('data-points', '90')
  })

  it('uses year period when specified (365 days)', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} period="year" />)
    const chart = screen.getByTestId('area-chart')
    expect(chart).toHaveAttribute('data-points', '365')
  })

  it('handles empty datasets', () => {
    render(<TrendChart title="Empty Chart" datasets={[]} />)
    expect(screen.getByText('Empty Chart')).toBeInTheDocument()
    expect(screen.getByTestId('area-chart')).toBeInTheDocument()
  })

  it('YAxis domain function calculates correct max value', () => {
    render(<TrendChart title="Test Chart" datasets={mockDatasets} />)

    // The domain function should have been captured
    expect(lastDomainFunction).not.toBeNull()

    // Test the domain function with various data max values
    if (lastDomainFunction) {
      // Math.ceil(100 * 1.1) can be 110 or 111 due to floating point
      expect(lastDomainFunction(100)).toBeGreaterThanOrEqual(110)
      expect(lastDomainFunction(100)).toBeLessThanOrEqual(111)
      expect(lastDomainFunction(50)).toBeGreaterThanOrEqual(55)
      expect(lastDomainFunction(0)).toBe(1) // fallback to 1 when dataMax is 0
    }
  })
})
