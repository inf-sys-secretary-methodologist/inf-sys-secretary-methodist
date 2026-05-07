'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  Assignment,
  AssignmentListResponse,
  AssignmentListFilter,
  SubmissionListResponse,
  SubmissionStatus,
  SaveGradeRequest,
  SaveGradeResponse,
  ReturnSubmissionRequest,
  ReturnSubmissionResponse,
} from '@/types/assignments'

const ASSIGNMENTS_URL = '/api/assignments'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

function buildAssignmentsUrl(filter?: AssignmentListFilter): string {
  if (!filter) return ASSIGNMENTS_URL
  const params = new URLSearchParams()
  if (filter.subject) params.append('subject', filter.subject)
  if (filter.group_name) params.append('group_name', filter.group_name)
  if (typeof filter.page_size === 'number') params.append('page_size', String(filter.page_size))
  if (typeof filter.offset === 'number') params.append('offset', String(filter.offset))
  const qs = params.toString()
  return qs ? `${ASSIGNMENTS_URL}?${qs}` : ASSIGNMENTS_URL
}

// useAssignments returns the page of assignments visible to the caller
// according to backend caller-scope (teacher: own only; methodist /
// secretary / admin: any).
export function useAssignments(filter?: AssignmentListFilter) {
  const url = buildAssignmentsUrl(filter)
  const { data, error, isLoading, mutate } = useSWR<AssignmentListResponse>(url, fetcher, {
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

// useAssignment fetches a single assignment by id. Passing null
// short-circuits the fetch so callers can keep the hook at the top of
// a component without firing requests until the id is known.
export function useAssignment(id: number | null) {
  const key = id == null ? null : `${ASSIGNMENTS_URL}/${id}`
  const { data, error, isLoading, mutate } = useSWR<Assignment>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { assignment: data, isLoading, error, mutate }
}

// useSubmissions fetches the submission list for a single assignment.
// Optional status pins to a single lifecycle state ("pending", "graded",
// "returned"). Passing null assignmentId short-circuits the fetch.
export function useSubmissions(assignmentId: number | null, status?: SubmissionStatus) {
  let key: string | null = null
  if (assignmentId != null) {
    key = `${ASSIGNMENTS_URL}/${assignmentId}/submissions`
    if (status) key += `?status=${encodeURIComponent(status)}`
  }
  const { data, error, isLoading, mutate } = useSWR<SubmissionListResponse>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { items: data?.items || [], isLoading, error, mutate }
}

// saveGrade POSTs the grade payload to the backend SaveGrade endpoint.
// On error the underlying axios error propagates so the caller can
// distinguish 409 / 422 / 403 by status.
export async function saveGrade(
  assignmentId: number,
  body: SaveGradeRequest
): Promise<SaveGradeResponse> {
  const response = await apiClient.post<ApiResponse<SaveGradeResponse>>(
    `${ASSIGNMENTS_URL}/${assignmentId}/grades`,
    body
  )
  return response.data
}

// returnSubmission POSTs the return-for-revision payload to the backend
// ReturnHandler endpoint. On error the underlying axios error
// propagates so the caller can distinguish 409 (already returned) /
// 422 (invalid reason) / 403 (forbidden) by HTTP status — mirrors
// saveGrade's error contract.
export async function returnSubmission(
  assignmentId: number,
  body: ReturnSubmissionRequest
): Promise<ReturnSubmissionResponse> {
  const response = await apiClient.post<ApiResponse<ReturnSubmissionResponse>>(
    `${ASSIGNMENTS_URL}/${assignmentId}/returns`,
    body
  )
  return response.data
}
