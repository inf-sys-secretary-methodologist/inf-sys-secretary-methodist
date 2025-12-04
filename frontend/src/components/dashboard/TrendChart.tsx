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
import type { TrendPoint } from '@/types/dashboard'

interface TrendDataset {
  name: string
  data: TrendPoint[]
  color: string
}

interface TrendChartProps {
  title: string
  datasets: TrendDataset[]
  className?: string
}

interface ChartDataPoint {
  date: string
  [key: string]: string | number
}

export function TrendChart({ title, datasets, className }: TrendChartProps) {
  // Merge all datasets into a single array for recharts
  const chartData: ChartDataPoint[] = []
  const dateMap = new Map<string, ChartDataPoint>()

  datasets.forEach((dataset) => {
    dataset.data.forEach((point) => {
      const dateStr = new Date(point.date).toLocaleDateString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
      })
      if (!dateMap.has(dateStr)) {
        dateMap.set(dateStr, { date: dateStr })
      }
      const existing = dateMap.get(dateStr)!
      existing[dataset.name] = point.value
    })
  })

  // Sort by date
  const sortedDates = Array.from(dateMap.keys()).sort((a, b) => {
    const [dayA, monthA] = a.split('.').map(Number)
    const [dayB, monthB] = b.split('.').map(Number)
    if (monthA !== monthB) return monthA - monthB
    return dayA - dayB
  })

  sortedDates.forEach((date) => {
    chartData.push(dateMap.get(date)!)
  })

  return (
    <div
      className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
    >
      <div className="relative z-10">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">{title}</h3>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
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
              <YAxis stroke="#9CA3AF" fontSize={12} tickLine={false} axisLine={false} />
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
