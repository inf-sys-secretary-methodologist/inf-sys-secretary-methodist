import { render, screen, fireEvent } from '@testing-library/react'
import { GroupSummaryCard } from '../GroupSummaryCard'
import type { GroupSummaryInfo } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: Record<string, unknown>) => {
    const translations: Record<string, string> = {
      studentsCount: `${params?.count} students`,
      avgAttendance: 'Avg Attendance',
      avgGrade: 'Avg Grade',
      riskDistribution: 'Risk Distribution',
      atRisk: 'at risk',
      'riskLevel.low': 'Low',
      'riskLevel.medium': 'Medium',
      'riskLevel.high': 'High',
      'riskLevel.critical': 'Critical',
    }
    return translations[key] || key
  },
}))

// Mock GlowingEffect
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockGroup: GroupSummaryInfo = {
  group_name: 'CS-101',
  total_students: 30,
  avg_attendance_rate: 85.5,
  avg_grade: 3.75,
  risk_distribution: {
    critical: 2,
    high: 5,
    medium: 8,
    low: 15,
  },
  at_risk_percentage: 50,
}

describe('GroupSummaryCard', () => {
  it('renders group name', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    expect(screen.getByText('CS-101')).toBeInTheDocument()
  })

  it('renders total students count', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    expect(screen.getByText('30 students')).toBeInTheDocument()
  })

  it('renders average attendance rate', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    expect(screen.getByText('85.5%')).toBeInTheDocument()
  })

  it('renders average grade', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    expect(screen.getByText('3.75')).toBeInTheDocument()
  })

  it('renders at-risk count', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    // 2 + 5 + 8 = 15 at risk (critical + high + medium)
    expect(screen.getByText(/15.*at risk/)).toBeInTheDocument()
  })

  it('renders risk distribution legend values', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    // Check legend values
    expect(screen.getByText('2')).toBeInTheDocument() // critical
    expect(screen.getByText('5')).toBeInTheDocument() // high
    expect(screen.getByText('8')).toBeInTheDocument() // medium
    expect(screen.getByText('15')).toBeInTheDocument() // low
  })

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn()
    render(<GroupSummaryCard group={mockGroup} onClick={handleClick} />)

    fireEvent.click(screen.getByText('CS-101').closest('div')!)
    expect(handleClick).toHaveBeenCalledWith(mockGroup)
  })

  it('does not crash when onClick not provided', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    fireEvent.click(screen.getByText('CS-101'))
  })

  it('applies custom className', () => {
    const { container } = render(<GroupSummaryCard group={mockGroup} className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('renders labels', () => {
    render(<GroupSummaryCard group={mockGroup} />)
    expect(screen.getByText('Avg Attendance')).toBeInTheDocument()
    expect(screen.getByText('Avg Grade')).toBeInTheDocument()
    expect(screen.getByText('Risk Distribution')).toBeInTheDocument()
  })

  it('handles zero risk counts', () => {
    const groupWithZeros: GroupSummaryInfo = {
      ...mockGroup,
      risk_distribution: {
        critical: 0,
        high: 0,
        medium: 0,
        low: 30,
      },
    }
    render(<GroupSummaryCard group={groupWithZeros} />)
    // Should render 0 for at risk count (0 + 0 + 0)
    expect(screen.getByText(/0.*at risk/)).toBeInTheDocument()
  })
})
