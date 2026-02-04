'use client'

import { useTranslations } from 'next-intl'
import { User, TrendingDown, GraduationCap, AlertTriangle } from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { RiskLevelBadge } from './RiskLevelBadge'
import { cn } from '@/lib/utils'
import type { StudentRiskInfo } from '@/lib/api/analytics'

interface StudentRiskCardProps {
  student: StudentRiskInfo
  onClick?: (student: StudentRiskInfo) => void
  className?: string
}

export function StudentRiskCard({ student, onClick, className }: StudentRiskCardProps) {
  const t = useTranslations('analytics')

  const formatPercent = (value?: number) => {
    if (value === undefined || value === null) return '—'
    return `${value.toFixed(1)}%`
  }

  const formatGrade = (value?: number) => {
    if (value === undefined || value === null) return '—'
    return value.toFixed(2)
  }

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-5 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:shadow-lg cursor-pointer',
        className
      )}
      onClick={() => onClick?.(student)}
    >
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={3}
      />
      <div className="relative z-10">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800">
              <User className="h-5 w-5 text-gray-600 dark:text-gray-400" />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white">
                {student.student_name}
              </h3>
              {student.group_name && (
                <p className="text-sm text-gray-500 dark:text-gray-400">{student.group_name}</p>
              )}
            </div>
          </div>
          <RiskLevelBadge level={student.risk_level} />
        </div>

        {/* Metrics */}
        <div className="grid grid-cols-2 gap-3">
          <div className="flex items-center gap-2">
            <TrendingDown className="h-4 w-4 text-gray-400" />
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('attendance')}</p>
              <p className="font-medium text-gray-900 dark:text-white">
                {formatPercent(student.attendance_rate)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <GraduationCap className="h-4 w-4 text-gray-400" />
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('gradeAverage')}</p>
              <p className="font-medium text-gray-900 dark:text-white">
                {formatGrade(student.grade_average)}
              </p>
            </div>
          </div>
        </div>

        {/* Risk Score */}
        <div className="mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-gray-400" />
              <span className="text-sm text-gray-600 dark:text-gray-400">{t('riskScore')}</span>
            </div>
            <span className="font-bold text-lg text-gray-900 dark:text-white">
              {student.risk_score.toFixed(0)}
            </span>
          </div>
          {/* Risk Score Progress Bar */}
          <div className="mt-2 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full transition-all',
                student.risk_score >= 70
                  ? 'bg-red-500'
                  : student.risk_score >= 50
                    ? 'bg-orange-500'
                    : student.risk_score >= 30
                      ? 'bg-yellow-500'
                      : 'bg-green-500'
              )}
              style={{ width: `${Math.min(student.risk_score, 100)}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
