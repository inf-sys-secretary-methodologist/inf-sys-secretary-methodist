import { apiClient } from '../api'

// Types
export interface UserWithOrg {
  id: number
  email: string
  name: string
  role: string
  status: string
  phone?: string
  avatar?: string
  department_id?: number
  department_name?: string
  position_id?: number
  position_name?: string
  created_at: string
  updated_at: string
}

export interface Department {
  id: number
  name: string
  code: string
  description?: string
  parent_id?: number
  head_id?: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Position {
  id: number
  name: string
  code: string
  description?: string
  level: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UserListFilter {
  department_id?: number
  position_id?: number
  role?: string
  status?: string
  search?: string
  page?: number
  limit?: number
}

export interface PaginatedResponse<T> {
  data: {
    users?: T[]
    departments?: T[]
    positions?: T[]
    total: number
    page: number
    limit: number
    total_pages: number
  }
}

export interface ApiResponse<T> {
  data: T
}

// Users API
export const usersApi = {
  // List users with filters
  list: async (filter?: UserListFilter): Promise<PaginatedResponse<UserWithOrg>> => {
    const params = new URLSearchParams()
    if (filter?.department_id) params.append('department_id', filter.department_id.toString())
    if (filter?.position_id) params.append('position_id', filter.position_id.toString())
    if (filter?.role) params.append('role', filter.role)
    if (filter?.status) params.append('status', filter.status)
    if (filter?.search) params.append('search', filter.search)
    if (filter?.page) params.append('page', filter.page.toString())
    if (filter?.limit) params.append('limit', filter.limit.toString())

    const query = params.toString() ? `?${params.toString()}` : ''
    return apiClient.get(`/api/users${query}`)
  },

  // Get single user
  getById: async (id: number): Promise<ApiResponse<UserWithOrg>> => {
    return apiClient.get(`/api/users/${id}`)
  },

  // Update user profile
  updateProfile: async (
    id: number,
    data: {
      department_id?: number | null
      position_id?: number | null
      phone?: string
      avatar?: string
      bio?: string
    }
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.put(`/api/users/${id}/profile`, data)
  },

  // Update user role
  updateRole: async (id: number, role: string): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.put(`/api/users/${id}/role`, { role })
  },

  // Update user status
  updateStatus: async (id: number, status: string): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.put(`/api/users/${id}/status`, { status })
  },

  // Delete user
  delete: async (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.delete(`/api/users/${id}`)
  },

  // Bulk update department
  bulkUpdateDepartment: async (
    userIds: number[],
    departmentId: number | null
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post('/api/users/bulk/department', {
      user_ids: userIds,
      department_id: departmentId,
    })
  },

  // Bulk update position
  bulkUpdatePosition: async (
    userIds: number[],
    positionId: number | null
  ): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.post('/api/users/bulk/position', {
      user_ids: userIds,
      position_id: positionId,
    })
  },

  // Get users by department
  getByDepartment: async (departmentId: number): Promise<ApiResponse<{ users: UserWithOrg[] }>> => {
    return apiClient.get(`/api/users/by-department/${departmentId}`)
  },

  // Get users by position
  getByPosition: async (positionId: number): Promise<ApiResponse<{ users: UserWithOrg[] }>> => {
    return apiClient.get(`/api/users/by-position/${positionId}`)
  },
}

// Departments API
export const departmentsApi = {
  // List departments
  list: async (
    page = 1,
    limit = 10,
    activeOnly = false
  ): Promise<PaginatedResponse<Department>> => {
    return apiClient.get(`/api/departments?page=${page}&limit=${limit}&active_only=${activeOnly}`)
  },

  // Get single department
  getById: async (id: number): Promise<ApiResponse<Department>> => {
    return apiClient.get(`/api/departments/${id}`)
  },

  // Create department
  create: async (data: {
    name: string
    code: string
    description?: string
    parent_id?: number
  }): Promise<ApiResponse<Department>> => {
    return apiClient.post('/api/departments', data)
  },

  // Update department
  update: async (
    id: number,
    data: {
      name: string
      code: string
      description?: string
      parent_id?: number | null
      head_id?: number | null
      is_active?: boolean
    }
  ): Promise<ApiResponse<Department>> => {
    return apiClient.put(`/api/departments/${id}`, data)
  },

  // Delete department
  delete: async (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.delete(`/api/departments/${id}`)
  },

  // Get child departments
  getChildren: async (parentId: number): Promise<ApiResponse<{ departments: Department[] }>> => {
    return apiClient.get(`/api/departments/${parentId}/children`)
  },
}

// Positions API
export const positionsApi = {
  // List positions
  list: async (page = 1, limit = 10, activeOnly = false): Promise<PaginatedResponse<Position>> => {
    return apiClient.get(`/api/positions?page=${page}&limit=${limit}&active_only=${activeOnly}`)
  },

  // Get single position
  getById: async (id: number): Promise<ApiResponse<Position>> => {
    return apiClient.get(`/api/positions/${id}`)
  },

  // Create position
  create: async (data: {
    name: string
    code: string
    description?: string
    level?: number
  }): Promise<ApiResponse<Position>> => {
    return apiClient.post('/api/positions', data)
  },

  // Update position
  update: async (
    id: number,
    data: {
      name: string
      code: string
      description?: string
      level?: number
      is_active?: boolean
    }
  ): Promise<ApiResponse<Position>> => {
    return apiClient.put(`/api/positions/${id}`, data)
  },

  // Delete position
  delete: async (id: number): Promise<ApiResponse<{ message: string }>> => {
    return apiClient.delete(`/api/positions/${id}`)
  },
}
