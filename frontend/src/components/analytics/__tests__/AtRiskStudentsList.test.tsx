import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AtRiskStudentsList } from '../AtRiskStudentsList'
import { analyticsApi } from '@/lib/api/analytics'
import type { AtRiskStudentsResponse } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      loadError: 'Failed to load data',
      noAtRiskStudents: 'No at-risk students found',
      showingStudents: `Showing ${params?.from}-${params?.to} of ${params?.total}`,
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
    getAtRiskStudents: jest.fn(),
    getStudentsByRiskLevel: jest.fn(),
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockApiResponse: AtRiskStudentsResponse = {
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
    {
      student_id: 2,
      student_name: 'Jane Smith',
      group_name: 'CS-102',
      risk_level: 'critical',
      risk_score: 90,
      attendance_rate: 40,
      grade_average: 1.8,
    },
  ],
  total: 2,
  page: 1,
  page_size: 9,
}

const mockApiClient = analyticsApi as jest.Mocked<typeof analyticsApi>

describe('AtRiskStudentsList', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockApiClient.getAtRiskStudents.mockImplementation(() => new Promise(() => {}))
    render(<AtRiskStudentsList />)

    expect(screen.getByText('', { selector: '.animate-spin' })).toBeInTheDocument()
  })

  it('renders students after loading', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockApiResponse)

    render(<AtRiskStudentsList />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
      expect(screen.getByText('Jane Smith')).toBeInTheDocument()
    })
  })

  it('shows error message on API failure', async () => {
    mockApiClient.getAtRiskStudents.mockRejectedValueOnce(new Error('API Error'))

    render(<AtRiskStudentsList />)

    await waitFor(() => {
      expect(screen.getByText('Failed to load data')).toBeInTheDocument()
    })
  })

  it('shows empty state when no students', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce({
      students: [],
      total: 0,
      page: 1,
      page_size: 9,
    })

    render(<AtRiskStudentsList />)

    await waitFor(() => {
      expect(screen.getByText('No at-risk students found')).toBeInTheDocument()
    })
  })

  it('calls getStudentsByRiskLevel when riskLevel prop provided', async () => {
    mockApiClient.getStudentsByRiskLevel.mockResolvedValueOnce(mockApiResponse)

    render(<AtRiskStudentsList riskLevel="critical" />)

    await waitFor(() => {
      expect(mockApiClient.getStudentsByRiskLevel).toHaveBeenCalledWith('critical', 1, 9)
    })
  })

  it('uses custom pageSize', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockApiResponse)

    render(<AtRiskStudentsList pageSize={12} />)

    await waitFor(() => {
      expect(mockApiClient.getAtRiskStudents).toHaveBeenCalledWith(1, 12)
    })
  })

  it('calls onStudentClick when student card clicked', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce(mockApiResponse)
    const handleClick = jest.fn()

    render(<AtRiskStudentsList onStudentClick={handleClick} />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByText('John Doe').closest('div')!)
    expect(handleClick).toHaveBeenCalledWith(mockApiResponse.students[0])
  })

  it('shows pagination when multiple pages', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce({
      ...mockApiResponse,
      total: 20,
    })

    render(<AtRiskStudentsList pageSize={9} />)

    await waitFor(() => {
      expect(screen.getByText(/1 \/ \d/)).toBeInTheDocument()
    })
  })

  it('navigates to next page', async () => {
    mockApiClient.getAtRiskStudents
      .mockResolvedValueOnce({ ...mockApiResponse, total: 20 })
      .mockResolvedValueOnce({ ...mockApiResponse, page: 2, total: 20 })

    render(<AtRiskStudentsList pageSize={9} />)

    await waitFor(() => {
      expect(screen.getByText('1 / 3')).toBeInTheDocument()
    })

    // Get all buttons and click the last one (next button)
    const buttons = screen.getAllByRole('button')
    const nextButton = buttons[buttons.length - 1]
    fireEvent.click(nextButton)

    await waitFor(() => {
      expect(mockApiClient.getAtRiskStudents).toHaveBeenCalledWith(2, 9)
    })
  })

  it('hides pagination for single page', async () => {
    mockApiClient.getAtRiskStudents.mockResolvedValueOnce({
      ...mockApiResponse,
      total: 2,
    })

    render(<AtRiskStudentsList pageSize={9} />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    expect(screen.queryByText(/1 \/ \d/)).not.toBeInTheDocument()
  })

  it('navigates to previous page', async () => {
    // Clear mocks at the start of this test
    mockApiClient.getAtRiskStudents.mockClear()

    // First load page 1, then go to page 2, then back to page 1
    mockApiClient.getAtRiskStudents
      .mockResolvedValueOnce({ ...mockApiResponse, page: 1, total: 20 })
      .mockResolvedValueOnce({ ...mockApiResponse, page: 2, total: 20 })
      .mockResolvedValueOnce({ ...mockApiResponse, page: 1, total: 20 })

    render(<AtRiskStudentsList pageSize={9} />)

    await waitFor(() => {
      expect(screen.getByText('1 / 3')).toBeInTheDocument()
    })

    // Navigate to page 2 first
    const buttons = screen.getAllByRole('button')
    const nextButton = buttons[buttons.length - 1]
    fireEvent.click(nextButton)

    await waitFor(() => {
      expect(screen.getByText('2 / 3')).toBeInTheDocument()
    })

    // Now navigate back to page 1
    const buttonsAfter = screen.getAllByRole('button')
    const prevButton = buttonsAfter[0]
    fireEvent.click(prevButton)

    // Verify it calls with page 1 (going back)
    await waitFor(() => {
      expect(mockApiClient.getAtRiskStudents).toHaveBeenLastCalledWith(1, 9)
    })
  })
})
