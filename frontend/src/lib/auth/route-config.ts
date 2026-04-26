import { UserRole } from '@/types/auth'

/**
 * Route configuration for role-based access control
 */

export interface RouteConfig {
  path: string
  allowedRoles: UserRole[]
  exactMatch?: boolean
}

/**
 * Public routes accessible without authentication
 */
export const publicRoutes = [
  '/',
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
  '/forbidden',
  '/offline',
]

/**
 * Auth routes that redirect to home if already authenticated
 */
export const authRoutes = ['/login', '/register', '/forgot-password', '/reset-password']

/**
 * Protected routes with role-based access control
 * Routes are checked in order - first match wins
 */
export const protectedRoutes: RouteConfig[] = [
  // Admin-only routes
  {
    path: '/admin',
    allowedRoles: [UserRole.SYSTEM_ADMIN],
  },
  // Users management - accessible to admins, methodists, secretaries and teachers
  {
    path: '/users',
    allowedRoles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
    ],
  },
  // 1C Integration - admin and methodist only
  {
    path: '/integration',
    allowedRoles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
  },

  // Methodist routes
  {
    path: '/documents',
    allowedRoles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT, // view only
    ],
  },
  {
    path: '/templates',
    allowedRoles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST],
  },
  {
    path: '/reports',
    allowedRoles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY],
  },

  // Secretary routes
  {
    path: '/schedule',
    allowedRoles: [UserRole.SYSTEM_ADMIN, UserRole.ACADEMIC_SECRETARY, UserRole.METHODIST],
  },
  {
    path: '/tasks',
    allowedRoles: [UserRole.SYSTEM_ADMIN, UserRole.ACADEMIC_SECRETARY, UserRole.METHODIST],
  },

  // Calendar routes
  {
    path: '/calendar',
    allowedRoles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT, // view only
    ],
  },

  // Files - all authenticated users
  {
    path: '/files',
    allowedRoles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },

  // Announcements - all authenticated users; STUDENT view-only
  {
    path: '/announcements',
    allowedRoles: [
      UserRole.SYSTEM_ADMIN,
      UserRole.METHODIST,
      UserRole.ACADEMIC_SECRETARY,
      UserRole.TEACHER,
      UserRole.STUDENT,
    ],
  },

  // Common protected routes (all authenticated users)
  {
    path: '/dashboard',
    allowedRoles: Object.values(UserRole),
  },
  {
    path: '/profile',
    allowedRoles: Object.values(UserRole),
  },
  {
    path: '/settings',
    allowedRoles: Object.values(UserRole),
  },
]

/**
 * Check if a path matches a route pattern
 */
export function matchesRoute(pathname: string, routePath: string, exactMatch = false): boolean {
  if (exactMatch) {
    return pathname === routePath
  }

  // For non-exact match, check if pathname starts with routePath
  return pathname === routePath || pathname.startsWith(`${routePath}/`)
}

/**
 * Check if user has access to a specific route
 */
export function hasRouteAccess(pathname: string, userRole: UserRole): boolean {
  // Check if it's a public route
  if (publicRoutes.some((route) => matchesRoute(pathname, route, true))) {
    return true
  }

  // Check protected routes
  for (const route of protectedRoutes) {
    if (matchesRoute(pathname, route.path, route.exactMatch)) {
      return route.allowedRoles.includes(userRole)
    }
  }

  // If no specific rule found, allow access (default protected route)
  return true
}

/**
 * Get route config for a specific path
 */
export function getRouteConfig(pathname: string): RouteConfig | null {
  for (const route of protectedRoutes) {
    if (matchesRoute(pathname, route.path, route.exactMatch)) {
      return route
    }
  }
  return null
}
