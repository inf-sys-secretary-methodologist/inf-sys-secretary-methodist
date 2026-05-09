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
// by RequireNonStudent. Pair 1 RED: stub key always null so apiClient
// is never invoked — tests asserting "called with correct URL" fail.
// Pair 1 GREEN replaces the stub key with real URL building.
export function useSections(curriculumID: number | null, opts?: FetchOpts) {
  void curriculumID
  void opts
  const key = null
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
