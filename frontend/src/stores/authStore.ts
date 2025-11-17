/**
 * Authentication Store
 *
 * Global state management for authentication using Zustand
 */

import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import { User, AuthState, UserRole } from '@/types/auth'

interface AuthStore extends AuthState {
  setUser: (user: User | null) => void
  setToken: (token: string | null) => void
  login: (user: User, token: string) => void
  logout: () => void
  setLoading: (isLoading: boolean) => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,

      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user
        }),

      setToken: (token) =>
        set({ token }),

      login: (user, token) =>
        set({
          user,
          token,
          isAuthenticated: true,
          isLoading: false
        }),

      logout: () =>
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          isLoading: false
        }),

      setLoading: (isLoading) =>
        set({ isLoading })
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage)
    }
  )
)

// Mock user for development
export const mockUser: User = {
  id: '1',
  email: 'admin@example.com',
  name: 'Администратор',
  role: UserRole.SYSTEM_ADMIN,
  avatar: undefined,
  createdAt: new Date('2024-01-01'),
  updatedAt: new Date()
}
