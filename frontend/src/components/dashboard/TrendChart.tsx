'use client'

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
import type { TrendPoint } from '@/types/dashboard'

interface TrendDataset {
  name: string
  data: TrendPoint[]
  color: string
}

interface TrendChartProps {
  title: string
  datasets: TrendDataset[]
  period?: 'week' | 'month' | 'quarter' | 'year'
  className?: string
}

interface ChartDataPoint {
  date: string
  [key: string]: string | number
}

// Generate date range for the last N days
function generateDateRange(days: number): Date[] {
  const dates: Date[] = []
  const today = new Date()
  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    dates.push(date)
  }
  return dates
}

export function TrendChart({ title, datasets, period = 'month', className }: TrendChartProps) {
  /* c8 ignore start - Period calculation helper, tested in e2e */
  // Generate date range based on period
  const getDaysForPeriod = (p: string): number => {
    switch (p) {
      case 'week':
        return 7
      case 'month':
        return 30
      case 'quarter':
        return 90
      case 'year':
        return 365
      default:
        return 30
    }
  }
  /* c8 ignore stop */
  const dateRange = generateDateRange(getDaysForPeriod(period))

  // Create map for quick lookup of data points
  const dataByDate = new Map<string, Map<string, number>>()

  datasets.forEach((dataset) => {
    dataset.data.forEach((point) => {
      const dateKey = new Date(point.date).toISOString().split('T')[0]
      if (!dataByDate.has(dateKey)) {
        dataByDate.set(dateKey, new Map())
      }
      dataByDate.get(dateKey)!.set(dataset.name, point.value)
    })
  })

  // Build chart data with all dates in range, filling missing with 0
  const chartData: ChartDataPoint[] = dateRange.map((date) => {
    const dateKey = date.toISOString().split('T')[0]
    const dateStr = date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
    })

    const dataPoint: ChartDataPoint = { date: dateStr }
    datasets.forEach((dataset) => {
      dataPoint[dataset.name] = dataByDate.get(dateKey)?.get(dataset.name) || 0
    })
    return dataPoint
  })

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
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">{title}</h3>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
              <defs>
                {datasets.map((dataset, index) => (
                  <linearGradient
                    key={dataset.name}
                    id={`gradient-${index}`}
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="5%" stopColor={dataset.color} stopOpacity={0.3} />
                    <stop offset="95%" stopColor={dataset.color} stopOpacity={0} />
                  </linearGradient>
                ))}
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.3} />
              <XAxis
                dataKey="date"
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
                domain={[0, (dataMax: number) => Math.ceil(dataMax * 1.1) || 1]}
                allowDecimals={false}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'rgba(0, 0, 0, 0.8)',
                  border: '1px solid #374151',
                  borderRadius: '8px',
                  color: '#fff',
                }}
              />
              <Legend />
              {datasets.map((dataset, index) => (
                <Area
                  key={dataset.name}
                  type="monotone"
                  dataKey={dataset.name}
                  stroke={dataset.color}
                  fill={`url(#gradient-${index})`}
                  strokeWidth={2}
                />
              ))}
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
