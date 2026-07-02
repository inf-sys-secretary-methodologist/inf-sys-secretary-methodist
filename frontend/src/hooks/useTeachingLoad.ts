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

// useTeachingLoads returns the load registry, optionally filtered. The key is
// null while disabled so the fetch is skipped until auth resolves.
export function useTeachingLoads(filter?: TeachingLoadFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildTeachingLoadUrl(filter) : null
  const { data, error, isLoading, mutate } = useSWR<TeachingLoadListResponse>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { items: data?.teaching_loads ?? [], isLoading, error, mutate }
}

// createTeachingLoad posts a new load line and returns the created record.
// apiClient does not unwrap the {success,data} envelope, so we read .data.
export async function createTeachingLoad(input: TeachingLoadInput): Promise<TeachingLoad> {
  const response = await apiClient.post<ApiResponse<TeachingLoad>>(TEACHING_LOAD_URL, input)
  return response.data
}

// updateTeachingLoad puts changes to an existing line and returns it.
export async function updateTeachingLoad(
  id: number,
  input: TeachingLoadInput
): Promise<TeachingLoad> {
  const response = await apiClient.put<ApiResponse<TeachingLoad>>(
    `${TEACHING_LOAD_URL}/${id}`,
    input
  )
  return response.data
}

// deleteTeachingLoad removes a load line.
export async function deleteTeachingLoad(id: number): Promise<void> {
  await apiClient.delete(`${TEACHING_LOAD_URL}/${id}`)
}
