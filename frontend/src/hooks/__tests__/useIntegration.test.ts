import { renderHook, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useSyncStats,
  useSyncLogs,
  useSyncStatus,
  useConflictStats,
  useConflicts,
  usePendingConflicts,
  useConflict,
  useExternalEmployees,
  useExternalStudents,
  startSync,
  cancelSync,
  resolveConflict,
  bulkResolveConflicts,
  deleteConflict,
} from '../useIntegration'
import { apiClient } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

// Mock the auth store
jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(),
}))

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

const mockedApiClient = jest.mocked(apiClient)
const mockedUseAuthStore = jest.mocked(useAuthStore)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useIntegration hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockLocalStorage.getItem.mockReturnValue('test-token')
    mockedUseAuthStore.mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
    } as ReturnType<typeof useAuthStore>)
  })

  describe('useSyncStats', () => {
    it('returns sync stats', async () => {
      const mockStats = {
        total_syncs: 100,
        successful_syncs: 95,
        failed_syncs: 5,
        last_sync_at: '2024-01-01T00:00:00Z',
      }

      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const { result } = renderHook(() => useSyncStats(), { wrapper })

      await waitFor(() => {
        expect(result.current.stats).toEqual(mockStats)
      })
    })

    it('passes entity type filter', async () => {
      mockedApiClient.get.mockResolvedValue({ data: {} })

      renderHook(() => useSyncStats('employee'), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(
          expect.stringContaining('entity_type=employee')
        )
      })
    })

    it('skips fetch when not authenticated', async () => {
      mockLocalStorage.getItem.mockReturnValue(null)
      mockedUseAuthStore.mockReturnValue({
        isAuthenticated: false,
        isLoading: false,
      } as ReturnType<typeof useAuthStore>)

      renderHook(() => useSyncStats(), { wrapper })

      // Should not call API when not authenticated
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useSyncLogs', () => {
    it('returns sync logs', async () => {
      const mockLogs = {
        logs: [
          { id: 1, entity_type: 'employee', status: 'completed' },
          { id: 2, entity_type: 'student', status: 'running' },
        ],
        total: 2,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockLogs })

      const { result } = renderHook(() => useSyncLogs(), { wrapper })

      await waitFor(() => {
        expect(result.current.logs).toHaveLength(2)
      })

      expect(result.current.total).toBe(2)
    })

    it('passes filter parameters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { logs: [], total: 0 } })

      renderHook(() => useSyncLogs({ entity_type: 'employee', status: 'completed', limit: 10 }), {
        wrapper,
      })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(
          expect.stringContaining('entity_type=employee')
        )
      })
    })
  })

  describe('useSyncStatus', () => {
    it('fetches sync status by id', async () => {
      const mockLog = { id: 1, status: 'running', progress: 50 }

      mockedApiClient.get.mockResolvedValue({ data: mockLog })

      const { result } = renderHook(() => useSyncStatus(1), { wrapper })

      await waitFor(() => {
        expect(result.current.syncLog).toEqual(mockLog)
      })
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useSyncStatus(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useConflictStats', () => {
    it('returns conflict stats', async () => {
      const mockStats = {
        total_conflicts: 10,
        pending_conflicts: 3,
        resolved_conflicts: 7,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockStats })

      const { result } = renderHook(() => useConflictStats(), { wrapper })

      await waitFor(() => {
        expect(result.current.stats).toEqual(mockStats)
      })
    })
  })

  describe('useConflicts', () => {
    it('returns conflicts list', async () => {
      const mockData = {
        conflicts: [{ id: 1, entity_type: 'employee', resolution: 'pending' }],
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockData })

      const { result } = renderHook(() => useConflicts(), { wrapper })

      await waitFor(() => {
        expect(result.current.conflicts).toHaveLength(1)
      })
    })
  })

  describe('usePendingConflicts', () => {
    it('returns pending conflicts', async () => {
      const mockData = {
        conflicts: [{ id: 1 }, { id: 2 }],
        total: 2,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockData })

      const { result } = renderHook(() => usePendingConflicts(20, 0), { wrapper })

      await waitFor(() => {
        expect(result.current.conflicts).toHaveLength(2)
      })
    })
  })

  describe('useConflict', () => {
    it('fetches single conflict', async () => {
      const mockConflict = { id: 1, entity_type: 'employee' }

      mockedApiClient.get.mockResolvedValue({ data: mockConflict })

      const { result } = renderHook(() => useConflict(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conflict).toEqual(mockConflict)
      })
    })

    it('does not fetch when id is null', () => {
      renderHook(() => useConflict(null), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useExternalEmployees', () => {
    it('returns external employees', async () => {
      const mockData = {
        employees: [{ id: 1, name: 'John Doe' }],
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockData })

      const { result } = renderHook(() => useExternalEmployees(), { wrapper })

      await waitFor(() => {
        expect(result.current.employees).toHaveLength(1)
      })
    })

    it('passes filter parameters', async () => {
      mockedApiClient.get.mockResolvedValue({ data: { employees: [], total: 0 } })

      renderHook(() => useExternalEmployees({ search: 'John', is_active: true }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('search=John'))
      })
    })
  })

  describe('useExternalStudents', () => {
    it('returns external students', async () => {
      const mockData = {
        students: [{ id: 1, name: 'Jane Student' }],
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({ data: mockData })

      const { result } = renderHook(() => useExternalStudents(), { wrapper })

      await waitFor(() => {
        expect(result.current.students).toHaveLength(1)
      })
    })
  })

  describe('startSync', () => {
    it('starts sync', async () => {
      const mockResponse = { sync_log_id: 1, status: 'started' }
      mockedApiClient.post.mockResolvedValue({ data: mockResponse })

      const result = await startSync({ entity_type: 'employee' })

      expect(result).toEqual(mockResponse)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/sync/start', {
        entity_type: 'employee',
      })
    })
  })

  describe('cancelSync', () => {
    it('cancels sync', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      await cancelSync(1)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/sync/cancel/1')
    })
  })

  describe('resolveConflict', () => {
    it('resolves conflict', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      await resolveConflict(1, { resolution: 'use_local' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/conflicts/1/resolve', {
        resolution: 'use_local',
      })
    })
  })

  describe('bulkResolveConflicts', () => {
    it('bulk resolves conflicts', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { count: 5 } })

      const result = await bulkResolveConflicts({
        ids: [1, 2, 3, 4, 5],
        resolution: 'use_external',
      })

      expect(result).toEqual({ count: 5 })
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/integration/conflicts/bulk-resolve', {
        ids: [1, 2, 3, 4, 5],
        resolution: 'use_external',
      })
    })
  })

  describe('deleteConflict', () => {
    it('deletes conflict', async () => {
      mockedApiClient.delete.mockResolvedValue({ success: true })

      await deleteConflict(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/integration/conflicts/1')
    })
  })
})
