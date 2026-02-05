import { render, screen, waitFor } from '@testing-library/react'
import { RiskDistributionChart } from '../RiskDistributionChart'
import { analyticsApi } from '@/lib/api/analytics'
import type { GroupSummaryInfo } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      loadError: 'Failed to load data',
      riskDistribution: 'Risk Distribution',
      allGroups: 'All Groups',
      noData: 'No data available',
      students: 'students',
      totalStudents: 'Total Students',
      groupsCount: 'Groups',
      'riskLevel.low': 'Low',
      'riskLevel.medium': 'Medium',
      'riskLevel.high': 'High',
      'riskLevel.critical': 'Critical',
    }
    return translations[key] || key
  },
}))

// Mock API
jest.mock('@/lib/api/analytics', () => ({
  analyticsApi: {
    getAllGroupsSummary: jest.fn(),
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

// Mock recharts - capture and execute formatter functions for coverage
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="chart-container">{children}</div>
  ),
  PieChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="pie-chart">{children}</div>
  ),
  Pie: ({ children }: { children?: React.ReactNode }) => <div data-testid="pie">{children}</div>,
  Cell: () => null,
  Legend: ({ formatter }: { formatter?: (value: string) => React.ReactNode }) => {
    // Execute the formatter to cover it
    const result = formatter ? formatter('Test Legend') : null
    return <div data-testid="legend">{result}</div>
  },
  Tooltip: ({ formatter }: { formatter?: (value: number) => [string, string] }) => {
    // Execute the formatter to cover it
    if (formatter) {
      const result = formatter(10)
      return (
        <div data-testid="tooltip-formatter">
          {result[0]} - {result[1]}
        </div>
      )
    }
    return null
  },
}))

const mockGroups: GroupSummaryInfo[] = [
  {
    group_name: 'CS-101',
    total_students: 30,
    avg_attendance_rate: 85,
    avg_grade: 3.5,
    risk_distribution: { critical: 2, high: 5, medium: 8, low: 15 },
    at_risk_percentage: 50,
  },
  {
    group_name: 'CS-102',
    total_students: 25,
    avg_attendance_rate: 90,
    avg_grade: 3.8,
    risk_distribution: { critical: 1, high: 3, medium: 5, low: 16 },
    at_risk_percentage: 36,
  },
]

const mockApiClient = analyticsApi as jest.Mocked<typeof analyticsApi>

describe('RiskDistributionChart', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockApiClient.getAllGroupsSummary.mockImplementation(() => new Promise(() => {}))
    render(<RiskDistributionChart />)

    expect(screen.getByText('', { selector: '.animate-spin' })).toBeInTheDocument()
  })

  it('renders chart after loading', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      expect(screen.getByText('Risk Distribution')).toBeInTheDocument()
      expect(screen.getByText('All Groups')).toBeInTheDocument()
    })
  })

  it('shows error message on API failure', async () => {
    mockApiClient.getAllGroupsSummary.mockRejectedValueOnce(new Error('API Error'))

    render(<RiskDistributionChart />)

    await waitFor(() => {
      expect(screen.getByText('Failed to load data')).toBeInTheDocument()
    })
  })

  it('shows no data message when no groups', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce([])

    render(<RiskDistributionChart />)

    await waitFor(() => {
      expect(screen.getByText('No data available')).toBeInTheDocument()
    })
  })

  it('renders total students count', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      // Total: 30 + 25 = 55 students
      expect(screen.getByText('55')).toBeInTheDocument()
    })
  })

  it('renders groups count', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      // 2 groups
      expect(screen.getByText('2')).toBeInTheDocument()
    })
  })

  it('renders pie chart', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
    })
  })

  it('applies custom className', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    const { container } = render(<RiskDistributionChart className="custom-chart" />)

    await waitFor(() => {
      expect(container.querySelector('.custom-chart')).toBeInTheDocument()
    })
  })

  it('renders tooltip formatter output', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      // The mock Tooltip executes the formatter and renders its output
      expect(screen.getByTestId('tooltip-formatter')).toBeInTheDocument()
    })
  })

  it('renders legend formatter output', async () => {
    mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)

    render(<RiskDistributionChart />)

    await waitFor(() => {
      // The mock Legend executes the formatter and renders its output
      expect(screen.getByTestId('legend')).toBeInTheDocument()
      expect(screen.getByText('Test Legend')).toBeInTheDocument()
    })
  })
})
