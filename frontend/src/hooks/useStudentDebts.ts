'use client'

// Student debts (Долги студентов) SWR hooks + JSON mutation actions. Talks
// to /api/student-debts (see
// internal/modules/student_debts/interfaces/http/handlers). Mirrors
// useWorkPrograms (list/detail + mutations + FetchOpts.enabled gate). File
// transfer (import multipart / export blob) lives in lib/api/studentDebts.ts.

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  StudentDebt,
  StudentDebtListResponse,
  StudentDebtStats,
  StudentDebtsFilter,
  ScheduleResitInput,
  RecordResitResultInput,
} from '@/types/studentDebts'

const BASE_URL = '/api/student-debts'

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
  // key to null, preventing the fetch — used by pages to skip the round-trip
  // while auth is still resolving. Mirrors useWorkPrograms.
  enabled?: boolean
}

function buildListUrl(base: string, filter?: StudentDebtsFilter): string {
  if (!filter) return base
  const params = new URLSearchParams()
  Object.entries(filter).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    params.append(key, String(value))
  })
  const qs = params.toString()
  return qs ? `${base}?${qs}` : base
}

// useStudentDebts returns the role-scoped registry page. The backend List
// use case scopes the result (staff → all / teacher → own disciplines) and
// clamps pagination, so the hook passes the raw filter through.
export function useStudentDebts(filter?: StudentDebtsFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildListUrl(BASE_URL, filter) : null
  const { data, error, isLoading, mutate } = useSWR<StudentDebtListResponse>(key, fetcher, {
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

// useStudentDebt fetches one debt by id with its attempt timeline hydrated.
// Passing null (or enabled=false) short-circuits the fetch.
export function useStudentDebt(id: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || id == null ? null : `${BASE_URL}/${id}`
  const { data, error, isLoading, mutate } = useSWR<StudentDebt>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { debt: data, isLoading, error, mutate }
}

// useMyStudentDebts returns the caller's own debts (student self-view).
export function useMyStudentDebts(filter?: StudentDebtsFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildListUrl(`${BASE_URL}/my`, filter) : null
  const { data, error, isLoading, mutate } = useSWR<StudentDebtListResponse>(key, fetcher, {
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

// useDebtStats returns the dashboard aggregate (counts per FSM status).
export function useDebtStats(filter?: StudentDebtsFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildListUrl(`${BASE_URL}/stats`, filter) : null
  const { data, error, isLoading, mutate } = useSWR<StudentDebtStats>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { stats: data, isLoading, error, mutate }
}

// === Mutations (resit lifecycle) ===
//
// Each POSTs to its endpoint and returns the updated aggregate. Axios errors
// propagate so callers branch via pickStudentDebtErrorKey. EDIT_ROLES gating
// + ownership are enforced server-side.

export async function scheduleResit(id: number, input: ScheduleResitInput): Promise<StudentDebt> {
  const response = await apiClient.post<ApiResponse<StudentDebt>>(`${BASE_URL}/${id}/resit`, input)
  return response.data
}

export async function recordResitResult(
  id: number,
  attemptNo: number,
  input: RecordResitResultInput
): Promise<StudentDebt> {
  const response = await apiClient.post<ApiResponse<StudentDebt>>(
    `${BASE_URL}/${id}/attempts/${attemptNo}/result`,
    input
  )
  return response.data
}

// === Error mapping ===
//
// Translates backend sentinel codes (mapDebtError) to camelCase i18n keys
// under the studentDebts.errors.* namespace. Status-aware fallback for codes
// the backend omits (plain 403 / 404 / 5xx). For non-manager callers the
// backend collapses scope-forbidden to 404 (IDOR), so 'forbidden' surfaces
// mainly for managers.

const ERROR_CODE_MAP: Record<string, string> = {
  VERSION_CONFLICT: 'versionConflict',
  IDENTITY_EXISTS: 'identityExists',
  DEBT_CLOSED: 'debtClosed',
  NO_SCHEDULED_RESIT: 'noScheduledResit',
  ALREADY_RECORDED: 'alreadyRecorded',
  INVALID_TRANSITION: 'invalidTransition',
  VALIDATION_ERROR: 'validationError',
}

export function pickStudentDebtErrorKey(err: unknown): string {
  if (!err) return 'generic'
  const e = err as {
    response?: { status?: number; data?: { error?: { code?: string } } }
  }
  const code = e.response?.data?.error?.code
  if (code && ERROR_CODE_MAP[code]) return ERROR_CODE_MAP[code]
  const status = e.response?.status
  if (status === 403) return 'forbidden'
  if (status === 404) return 'notFound'
  return 'generic'
}
