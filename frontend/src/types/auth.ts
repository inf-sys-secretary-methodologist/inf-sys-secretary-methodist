/**
 * Authentication and Authorization Types
 *
 * Defines user roles, permissions, and authentication state
 */

export enum UserRole {
  SYSTEM_ADMIN = 'system_admin',
  METHODIST = 'methodist',
  ACADEMIC_SECRETARY = 'academic_secretary',
  TEACHER = 'teacher',
  STUDENT = 'student'
}

export interface User {
  id: string
  email: string
  name: string
  role: UserRole
  avatar?: string
  createdAt?: Date
  updatedAt?: Date
}

export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterData {
  email: string
  password: string
  name: string
  role?: UserRole
}

export interface AuthResponse {
  user: User
  token: string
}

// Role labels for display
export const UserRoleLabels: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Учебный секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент'
}

// Permission helpers
export function hasRole(user: User | null, roles: UserRole[]): boolean {
  if (!user) return false
  return roles.includes(user.role)
}

export function isAdmin(user: User | null): boolean {
  return user?.role === UserRole.SYSTEM_ADMIN
}

export function canManageDocuments(user: User | null): boolean {
  return hasRole(user, [
    UserRole.SYSTEM_ADMIN,
    UserRole.METHODIST,
    UserRole.ACADEMIC_SECRETARY
  ])
}

export function canManageUsers(user: User | null): boolean {
  return user?.role === UserRole.SYSTEM_ADMIN
}
