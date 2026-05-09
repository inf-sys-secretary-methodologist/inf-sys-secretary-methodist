'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { DisciplineItem, DisciplineItemListResponse } from '@/types/disciplineItem'

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

// useDisciplineItems returns the read-only list of items for the given
// section. Backend GET /api/sections/:sectionID/items — gated by
// RequireNonStudent. Pair 2 RED: stub key=null → no fetch fires; tests
// asserting URL fail. Pair 2 GREEN replaces stub с real URL builder.
export function useDisciplineItems(sectionID: number | null, opts?: FetchOpts) {
  void sectionID
  void opts
  const key = null
  const { data, error, isLoading, mutate } = useSWR<DisciplineItemListResponse>(key, fetcher, {
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

// fetchDisciplineItem imperatively reads a single item by id. Used by
// the 409 VERSION_CONFLICT recovery flow per plan ADR-12: backend
// returns CurrentVersion=0 under Repeatable Read isolation; frontend
// MUST refetch outside the failed tx via GET /api/items/:id для
// accurate current state перед showing merge UI per row. Imperative
// (not a hook) since the call is triggered by post-await state machine
// transitions, не render-time subscriptions. Pair 2 RED stub throws.
export async function fetchDisciplineItem(_id: number): Promise<DisciplineItem> {
  void _id
  throw new Error('fetchDisciplineItem not implemented (Pair 2 RED)')
}
