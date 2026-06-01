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
  CreateRevisionInput,
  GoalInput,
  CompetenceInput,
  TopicInput,
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

// generateWorkProgram asks the backend to fill an empty draft from an LLM
// (see GenerateDraftUseCase). Empty body — path id + JWT subject identify
// the row + actor. The backend enforces the invariants (draft must be
// empty → DRAFT_NOT_EMPTY/409, hourly quota → RATE_LIMITED/429); callers
// branch on those via pickWorkProgramErrorKey.
export async function generateWorkProgram(id: number): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${id}/generate`, {})
  return response.data
}

// === Revision (лист актуализации) write-workflow mutations ===
//
// The revision endpoints are nested under a parent РПД
// (/work-programs/:id/revisions/...). Every mutation returns the updated
// parent aggregate (the revision read-projection rides along in
// wp.revisions), so callers can `mutate()` the detail SWR cache directly.
// Errors propagate so dialogs branch via pickWorkProgramErrorKey
// (REVISION_NOT_PERMITTED / INVALID_TRANSITION / VERSION_CONFLICT / ...).

// createRevision proposes a draft лист актуализации on an approved /
// needs_revision РПД. The author derives from the JWT subject server-side.
export async function createRevision(
  workProgramId: number,
  input: CreateRevisionInput
): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${workProgramId}/revisions`,
    input
  )
  return response.data
}

// submitRevision moves a draft revision to pending_approval. Empty body —
// path ids + JWT subject identify the row + actor.
export async function submitRevision(
  workProgramId: number,
  revisionId: number
): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${workProgramId}/revisions/${revisionId}/submit`,
    {}
  )
  return response.data
}

// approveRevision moves a pending revision to approved (approver-side).
// Empty body — the approver derives from the JWT subject server-side.
export async function approveRevision(
  workProgramId: number,
  revisionId: number
): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${workProgramId}/revisions/${revisionId}/approve`,
    {}
  )
  return response.data
}

// rejectRevision rejects a pending revision with a mandatory reason
// (approver-side). Reuses RejectWorkProgramInput — identical { reason }
// shape as the РПД reject.
export async function rejectRevision(
  workProgramId: number,
  revisionId: number,
  body: RejectWorkProgramInput
): Promise<WorkProgram> {
  const response = await apiClient.post<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${workProgramId}/revisions/${revisionId}/reject`,
    body
  )
  return response.data
}

// === Collection-edit mutations (slice 12c) ===
//
// Manual editing of the five inner collections (goals / competences /
// topics / assessments / references). Each add/update/delete hits the
// 12b content endpoints and returns the full updated parent aggregate,
// so callers `mutate()` the detail SWR cache with the result directly.
// Author-scoped + status-gated server-side; errors propagate so dialogs
// branch via pickWorkProgramErrorKey.

// --- Goals ---
export async function addGoal(wpId: number, input: GoalInput): Promise<WorkProgram> {
  const r = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${wpId}/goals`, input)
  return r.data
}
export async function updateGoal(
  wpId: number,
  goalId: number,
  input: GoalInput
): Promise<WorkProgram> {
  const r = await apiClient.put<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${wpId}/goals/${goalId}`,
    input
  )
  return r.data
}
export async function deleteGoal(wpId: number, goalId: number): Promise<WorkProgram> {
  const r = await apiClient.delete<ApiResponse<WorkProgram>>(`${BASE_URL}/${wpId}/goals/${goalId}`)
  return r.data
}

// --- Competences ---
export async function addCompetence(wpId: number, input: CompetenceInput): Promise<WorkProgram> {
  const r = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${wpId}/competences`, input)
  return r.data
}
export async function updateCompetence(
  wpId: number,
  competenceId: number,
  input: CompetenceInput
): Promise<WorkProgram> {
  const r = await apiClient.put<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${wpId}/competences/${competenceId}`,
    input
  )
  return r.data
}
export async function deleteCompetence(wpId: number, competenceId: number): Promise<WorkProgram> {
  const r = await apiClient.delete<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${wpId}/competences/${competenceId}`
  )
  return r.data
}

// --- Topics ---
export async function addTopic(wpId: number, input: TopicInput): Promise<WorkProgram> {
  const r = await apiClient.post<ApiResponse<WorkProgram>>(`${BASE_URL}/${wpId}/topics`, input)
  return r.data
}
export async function updateTopic(
  wpId: number,
  topicId: number,
  input: TopicInput
): Promise<WorkProgram> {
  const r = await apiClient.put<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${wpId}/topics/${topicId}`,
    input
  )
  return r.data
}
export async function deleteTopic(wpId: number, topicId: number): Promise<WorkProgram> {
  const r = await apiClient.delete<ApiResponse<WorkProgram>>(
    `${BASE_URL}/${wpId}/topics/${topicId}`
  )
  return r.data
}

// Assessments / references mutations land in 12c-2b alongside their section
// wiring.

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
  RATE_LIMITED: 'rateLimited',
  DRAFT_NOT_EMPTY: 'draftNotEmpty',
  REVISION_NOT_PERMITTED: 'revisionNotPermitted',
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
