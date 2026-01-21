import { dashboardApi } from '../dashboard'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('dashboardApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('getStats', () => {
    it('fetches stats with default period', async () => {
      const mockStats = { total_users: 100, total_documents: 50 }
      mockedApiClient.get.mockResolvedValue(mockStats)

      const result = await dashboardApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/stats?period=month')
      expect(result).toEqual(mockStats)
    })

    it('fetches stats with custom period', async () => {
      const mockStats = { total_users: 100, total_documents: 50 }
      mockedApiClient.get.mockResolvedValue(mockStats)

      const result = await dashboardApi.getStats('week')

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/stats?period=week')
      expect(result).toEqual(mockStats)
    })
  })

  describe('getTrends', () => {
    it('fetches trends with default period', async () => {
      const mockTrends = { labels: ['Jan', 'Feb'], data: [10, 20] }
      mockedApiClient.get.mockResolvedValue(mockTrends)

      const result = await dashboardApi.getTrends()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/trends?period=month')
      expect(result).toEqual(mockTrends)
    })

    it('fetches trends with custom period and dates', async () => {
      const mockTrends = { labels: ['Jan'], data: [10] }
      mockedApiClient.get.mockResolvedValue(mockTrends)

      await dashboardApi.getTrends('week', '2024-01-01', '2024-01-07')

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('start_date=2024-01-01')
      )
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('end_date=2024-01-07')
      )
    })
  })

  describe('getActivity', () => {
    it('fetches activity with default limit', async () => {
      const mockActivity = { items: [{ id: 1, action: 'created' }] }
      mockedApiClient.get.mockResolvedValue(mockActivity)

      const result = await dashboardApi.getActivity()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/activity?limit=10')
      expect(result).toEqual(mockActivity)
    })

    it('fetches activity with custom limit', async () => {
      const mockActivity = { items: [] }
      mockedApiClient.get.mockResolvedValue(mockActivity)

      await dashboardApi.getActivity(5)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/dashboard/activity?limit=5')
    })
  })

  describe('exportDashboard', () => {
    it('exports dashboard data', async () => {
      const mockExport = { url: '/exports/dashboard.pdf' }
      mockedApiClient.post.mockResolvedValue(mockExport)

      const input = { format: 'pdf' as const, include_charts: true }
      const result = await dashboardApi.exportDashboard(input)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/dashboard/export', input)
      expect(result).toEqual(mockExport)
    })
  })
})
