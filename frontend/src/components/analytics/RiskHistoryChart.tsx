'use client'

import { useEffect, useState } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { useTranslations } from 'next-intl'
import { analyticsApi, RiskHistoryEntry } from '@/lib/api/analytics'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface RiskHistoryChartProps {
  studentId: number
  limit?: number
}

export function RiskHistoryChart({ studentId, limit = 90 }: RiskHistoryChartProps) {
  const t = useTranslations('analytics')
  const [data, setData] = useState<RiskHistoryEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const response = await analyticsApi.getStudentRiskHistory(studentId, limit)
        // Reverse to show oldest first (chronological)
        setData([...response.history].reverse())
      } catch {
        setData([])
      } finally {
        setLoading(false)
      }
    }
    fetchHistory()
  }, [studentId, limit])

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-64 w-full" />
        </CardContent>
      </Card>
    )
  }

  if (data.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{t('riskHistory.title')}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">{t('riskHistory.noData')}</p>
        </CardContent>
      </Card>
    )
  }

  const chartData = data.map((entry) => ({
    date: new Date(entry.calculated_at).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
    }),
    score: entry.risk_score,
    attendance: entry.attendance_rate ?? 0,
    grade: entry.grade_average ?? 0,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('riskHistory.title')}</CardTitle>
      </CardHeader>
      <CardContent>
        <div style={{ width: '100%', height: 300 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData} margin={{ top: 5, right: 20, bottom: 5, left: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
              <XAxis dataKey="date" tick={{ fontSize: 12 }} />
              <YAxis domain={[0, 100]} tick={{ fontSize: 12 }} />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'hsl(var(--card))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '8px',
                }}
              />
              <ReferenceLine y={70} stroke="hsl(var(--destructive))" strokeDasharray="5 5" label="High Risk" />
              <Line
                type="monotone"
                dataKey="score"
                stroke="hsl(var(--primary))"
                strokeWidth={2}
                dot={{ r: 3 }}
                name={t('riskHistory.riskScore')}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}
