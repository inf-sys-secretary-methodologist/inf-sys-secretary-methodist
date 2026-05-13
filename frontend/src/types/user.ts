// User mirrors the entities.UserWithOrg JSON projection returned by
// GET /api/users. The optional department_id / position_id pair lets
// the list page show "—" for unassigned users without a numeric 0
// fallback.
export interface User {
  id: number
  email: string
  name: string
  role: UserRole
  status: UserStatus
  phone?: string
  avatar?: string
  bio?: string
  department_id?: number | null
  department_name?: string
  position_id?: number | null
  position_name?: string
  created_at: string
  updated_at: string
}

export type UserRole = 'system_admin' | 'methodist' | 'academic_secretary' | 'teacher' | 'student'

export type UserStatus = 'active' | 'inactive' | 'blocked'

export interface UserListFilter {
  page?: number
  limit?: number
  search?: string
  role?: UserRole | ''
  status?: UserStatus | ''
  department_id?: number
  position_id?: number
}

export interface UserListResponse {
  users: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}
