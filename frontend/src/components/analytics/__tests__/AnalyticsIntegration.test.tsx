import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { RiskDistributionChart } from '../RiskDistributionChart'
import { AttendanceTrendChart } from '../AttendanceTrendChart'
import { AtRiskStudentsList } from '../AtRiskStudentsList'
import { GroupSummaryCard } from '../GroupSummaryCard'
import { analyticsApi } from '@/lib/api/analytics'
import type {
  GroupSummaryInfo,
  MonthlyTrendInfo,
  AtRiskStudentsResponse,
} from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      loadError: 'Failed to load data',
      riskDistribution: 'Risk Distribution',
      allGroups: 'All Groups',
      noData: 'No data available',
      students: 'students',
      totalStudents: 'Total Students',
      groupsCount: 'Groups',
      attendanceTrend: 'Attendance Trend',
      lastMonths: `Last ${params?.months} months`,
      attendanceRate: 'Attendance Rate',
      totalRecords: 'Total Records',
      avgAttendance: 'Avg Attendance',
      uniqueStudents: 'Unique Students',
      noAtRiskStudents: 'No at-risk students found',
      showingStudents: `Showing ${params?.from}-${params?.to} of ${params?.total}`,
      attendance: 'Attendance',
      gradeAverage: 'Grade Average',
      riskScore: 'Risk Score',
      avgRisk: 'Avg Risk',
      atRiskPercentage: 'At Risk',
      totalStudentsLabel: 'Total Students',
      avgAttendanceLabel: 'Avg Attendance',
      avgGradeLabel: 'Avg Grade',
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
    getAttendanceTrend: jest.fn(),
    getAtRiskStudents: jest.fn(),
    getStudentsByRiskLevel: jest.fn(),
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

// Mock recharts
jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="chart-container">{children}</div>
  ),
  PieChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="pie-chart">{children}</div>
  ),
  AreaChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="area-chart">{children}</div>
  ),
  Pie: ({ children }: { children?: React.ReactNode }) => <div data-testid="pie">{children}</div>,
  Area: () => <div data-testid="area" />,
  Cell: () => null,
  XAxis: () => null,
  YAxis: ({ tickFormatter }: { tickFormatter?: (value: number) => string }) => {
    const result = tickFormatter ? tickFormatter(85) : null
    return <div data-testid="y-axis">{result}</div>
  },
  CartesianGrid: () => null,
  Legend: ({ formatter }: { formatter?: (value: string) => React.ReactNode }) => {
    const result = formatter ? formatter('Test Legend') : null
    return <div data-testid="legend">{result}</div>
  },
  Tooltip: ({ formatter }: { formatter?: (value: number) => [string, string] }) => {
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
]

const mockTrends: MonthlyTrendInfo[] = [
  {
    month: '2024-01',
    unique_students: 100,
    total_records: 500,
    present_count: 450,
    absent_count: 50,
    attendance_rate: 90,
  },
]

const mockStudents: AtRiskStudentsResponse = {
  students: [
    {
      student_id: 1,
      student_name: 'John Doe',
      group_name: 'CS-101',
      risk_level: 'high',
      risk_score: 75,
      attendance_rate: 60,
      grade_average: 2.5,
    },
  ],
  total: 1,
  page: 1,
  page_size: 9,
}

const mockApiClient = analyticsApi as jest.Mocked<typeof analyticsApi>

describe('Analytics Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('Dashboard Components Integration', () => {
    it('renders risk distribution and attendance trend charts together', async () => {
      mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)
      mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)

      render(
        <div>
          <RiskDistributionChart />
          <AttendanceTrendChart />
        </div>
      )

      await waitFor(() => {
        expect(screen.getByText('Risk Distribution')).toBeInTheDocument()
        expect(screen.getByText('Attendance Trend')).toBeInTheDocument()
      })
    })

    it('charts and student list load data independently', async () => {
      mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)
      mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockStudents)

      render(
        <div>
          <RiskDistributionChart />
          <AttendanceTrendChart />
          <AtRiskStudentsList />
        </div>
      )

      await waitFor(() => {
        expect(mockApiClient.getAllGroupsSummary).toHaveBeenCalled()
        expect(mockApiClient.getAttendanceTrend).toHaveBeenCalled()
        expect(mockApiClient.getAtRiskStudents).toHaveBeenCalled()
      })
    })

    it('one API failure does not affect other components', async () => {
      mockApiClient.getAllGroupsSummary.mockRejectedValueOnce(new Error('API Error'))
      mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockStudents)

      render(
        <div>
          <RiskDistributionChart />
          <AttendanceTrendChart />
          <AtRiskStudentsList />
        </div>
      )

      await waitFor(() => {
        // Risk distribution shows error
        expect(screen.getByText('Failed to load data')).toBeInTheDocument()
        // But other components still render
        expect(screen.getByText('Attendance Trend')).toBeInTheDocument()
        expect(screen.getByText('John Doe')).toBeInTheDocument()
      })
    })
  })

  describe('GroupSummaryCard with StudentList', () => {
    it('displays group card alongside at-risk students list', async () => {
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockStudents)

      const { container } = render(
        <div>
          <GroupSummaryCard group={mockGroups[0]} />
          <AtRiskStudentsList />
        </div>
      )

      // GroupSummaryCard renders synchronously
      expect(container.textContent).toContain('CS-101')

      // AtRiskStudentsList loads async
      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument()
      })
    })

    it('clicking student triggers callback', async () => {
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockStudents)
      const handleClick = jest.fn()

      render(<AtRiskStudentsList onStudentClick={handleClick} />)

      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument()
      })

      fireEvent.click(screen.getByText('John Doe').closest('div')!)
      expect(handleClick).toHaveBeenCalledWith(mockStudents.students[0])
    })
  })

  describe('Data Flow', () => {
    it('all components receive and display their data correctly', async () => {
      mockApiClient.getAllGroupsSummary.mockResolvedValueOnce(mockGroups)
      mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockStudents)

      render(
        <div>
          <RiskDistributionChart />
          <AttendanceTrendChart />
          <AtRiskStudentsList />
        </div>
      )

      await waitFor(() => {
        // All titles are visible
        expect(screen.getByText('Risk Distribution')).toBeInTheDocument()
        expect(screen.getByText('Attendance Trend')).toBeInTheDocument()
        // Student data is visible
        expect(screen.getByText('John Doe')).toBeInTheDocument()
        expect(screen.getByText('High')).toBeInTheDocument()
      })
    })

    it('handles empty data states across components', async () => {
      mockApiClient.getAllGroupsSummary.mockResolvedValueOnce([])
      mockApiClient.getAttendanceTrend.mockResolvedValueOnce(mockTrends)
      mockApiClient.getAtRiskStudents.mockResolvedValueOnce({
        students: [],
        total: 0,
        page: 1,
        page_size: 9,
      })

      render(
        <div>
          <RiskDistributionChart />
          <AtRiskStudentsList />
        </div>
      )

      await waitFor(() => {
        expect(screen.getByText('No data available')).toBeInTheDocument()
        expect(screen.getByText('No at-risk students found')).toBeInTheDocument()
      })
    })
  })
})
