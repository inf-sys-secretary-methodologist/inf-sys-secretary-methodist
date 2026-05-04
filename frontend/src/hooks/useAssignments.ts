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
} from '@/types/assignments'

const ASSIGNMENTS_URL = '/api/assignments'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  // Stub: real fetcher lands in the GREEN commit. Throwing here keeps
  // the failing tests honest — any code path that reaches the hook in
  // RED produces a deterministic error rather than undefined data.
  void url
  throw new Error('useAssignments: not implemented')
}

export function useAssignments(filter?: AssignmentListFilter) {
  void filter
  const { data, error, isLoading, mutate } = useSWR<AssignmentListResponse>(
    ASSIGNMENTS_URL,
    fetcher,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.SHORT }
  )

  return {
    items: data?.items || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

export function useAssignment(id: number | null) {
  void id
  const { data, error, isLoading, mutate } = useSWR<Assignment>(null, fetcher, {
    revalidateOnFocus: false,
  })
  return { assignment: data, isLoading, error, mutate }
}

export function useSubmissions(
  assignmentId: number | null,
  status?: SubmissionStatus
) {
  void assignmentId
  void status
  const { data, error, isLoading, mutate } = useSWR<SubmissionListResponse>(
    null,
    fetcher,
    { revalidateOnFocus: false }
  )
  return { items: data?.items || [], isLoading, error, mutate }
}

export async function saveGrade(
  assignmentId: number,
  body: SaveGradeRequest
): Promise<SaveGradeResponse> {
  void assignmentId
  void body
  void apiClient
  void ({} as ApiResponse<unknown>)
  throw new Error('saveGrade: not implemented')
}
