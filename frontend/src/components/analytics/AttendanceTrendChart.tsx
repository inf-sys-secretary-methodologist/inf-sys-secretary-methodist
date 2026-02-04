'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Loader2, TrendingUp } from 'lucide-react'
import { analyticsApi, MonthlyTrendInfo } from '@/lib/api/analytics'

interface AttendanceTrendChartProps {
  months?: number
  className?: string
}

export function AttendanceTrendChart({ months = 6, className }: AttendanceTrendChartProps) {
  const t = useTranslations('analytics')
  const [trends, setTrends] = useState<MonthlyTrendInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchTrends = async () => {
      try {
        setIsLoading(true)
        setError(null)
        const data = await analyticsApi.getAttendanceTrend(months)
        setTrends(data)
      } catch (err) {
        console.error('Failed to fetch attendance trend:', err)
        setError(t('loadError'))
      } finally {
        setIsLoading(false)
      }
    }

    fetchTrends()
  }, [months, t])

  // Transform data for chart
  const chartData = trends.map((trend) => ({
    month: trend.month,
    attendance: trend.attendance_rate,
    present: trend.present_count,
    absent: trend.absent_count,
    students: trend.unique_students,
  }))

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
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
            <TrendingUp className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              {t('attendanceTrend')}
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {t('lastMonths', { months })}
            </p>
          </div>
        </div>

        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
              <defs>
                <linearGradient id="gradientAttendance" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3B82F6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
              <XAxis
                dataKey="month"
                stroke="#9CA3AF"
                fontSize={12}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke="#9CA3AF"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                domain={[0, 100]}
                tickFormatter={(value) => `${value}%`}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(0, 0, 0, 0.8)',
                  border: '1px solid #374151',
                  borderRadius: '8px',
                  color: '#fff',
                }}
                formatter={(value: number) => [`${value.toFixed(1)}%`, t('attendanceRate')]}
              />
              <Legend />
              <Area
                type="monotone"
                dataKey="attendance"
                name={t('attendanceRate')}
                stroke="#3B82F6"
                fill="url(#gradientAttendance)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Summary Stats */}
        {trends.length > 0 && (
          <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 grid grid-cols-3 gap-4">
            <div className="text-center">
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('totalRecords')}</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">
                {trends.reduce((sum, t) => sum + t.total_records, 0).toLocaleString()}
              </p>
            </div>
            <div className="text-center">
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('avgAttendance')}</p>
              <p className="text-lg font-semibold text-blue-600 dark:text-blue-400">
                {(trends.reduce((sum, t) => sum + t.attendance_rate, 0) / trends.length).toFixed(1)}
                %
              </p>
            </div>
            <div className="text-center">
              <p className="text-xs text-gray-500 dark:text-gray-400">{t('uniqueStudents')}</p>
              <p className="text-lg font-semibold text-gray-900 dark:text-white">
                {Math.max(...trends.map((t) => t.unique_students)).toLocaleString()}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
