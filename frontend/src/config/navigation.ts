/**
 * Navigation Configuration
 * Uses permissions from config/permissions.ts for role-based filtering
 */

import {
  LucideIcon,
  LayoutDashboard,
  Users,
  FileText,
  Calendar,
  ListTodo,
  Megaphone,
} from 'lucide-react'
import { UserRole } from '@/types/auth'
import { Resource, Action, hasPermission } from './permissions'

export interface NavItem {
  name: string
  url: string
  icon: LucideIcon
  resource?: Resource // If defined, will check permission
  action?: Action // Action to check (defaults to READ)
  roles?: UserRole[] // Legacy: direct role check (deprecated, use resource/action instead)
}

// Navigation items with permission-based access
export const navigationConfig: NavItem[] = [
  {
    name: 'Главная',
    url: '/dashboard',
    icon: LayoutDashboard,
    // Available to all authenticated users - no resource check
  },
  {
    name: 'Студенты',
    url: '/students',
    icon: Users,
    resource: Resource.STUDENTS,
    action: Action.READ,
  },
  {
    name: 'Документы',
    url: '/documents',
    icon: FileText,
    resource: Resource.DOCUMENTS,
    action: Action.READ,
  },
  {
    name: 'Календарь',
    url: '/calendar',
    icon: Calendar,
    resource: Resource.EVENTS,
    action: Action.READ,
  },
  {
    name: 'Задачи',
    url: '/tasks',
    icon: ListTodo,
    resource: Resource.TASKS,
    action: Action.READ,
  },
  {
    name: 'Объявления',
    url: '/announcements',
    icon: Megaphone,
    resource: Resource.ANNOUNCEMENTS,
    action: Action.READ,
  },
]

/**
 * Filter navigation items based on user role and permissions
 */
export function getAvailableNavItems(userRole?: UserRole | string): NavItem[] {
  if (!userRole) return []

  return navigationConfig.filter((item) => {
    // If no resource specified, item is available to all authenticated users
    if (!item.resource) {
      return true
    }

    // Check permission using the new permission system
    const action = item.action || Action.READ
    return hasPermission(userRole, item.resource, action)
  })
}
