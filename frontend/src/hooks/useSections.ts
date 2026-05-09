'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { SectionListResponse } from '@/types/section'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

interface FetchOpts {
  enabled?: boolean
}

// useSections returns the read-only list of sections for the given
// curriculum. Backend GET /api/curricula/:curriculumID/sections — gated
// by RequireNonStudent. Passing null curriculumID short-circuits the
// SWR key, mirroring the useCurriculum / useMyAssignments hook
// convention so callers can keep the hook at the top of a component
// without firing requests until the id is known. opts.enabled=false
// achieves the same suppression for the role-guard redirect path.
export function useSections(curriculumID: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || curriculumID == null ? null : `/api/curricula/${curriculumID}/sections`
  const { data, error, isLoading, mutate } = useSWR<SectionListResponse>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return {
    items: data?.items || [],
    isLoading,
    error,
    mutate,
  }
}
