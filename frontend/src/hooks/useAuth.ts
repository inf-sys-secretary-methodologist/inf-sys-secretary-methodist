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

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push(redirectTo)
    }
  }, [isAuthenticated, isLoading, router, redirectTo])

  return {
    isAuthenticated,
    isLoading,
  }
}
