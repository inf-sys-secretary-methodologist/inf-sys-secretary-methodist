'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  StudentAssignmentView,
  MyAssignmentListResponse,
  SubmissionStatus,
} from '@/types/assignments'

const MY_ASSIGNMENTS_URL = '/api/assignments/my'

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
  // pages to skip the round-trip when the caller is not a student
  // (the redirect to /forbidden is in flight; firing a request that
  // is guaranteed to 401 is wasted bandwidth and a security smell).
  enabled?: boolean
}

// useMyAssignments returns the student's "My Assignments" view — every
// submission the student owns joined with its parent assignment in a
// single round-trip. Optional status pin filters by lifecycle state.
export function useMyAssignments(status?: SubmissionStatus, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  let derivedKey: string | null = null
  if (enabled) {
    derivedKey = MY_ASSIGNMENTS_URL
    if (status) derivedKey += `?status=${encodeURIComponent(status)}`
  }

  const { data, error, isLoading, mutate } = useSWR<MyAssignmentListResponse>(derivedKey, fetcher, {
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

// useMyAssignment fetches the student's combined view for a single
// assignment. Passing null short-circuits the fetch — same gating
// pattern as useAssignment / useSubmissions, so the page can declare
// the hook at the top of the component before the path id is known.
// opts.enabled=false also short-circuits, mirroring useMyAssignments.
export function useMyAssignment(id: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || id == null ? null : `/api/assignments/${id}/my`
  const { data, error, isLoading, mutate } = useSWR<StudentAssignmentView>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { view: data, isLoading, error, mutate }
}
