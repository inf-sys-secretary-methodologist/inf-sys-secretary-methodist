import { render, screen, waitFor } from '@testing-library/react'
import { AttendanceTrendChart } from '../AttendanceTrendChart'
import { analyticsApi } from '@/lib/api/analytics'
import type { MonthlyTrendInfo } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      loadError: 'Failed to load data',
      attendanceTrend: 'Attendance Trend',
      lastMonths: `Last ${params?.months} months`,
      attendanceRate: 'Attendance Rate',
      totalRecords: 'Total Records',
      avgAttendance: 'Avg Attendance',
      uniqueStudents: 'Unique Students',
    }
    return translations[key] || key
  },
}))

// Mock API
jest.mock('@/lib/api/analytics', () => ({
  analyticsApi: {
    getAttendanceTrend: jest.fn(),
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

// Mock recharts - execute formatter functions for coverage
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="chart-container">{children}</div>
  ),
  AreaChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="area-chart">{children}</div>
  ),
  Area: () => <div data-testid="area" />,
  XAxis: () => null,
  YAxis: ({ tickFormatter }: { tickFormatter?: (value: number) => string }) => {
    // Execute the tickFormatter to cover it
    const result = tickFormatter ? tickFormatter(85) : null
    return <div data-testid="y-axis">{result}</div>
  },
  CartesianGrid: () => null,
  Tooltip: ({ formatter }: { formatter?: (value: number) => [string, string] }) => {
    // Execute the formatter to cover it
    if (formatter) {
      const result = formatter(85.5)
      return (
        <div data-testid="tooltip-formatter">
          {result[0]} - {result[1]}
        </div>
      )
    }
    return null
  },
  Legend: () => <div data-testid="legend" />,
}))

const mockTrends: MonthlyTrendInfo[] = [
  {
    month: '2024-01',
    unique_students: 100,
    total_records: 500,
    present_count: 450,
    absent_count: 50,
    attendance_rate: 90,
  },
  {
    month: '2024-02',
    unique_students: 105,
    total_records: 520,
    present_count: 468,
    absent_count: 52,
    attendance_rate: 90,
  },
  {
    month: '2024-03',
    unique_students: 110,
    total_records: 550,
    present_count: 495,
    absent_count: 55,
    attendance_rate: 90,
  },
]

const mockApiClient = analyticsApi as jest.Mocked<typeof analyticsApi>

describe('AttendanceTrendChart', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockApiClient.getAttendanceTrend.mockImplementation(() => new Promise(() => {}))
    render(<AttendanceTrendChart />)

    expect(screen.getByText('', { selector: '.animate-spin' })).toBeInTheDocument()
  })

  it('renders chart after loading', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      expect(screen.getByText('Attendance Trend')).toBeInTheDocument()
    })
  })

  it('uses default 6 months', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      expect(mockApiClient.getAttendanceTrend).toHaveBeenCalledWith(6)
      expect(screen.getByText('Last 6 months')).toBeInTheDocument()
    })
  })

  it('uses custom months prop', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart months={12} />)

    await waitFor(() => {
      expect(mockApiClient.getAttendanceTrend).toHaveBeenCalledWith(12)
      expect(screen.getByText('Last 12 months')).toBeInTheDocument()
    })
  })

  it('shows error message on API failure', async () => {
    mockApiClient.getAttendanceTrend.mockRejectedValueOnce(new Error('API Error'))

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      expect(screen.getByText('Failed to load data')).toBeInTheDocument()
    })
  })

  it('renders total records', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      // Total: 500 + 520 + 550 = 1570
      expect(screen.getByText('1,570')).toBeInTheDocument()
    })
  })

  it('renders average attendance', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      // Average: (90 + 90 + 90) / 3 = 90
      expect(screen.getByText('90.0%')).toBeInTheDocument()
    })
  })

  it('renders max unique students', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      // Max: 110
      expect(screen.getByText('110')).toBeInTheDocument()
    })
  })

  it('renders area chart', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      expect(screen.getByTestId('area-chart')).toBeInTheDocument()
    })
  })

  it('applies custom className', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    const { container } = render(<AttendanceTrendChart className="custom-chart" />)

    await waitFor(() => {
      expect(container.querySelector('.custom-chart')).toBeInTheDocument()
    })
  })

  it('renders summary labels', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      expect(screen.getByText('Total Records')).toBeInTheDocument()
      expect(screen.getByText('Avg Attendance')).toBeInTheDocument()
      expect(screen.getByText('Unique Students')).toBeInTheDocument()
    })
  })

  it('renders YAxis tick formatter output', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      // The mock YAxis executes the tickFormatter
      expect(screen.getByTestId('y-axis')).toHaveTextContent('85%')
    })
  })

  it('renders tooltip formatter output', async () => {
    mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

    render(<AttendanceTrendChart />)

    await waitFor(() => {
      // The mock Tooltip executes the formatter
      expect(screen.getByTestId('tooltip-formatter')).toBeInTheDocument()
    })
  })
})
