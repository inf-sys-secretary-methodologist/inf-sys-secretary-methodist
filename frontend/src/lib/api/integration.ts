import { apiClient } from '../api'

// Types
export type SyncEntityType = 'employee' | 'student'
export type SyncDirection = 'import' | 'export' | 'bidirectional'
export type SyncStatus = 'pending' | 'in_progress' | 'completed' | 'failed' | 'cancelled'
export type ConflictResolution = 'pending' | 'use_local' | 'use_external' | 'merge' | 'skip'

export interface SyncLog {
  id: number
  entity_type: SyncEntityType
  direction: SyncDirection
  status: SyncStatus
  started_at: string
  completed_at?: string
  total_records: number
  processed_count: number
  success_count: number
  error_count: number
  conflict_count: number
  error_message?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface SyncConflict {
  id: number
  sync_log_id: number
  entity_type: SyncEntityType
  entity_id: string
  local_data: string
  external_data: string
  conflict_type: string
  conflict_fields: string[]
  resolution: ConflictResolution
  resolved_by?: number
  resolved_at?: string
  resolved_data?: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface ExternalEmployee {
  id: number
  external_id: string
  first_name: string
  last_name: string
  middle_name?: string
  email?: string
  phone?: string
  position?: string
  department?: string
  hire_date?: string
  is_active: boolean
  data_hash: string
  last_sync_at: string
  created_at: string
  updated_at: string
}

export interface ExternalStudent {
  id: number
  external_id: string
  first_name: string
  last_name: string
  middle_name?: string
  email?: string
  phone?: string
  student_id?: string
  group_name?: string
  faculty?: string
  specialization?: string
  course?: number
  enrollment_date?: string
  graduation_date?: string
  status?: string
  is_active: boolean
  data_hash: string
  last_sync_at: string
  created_at: string
  updated_at: string
}

export interface SyncStats {
  total_syncs: number
  successful_syncs: number
  failed_syncs: number
  total_records: number
  total_conflicts: number
  last_sync_at: string
}

export interface ConflictStats {
  total_conflicts: number
  pending_conflicts: number
  resolved_conflicts: number
  by_entity_type: Record<SyncEntityType, number>
}

export interface StartSyncRequest {
  entity_type: SyncEntityType
  direction?: SyncDirection
  force?: boolean
}

export interface StartSyncResponse {
  sync_log_id: number
  status: SyncStatus
  message: string
  total_records: number
  success_count: number
  error_count: number
  conflict_count: number
}

export interface ResolveConflictRequest {
  resolution: ConflictResolution
  resolved_data?: string
  notes?: string
}

export interface BulkResolveRequest {
  ids: number[]
  resolution: ConflictResolution
}

export interface SyncLogFilter {
  entity_type?: SyncEntityType
  direction?: SyncDirection
  status?: SyncStatus
  start_date?: string
  end_date?: string
  limit?: number
  offset?: number
}

export interface EmployeeFilter {
  search?: string
  department?: string
  is_active?: boolean
  limit?: number
  offset?: number
}

export interface StudentFilter {
  search?: string
  faculty?: string
  group?: string
  course?: number
  is_active?: boolean
  limit?: number
  offset?: number
}

export interface ConflictFilter {
  sync_log_id?: number
  entity_type?: SyncEntityType
  resolution?: ConflictResolution
  limit?: number
  offset?: number
}

interface ApiResponse<T> {
  data: T
}

interface PaginatedList<T> {
  items: T[]
  total: number
}

// Sync API
export const syncApi = {
  // Start synchronization
  start: async (request: StartSyncRequest): Promise<ApiResponse<StartSyncResponse>> => {
    return apiClient.post('/api/integration/sync/start', request)
  },

  // Get sync status
  getStatus: async (syncLogId: number): Promise<ApiResponse<SyncLog>> => {
    return apiClient.get(`/api/integration/sync/status/${syncLogId}`)
  },

  // Cancel sync
  cancel: async (syncLogId: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post(`/api/integration/sync/cancel/${syncLogId}`)
  },

  // Get sync logs
  getLogs: async (
    filter?: SyncLogFilter
  ): Promise<ApiResponse<PaginatedList<SyncLog> & { logs: SyncLog[] }>> => {
    const params = new URLSearchParams()
    if (filter?.entity_type) params.append('entity_type', filter.entity_type)
    if (filter?.direction) params.append('direction', filter.direction)
    if (filter?.status) params.append('status', filter.status)
    if (filter?.start_date) params.append('start_date', filter.start_date)
    if (filter?.end_date) params.append('end_date', filter.end_date)
    if (filter?.limit) params.append('limit', filter.limit.toString())
    if (filter?.offset) params.append('offset', filter.offset.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    return apiClient.get(`/api/integration/sync/logs${query}`)
  },

  // Get sync stats
  getStats: async (entityType?: SyncEntityType): Promise<ApiResponse<SyncStats>> => {
    const query = entityType ? `?entity_type=${entityType}` : ''
    return apiClient.get(`/api/integration/sync/stats${query}`)
  },
}

// Employees API
export const employeesApi = {
  // List employees
  list: async (
    filter?: EmployeeFilter
  ): Promise<ApiResponse<PaginatedList<ExternalEmployee> & { employees: ExternalEmployee[] }>> => {
    const params = new URLSearchParams()
    if (filter?.search) params.append('search', filter.search)
    if (filter?.department) params.append('department', filter.department)
    if (filter?.is_active !== undefined) params.append('is_active', filter.is_active.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())
    if (filter?.offset) params.append('offset', filter.offset.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    return apiClient.get(`/api/integration/employees${query}`)
  },

  // Get employee by ID
  getById: async (id: number): Promise<ApiResponse<ExternalEmployee>> => {
    return apiClient.get(`/api/integration/employees/${id}`)
  },

  // Get employee by external ID
  getByExternalId: async (externalId: string): Promise<ApiResponse<ExternalEmployee>> => {
    return apiClient.get(`/api/integration/employees/external/${externalId}`)
  },

  // Get employee stats
  getStats: async (): Promise<
    ApiResponse<{
      total: number
      active: number
      inactive: number
      by_department: Record<string, number>
    }>
  > => {
    return apiClient.get('/api/integration/employees/stats')
  },
}

// Students API
export const studentsApi = {
  // List students
  list: async (
    filter?: StudentFilter
  ): Promise<ApiResponse<PaginatedList<ExternalStudent> & { students: ExternalStudent[] }>> => {
    const params = new URLSearchParams()
    if (filter?.search) params.append('search', filter.search)
    if (filter?.faculty) params.append('faculty', filter.faculty)
    if (filter?.group) params.append('group', filter.group)
    if (filter?.course) params.append('course', filter.course.toString())
    if (filter?.is_active !== undefined) params.append('is_active', filter.is_active.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())
    if (filter?.offset) params.append('offset', filter.offset.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    return apiClient.get(`/api/integration/students${query}`)
  },

  // Get student by ID
  getById: async (id: number): Promise<ApiResponse<ExternalStudent>> => {
    return apiClient.get(`/api/integration/students/${id}`)
  },

  // Get student by external ID
  getByExternalId: async (externalId: string): Promise<ApiResponse<ExternalStudent>> => {
    return apiClient.get(`/api/integration/students/external/${externalId}`)
  },

  // Get students stats
  getStats: async (): Promise<
    ApiResponse<{
      total: number
      active: number
      inactive: number
      by_faculty: Record<string, number>
      by_course: Record<number, number>
    }>
  > => {
    return apiClient.get('/api/integration/students/stats')
  },
}

// Conflicts API
export const conflictsApi = {
  // List conflicts
  list: async (
    filter?: ConflictFilter
  ): Promise<ApiResponse<PaginatedList<SyncConflict> & { conflicts: SyncConflict[] }>> => {
    const params = new URLSearchParams()
    if (filter?.sync_log_id) params.append('sync_log_id', filter.sync_log_id.toString())
    if (filter?.entity_type) params.append('entity_type', filter.entity_type)
    if (filter?.resolution) params.append('resolution', filter.resolution)
    if (filter?.limit) params.append('limit', filter.limit.toString())
    if (filter?.offset) params.append('offset', filter.offset.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    return apiClient.get(`/api/integration/conflicts${query}`)
  },

  // Get pending conflicts
  getPending: async (
    limit = 20,
    offset = 0
  ): Promise<ApiResponse<PaginatedList<SyncConflict> & { conflicts: SyncConflict[] }>> => {
    return apiClient.get(`/api/integration/conflicts/pending?limit=${limit}&offset=${offset}`)
  },

  // Get conflict by ID
  getById: async (id: number): Promise<ApiResponse<SyncConflict>> => {
    return apiClient.get(`/api/integration/conflicts/${id}`)
  },

  // Resolve conflict
  resolve: async (
    id: number,
    request: ResolveConflictRequest
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post(`/api/integration/conflicts/${id}/resolve`, request)
  },

  // Bulk resolve conflicts
  bulkResolve: async (
    request: BulkResolveRequest
  ): Promise<ApiResponse<{ message: string; count: number }>> => {
    return apiClient.post('/api/integration/conflicts/bulk-resolve', request)
  },

  // Delete conflict
  delete: async (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.delete(`/api/integration/conflicts/${id}`)
  },

  // Get conflict stats
  getStats: async (): Promise<ApiResponse<ConflictStats>> => {
    return apiClient.get('/api/integration/conflicts/stats')
  },
}

// Combined integration API
export const integrationApi = {
  sync: syncApi,
  employees: employeesApi,
  students: studentsApi,
  conflicts: conflictsApi,
}
