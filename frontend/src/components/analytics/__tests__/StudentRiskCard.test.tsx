import { render, screen, fireEvent } from '@testing-library/react'
import { StudentRiskCard } from '../StudentRiskCard'
import type { StudentRiskInfo } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      attendance: 'Attendance',
      gradeAverage: 'Grade Average',
      riskScore: 'Risk Score',
      'riskLevel.low': 'Low',
      'riskLevel.medium': 'Medium',
      'riskLevel.high': 'High',
      'riskLevel.critical': 'Critical',
    }
    return translations[key] || key
  },
}))

// Mock GlowingEffect to avoid animation issues
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => null,
}))

const mockStudent: StudentRiskInfo = {
  student_id: 1,
  student_name: 'John Doe',
  group_name: 'CS-101',
  attendance_rate: 75.5,
  grade_average: 3.45,
  risk_level: 'high',
  risk_score: 65,
}

describe('StudentRiskCard', () => {
  it('renders student name', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('renders group name', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('CS-101')).toBeInTheDocument()
  })

  it('renders attendance rate formatted', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('75.5%')).toBeInTheDocument()
  })

  it('renders grade average formatted', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('3.45')).toBeInTheDocument()
  })

  it('renders risk score', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('65')).toBeInTheDocument()
  })

  it('renders dash for missing attendance', () => {
    const studentWithoutAttendance = { ...mockStudent, attendance_rate: undefined }
    render(<StudentRiskCard student={studentWithoutAttendance} />)
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('renders dash for missing grade', () => {
    const studentWithoutGrade = { ...mockStudent, grade_average: undefined }
    render(<StudentRiskCard student={studentWithoutGrade} />)
    expect(screen.getAllByText('—').length).toBeGreaterThan(0)
  })

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn()
    render(<StudentRiskCard student={mockStudent} onClick={handleClick} />)

    fireEvent.click(screen.getByText('John Doe').closest('div')!)
    expect(handleClick).toHaveBeenCalledWith(mockStudent)
  })

  it('does not call onClick when not provided', () => {
    render(<StudentRiskCard student={mockStudent} />)
    // Should not throw when clicking
    fireEvent.click(screen.getByText('John Doe'))
  })

  it('applies custom className', () => {
    const { container } = render(<StudentRiskCard student={mockStudent} className="custom-card" />)
    expect(container.querySelector('.custom-card')).toBeInTheDocument()
  })

  it('does not render group name when not provided', () => {
    const studentWithoutGroup = { ...mockStudent, group_name: undefined }
    render(<StudentRiskCard student={studentWithoutGroup} />)
    expect(screen.queryByText('CS-101')).not.toBeInTheDocument()
  })

  it('renders risk level badge', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('High')).toBeInTheDocument()
  })

  it('renders labels for metrics', () => {
    render(<StudentRiskCard student={mockStudent} />)
    expect(screen.getByText('Attendance')).toBeInTheDocument()
    expect(screen.getByText('Grade Average')).toBeInTheDocument()
    expect(screen.getByText('Risk Score')).toBeInTheDocument()
  })

  it('renders yellow progress bar for medium risk score (30-49)', () => {
    const mediumRiskStudent = { ...mockStudent, risk_score: 40 }
    const { container } = render(<StudentRiskCard student={mediumRiskStudent} />)
    const progressBar = container.querySelector('.bg-yellow-500')
    expect(progressBar).toBeInTheDocument()
  })

  it('renders green progress bar for low risk score (<30)', () => {
    const lowRiskStudent = { ...mockStudent, risk_score: 20 }
    const { container } = render(<StudentRiskCard student={lowRiskStudent} />)
    const progressBar = container.querySelector('.bg-green-500')
    expect(progressBar).toBeInTheDocument()
  })

  it('renders red progress bar for critical risk score (>=70)', () => {
    const criticalStudent = { ...mockStudent, risk_score: 85 }
    const { container } = render(<StudentRiskCard student={criticalStudent} />)
    const progressBar = container.querySelector('.bg-red-500')
    expect(progressBar).toBeInTheDocument()
  })

  it('renders orange progress bar for high risk score (50-69)', () => {
    const highRiskStudent = { ...mockStudent, risk_score: 55 }
    const { container } = render(<StudentRiskCard student={highRiskStudent} />)
    const progressBar = container.querySelector('.bg-orange-500')
    expect(progressBar).toBeInTheDocument()
  })
})
