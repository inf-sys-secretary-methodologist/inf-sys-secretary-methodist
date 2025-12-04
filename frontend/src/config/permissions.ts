/**
 * Permissions Configuration
 * Synchronized with backend: internal/modules/auth/domain/role.go & permission.go
 *
 * This is the single source of truth for all permissions on frontend.
 * When adding new resources/actions, update both this file and backend.
 */

import { UserRole } from '@/types/auth'

// ============================================================================
// Resource Types (synced with backend ResourceType)
// ============================================================================
export enum Resource {
  USERS = 'users',
  CURRICULUM = 'curriculum',
  SCHEDULE = 'schedule',
  ASSIGNMENTS = 'assignments',
  REPORTS = 'reports',
  DOCUMENTS = 'documents',
  STUDENTS = 'students',
  EVENTS = 'events',
  TASKS = 'tasks',
  ANNOUNCEMENTS = 'announcements',
}

// ============================================================================
// Action Types (synced with backend ActionType)
// ============================================================================
export enum Action {
  CREATE = 'create',
  READ = 'read',
  UPDATE = 'update',
  DELETE = 'delete',
  DEACTIVATE = 'deactivate',
  APPROVE = 'approve',
  EXECUTE = 'execute',
  EXPORT = 'export',
  UPLOAD = 'upload',
}

// ============================================================================
// Access Levels (synced with backend AccessLevel)
// ============================================================================
export enum AccessLevel {
  DENIED = 0,
  LIMITED = 1,
  OWN = 2,
  FULL = 3,
}

// ============================================================================
// Permission Matrix Type
// ============================================================================
type PermissionMatrix = {
  [role in UserRole]: {
    [resource in Resource]?: {
      [action in Action]?: AccessLevel
    }
  }
}

