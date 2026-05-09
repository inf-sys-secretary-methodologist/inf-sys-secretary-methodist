'use client'

import axios from 'axios'
import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { DisciplineItem, DisciplineItemListResponse } from '@/types/disciplineItem'
import type {
  BulkEditConflict,
  BulkEditConflictResponse,
  BulkEditRequest,
  BulkEditResult,
  BulkEditSuccessResponse,
} from '@/types/bulkEdit'

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
// RequireNonStudent. Passing null short-circuits the SWR key, mirroring
// useSections. opts.enabled=false suppresses the fetch для role-guard
// redirect path symmetry с useCurricula.
export function useDisciplineItems(sectionID: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || sectionID == null ? null : `/api/sections/${sectionID}/items`
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
// transitions, не render-time subscriptions.
export async function fetchDisciplineItem(id: number): Promise<DisciplineItem> {
  const response = await apiClient.get<ApiResponse<DisciplineItem>>(`/api/items/${id}`)
  return response.data
}

// bulkEditDisciplineItems POSTs a combined creates+updates+deletes
// request к /api/sections/:sectionID/items/bulk. Backend applies all
// operations within one tx under Repeatable Read isolation (plan
// ADR-12). Returns a discriminated union:
//   - kind='success' — 200 OK с created/updated/deleted lists
//   - kind='conflict' — 409 VERSION_CONFLICT с per-item conflict
//     entries (collect-all per ADR-12). Caller refetches each via
//     fetchDisciplineItem для accurate current_version display.
// Other axios errors (404 SECTION_NOT_FOUND, 422 EMPTY_BULK_INPUT /
// CROSS_SECTION / NOT_EDITABLE / INVALID_INPUT, 403, 500) propagate
// to caller для mapping via pickErrorKey (Pair 4 utility).
export async function bulkEditDisciplineItems(
  sectionID: number,
  body: BulkEditRequest
): Promise<BulkEditResult> {
  try {
    const response = await apiClient.post<ApiResponse<BulkEditSuccessResponse>>(
      `/api/sections/${sectionID}/items/bulk`,
      body
    )
    return { kind: 'success', data: response.data }
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 409) {
      const conflictBody = err.response.data as Partial<BulkEditConflictResponse> | undefined
      const conflicts: BulkEditConflict[] = Array.isArray(conflictBody?.conflicts)
        ? conflictBody.conflicts
        : []
      return { kind: 'conflict', conflicts }
    }
    throw err
  }
}
