import { render, screen } from '@testing-library/react'
import { RiskLevelBadge } from '../RiskLevelBadge'
import type { RiskLevel } from '@/lib/api/analytics'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      'riskLevel.low': 'Low',
      'riskLevel.medium': 'Medium',
      'riskLevel.high': 'High',
      'riskLevel.critical': 'Critical',
    }
    return translations[key] || key
  },
}))

describe('RiskLevelBadge', () => {
  it('renders low risk level', () => {
    render(<RiskLevelBadge level="low" />)
    expect(screen.getByText('Low')).toBeInTheDocument()
  })

  it('renders medium risk level', () => {
    render(<RiskLevelBadge level="medium" />)
    expect(screen.getByText('Medium')).toBeInTheDocument()
  })

  it('renders high risk level', () => {
    render(<RiskLevelBadge level="high" />)
    expect(screen.getByText('High')).toBeInTheDocument()
  })

  it('renders critical risk level', () => {
    render(<RiskLevelBadge level="critical" />)
    expect(screen.getByText('Critical')).toBeInTheDocument()
  })

  it('renders raw level when showLabel is false', () => {
    render(<RiskLevelBadge level="high" showLabel={false} />)
    expect(screen.getByText('high')).toBeInTheDocument()
  })

  it('applies correct color classes for low level', () => {
    const { container } = render(<RiskLevelBadge level="low" />)
    const badge = container.querySelector('[class*="bg-green"]')
    expect(badge).toBeInTheDocument()
  })

  it('applies correct color classes for critical level', () => {
    const { container } = render(<RiskLevelBadge level="critical" />)
    const badge = container.querySelector('[class*="bg-red"]')
    expect(badge).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<RiskLevelBadge level="medium" className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it.each([
    ['low', 'green'],
    ['medium', 'yellow'],
    ['high', 'orange'],
    ['critical', 'red'],
  ] as [RiskLevel, string][])('applies %s risk color', (level, color) => {
    const { container } = render(<RiskLevelBadge level={level} />)
    const badge = container.querySelector(`[class*="bg-${color}"]`)
    expect(badge).toBeInTheDocument()
  })
})
