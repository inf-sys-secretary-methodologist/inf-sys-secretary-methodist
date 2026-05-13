'use client'

import { useState } from 'react'
import { apiClient } from '@/lib/api'
import type { UserRole, UserStatus } from '@/types/user'

interface MutationState {
  isLoading: boolean
  error: Error | null
}

interface Envelope {
  success: boolean
  data?: unknown
  error?: { code: string; message: string }
}

// User management mutation hooks. Each wraps the corresponding
// /api/users/:id/* admin-gated endpoint and tracks isLoading + error
// so the dialog can render spinners + error toasts without owning
// fetch state. They intentionally do NOT auto-invalidate SWR — the
// caller passes mutate from useUsers and runs it on success so the
// table refreshes once.

export function useUpdateUserRole() {
  const [state, setState] = useState<MutationState>({ isLoading: false, error: null })

  const updateRole = async (userId: number, role: UserRole): Promise<void> => {
    setState({ isLoading: true, error: null })
    try {
      await apiClient.put<Envelope>(`/api/users/${userId}/role`, { role })
      setState({ isLoading: false, error: null })
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      setState({ isLoading: false, error: err })
      throw err
    }
  }

  return { updateRole, ...state }
}

export function useUpdateUserStatus() {
  const [state, setState] = useState<MutationState>({ isLoading: false, error: null })

  const updateStatus = async (userId: number, status: UserStatus): Promise<void> => {
    setState({ isLoading: true, error: null })
    try {
      await apiClient.put<Envelope>(`/api/users/${userId}/status`, { status })
      setState({ isLoading: false, error: null })
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      setState({ isLoading: false, error: err })
      throw err
    }
  }

  return { updateStatus, ...state }
}

export function useDeleteUser() {
  const [state, setState] = useState<MutationState>({ isLoading: false, error: null })

  const deleteUser = async (userId: number): Promise<void> => {
    setState({ isLoading: true, error: null })
    try {
      await apiClient.delete<Envelope>(`/api/users/${userId}`)
      setState({ isLoading: false, error: null })
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      setState({ isLoading: false, error: err })
      throw err
    }
  }

  return { deleteUser, ...state }
}
