'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { swrFetcher } from '@/lib/api/fetchers'
import { SWR_DEDUPING, SWR_REFRESH } from '@/config/swr'
import type {
  DashboardStats,
  DashboardTrends,
  DashboardActivity,
  ExportInput,
  ExportOutput,
} from '@/types/dashboard'

const DASHBOARD_BASE_URL = '/api/dashboard'

// API Response wrapper type from backend
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
  meta?: {
    request_id: string
    timestamp: string
    version: string
  }
}

// Hook for fetching dashboard stats
export function useDashboardStats(period: string = 'month') {
  const url = `${DASHBOARD_BASE_URL}/stats?period=${period}`

  const { data, error, isLoading, mutate } = useSWR<DashboardStats>(
    url,
    swrFetcher<DashboardStats>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.LONG,
      refreshInterval: SWR_REFRESH.STANDARD, // Auto-refresh every minute
    }
  )

  return {
    stats: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for fetching dashboard trends
export function useDashboardTrends(period: string = 'month', startDate?: string, endDate?: string) {
  const params = new URLSearchParams({ period })
  if (startDate) params.append('start_date', startDate)
  if (endDate) params.append('end_date', endDate)
  const url = `${DASHBOARD_BASE_URL}/trends?${params.toString()}`

  const { data, error, isLoading, mutate } = useSWR<DashboardTrends>(
    url,
    swrFetcher<DashboardTrends>,
    {
      revalidateOnFocus: true,
      revalidateOnMount: true,
      dedupingInterval: SWR_DEDUPING.SHORT,
      refreshInterval: SWR_REFRESH.STANDARD, // Auto-refresh every minute
      shouldRetryOnError: true,
      errorRetryCount: 3,
    }
  )

  return {
    trends: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for fetching recent activity
export function useDashboardActivity(limit: number = 10) {
  const url = `${DASHBOARD_BASE_URL}/activity?limit=${limit}`

  const { data, error, isLoading, mutate } = useSWR<DashboardActivity>(
    url,
    swrFetcher<DashboardActivity>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.MEDIUM,
      refreshInterval: SWR_REFRESH.REALTIME, // Auto-refresh every 30 seconds
    }
  )

  return {
    activities: data?.activities || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// Export dashboard function
export async function exportDashboard(input: ExportInput): Promise<ExportOutput> {
  const response = await apiClient.post<ApiResponse<ExportOutput>>(
    `${DASHBOARD_BASE_URL}/export`,
    input
  )
  return response.data
}
