import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useDashboardStats,
  useDashboardTrends,
  useDashboardActivity,
  exportDashboard,
} from '../useDashboard'
import { apiClient } from '@/lib/api'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useDashboard hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useDashboardStats', () => {
    it('returns dashboard stats with default period', async () => {
      const mockStats = {
        total_documents: 100,
        pending_documents: 10,
        total_students: 500,
        active_tasks: 25,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockStats,
      })

      const { result } = renderHook(() => useDashboardStats(), { wrapper })

      await waitFor(() => {
        expect(result.current.stats).toEqual(mockStats)
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/stats?period=month')
    })

    it('uses custom period parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: {},
      })

      renderHook(() => useDashboardStats('week'), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/stats?period=week')
      })
    })

    it('handles API error', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: false,
        error: { code: 'ERROR', message: 'Something went wrong' },
      })

      const { result } = renderHook(() => useDashboardStats(), { wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeDefined()
      })
    })
  })

  describe('useDashboardTrends', () => {
    it('returns dashboard trends with default parameters', async () => {
      const mockTrends = {
        documents: [{ date: '2024-01-01', count: 10 }],
        tasks: [{ date: '2024-01-01', count: 5 }],
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockTrends,
      })

      const { result } = renderHook(() => useDashboardTrends(), { wrapper })

      await waitFor(() => {
        expect(result.current.trends).toEqual(mockTrends)
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/trends?period=month')
    })

    it('includes date range parameters when provided', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: {},
      })

      renderHook(() => useDashboardTrends('week', '2024-01-01', '2024-01-07'), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(
          expect.stringContaining('start_date=2024-01-01')
        )
      })
    })

    it('returns isLoading while fetching', () => {
      mockedApiClient.get.mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      const { result } = renderHook(() => useDashboardTrends(), { wrapper })

      expect(result.current.isLoading).toBe(true)
    })
  })

  describe('useDashboardActivity', () => {
    it('returns recent activities with default limit', async () => {
      const mockActivity = {
        activities: [
          { id: 1, type: 'document_created', title: 'New document' },
          { id: 2, type: 'task_completed', title: 'Task done' },
        ],
        total: 2,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockActivity,
      })

      const { result } = renderHook(() => useDashboardActivity(), { wrapper })

      await waitFor(() => {
        expect(result.current.activities).toHaveLength(2)
      })

      expect(result.current.total).toBe(2)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/activity?limit=10')
    })

    it('uses custom limit parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { activities: [], total: 0 },
      })

      renderHook(() => useDashboardActivity(5), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/activity?limit=5')
      })
    })

    it('returns empty arrays when no data', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: null,
      })

      const { result } = renderHook(() => useDashboardActivity(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.activities).toEqual([])
      expect(result.current.total).toBe(0)
    })
  })

  describe('exportDashboard', () => {
    it('exports dashboard data', async () => {
      const mockExportResult = {
        url: 'https://example.com/export/123.pdf',
        filename: 'dashboard-export.pdf',
      }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockExportResult,
      })

      const result = await exportDashboard({
        format: 'pdf',
        start_date: '2024-01-01',
        end_date: '2024-01-31',
      })

      expect(result).toEqual(mockExportResult)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/dashboard/export', {
        format: 'pdf',
        start_date: '2024-01-01',
        end_date: '2024-01-31',
      })
    })
  })
})
