'use client'

// Минобрнауки order (приказ) SWR read hooks. Talks to
// /api/v1/minobrnauki-orders (see
// internal/modules/work_program/interfaces/http/handlers/minobrnauki_order_handler.go).
// Read-only browse — list + detail. Mirrors the useWorkPrograms query
// hooks (FetchOpts.enabled gate + buildListUrl). The backend applies a
// flat non-student role gate, so no row-level scoping is needed here.

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  MinobrnaukiOrder,
  MinobrnaukiOrdersListResponse,
  MinobrnaukiOrderListFilter,
  RecordMinobrnaukiOrderInput,
} from '@/types/minobrnaukiOrder'

const BASE_URL = '/api/v1/minobrnauki-orders'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
  meta?: { request_id: string; timestamp: string; version: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

interface FetchOpts {
  // enabled defaults to true. Setting it to false short-circuits the SWR
  // key to null, preventing the fetch entirely — used by the page to skip
  // the round-trip while auth is still resolving / for student callers the
  // backend would 403. Mirrors useWorkPrograms' FetchOpts.enabled.
  enabled?: boolean
}

function buildListUrl(filter?: MinobrnaukiOrderListFilter): string {
  if (!filter) return BASE_URL
  const params = new URLSearchParams()
  Object.entries(filter).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    params.append(key, String(value))
  })
  const qs = params.toString()
  return qs ? `${BASE_URL}?${qs}` : BASE_URL
}

// useMinobrnaukiOrders returns the page of orders visible to the caller.
// The backend list use case applies a flat non-student role gate +
// pagination defaults, so the hook passes the raw filter through.
// opts.enabled lets the page skip the fetch until auth resolves.
export function useMinobrnaukiOrders(filter?: MinobrnaukiOrderListFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildListUrl(filter) : null
  const { data, error, isLoading, mutate } = useSWR<MinobrnaukiOrdersListResponse>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    items: data?.items || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// useMinobrnaukiOrder fetches one order by id with its affected
// work-program ids hydrated. Passing null (or enabled=false)
// short-circuits the fetch.
export function useMinobrnaukiOrder(id: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || id == null ? null : `${BASE_URL}/${id}`
  const { data, error, isLoading, mutate } = useSWR<MinobrnaukiOrder>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { order: data, isLoading, error, mutate }
}

// === Mutations (STUB — RED state) ===
// Real implementation + error mapping land in the GREEN commit.

export async function recordMinobrnaukiOrder(
  _input: RecordMinobrnaukiOrderInput
): Promise<MinobrnaukiOrder> {
  return {} as MinobrnaukiOrder
}

export function pickMinobrnaukiOrderErrorKey(_err: unknown): string {
  return 'generic'
}
