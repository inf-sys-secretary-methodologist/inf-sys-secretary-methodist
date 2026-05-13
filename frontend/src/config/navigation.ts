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
  BookMarked,
  ClipboardCheck,
  HardDrive,
  Activity,
  UserCog,
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
      {
        // Annual methodist report (v0.129.0 B4). Read-only DOCX download
        // за календарный год. Methodist + system_admin only — backend
        // ADR-6 excludes academic_secretary (observer, not decision-maker).
        nameKey: 'annualReport',
        url: '/reports/annual',
        icon: FileCheck,
        roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
      },
    ],
  },
  // Education group — schedule + calendar + tasks
  {
    nameKey: 'educationGroup',
    icon: GraduationCap,
    roles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
    items: [
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
      {
        // Academic assignments — separate from project-management tasks.
        // Hidden from students because the v0.110.0 page is the
        // grading view (the read-side endpoint is gated by the
        // RequireNonStudent middleware on the backend).
        nameKey: 'assignments',
        url: '/assignments',
        icon: GraduationCap,
        roles: [
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
          UserRole.TEACHER,
        ],
      },
      {
        // Student-only mirror of /assignments. The student sees only
        // their own work — backend GET /api/assignments/my is gated by
        // RequireRole("student"), so other roles never reach the data.
        // Showing the entry alongside (rather than swapping) keeps the
        // navigation layout stable when a developer hops accounts.
        nameKey: 'myAssignments',
        url: '/my-assignments',
        icon: GraduationCap,
        roles: [UserRole.STUDENT],
      },
      {
        // Curriculum module list — methodist authors / admin approves
        // / secretary + teacher reads. Backend GET /api/curriculum is
        // gated by RequireNonStudent (v0.116.0), so the navigation
        // mirrors that role list to avoid a dead-link round-trip.
        nameKey: 'curriculum',
        url: '/curriculum',
        icon: BookMarked,
        roles: [
          UserRole.SYSTEM_ADMIN,
          UserRole.METHODIST,
          UserRole.ACADEMIC_SECRETARY,
          UserRole.TEACHER,
        ],
      },
    ],
  },
  // Communication group — announcements + messages + AI
  {
    nameKey: 'communicationGroup',
    icon: MessageCircle,
    items: [
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
      {
        nameKey: 'messages',
        url: '/messages',
        icon: MessageCircle,
      },
      {
        nameKey: 'aiAssistant',
        url: '/ai',
        icon: Sparkles,
      },
    ],
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
        // Curriculum admin queue — pending_approval list с Approve / Reject
        // dialogs. Backend endpoints (POST /api/curriculum/:id/approve и
        // /:id/reject) gated by RequireRole(SystemAdmin); navigation mirror
        // single-role allowlist чтобы non-admins не видели dead-link.
        nameKey: 'curriculumApprove',
        url: '/admin/curriculum/approve',
        icon: ClipboardCheck,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        nameKey: 'adminSettings',
        url: '/admin/settings/appearance',
        icon: Shield,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        // Admin audit-log timeline (v0.131.0). Backend GET /api/admin/
        // audit-logs gated by RequireRole(system_admin); navigation
        // mirror single-role allowlist so non-admins never see a
        // dead-link entry.
        nameKey: 'auditLogs',
        url: '/admin/audit-logs',
        icon: FileText,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        // Admin backup observability (v0.132.0). Read-only surface
        // over /backup sidecar's shared volumes.
        nameKey: 'backups',
        url: '/admin/backups',
        icon: HardDrive,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        // Admin Sentry config view (v0.133.0). Read-only mirror of
        // initSentry runtime configuration — DSN-as-boolean only.
        nameKey: 'sentry',
        url: '/admin/sentry',
        icon: Activity,
        roles: [UserRole.SYSTEM_ADMIN],
      },
      {
        // Admin user management (v0.133.0). List/filter/edit; write
        // endpoints gated by RequireRole(system_admin) since v0.133.0.
        nameKey: 'adminUsers',
        url: '/admin/users',
        icon: UserCog,
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
