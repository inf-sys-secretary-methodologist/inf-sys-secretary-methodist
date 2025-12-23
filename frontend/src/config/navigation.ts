import {
  LucideIcon,
  LayoutDashboard,
  Users,
  FileText,
  BarChart3,
  Calendar,
  MessageCircle,
  Database,
} from 'lucide-react'
import { UserRole } from '@/types/auth'

export interface NavItem {
  /** Translation key for the nav item name (e.g., 'dashboard' -> t('nav.dashboard')) */
  nameKey: string
  url: string
  icon: LucideIcon
  roles?: UserRole[] // If undefined, available to all authenticated users
}

// Define which roles can access which pages
// nameKey corresponds to keys in messages/*.json under "nav" namespace
export const navigationConfig: NavItem[] = [
  {
    nameKey: 'dashboard',
    url: '/dashboard',
    icon: LayoutDashboard,
    // Available to all authenticated users
  },
  {
    nameKey: 'users',
    url: '/users',
    icon: Users,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
  },
  {
    nameKey: 'documents',
    url: '/documents',
    icon: FileText,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },
  {
    nameKey: 'reports',
    url: '/reports',
    icon: BarChart3,
    roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY],
  },
  {
    nameKey: 'calendar',
    url: '/calendar',
    icon: Calendar,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },
  {
    nameKey: 'messages',
    url: '/messages',
    icon: MessageCircle,
    // Available to all authenticated users
  },
  {
    nameKey: 'integration',
    url: '/integration',
    icon: Database,
    roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
  },
]

/**
 * Filter navigation items based on user role
 */
export function getAvailableNavItems(userRole?: UserRole | string): NavItem[] {
  if (!userRole) return []

  return navigationConfig.filter((item) => {
    // If no roles specified, item is available to all authenticated users
    if (!item.roles || item.roles.length === 0) {
      return true
    }

    // Check if user's role is in the allowed roles
    return item.roles.includes(userRole as UserRole)
  })
}
