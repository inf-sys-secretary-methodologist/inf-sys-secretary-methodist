'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { Curriculum, CurriculumListResponse, CurriculumListFilter } from '@/types/curriculum'

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
