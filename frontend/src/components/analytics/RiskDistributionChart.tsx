'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { ResponsiveContainer, PieChart, Pie, Cell, Legend, Tooltip } from 'recharts'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Loader2, PieChartIcon } from 'lucide-react'
import { analyticsApi, GroupSummaryInfo } from '@/lib/api/analytics'

interface RiskDistributionChartProps {
  className?: string
}

const COLORS = {
  critical: '#EF4444',
  high: '#F97316',
  medium: '#EAB308',
  low: '#22C55E',
}

export function RiskDistributionChart({ className }: RiskDistributionChartProps) {
  const t = useTranslations('analytics')
  const [groups, setGroups] = useState<GroupSummaryInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        setIsLoading(true)
        setError(null)
        const data = await analyticsApi.getAllGroupsSummary()
        setGroups(data)
      } catch (err) {
        console.error('Failed to fetch groups summary:', err)
        setError(t('loadError'))
      } finally {
        setIsLoading(false)
      }
    }

    fetchGroups()
  }, [t])

  // Aggregate risk distribution across all groups
  const aggregateRisk = groups.reduce(
    (acc, group) => ({
      critical: acc.critical + group.risk_distribution.critical,
      high: acc.high + group.risk_distribution.high,
      medium: acc.medium + group.risk_distribution.medium,
      low: acc.low + group.risk_distribution.low,
    }),
    { critical: 0, high: 0, medium: 0, low: 0 }
  )

  const pieData = [
    { name: t('riskLevel.critical'), value: aggregateRisk.critical, color: COLORS.critical },
    { name: t('riskLevel.high'), value: aggregateRisk.high, color: COLORS.high },
    { name: t('riskLevel.medium'), value: aggregateRisk.medium, color: COLORS.medium },
    { name: t('riskLevel.low'), value: aggregateRisk.low, color: COLORS.low },
  ].filter((item) => item.value > 0)

  const total = pieData.reduce((sum, item) => sum + item.value, 0)

  if (isLoading) {
    return (
      <div
        className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
      >
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div
        className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
      >
        <div className="text-center py-12">
          <p className="text-red-500">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div
      className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
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
        <div className="flex items-center gap-3 mb-6">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400">
            <PieChartIcon className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              {t('riskDistribution')}
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">{t('allGroups')}</p>
          </div>
        </div>

        {total === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">{t('noData')}</p>
          </div>
        ) : (
          <>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={90}
                    paddingAngle={2}
                    dataKey="value"
                  >
                    {pieData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'rgba(0, 0, 0, 0.8)',
                      border: '1px solid #374151',
                      borderRadius: '8px',
                      color: '#fff',
                    }}
                    formatter={(value: number) => [
                      `${value} (${((value / total) * 100).toFixed(1)}%)`,
                      t('students'),
                    ]}
                  />
                  <Legend
                    formatter={(value) => (
                      <span className="text-gray-700 dark:text-gray-300">{value}</span>
                    )}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* Summary */}
            <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 grid grid-cols-2 gap-4">
              <div className="text-center">
                <p className="text-xs text-gray-500 dark:text-gray-400">{t('totalStudents')}</p>
                <p className="text-lg font-semibold text-gray-900 dark:text-white">
                  {total.toLocaleString()}
                </p>
              </div>
              <div className="text-center">
                <p className="text-xs text-gray-500 dark:text-gray-400">{t('groupsCount')}</p>
                <p className="text-lg font-semibold text-gray-900 dark:text-white">
                  {groups.length}
                </p>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
