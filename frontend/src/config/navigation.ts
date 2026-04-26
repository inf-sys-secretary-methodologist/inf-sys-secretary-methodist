import {
  LucideIcon,
  LayoutDashboard,
  Users,
  FileText,
  BarChart3,
  Calendar,
  MessageCircle,
  Database,
  Shield,
  TrendingUp,
  FileCheck,
  Settings,
  Sparkles,
  ListTodo,
  Megaphone,
  FolderOpen,
  GraduationCap,
} from 'lucide-react'
import { UserRole } from '@/types/auth'

export interface NavItem {
  /** Translation key for the nav item name (e.g., 'dashboard' -> t('nav.dashboard')) */
  nameKey: string
  url: string
  icon: LucideIcon
  roles?: UserRole[] // If undefined, available to all authenticated users
}

export interface NavGroup {
  /** Translation key for the group name */
  nameKey: string
  icon: LucideIcon
  items: NavItem[]
  roles?: UserRole[] // If undefined, available to all authenticated users
}

export type NavEntry = NavItem | NavGroup

export function isNavGroup(entry: NavEntry): entry is NavGroup {
  return 'items' in entry
}

// Define which roles can access which pages
// nameKey corresponds to keys in messages/*.json under "nav" namespace
export const navigationConfig: NavEntry[] = [
  // Dashboard - standalone
  {
    nameKey: 'dashboard',
    url: '/dashboard',
    icon: LayoutDashboard,
  },
  // Documents group
  {
    nameKey: 'documentsGroup',
    icon: FileText,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
    items: [
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
        nameKey: 'files',
        url: '/files',
        icon: FolderOpen,
        roles: [
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
          UserRole.TEACHER,
          UserRole.STUDENT,
        ],
      },
      {
        nameKey: 'templates',
        url: '/documents/templates',
        icon: FileCheck,
        roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY],
      },
    ],
  },
  // Analytics group — teacher gets limited reports access
  {
    nameKey: 'analyticsGroup',
    icon: BarChart3,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
    items: [
      {
        nameKey: 'reports',
        url: '/reports',
        icon: BarChart3,
        roles: [
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
          UserRole.TEACHER,
        ],
      },
      {
        nameKey: 'analytics',
        url: '/analytics',
        icon: TrendingUp,
        roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY],
      },
    ],
  },
  // Schedule - class timetable (all roles per 0.102.2 matrix)
  {
    nameKey: 'schedule',
    url: '/schedule',
    icon: GraduationCap,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },
  // Calendar - standalone
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
  // Tasks/assignments - all roles; student & teacher: own scope
  {
    nameKey: 'tasks',
    url: '/tasks',
    icon: ListTodo,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },
  // Announcements - standalone (all authenticated users; STUDENT view-only)
  {
    nameKey: 'announcements',
    url: '/announcements',
    icon: Megaphone,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },
  // Messages - standalone
  {
    nameKey: 'messages',
    url: '/messages',
    icon: MessageCircle,
  },
  // AI Assistant - standalone
  {
    nameKey: 'aiAssistant',
    url: '/ai',
    icon: Sparkles,
  },
  // Admin group
  {
    nameKey: 'adminGroup',
    icon: Settings,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
    items: [
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
        nameKey: 'integration',
        url: '/integration',
        icon: Database,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        nameKey: 'adminSettings',
        url: '/admin/settings/appearance',
        icon: Shield,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        nameKey: 'settingsPage',
        url: '/settings/appearance',
        icon: Settings,
      },
    ],
  },
]

/**
 * Filter navigation entries based on user role
 */
export function getAvailableNavEntries(userRole?: UserRole | string): NavEntry[] {
  if (!userRole) return []

  return navigationConfig
    .filter((entry) => {
      // Check if entry is available for user's role
      if (entry.roles && entry.roles.length > 0) {
        if (!entry.roles.includes(userRole as UserRole)) {
          return false
        }
      }
      return true
    })
    .map((entry) => {
      // If it's a group, filter its items too
      if (isNavGroup(entry)) {
        const filteredItems = entry.items.filter((item) => {
          if (item.roles && item.roles.length > 0) {
            return item.roles.includes(userRole as UserRole)
          }
          /* c8 ignore next 2 - defensive: all config items have roles, but handle undefined for future-proofing */
          return true
        })
        // Only return group if it has available items
        if (filteredItems.length === 0) return null
        // If only one item, return as direct link instead of group
        if (filteredItems.length === 1) {
          return filteredItems[0]
        }
        return { ...entry, items: filteredItems }
      }
      return entry
    })
    .filter((entry): entry is NavEntry => entry !== null)
}

// Legacy function for backwards compatibility
export function getAvailableNavItems(userRole?: UserRole | string): NavItem[] {
  const entries = getAvailableNavEntries(userRole)
  const items: NavItem[] = []

  for (const entry of entries) {
    if (isNavGroup(entry)) {
      items.push(...entry.items)
    } else {
      items.push(entry)
    }
  }

  return items
}
