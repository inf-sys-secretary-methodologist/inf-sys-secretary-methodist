import { UserRole } from '@/types/auth'

export enum Resource {
  USERS = 'users',
  CURRICULUM = 'curriculum',
  SCHEDULE = 'schedule',
  ASSIGNMENTS = 'assignments',
  REPORTS = 'reports',
  INTEGRATION = 'integration',
  SYSTEM_SETTINGS = 'system_settings',
  PERSONAL_SETTINGS = 'personal_settings',
}

export enum Action {
  READ = 'read',
  CREATE = 'create',
  UPDATE = 'update',
  DELETE = 'delete',
  APPROVE = 'approve',
}

export enum AccessLevel {
  DENIED = 0,
  LIMITED = 1,
  OWN = 2,
  FULL = 3,
}

type PermissionMatrix = Record<UserRole, Record<Resource, AccessLevel>>

const PERMISSION_MATRIX: PermissionMatrix = {
  [UserRole.SYSTEM_ADMIN]: {
    [Resource.USERS]: AccessLevel.FULL,
    [Resource.CURRICULUM]: AccessLevel.FULL,
    [Resource.SCHEDULE]: AccessLevel.FULL,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.FULL,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.FULL,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.METHODIST]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.FULL,
    [Resource.SCHEDULE]: AccessLevel.OWN,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.ACADEMIC_SECRETARY]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.FULL,
    [Resource.ASSIGNMENTS]: AccessLevel.LIMITED,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.TEACHER]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    [Resource.CURRICULUM]: AccessLevel.OWN,
    [Resource.SCHEDULE]: AccessLevel.LIMITED,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.LIMITED,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.STUDENT]: {
    [Resource.USERS]: AccessLevel.OWN,
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.LIMITED,
    [Resource.ASSIGNMENTS]: AccessLevel.OWN,
    [Resource.REPORTS]: AccessLevel.DENIED,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
}

const ACTION_MIN_LEVEL: Record<Action, AccessLevel> = {
  [Action.READ]: AccessLevel.LIMITED,
  [Action.CREATE]: AccessLevel.FULL,
  [Action.UPDATE]: AccessLevel.OWN,
  [Action.DELETE]: AccessLevel.FULL,
  [Action.APPROVE]: AccessLevel.FULL,
}

export function getAccessLevel(
  role: UserRole | string | undefined,
  resource: Resource
): AccessLevel {
  if (!role) return AccessLevel.DENIED
  const roleMatrix = PERMISSION_MATRIX[role as UserRole]
  if (!roleMatrix) return AccessLevel.DENIED
  return roleMatrix[resource] ?? AccessLevel.DENIED
}

export function can(
  role: UserRole | string | undefined,
  resource: Resource,
  action: Action
): boolean {
  if (!role) return false
  if (action === Action.APPROVE) {
    return role === UserRole.SYSTEM_ADMIN && resource === Resource.CURRICULUM
  }
  const level = getAccessLevel(role, resource)
  return level >= ACTION_MIN_LEVEL[action]
}

// --- Legacy functions (backward compat) ---

export const EDIT_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
  UserRole.TEACHER,
]

export const VIEW_ONLY_ROLES: UserRole[] = [UserRole.STUDENT]

/** @deprecated Use can(role, resource, action) instead */
export function canEdit(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return EDIT_ROLES.includes(userRole as UserRole)
}

/** @deprecated Use can(role, resource, action) instead */
export function canCreate(userRole?: UserRole | string): boolean {
  return canEdit(userRole)
}

/** @deprecated Use can(role, resource, action) instead */
export function canDelete(userRole?: UserRole | string): boolean {
  return canEdit(userRole)
}

/** @deprecated Use can(role, resource, action) instead */
export function isViewOnly(userRole?: UserRole | string): boolean {
  if (!userRole) return true
  return VIEW_ONLY_ROLES.includes(userRole as UserRole)
}

/** @deprecated Use can(role, resource, action) instead */
export function isAdmin(userRole?: UserRole | string): boolean {
  return userRole === UserRole.SYSTEM_ADMIN
}