// ============================================================================
// Permission Matrix (synced with backend PermissionMatrix in permission.go)
// ============================================================================
//
// БИЗНЕС-ЛОГИКА РОЛЕЙ:
//
// SYSTEM_ADMIN - Системный администратор
//   Полное управление системой, пользователями, всеми ресурсами
//
// METHODIST - Методист
//   Методическое обеспечение: документы, учебные планы, мероприятия
//   Может создавать объявления, работать с задачами
//
// ACADEMIC_SECRETARY - Академический секретарь
//   Административное сопровождение: студенты, расписание, отчёты
//   Полное управление студентами и расписанием
//
// TEACHER - Преподаватель
//   Реализация учебного процесса: просмотр студентов, задания
//   Загрузка своих документов, работа со своими задачами
//
// STUDENT - Студент
//   Участие в учебном процессе: просмотр расписания, задач, событий
//   Выполнение заданий, просмотр объявлений
//
// ============================================================================
export const permissionMatrix: PermissionMatrix = {
  // =========================================================================
  // SYSTEM_ADMIN - Полный доступ ко всему
  // =========================================================================
  [UserRole.SYSTEM_ADMIN]: {
    [Resource.USERS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
      [Action.DEACTIVATE]: AccessLevel.FULL,
    },
    [Resource.CURRICULUM]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.APPROVE]: AccessLevel.FULL,
    },
    [Resource.DOCUMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
      [Action.UPLOAD]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
    },
    [Resource.STUDENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.EVENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.TASKS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.ASSIGNMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.REPORTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
      [Action.APPROVE]: AccessLevel.FULL,
    },
    [Resource.SCHEDULE]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
    },
    [Resource.ANNOUNCEMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
  },

  // =========================================================================
  // METHODIST - Методическое обеспечение учебного процесса
  // =========================================================================
  [UserRole.METHODIST]: {
    [Resource.USERS]: {
      [Action.READ]: AccessLevel.LIMITED, // Только базовая информация
    },
    [Resource.CURRICULUM]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      // Нет APPROVE - утверждает только админ
    },
    [Resource.DOCUMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.UPLOAD]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
    },
    [Resource.STUDENTS]: {
      [Action.READ]: AccessLevel.FULL, // Просмотр всех студентов
      [Action.UPDATE]: AccessLevel.LIMITED, // Ограниченное редактирование
    },
    [Resource.EVENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
    },
    [Resource.TASKS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
    },
    [Resource.ASSIGNMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.LIMITED,
    },
    [Resource.REPORTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
    },
    [Resource.SCHEDULE]: {
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.LIMITED, // Может предлагать изменения
    },
    [Resource.ANNOUNCEMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.OWN, // Только свои объявления
      [Action.DELETE]: AccessLevel.OWN,
    },
  },

  // =========================================================================
  // ACADEMIC_SECRETARY - Административное сопровождение
  // =========================================================================
  [UserRole.ACADEMIC_SECRETARY]: {
    [Resource.USERS]: {
      [Action.READ]: AccessLevel.LIMITED,
    },
    [Resource.CURRICULUM]: {
      [Action.READ]: AccessLevel.FULL,
      // Не создаёт и не редактирует учебные планы
    },
    [Resource.DOCUMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.UPLOAD]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
    },
    [Resource.STUDENTS]: {
      [Action.CREATE]: AccessLevel.FULL, // Полное управление студентами
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.EVENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
      [Action.DELETE]: AccessLevel.FULL,
    },
    [Resource.TASKS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
    },
    [Resource.ASSIGNMENTS]: {
      [Action.READ]: AccessLevel.FULL,
      // Не создаёт задания
    },
    [Resource.REPORTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.EXPORT]: AccessLevel.FULL,
    },
    [Resource.SCHEDULE]: {
      [Action.CREATE]: AccessLevel.FULL, // Полное управление расписанием
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.FULL,
    },
    [Resource.ANNOUNCEMENTS]: {
      [Action.CREATE]: AccessLevel.FULL,
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.OWN,
    },
  },

  // =========================================================================
  // TEACHER - Преподаватель
  // =========================================================================
  [UserRole.TEACHER]: {
    [Resource.USERS]: {
      [Action.READ]: AccessLevel.LIMITED, // Базовая информация коллег
    },
    [Resource.CURRICULUM]: {
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.LIMITED, // Может предлагать изменения
    },
    [Resource.DOCUMENTS]: {
      [Action.READ]: AccessLevel.FULL,
      [Action.UPLOAD]: AccessLevel.OWN, // Только свои документы
      [Action.EXPORT]: AccessLevel.LIMITED,
    },
    [Resource.STUDENTS]: {
      [Action.READ]: AccessLevel.FULL, // Просмотр студентов для выставления оценок
    },
    [Resource.EVENTS]: {
      [Action.READ]: AccessLevel.FULL,
      [Action.CREATE]: AccessLevel.OWN, // Может создавать свои мероприятия
    },
    [Resource.TASKS]: {
      [Action.READ]: AccessLevel.FULL, // Просмотр всех задач
      [Action.UPDATE]: AccessLevel.OWN, // Редактирование своих
      [Action.CREATE]: AccessLevel.OWN, // Создание задач для студентов
    },
    [Resource.ASSIGNMENTS]: {
      [Action.CREATE]: AccessLevel.FULL, // Создание заданий
      [Action.READ]: AccessLevel.FULL,
      [Action.UPDATE]: AccessLevel.OWN, // Редактирование своих
    },
    [Resource.REPORTS]: {
      [Action.CREATE]: AccessLevel.FULL, // Отчёты по своим предметам
      [Action.READ]: AccessLevel.LIMITED, // Только свои отчёты
      [Action.EXPORT]: AccessLevel.LIMITED,
    },
    [Resource.SCHEDULE]: {
      [Action.READ]: AccessLevel.FULL,
    },
    [Resource.ANNOUNCEMENTS]: {
      [Action.READ]: AccessLevel.FULL,
      [Action.CREATE]: AccessLevel.OWN, // Может создавать для своих групп
    },
  },

  // =========================================================================
  // STUDENT - Студент (ВАЖНО: видит задачи в режиме просмотра!)
  // =========================================================================
  [UserRole.STUDENT]: {
    [Resource.USERS]: {
      [Action.READ]: AccessLevel.OWN, // Только свой профиль
      [Action.UPDATE]: AccessLevel.OWN, // Редактирование профиля
    },
    [Resource.CURRICULUM]: {
      [Action.READ]: AccessLevel.LIMITED, // Свой учебный план
    },
    [Resource.DOCUMENTS]: {
      [Action.READ]: AccessLevel.LIMITED, // Доступные документы
    },
    [Resource.STUDENTS]: {
      [Action.READ]: AccessLevel.OWN, // Только свои данные
      [Action.UPDATE]: AccessLevel.OWN, // Редактирование профиля
    },
    [Resource.EVENTS]: {
      [Action.READ]: AccessLevel.FULL, // Все мероприятия
    },
    [Resource.TASKS]: {
      [Action.READ]: AccessLevel.FULL, // ВАЖНО: Видит все задачи (только просмотр!)
      [Action.EXECUTE]: AccessLevel.FULL, // Может выполнять задачи
    },
    [Resource.ASSIGNMENTS]: {
      [Action.READ]: AccessLevel.OWN, // Свои задания
      [Action.EXECUTE]: AccessLevel.FULL, // Выполнение заданий
    },
    [Resource.REPORTS]: {
      [Action.READ]: AccessLevel.OWN, // Свои отчёты/оценки
    },
    [Resource.SCHEDULE]: {
      [Action.READ]: AccessLevel.FULL, // Всё расписание
    },
    [Resource.ANNOUNCEMENTS]: {
      [Action.READ]: AccessLevel.FULL, // Все объявления
    },
  },
}

// ============================================================================
// Permission Check Functions
// ============================================================================

/**
 * Check if user has permission to perform action on resource
 */
