'use client'

import { useEffect, ComponentType, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/stores/authStore'
import type { UserRole } from '@/types/auth'

export interface WithAuthOptions {
  /**
   * Allowed roles for this component
   * If not specified, any authenticated user can access
   */
  roles?: UserRole[]

  /**
   * Redirect path when not authenticated
   * Defaults to '/login'
   */
  redirectTo?: string

  /**
   * Custom loading component
   */
  LoadingComponent?: ComponentType
}

/**
 * Higher-Order Component for client-side route protection
 *
 * @example
 * ```tsx
 * // Protect component - any authenticated user
 * export default withAuth(DashboardPage)
 *
 * // Protect component - only admins
 * export default withAuth(AdminPage, { roles: [UserRole.SYSTEM_ADMIN] })
 *
 * // Protect component - admins and methodists
 * export default withAuth(DocumentsPage, {
 *   roles: [UserRole.SYSTEM_ADMIN, UserRole.METHODIST]
 * })
 * ```
 */
export function withAuth<P extends object>(Component: ComponentType<P>, options?: WithAuthOptions) {
  const displayName = Component.displayName || Component.name || 'Component'

  const ProtectedComponent = (props: P) => {
    const router = useRouter()
    const { isAuthenticated, user, isLoading, checkAuth } = useAuthStore()
    const [authChecked, setAuthChecked] = useState(false)

    const { roles, redirectTo = '/login', LoadingComponent } = options || {}

    // Run checkAuth on mount (same as useAuthCheck does for other pages)
    useEffect(() => {
      checkAuth().finally(() => {
        setAuthChecked(true)
      })
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []) // Only run once on mount

    useEffect(() => {
      // Wait for auth check to complete
      if (!authChecked || isLoading) {
        return
      }

      // Redirect if not authenticated
      if (!isAuthenticated) {
        const currentPath = window.location.pathname
        const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(currentPath)}`
        router.push(redirectUrl)
        return
      }

      // Check role-based access if roles are specified
      if (roles && roles.length > 0) {
        if (!user || !roles.includes(user.role)) {
          router.push('/forbidden')
          return
        }
      }
    }, [isAuthenticated, isLoading, user, router, roles, redirectTo, authChecked])

    // Show loading state
    if (!authChecked || isLoading) {
      if (LoadingComponent) {
        return <LoadingComponent />
      }

      return (
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center space-y-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
            <p className="text-muted-foreground">Загрузка...</p>
          </div>
        </div>
      )
    }

    // Don't render if not authenticated
    if (!isAuthenticated) {
      return null
    }

    // Don't render if user doesn't have required role
    if (roles && roles.length > 0 && user && !roles.includes(user.role)) {
      return null
    }

    // Render protected component
    return <Component {...props} />
  }

  ProtectedComponent.displayName = `withAuth(${displayName})`

  return ProtectedComponent
}
