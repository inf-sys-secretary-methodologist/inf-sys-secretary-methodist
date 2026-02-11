'use client'

import { useAuthStore } from '@/stores/authStore'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import type { LoginRequest, RegisterRequest } from '@/types/auth'

/**
 * Hook for accessing auth state and actions
 */
export function useAuth() {
  const {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    logout,
    checkAuth,
    clearError,
  } = useAuthStore()

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    logout,
    checkAuth,
    clearError,
  }
}

/**
 * Hook for login with redirect
 */
export function useLogin() {
  const router = useRouter()
  const { login, isLoading, error, clearError } = useAuthStore()

  const handleLogin = async (credentials: LoginRequest, redirectTo: string = '/') => {
    try {
      await login(credentials)
      // Small delay to ensure cookie is written before redirect
      await new Promise((resolve) => setTimeout(resolve, 100))
      router.push(redirectTo)
    } catch (error) {
      // Error is already set in store
      throw error
    }
  }

  return {
    login: handleLogin,
    isLoading,
    error,
    clearError,
  }
}

/**
 * Hook for registration with redirect
 */
export function useRegister() {
  const router = useRouter()
  const { register, isLoading, error, clearError } = useAuthStore()

  const handleRegister = async (data: RegisterRequest, redirectTo: string = '/login') => {
    try {
      await register(data)
      router.push(redirectTo)
    } catch (error) {
      // Error is already set in store
      throw error
    }
  }

  return {
    register: handleRegister,
    isLoading,
    error,
    clearError,
  }
}

/**
 * Hook for logout with redirect
 */
export function useLogout() {
  const router = useRouter()
  const { logout } = useAuthStore()

  const handleLogout = (redirectTo: string = '/login') => {
    logout()
    router.push(redirectTo)
  }

  return {
    logout: handleLogout,
    isLoading: false,
  }
}

/**
 * Hook to check auth status on mount
 * Useful for layout components
 */
export function useAuthCheck() {
  const { checkAuth, user, isAuthenticated, isLoading } = useAuthStore()

  useEffect(() => {
    checkAuth()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []) // Only run once on mount - checkAuth is stable from Zustand

  return {
    user,
    isAuthenticated,
    isLoading,
  }
}

/**
 * Hook to require authentication
 * Redirects to login if not authenticated
 */
export function useRequireAuth(redirectTo: string = '/login') {
  const router = useRouter()
  const { isAuthenticated, isLoading } = useAuthCheck()

  /* c8 ignore start - Auth redirect, tested in e2e */
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push(redirectTo)
    }
  }, [isAuthenticated, isLoading, router, redirectTo])
  /* c8 ignore stop */

  return {
    isAuthenticated,
    isLoading,
  }
}

// ============================================================================
// Granular Selector Hooks
// These hooks use Zustand's selector pattern to prevent cascading re-renders
// by only subscribing to specific slices of state.
// ============================================================================

/**
 * Hook to access only the user object
 * Components using this will only re-render when user changes
 */
export function useUser() {
  return useAuthStore((state) => state.user)
}

/**
 * Hook to access only authentication status
 * Components using this will only re-render when isAuthenticated changes
 */
export function useIsAuthenticated() {
  return useAuthStore((state) => state.isAuthenticated)
}

/**
 * Hook to access only loading state
 * Components using this will only re-render when isLoading changes
 */
export function useAuthLoading() {
  return useAuthStore((state) => state.isLoading)
}

/**
 * Hook to access only error state
 * Components using this will only re-render when error changes
 */
export function useAuthError() {
  return useAuthStore((state) => state.error)
}

/**
 * Hook to access only auth action functions
 * Returns a stable object of action functions that won't cause re-renders
 * since Zustand action functions are stable references
 */
export function useAuthActions() {
  return useAuthStore((state) => ({
    login: state.login,
    register: state.register,
    logout: state.logout,
    checkAuth: state.checkAuth,
    clearError: state.clearError,
  }))
}
