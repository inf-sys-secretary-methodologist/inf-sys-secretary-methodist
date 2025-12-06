import { UserRole } from '@/types/auth'

/**
 * Roles that have edit permissions (can create, update, delete)
 * STUDENT is explicitly excluded - view only
 */
export const EDIT_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
  UserRole.TEACHER,
]

/**
 * Roles that can only view content (no create, update, delete)
 */
export const VIEW_ONLY_ROLES: UserRole[] = [UserRole.STUDENT]

/**
 * Check if user role has edit permissions
 */
export function canEdit(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return EDIT_ROLES.includes(userRole as UserRole)
}

/**
 * Check if user role can create new items
 */
export function canCreate(userRole?: UserRole | string): boolean {
  return canEdit(userRole)
}

/**
 * Check if user role can delete items
 */
export function canDelete(userRole?: UserRole | string): boolean {
  return canEdit(userRole)
}

/**
 * Check if user role is view-only
 */
export function isViewOnly(userRole?: UserRole | string): boolean {
  if (!userRole) return true
  return VIEW_ONLY_ROLES.includes(userRole as UserRole)
}

/**
 * Check if user is admin
 */
export function isAdmin(userRole?: UserRole | string): boolean {
  return userRole === UserRole.SYSTEM_ADMIN
}
