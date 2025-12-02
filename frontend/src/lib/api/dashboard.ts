import { apiClient } from '../api'
import type {
  DashboardStats,
  DashboardTrends,
  DashboardActivity,
  ExportInput,
  ExportOutput,
} from '@/types/dashboard'

export const dashboardApi = {
  getStats: async (period: string = 'month'): Promise<DashboardStats> => {
    return apiClient.get<DashboardStats>(`/api/dashboard/stats?period=${period}`)
  },

  getTrends: async (
    period: string = 'month',
    startDate?: string,
    endDate?: string
  ): Promise<DashboardTrends> => {
    const params = new URLSearchParams({ period })
    if (startDate) params.append('start_date', startDate)
    if (endDate) params.append('end_date', endDate)
    return apiClient.get<DashboardTrends>(`/api/dashboard/trends?${params.toString()}`)
  },

  getActivity: async (limit: number = 10): Promise<DashboardActivity> => {
    return apiClient.get<DashboardActivity>(`/api/dashboard/activity?limit=${limit}`)
  },

  exportDashboard: async (input: ExportInput): Promise<ExportOutput> => {
    return apiClient.post<ExportOutput>('/api/dashboard/export', input)
  },
}
