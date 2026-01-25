'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
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

// Fetcher for SWR - extracts data from wrapped response
const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)

  // Check if response is the API wrapper format
  if (response && typeof response === 'object' && 'success' in response) {
    if (response.success && response.data !== undefined) {
      return response.data
      /* c8 ignore start - Error handling and fallback paths */
    } else {
      throw new Error(response.error?.message || 'API returned error')
    }
  }

  // Response is already the data (shouldn't happen but handle it)
  return response as T
  /* c8 ignore stop */
}

// Hook for fetching dashboard stats
export function useDashboardStats(period: string = 'month') {
  const url = `${DASHBOARD_BASE_URL}/stats?period=${period}`

  const { data, error, isLoading, mutate } = useSWR<DashboardStats>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000, // 30 seconds
    refreshInterval: 60000, // Auto-refresh every minute
  })

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

  const { data, error, isLoading, mutate } = useSWR<DashboardTrends>(url, fetcher, {
    revalidateOnFocus: true,
    revalidateOnMount: true,
    dedupingInterval: 5000,
    refreshInterval: 60000, // Auto-refresh every minute
    shouldRetryOnError: true,
    errorRetryCount: 3,
  })

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

  const { data, error, isLoading, mutate } = useSWR<DashboardActivity>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 10000,
    refreshInterval: 30000, // Auto-refresh every 30 seconds
  })

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
