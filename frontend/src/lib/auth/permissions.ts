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
    // v0.158.0: methodist is the curriculum APPROVER, not author. Read +
    // approve/reject via Action.APPROVE branch in can(); cannot create
    // or edit drafts (those belong to academic_secretary).
    [Resource.CURRICULUM]: AccessLevel.LIMITED,
    [Resource.SCHEDULE]: AccessLevel.OWN,
    [Resource.ASSIGNMENTS]: AccessLevel.FULL,
    [Resource.REPORTS]: AccessLevel.FULL,
    [Resource.INTEGRATION]: AccessLevel.DENIED,
    [Resource.SYSTEM_SETTINGS]: AccessLevel.DENIED,
    [Resource.PERSONAL_SETTINGS]: AccessLevel.OWN,
  },
  [UserRole.ACADEMIC_SECRETARY]: {
    [Resource.USERS]: AccessLevel.LIMITED,
    // v0.158.0: academic_secretary is the curriculum AUTHOR. Owns the
    // full authoring lifecycle (create / edit drafts / submit) plus
    // sections + discipline items. Approval belongs to methodist.
    [Resource.CURRICULUM]: AccessLevel.FULL,
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

const ACTION_MIN_LEVEL: Record<Exclude<Action, Action.APPROVE>, AccessLevel> = {
  [Action.READ]: AccessLevel.LIMITED,
  [Action.CREATE]: AccessLevel.FULL,
  [Action.UPDATE]: AccessLevel.OWN,
  [Action.DELETE]: AccessLevel.FULL,
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
    // v0.158.0: curriculum approval/rejection — methodist + system_admin.
    // Academic_secretary authors; methodist approves; admin retains
    // emergency override. Mirror backend canApprove whitelist.
    if (resource !== Resource.CURRICULUM) return false
    return role === UserRole.METHODIST || role === UserRole.SYSTEM_ADMIN
  }
  const level = getAccessLevel(role, resource)
  return level >= ACTION_MIN_LEVEL[action as Exclude<Action, Action.APPROVE>]
}

// --- Legacy functions (backward compat) ---

export const EDIT_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
  UserRole.TEACHER,
]

export const VIEW_ONLY_ROLES: UserRole[] = [UserRole.STUDENT]

// canSignDocuments — non-student staff may apply a cryptographic signature to
// a document. Mirrors the backend SignDocumentUseCase, which denies students
// (#140). Students are viewers of the signatures panel only.
export function canSignDocuments(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return EDIT_ROLES.includes(userRole as UserRole)
}

// CURRICULUM_WRITE_ROLES — roles permitted to create/update/submit a
// curriculum через UI. Mirrors the backend write-whitelist enforced by
// the POST /api/curriculum + PUT /api/curriculum/:id handlers.
// v0.158.0: academic_secretary owns the authoring lifecycle; methodist
// is the approver (separate canApprove path); admin retains override.
export const CURRICULUM_WRITE_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.ACADEMIC_SECRETARY,
]

export function canWriteCurriculum(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return CURRICULUM_WRITE_ROLES.includes(userRole as UserRole)
}

// WORK_PROGRAM_CREATE_ROLES — roles permitted to create a РПД (рабочая
// программа дисциплины) draft via UI. Mirrors the backend create
// authorization (work-program ADR-5): teacher is the primary author of
// their own discipline, methodist is the reserve author, admin retains
// override. academic_secretary owns curriculum (not РПД) and student is
// view-only, so both are excluded.
export const WORK_PROGRAM_CREATE_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.TEACHER,
]

// WORK_PROGRAM_APPROVE_ROLES — roles permitted to approve/reject a
// pending РПД. The teacher authors but cannot approve their own work;
// methodist is the approver (методотдел/проректор combined per ADR-5),
// admin retains emergency override.
export const WORK_PROGRAM_APPROVE_ROLES: UserRole[] = [UserRole.SYSTEM_ADMIN, UserRole.METHODIST]

