'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { UserListFilter, UserListResponse } from '@/types/user'

interface FetchOpts {
  enabled?: boolean
}

interface Envelope {
  success: boolean
  data: UserListResponse
  meta?: unknown
  error?: { code: string; message: string }
}

// buildURL assembles the /api/users query string skipping empty
// values so the backend defaults kick in for unfilled filter
// dimensions. ListUsers clamps page/limit anyway; we omit them when
// the caller hasn't set them.
function buildURL(filter: UserListFilter): string {
  const params = new URLSearchParams()
  if (filter.page) params.set('page', String(filter.page))
  if (filter.limit) params.set('limit', String(filter.limit))
  if (filter.search) params.set('search', filter.search)
  if (filter.role) params.set('role', filter.role)
  if (filter.status) params.set('status', filter.status)
  if (filter.department_id) params.set('department_id', String(filter.department_id))
  if (filter.position_id) params.set('position_id', String(filter.position_id))
  const qs = params.toString()
  return qs ? `/api/users?${qs}` : '/api/users'
}

const usersFetcher = async (url: string): Promise<UserListResponse> => {
  const envelope = await apiClient.get<Envelope>(url)
  return envelope.data
}

// useUsers reads GET /api/users with filter + pagination params. The
// read endpoint is permissive (any authenticated caller may read);
// the page-level role guard enforces system_admin separately.
export function useUsers(filter: UserListFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildURL(filter) : null
  const { data, error, isLoading, mutate } = useSWR<UserListResponse>(key, usersFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    users: data?.users ?? [],
    total: data?.total ?? 0,
    page: data?.page ?? 1,
    limit: data?.limit ?? 0,
    totalPages: data?.total_pages ?? 0,
    isLoading,
    error,
    mutate,
  }
}
