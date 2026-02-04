'use client'

import { useTranslations } from 'next-intl'
import { Users, TrendingDown, GraduationCap, AlertTriangle } from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { cn } from '@/lib/utils'
import type { GroupSummaryInfo } from '@/lib/api/analytics'

interface GroupSummaryCardProps {
  group: GroupSummaryInfo
  onClick?: (group: GroupSummaryInfo) => void
  className?: string
}

export function GroupSummaryCard({ group, onClick, className }: GroupSummaryCardProps) {
  const t = useTranslations('analytics')

  const formatPercent = (value: number) => `${value.toFixed(1)}%`
  const formatGrade = (value: number) => value.toFixed(2)

  const totalAtRisk =
    group.risk_distribution.critical + group.risk_distribution.high + group.risk_distribution.medium

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:shadow-xl cursor-pointer',
        className
      )}
      onClick={() => onClick?.(group)}
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
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400">
              <Users className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                {group.group_name}
              </h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {t('studentsCount', { count: group.total_students })}
              </p>
            </div>
          </div>
        </div>

        {/* Metrics */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div className="flex items-center gap-2">
            <TrendingDown className="h-4 w-4 text-blue-500" />
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('avgAttendance')}</p>
              <p className="font-semibold text-gray-900 dark:text-white">
                {formatPercent(group.avg_attendance_rate)}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <GraduationCap className="h-4 w-4 text-green-500" />
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('avgGrade')}</p>
              <p className="font-semibold text-gray-900 dark:text-white">
                {formatGrade(group.avg_grade)}
              </p>
            </div>
          </div>
        </div>

        {/* Risk Distribution */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-orange-500" />
              <span className="text-sm text-gray-600 dark:text-gray-400">
                {t('riskDistribution')}
              </span>
            </div>
            <span className="text-sm font-medium text-orange-600 dark:text-orange-400">
              {totalAtRisk} {t('atRisk')}
            </span>
          </div>

          {/* Risk Distribution Bar */}
          <div className="flex h-3 rounded-full overflow-hidden bg-gray-200 dark:bg-gray-700">
            {group.risk_distribution.critical > 0 && (
              <div
                className="bg-red-500"
                style={{
                  width: `${(group.risk_distribution.critical / group.total_students) * 100}%`,
                }}
                title={`${t('riskLevel.critical')}: ${group.risk_distribution.critical}`}
              />
            )}
            {group.risk_distribution.high > 0 && (
              <div
                className="bg-orange-500"
                style={{
                  width: `${(group.risk_distribution.high / group.total_students) * 100}%`,
                }}
                title={`${t('riskLevel.high')}: ${group.risk_distribution.high}`}
              />
            )}
            {group.risk_distribution.medium > 0 && (
              <div
                className="bg-yellow-500"
                style={{
                  width: `${(group.risk_distribution.medium / group.total_students) * 100}%`,
                }}
                title={`${t('riskLevel.medium')}: ${group.risk_distribution.medium}`}
              />
            )}
            {group.risk_distribution.low > 0 && (
              <div
                className="bg-green-500"
                style={{
                  width: `${(group.risk_distribution.low / group.total_students) * 100}%`,
                }}
                title={`${t('riskLevel.low')}: ${group.risk_distribution.low}`}
              />
            )}
          </div>

          {/* Legend */}
          <div className="flex flex-wrap gap-3 mt-2 text-xs">
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-red-500" />
              <span className="text-gray-500">{group.risk_distribution.critical}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-orange-500" />
              <span className="text-gray-500">{group.risk_distribution.high}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-yellow-500" />
              <span className="text-gray-500">{group.risk_distribution.medium}</span>
            </div>
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-gray-500">{group.risk_distribution.low}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