export function hasPermission(
  userRole: UserRole | string | undefined,
  resource: Resource,
  action: Action
): boolean {
  if (!userRole) return false

  const role = userRole as UserRole
  const rolePermissions = permissionMatrix[role]
  if (!rolePermissions) return false

  const resourcePermissions = rolePermissions[resource]
  if (!resourcePermissions) return false

  const accessLevel = resourcePermissions[action]
  return accessLevel !== undefined && accessLevel > AccessLevel.DENIED
}

/**
 * Get access level for user on resource/action
 */
export function getAccessLevel(
  userRole: UserRole | string | undefined,
  resource: Resource,
  action: Action
): AccessLevel {
  if (!userRole) return AccessLevel.DENIED

  const role = userRole as UserRole
  const rolePermissions = permissionMatrix[role]
  if (!rolePermissions) return AccessLevel.DENIED

  const resourcePermissions = rolePermissions[resource]
  if (!resourcePermissions) return AccessLevel.DENIED

  return resourcePermissions[action] ?? AccessLevel.DENIED
}

/**
 * Check if user has full access to resource/action
 */
export function hasFullAccess(
  userRole: UserRole | string | undefined,
  resource: Resource,
  action: Action
): boolean {
  return getAccessLevel(userRole, resource, action) === AccessLevel.FULL
}

/**
 * Check if user can only access own resources
 */
export function hasOwnAccess(
  userRole: UserRole | string | undefined,
  resource: Resource,
  action: Action
): boolean {
  return getAccessLevel(userRole, resource, action) === AccessLevel.OWN
}

// ============================================================================
// Quick Actions Configuration
// ============================================================================
export interface QuickAction {
  id: string
  label: string
  description?: string
  icon: string // Icon name from lucide-react
  path: string
  resource: Resource
  action: Action
}

export const quickActionsConfig: QuickAction[] = [
  {
    id: 'upload-document',
    label: 'Загрузить документ',
    description: 'Добавить новый документ в систему',
    icon: 'Upload',
    path: '/documents?action=upload',
    resource: Resource.DOCUMENTS,
    action: Action.UPLOAD,
  },
  {
    id: 'add-student',
    label: 'Добавить студента',
    description: 'Зарегистрировать нового студента',
    icon: 'UserPlus',
    path: '/students?action=create',
    resource: Resource.STUDENTS,
    action: Action.CREATE,
  },
  {
    id: 'create-event',
    label: 'Создать мероприятие',
    description: 'Запланировать новое мероприятие',
    icon: 'CalendarPlus',
    path: '/calendar?action=create',
    resource: Resource.EVENTS,
    action: Action.CREATE,
  },
  {
    id: 'create-task',
    label: 'Создать задачу',
    description: 'Добавить новую задачу',
    icon: 'ListTodo',
    path: '/tasks?action=create',
    resource: Resource.TASKS,
    action: Action.CREATE,
  },
]

/**
 * Get available quick actions for user based on their role
 */
export function getAvailableQuickActions(userRole: UserRole | string | undefined): QuickAction[] {
  if (!userRole) return []

  return quickActionsConfig.filter((action) =>
    hasPermission(userRole, action.resource, action.action)
  )
}

// ============================================================================
// Role Display Names
// ============================================================================
export const roleDisplayNames: Record<UserRole, string> = {
  [UserRole.SYSTEM_ADMIN]: 'Системный администратор',
  [UserRole.METHODIST]: 'Методист',
  [UserRole.ACADEMIC_SECRETARY]: 'Академический секретарь',
  [UserRole.TEACHER]: 'Преподаватель',
  [UserRole.STUDENT]: 'Студент',
}

export function getRoleDisplayName(role: UserRole | string | undefined): string {
  if (!role) return 'Неизвестная роль'
  return roleDisplayNames[role as UserRole] || role
}

// ============================================================================
// Resource Display Names
// ============================================================================
export const resourceDisplayNames: Record<Resource, string> = {
  [Resource.USERS]: 'Пользователи',
  [Resource.CURRICULUM]: 'Учебный план',
  [Resource.SCHEDULE]: 'Расписание',
  [Resource.ASSIGNMENTS]: 'Задания',
  [Resource.REPORTS]: 'Отчеты',
  [Resource.DOCUMENTS]: 'Документы',
  [Resource.STUDENTS]: 'Студенты',
  [Resource.EVENTS]: 'Мероприятия',
  [Resource.TASKS]: 'Задачи',
  [Resource.ANNOUNCEMENTS]: 'Объявления',
}

// ============================================================================
// Action Display Names
// ============================================================================
export const actionDisplayNames: Record<Action, string> = {
  [Action.CREATE]: 'Создание',
  [Action.READ]: 'Просмотр',
  [Action.UPDATE]: 'Редактирование',
  [Action.DELETE]: 'Удаление',
  [Action.DEACTIVATE]: 'Деактивация',
  [Action.APPROVE]: 'Утверждение',
  [Action.EXECUTE]: 'Выполнение',
  [Action.EXPORT]: 'Экспорт',
  [Action.UPLOAD]: 'Загрузка',
}
