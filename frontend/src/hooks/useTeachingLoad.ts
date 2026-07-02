'use client'

// Teaching load (Нагрузка) SWR hook + JSON mutation actions. Talks to
// /api/schedule/teaching-load (see
// internal/modules/schedule/interfaces/http/handlers/teaching_load_handler.go).
// Mirrors useStudentDebts (list + FetchOpts.enabled gate + mutations that
// unwrap the {success,data} envelope manually, since apiClient does not).

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  TeachingLoad,
  TeachingLoadInput,
  TeachingLoadFilter,
  TeachingLoadListResponse,
} from '@/types/teachingLoad'

export const TEACHING_LOAD_URL = '/api/schedule/teaching-load'

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

export function buildTeachingLoadUrl(filter?: TeachingLoadFilter): string {
  if (!filter) return TEACHING_LOAD_URL
  const params = new URLSearchParams()
  Object.entries(filter).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    params.append(key, String(value))
  })
  const qs = params.toString()
  return qs ? `${TEACHING_LOAD_URL}?${qs}` : TEACHING_LOAD_URL
}

// useTeachingLoads returns the load registry, optionally filtered.
export function useTeachingLoads(_filter?: TeachingLoadFilter, _opts?: FetchOpts) {
  const { data, error, isLoading, mutate } = useSWR<TeachingLoadListResponse>(null, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { items: data?.teaching_loads ?? [], isLoading, error, mutate }
}

// createTeachingLoad — STUB, see GREEN commit.
export async function createTeachingLoad(_input: TeachingLoadInput): Promise<TeachingLoad> {
  return {} as TeachingLoad
}

// updateTeachingLoad — STUB, see GREEN commit.
export async function updateTeachingLoad(
  _id: number,
  _input: TeachingLoadInput
): Promise<TeachingLoad> {
  return {} as TeachingLoad
}

// deleteTeachingLoad — STUB, see GREEN commit.
export async function deleteTeachingLoad(_id: number): Promise<void> {
  return
}
