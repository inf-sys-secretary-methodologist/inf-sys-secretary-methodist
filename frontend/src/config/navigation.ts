import { LucideIcon, LayoutDashboard, Users, FileText, Calendar } from 'lucide-react'
import { UserRole } from '@/types/auth'

export interface NavItem {
  name: string
  url: string
  icon: LucideIcon
  roles?: UserRole[] // If undefined, available to all authenticated users
}

// Define which roles can access which pages
export const navigationConfig: NavItem[] = [
  {
    name: 'Главная',
    url: '/dashboard',
    icon: LayoutDashboard,
    // Available to all authenticated users
  },
  {
    name: 'Студенты',
    url: '/students',
    icon: Users,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
  },
  {
    name: 'Документы',
    url: '/documents',
    icon: FileText,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
  },
  {
    name: 'Календарь',
    url: '/calendar',
    icon: Calendar,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
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