export function canCreateWorkProgram(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return WORK_PROGRAM_CREATE_ROLES.includes(userRole as UserRole)
}

export function canApproveWorkProgram(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return WORK_PROGRAM_APPROVE_ROLES.includes(userRole as UserRole)
}

// MINOBRNAUKI_ORDER_VIEW_ROLES — roles permitted to browse Минобрнауки
// orders (приказы). Mirrors the backend isAllowedToViewMinobrnaukiOrders
// read gate (ADR-11): every non-student staff role may view orders
// (teachers need to see orders affecting their disciplines). Students have
// no business reason to read internal regulatory-tracking artifacts.
export const MINOBRNAUKI_ORDER_VIEW_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
  UserRole.TEACHER,
]

export function canViewMinobrnaukiOrders(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return MINOBRNAUKI_ORDER_VIEW_ROLES.includes(userRole as UserRole)
}

// MINOBRNAUKI_ORDER_RECORD_ROLES — roles permitted to record (create) a
// приказ + trigger the bulk-revision flow. Mirrors the backend
// isAllowedToRecordMinobrnaukiOrder write gate (ADR-11): methodist +
// academic_secretary + admin. Teacher may view but not record (they author
// the РПД revisions the order triggers, not the regulatory record itself).
export const MINOBRNAUKI_ORDER_RECORD_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
]

export function canRecordMinobrnaukiOrder(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return MINOBRNAUKI_ORDER_RECORD_ROLES.includes(userRole as UserRole)
}

// STUDENT_DEBTS_MANAGE_ROLES — roles permitted to manage the student debt
// registry (import/export, schedule resit, record resit result). Mirrors
// the backend isDebtManager predicate (student_debts design §5,
// application/usecases/role_predicates.go): the staff set
// {admin, methodist, academic_secretary} holds every write/import right
// (EDIT_ROLES for debts). teacher has read-only access scoped to owned
// disciplines; student reads only their own debts — both are excluded.
export const STUDENT_DEBTS_MANAGE_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
]

// STUDENT_DEBTS_REGISTRY_ROLES — roles permitted to browse the full debt
// registry (/student-debts). Mirrors the backend read scope (read_scope.go
// + get_debt_usecase.go): staff read all; teacher reads the registry
// server-scoped to owned disciplines. Students are denied the registry
// endpoint and read only their own debts via /student-debts/my, so the
// list page redirects them there.
export const STUDENT_DEBTS_REGISTRY_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
  UserRole.TEACHER,
]

export function canManageStudentDebts(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return STUDENT_DEBTS_MANAGE_ROLES.includes(userRole as UserRole)
}

export function canViewDebtRegistry(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return STUDENT_DEBTS_REGISTRY_ROLES.includes(userRole as UserRole)
}

// TEACHING_LOAD_MANAGE_ROLES — roles allowed to manage the planned teaching
// load (/teaching-load), mirroring the backend write gate in
// teaching_load_handler.go (requireLoadWrite). Academic planning is the
// methodist's job, so methodist is included alongside secretary and admin.
export const TEACHING_LOAD_MANAGE_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
]

export function canManageTeachingLoad(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return TEACHING_LOAD_MANAGE_ROLES.includes(userRole as UserRole)
}

// SCHEDULE_GENERATE_ROLES — roles allowed to run the automatic schedule
// generator (preview + apply), mirroring the backend write gate in
// generate_schedule_handler.go (requireGenerateWrite). Building the timetable
// is academic planning, so methodist + academic_secretary manage it and admin
// retains override. Teacher and student only read the resulting timetable.
export const SCHEDULE_GENERATE_ROLES: UserRole[] = [
  UserRole.SYSTEM_ADMIN,
  UserRole.METHODIST,
  UserRole.ACADEMIC_SECRETARY,
]

export function canGenerateSchedule(userRole?: UserRole | string): boolean {
  if (!userRole) return false
  return SCHEDULE_GENERATE_ROLES.includes(userRole as UserRole)
}

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
