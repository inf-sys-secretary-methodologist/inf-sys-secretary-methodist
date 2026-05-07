'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  Curriculum,
  CurriculumListResponse,
  CurriculumListFilter,
  UpdateCurriculumRequest,
  RejectCurriculumRequest,
} from '@/types/curriculum'

const CURRICULUM_URL = '/api/curriculum'

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
  // enabled defaults to true. Setting it to false short-circuits the
  // SWR key to null, which prevents the fetch entirely — used by the
  // /curriculum page to skip the round-trip when the caller is a
  // student (the redirect to /forbidden is in flight; firing a request
  // that is guaranteed to 401 is wasted bandwidth and a security
  // smell). Mirrors useMyAssignments v0.114.0 SEC pattern.
  enabled?: boolean
}

function buildCurriculaUrl(filter?: CurriculumListFilter): string {
  if (!filter) return CURRICULUM_URL
  const params = new URLSearchParams()
  if (filter.status) params.append('status', filter.status)
  if (typeof filter.year === 'number') params.append('year', String(filter.year))
  if (filter.specialty) params.append('specialty', filter.specialty)
  if (typeof filter.created_by === 'number') {
    params.append('created_by', String(filter.created_by))
  }
  if (typeof filter.limit === 'number') params.append('limit', String(filter.limit))
  if (typeof filter.offset === 'number') params.append('offset', String(filter.offset))
  const qs = params.toString()
  return qs ? `${CURRICULUM_URL}?${qs}` : CURRICULUM_URL
}

// useCurricula returns the page of curricula visible to the caller.
// Backend GET /api/curriculum is gated by RequireNonStudent (v0.116.0)
// — methodist / system_admin / academic_secretary / teacher all see the
// full list. The optional opts.enabled flag lets the page skip the
// fetch entirely when the role guard hasn't yet redirected the
// student.
export function useCurricula(filter?: CurriculumListFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildCurriculaUrl(filter) : null
  const { data, error, isLoading, mutate } = useSWR<CurriculumListResponse>(key, fetcher, {
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

// useCurriculum fetches a single curriculum by id. Passing null
// short-circuits the fetch so callers can keep the hook at the top of
// a component without firing requests until the id is known. The
// optional opts.enabled flag mirrors useCurricula for symmetry with
// the page-level role guard.
export function useCurriculum(id: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || id == null ? null : `${CURRICULUM_URL}/${id}`
  const { data, error, isLoading, mutate } = useSWR<Curriculum>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { curriculum: data, isLoading, error, mutate }
}

// updateCurriculum PUTs the edit payload to PUT /api/curriculum/:id
// and returns the unwrapped Curriculum. On error the underlying axios
// error propagates so the caller (EditCurriculumDialog) can branch
// on 409 (CODE_EXISTS) / 422 (NOT_EDITABLE / INVALID_INPUT) / 403
// (forbidden) by HTTP status — mirrors saveGrade's error contract.
export async function updateCurriculum(
  id: number,
  body: UpdateCurriculumRequest
): Promise<Curriculum> {
  const response = await apiClient.put<ApiResponse<Curriculum>>(`${CURRICULUM_URL}/${id}`, body)
  return response.data
}

// submitCurriculum POSTs to the submit endpoint to transition a
// draft → pending_approval. Body is empty per the backend contract
// (path id + JWT subject identify the row + actor). Returns the
// updated Curriculum with status='pending_approval'. Axios errors
// propagate — caller distinguishes 422 (NOT_DRAFT) / 403 (forbidden)
// by HTTP status.
export async function submitCurriculum(id: number): Promise<Curriculum> {
  const response = await apiClient.post<ApiResponse<Curriculum>>(
    `${CURRICULUM_URL}/${id}/submit`,
    {}
  )
  return response.data
}

// approveCurriculum POSTs to the admin-only approve endpoint to
// transition pending_approval → approved. Body is empty per the
// backend contract (path id + JWT subject identify the row + admin).
// Returns the updated Curriculum со status='approved' и populated
// approved_by / approved_at fields. Axios errors propagate — caller
// distinguishes 422 (NOT_PENDING) / 403 (forbidden, defended даже
// unreachable за RequireRole(SystemAdmin) middleware) by HTTP status.
export async function approveCurriculum(id: number): Promise<Curriculum> {
  const response = await apiClient.post<ApiResponse<Curriculum>>(
    `${CURRICULUM_URL}/${id}/approve`,
    {}
  )
  return response.data
}

// rejectCurriculum POSTs to the admin-only reject endpoint to
// transition pending_approval → draft. Body carries the rejection
// reason — backend audits it verbatim (ADR-3 v0.117.0: audit-only,
// not stored on the entity, so a future rework cycle starts clean).
// Returns the updated Curriculum со status='draft'. Axios errors
// propagate — caller distinguishes 422 (NOT_PENDING) / 403 (forbidden)
// / 400 (empty reason) by HTTP status.
export async function rejectCurriculum(
  id: number,
  body: RejectCurriculumRequest
): Promise<Curriculum> {
  const response = await apiClient.post<ApiResponse<Curriculum>>(
    `${CURRICULUM_URL}/${id}/reject`,
    body
  )
  return response.data
}
