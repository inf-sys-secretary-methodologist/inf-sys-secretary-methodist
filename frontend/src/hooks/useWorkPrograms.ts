'use client'

// Work program (РПД) SWR hooks + mutation actions. Talks to
// /api/v1/work-programs (see
// internal/modules/work_program/interfaces/http/handlers/work_program_handler.go).
// Mirrors useExtracurricularEvents (list/detail + mutations) and the
// useCurricula FetchOpts.enabled gate.

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  WorkProgram,
  WorkProgramListResponse,
  WorkProgramListFilter,
  CreateWorkProgramInput,
  RejectWorkProgramInput,
} from '@/types/workProgram'

const BASE_URL = '/api/v1/work-programs'

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
  // key to null, preventing the fetch entirely — used by the pages to
  // skip the round-trip while auth is still resolving. Mirrors
  // useCurricula's FetchOpts.enabled SEC pattern (v0.114.0).
  enabled?: boolean
}

function buildListUrl(filter?: WorkProgramListFilter): string {
  if (!filter) return BASE_URL
  const params = new URLSearchParams()
  Object.entries(filter).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    params.append(key, String(value))
  })
  const qs = params.toString()
  return qs ? `${BASE_URL}?${qs}` : BASE_URL
}

// useWorkPrograms returns the page of РПД visible to the caller. The
// backend List use case role-scopes the result (teacher → own /
// student → approved only / methodist+secretary+admin → all) and
// applies pagination defaults, so the hook passes the raw filter
// through. opts.enabled lets a page skip the fetch until auth resolves.
export function useWorkPrograms(filter?: WorkProgramListFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildListUrl(filter) : null
  const { data, error, isLoading, mutate } = useSWR<WorkProgramListResponse>(key, fetcher, {
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

// useWorkProgram fetches one РПД by id with all six inner collections
// hydrated. Passing null (or enabled=false) short-circuits the fetch.
export function useWorkProgram(id: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || id == null ? null : `${BASE_URL}/${id}`
  const { data, error, isLoading, mutate } = useSWR<WorkProgram>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { workProgram: data, isLoading, error, mutate }
}

// === Mutations ===
//
// Each transition POSTs to its endpoint and returns the updated
// aggregate. Submit/approve/discard carry an empty body (path id + JWT
// subject identify the row + actor); reject carries the mandatory
// reason. Axios errors propagate so callers can branch via
// pickWorkProgramErrorKey.

export async function createWorkProgram(input: CreateWorkProgramInput): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(BASE_URL, input)
  return response.data
}

export async function submitWorkProgram(id: number): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${id}/submit`, {})
  return response.data
}

export async function approveWorkProgram(id: number): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${id}/approve`, {})
  return response.data
}

export async function rejectWorkProgram(
  id: number,
  body: RejectWorkProgramInput
): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${id}/reject`, body)
  return response.data
}

export async function discardWorkProgram(id: number): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${id}/discard`, {})
  return response.data
}

// === Error mapping ===
//
// Translates backend sentinel codes (mapWorkProgramError) to camelCase
// i18n keys under the workProgram.errors.* namespace. Status-aware
// fallback for codes the backend omits (plain 403 / 404 / 5xx). Note:
// for non-admin callers the backend collapses scope-forbidden to 404
// (IDOR mitigation), so 'forbidden' surfaces mainly for admins.

const ERROR_CODE_MAP: Record<string, string> = {
  IDENTITY_EXISTS: 'identityExists',
  VERSION_CONFLICT: 'versionConflict',
  INVALID_TRANSITION: 'invalidTransition',
  REJECT_REASON_REQUIRED: 'rejectReasonRequired',
  INVALID_WORK_PROGRAM: 'invalidWorkProgram',
}

export function pickWorkProgramErrorKey(err: unknown): string {
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
