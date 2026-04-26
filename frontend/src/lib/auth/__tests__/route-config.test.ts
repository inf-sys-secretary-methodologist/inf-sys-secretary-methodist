import {
  publicRoutes,
  authRoutes,
  protectedRoutes,
  matchesRoute,
  hasRouteAccess,
  getRouteConfig,
} from '../route-config'
import { UserRole } from '@/types/auth'

describe('route-config', () => {
  describe('publicRoutes', () => {
    it('includes login page', () => {
      expect(publicRoutes).toContain('/login')
    })

    it('includes register page', () => {
      expect(publicRoutes).toContain('/register')
    })

    it('includes home page', () => {
      expect(publicRoutes).toContain('/')
    })

    it('includes forgot-password page', () => {
      expect(publicRoutes).toContain('/forgot-password')
    })

    it('includes reset-password page', () => {
      expect(publicRoutes).toContain('/reset-password')
    })
  })

  describe('authRoutes', () => {
    it('includes login', () => {
      expect(authRoutes).toContain('/login')
    })

    it('includes register', () => {
      expect(authRoutes).toContain('/register')
    })
  })

  describe('protectedRoutes', () => {
    it('includes admin route for SYSTEM_ADMIN only', () => {
      const adminRoute = protectedRoutes.find((r) => r.path === '/admin')
      expect(adminRoute?.allowedRoles).toEqual([UserRole.SYSTEM_ADMIN])
    })

    it('includes dashboard for all roles', () => {
      const dashboardRoute = protectedRoutes.find((r) => r.path === '/dashboard')
      expect(dashboardRoute?.allowedRoles).toEqual(Object.values(UserRole))
    })

    it('includes /files for all authenticated roles', () => {
      const filesRoute = protectedRoutes.find((r) => r.path === '/files')
      expect(filesRoute).toBeDefined()
      expect(filesRoute?.allowedRoles).toEqual([
        UserRole.SYSTEM_ADMIN,
        UserRole.METHODIST,
        UserRole.ACADEMIC_SECRETARY,
        UserRole.TEACHER,
        UserRole.STUDENT,
      ])
    })
  })

  describe('matchesRoute', () => {
    it('matches exact route when exactMatch is true', () => {
      expect(matchesRoute('/login', '/login', true)).toBe(true)
      expect(matchesRoute('/login/extra', '/login', true)).toBe(false)
    })

    it('matches route prefix when exactMatch is false', () => {
      expect(matchesRoute('/documents', '/documents', false)).toBe(true)
      expect(matchesRoute('/documents/123', '/documents', false)).toBe(true)
      expect(matchesRoute('/documents-other', '/documents', false)).toBe(false)
    })

    it('matches exact route even without exactMatch flag', () => {
      expect(matchesRoute('/profile', '/profile', false)).toBe(true)
    })

    it('does not match different routes', () => {
      expect(matchesRoute('/users', '/admin', false)).toBe(false)
    })
  })

  describe('hasRouteAccess', () => {
    describe('public routes', () => {
      it('allows access to login for any role', () => {
        expect(hasRouteAccess('/login', UserRole.STUDENT)).toBe(true)
        expect(hasRouteAccess('/login', UserRole.SYSTEM_ADMIN)).toBe(true)
      })

      it('allows access to home page', () => {
        expect(hasRouteAccess('/', UserRole.STUDENT)).toBe(true)
      })
    })

    describe('admin routes', () => {
      it('allows SYSTEM_ADMIN access to admin routes', () => {
        expect(hasRouteAccess('/admin', UserRole.SYSTEM_ADMIN)).toBe(true)
      })

      it('denies non-admin access to admin routes', () => {
        expect(hasRouteAccess('/admin', UserRole.STUDENT)).toBe(false)
        expect(hasRouteAccess('/admin', UserRole.TEACHER)).toBe(false)
        expect(hasRouteAccess('/admin', UserRole.METHODIST)).toBe(false)
      })
    })

    describe('documents routes', () => {
      it('allows all main roles to access documents', () => {
        expect(hasRouteAccess('/documents', UserRole.SYSTEM_ADMIN)).toBe(true)
        expect(hasRouteAccess('/documents', UserRole.METHODIST)).toBe(true)
        expect(hasRouteAccess('/documents', UserRole.ACADEMIC_SECRETARY)).toBe(true)
        expect(hasRouteAccess('/documents', UserRole.TEACHER)).toBe(true)
        expect(hasRouteAccess('/documents', UserRole.STUDENT)).toBe(true)
      })
    })

    describe('reports routes (per 0.102.2 matrix)', () => {
      it('allows admin, methodist, secretary, and teacher to access reports', () => {
        expect(hasRouteAccess('/reports', UserRole.SYSTEM_ADMIN)).toBe(true)
        expect(hasRouteAccess('/reports', UserRole.METHODIST)).toBe(true)
        expect(hasRouteAccess('/reports', UserRole.ACADEMIC_SECRETARY)).toBe(true)
        expect(hasRouteAccess('/reports', UserRole.TEACHER)).toBe(true)
      })

      it('denies student access to reports', () => {
        expect(hasRouteAccess('/reports', UserRole.STUDENT)).toBe(false)
      })
    })

    describe('integration routes (per 0.102.2 matrix)', () => {
      it('allows only admin', () => {
        expect(hasRouteAccess('/integration', UserRole.SYSTEM_ADMIN)).toBe(true)
      })

      it('denies all other roles including methodist', () => {
        expect(hasRouteAccess('/integration', UserRole.METHODIST)).toBe(false)
        expect(hasRouteAccess('/integration', UserRole.ACADEMIC_SECRETARY)).toBe(false)
        expect(hasRouteAccess('/integration', UserRole.TEACHER)).toBe(false)
        expect(hasRouteAccess('/integration', UserRole.STUDENT)).toBe(false)
      })
    })

    describe('common protected routes', () => {
      it('allows all authenticated users to access dashboard', () => {
        Object.values(UserRole).forEach((role) => {
          expect(hasRouteAccess('/dashboard', role)).toBe(true)
        })
      })

      it('allows all authenticated users to access profile', () => {
        Object.values(UserRole).forEach((role) => {
          expect(hasRouteAccess('/profile', role)).toBe(true)
        })
      })
    })

    describe('unknown routes', () => {
      it('allows access to routes not in config', () => {
        expect(hasRouteAccess('/unknown-route', UserRole.STUDENT)).toBe(true)
      })
    })
  })

  describe('getRouteConfig', () => {
    it('returns config for known protected route', () => {
      const config = getRouteConfig('/admin')
      expect(config).not.toBeNull()
      expect(config?.path).toBe('/admin')
      expect(config?.allowedRoles).toContain(UserRole.SYSTEM_ADMIN)
    })

    it('returns config for nested route', () => {
      const config = getRouteConfig('/documents/123')
      expect(config).not.toBeNull()
      expect(config?.path).toBe('/documents')
    })

    it('returns null for public route', () => {
      const config = getRouteConfig('/login')
      expect(config).toBeNull()
    })

    it('returns null for unknown route', () => {
      const config = getRouteConfig('/some-unknown-path')
      expect(config).toBeNull()
    })

    it('returns dashboard config for dashboard route', () => {
      const config = getRouteConfig('/dashboard')
      expect(config).not.toBeNull()
      expect(config?.allowedRoles).toEqual(Object.values(UserRole))
    })
  })
})
