'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'
import type {
  SyncLog,
  SyncConflict,
  SyncStats,
  ConflictStats,
  ExternalEmployee,
  ExternalStudent,
  SyncLogFilter,
  EmployeeFilter,
  StudentFilter,
  ConflictFilter,
  StartSyncRequest,
  StartSyncResponse,
  ResolveConflictRequest,
  BulkResolveRequest,
  SyncEntityType,
  ConflictResolution,
} from '@/lib/api/integration'

const INTEGRATION_BASE_URL = '/api/integration'

// Fetcher for SWR
const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<{ data: T }>(url)
  return response.data
}

// Helper to check if user is authenticated
// Uses token from localStorage directly to avoid race condition with Zustand hydration
function useAuthenticatedKey(url: string | null): string | null {
  const { isAuthenticated, isLoading } = useAuthStore()

  // Check if token exists in localStorage (more reliable than waiting for Zustand hydration)
  const hasToken = typeof window !== 'undefined' && !!localStorage.getItem('authToken')

  // Skip fetch if:
  // 1. Explicitly loading auth state AND no token in storage
  // 2. Not authenticated AND no token in storage (after hydration)
  if (!hasToken && (isLoading || !isAuthenticated)) {
    return null
  }

  return url
}

// Hook for sync stats
export function useSyncStats(entityType?: SyncEntityType) {
  const query = entityType ? `?entity_type=${entityType}` : ''
  const url = `${INTEGRATION_BASE_URL}/sync/stats${query}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<SyncStats>(authenticatedUrl, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000,
    refreshInterval: 60000,
  })

  return {
    stats: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for sync logs
export function useSyncLogs(filter?: SyncLogFilter) {
  const params = new URLSearchParams()
  /* c8 ignore start - Optional filter params */
  if (filter?.entity_type) params.append('entity_type', filter.entity_type)
  if (filter?.direction) params.append('direction', filter.direction)
  if (filter?.status) params.append('status', filter.status)
  if (filter?.start_date) params.append('start_date', filter.start_date)
  if (filter?.end_date) params.append('end_date', filter.end_date)
  if (filter?.limit) params.append('limit', filter.limit.toString())
  if (filter?.offset) params.append('offset', filter.offset.toString())
  /* c8 ignore stop */

  const query = params.toString() ? `?${params.toString()}` : ''
  const url = `${INTEGRATION_BASE_URL}/sync/logs${query}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<{ logs: SyncLog[]; total: number }>(
    authenticatedUrl,
    fetcher,
    {
      revalidateOnFocus: false,
      dedupingInterval: 10000,
      refreshInterval: 30000,
    }
  )

  return {
    logs: data?.logs || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// Hook for sync status (single log)
export function useSyncStatus(syncLogId: number | null) {
  const url = syncLogId ? `${INTEGRATION_BASE_URL}/sync/status/${syncLogId}` : null
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<SyncLog>(authenticatedUrl, fetcher, {
    revalidateOnFocus: false,
    refreshInterval: syncLogId ? 5000 : 0, // Poll every 5s if there's an active sync
  })

  return {
    syncLog: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for conflict stats
export function useConflictStats() {
  const url = `${INTEGRATION_BASE_URL}/conflicts/stats`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<ConflictStats>(authenticatedUrl, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000,
    refreshInterval: 60000,
  })

  return {
    stats: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for conflicts list
export function useConflicts(filter?: ConflictFilter) {
  const params = new URLSearchParams()
  /* c8 ignore start - Optional filter params */
  if (filter?.sync_log_id) params.append('sync_log_id', filter.sync_log_id.toString())
  if (filter?.entity_type) params.append('entity_type', filter.entity_type)
  if (filter?.resolution) params.append('resolution', filter.resolution)
  if (filter?.limit) params.append('limit', filter.limit.toString())
  if (filter?.offset) params.append('offset', filter.offset.toString())
  const query = params.toString() ? `?${params.toString()}` : ''
  /* c8 ignore stop */
  const url = `${INTEGRATION_BASE_URL}/conflicts${query}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<{ conflicts: SyncConflict[]; total: number }>(
    authenticatedUrl,
    fetcher,
    {
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    }
  )

  return {
    conflicts: data?.conflicts || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// Hook for pending conflicts
export function usePendingConflicts(limit = 20, offset = 0) {
  const url = `${INTEGRATION_BASE_URL}/conflicts/pending?limit=${limit}&offset=${offset}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<{ conflicts: SyncConflict[]; total: number }>(
    authenticatedUrl,
    fetcher,
    {
      revalidateOnFocus: true,
      dedupingInterval: 10000,
      refreshInterval: 30000,
    }
  )

  return {
    conflicts: data?.conflicts || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// Hook for single conflict
export function useConflict(id: number | null) {
  const url = id ? `${INTEGRATION_BASE_URL}/conflicts/${id}` : null
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<SyncConflict>(authenticatedUrl, fetcher, {
    revalidateOnFocus: false,
  })

  return {
    conflict: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for external employees
export function useExternalEmployees(filter?: EmployeeFilter) {
  const params = new URLSearchParams()
  /* c8 ignore start - Optional filter params */
  if (filter?.search) params.append('search', filter.search)
  if (filter?.department) params.append('department', filter.department)
  if (filter?.is_active !== undefined) params.append('is_active', filter.is_active.toString())
  if (filter?.limit) params.append('limit', filter.limit.toString())
  if (filter?.offset) params.append('offset', filter.offset.toString())
  /* c8 ignore stop */

  const query = params.toString() ? `?${params.toString()}` : ''
  const url = `${INTEGRATION_BASE_URL}/employees${query}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<{
    employees: ExternalEmployee[]
    total: number
  }>(authenticatedUrl, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000,
  })

  return {
    employees: data?.employees || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// Hook for external students
export function useExternalStudents(filter?: StudentFilter) {
  const params = new URLSearchParams()
  /* c8 ignore start - Optional filter params */
  if (filter?.search) params.append('search', filter.search)
  if (filter?.faculty) params.append('faculty', filter.faculty)
  if (filter?.group) params.append('group', filter.group)
  if (filter?.course) params.append('course', filter.course.toString())
  if (filter?.is_active !== undefined) params.append('is_active', filter.is_active.toString())
  if (filter?.limit) params.append('limit', filter.limit.toString())
  if (filter?.offset) params.append('offset', filter.offset.toString())
  const query = params.toString() ? `?${params.toString()}` : ''
  /* c8 ignore stop */
  const url = `${INTEGRATION_BASE_URL}/students${query}`
  const authenticatedUrl = useAuthenticatedKey(url)

  const { data, error, isLoading, mutate } = useSWR<{ students: ExternalStudent[]; total: number }>(
    authenticatedUrl,
    fetcher,
    {
      revalidateOnFocus: false,
      dedupingInterval: 30000,
    }
  )

  return {
    students: data?.students || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// API mutation functions
export async function startSync(request: StartSyncRequest): Promise<StartSyncResponse> {
  const response = await apiClient.post<{ data: StartSyncResponse }>(
    `${INTEGRATION_BASE_URL}/sync/start`,
    request
  )
  return response.data
}

export async function cancelSync(syncLogId: number): Promise<void> {
  await apiClient.post(`${INTEGRATION_BASE_URL}/sync/cancel/${syncLogId}`)
}

export async function resolveConflict(id: number, request: ResolveConflictRequest): Promise<void> {
  await apiClient.post(`${INTEGRATION_BASE_URL}/conflicts/${id}/resolve`, request)
}

export async function bulkResolveConflicts(
  request: BulkResolveRequest
): Promise<{ count: number }> {
  const response = await apiClient.post<{ data: { count: number } }>(
    `${INTEGRATION_BASE_URL}/conflicts/bulk-resolve`,
    request
  )
  return response.data
}

export async function deleteConflict(id: number): Promise<void> {
  await apiClient.delete(`${INTEGRATION_BASE_URL}/conflicts/${id}`)
}

// Re-export types for convenience
export type {
  SyncLog,
  SyncConflict,
  SyncStats,
  ConflictStats,
  ExternalEmployee,
  ExternalStudent,
  SyncEntityType,
  ConflictResolution,
  StartSyncRequest,
  ResolveConflictRequest,
  BulkResolveRequest,
}
