/**
 * useAuth Hook
 *
 * Convenient hook for accessing authentication state and actions
 */

import { useEffect } from 'react'
import { useAuthStore, mockUser } from '@/stores/authStore'
import { User, UserRole } from '@/types/auth'

export function useAuth() {
  const { user, token, isAuthenticated, isLoading, login, logout, setLoading } =
    useAuthStore()

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    login,
    logout,
    setLoading
  }
}

/**
 * useAuthCheck Hook
 *
 * Automatically initializes mock user for development
 * In production, this would validate the token and fetch user data
 */
export function useAuthCheck() {
  const { user, isAuthenticated, login, setLoading } = useAuthStore()

  useEffect(() => {
    // Simulate auth check
    const checkAuth = async () => {
      setLoading(true)

      // In development, automatically log in with mock user
      // In production, this would validate the token from cookies/localStorage
      if (!isAuthenticated) {
        // Simulate API call delay
        await new Promise(resolve => setTimeout(resolve, 500))

        // Auto-login with mock user for development
        login(mockUser, 'mock-token-' + Date.now())
      }

      setLoading(false)
    }

    checkAuth()
  }, [isAuthenticated, login, setLoading])

  return {
    user,
    isLoading: false, // Set to false to avoid loading screen in development
    isAuthenticated
  }
}

/**
 * useRequireAuth Hook
 *
 * Requires authentication and optionally specific roles
 * Redirects to login if not authenticated or forbidden if no permission
 */
export function useRequireAuth(allowedRoles?: UserRole[]) {
  const { user, isAuthenticated, isLoading } = useAuth()

  useEffect(() => {
    if (isLoading) return

    if (!isAuthenticated) {
      // In production, redirect to login
      console.warn('User not authenticated')
    } else if (allowedRoles && user && !allowedRoles.includes(user.role)) {
      // In production, redirect to forbidden page
      console.warn('User does not have permission')
    }
  }, [user, isAuthenticated, isLoading, allowedRoles])

  return {
    user,
    isLoading,
    isAuthenticated,
    hasPermission:
      !allowedRoles || (user && allowedRoles.includes(user.role)) || false
  }
}
