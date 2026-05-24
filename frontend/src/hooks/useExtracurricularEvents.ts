'use client'

// Extracurricular events SWR hooks + mutation actions.
// Mirrors useAnnouncements pattern; talks to /api/v1/extracurricular/events
// (see internal/modules/extracurricular/interfaces/http/handlers).

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  ExtracurricularEvent,
  ExtracurricularEventListResponse,
  ExtracurricularEventFilterParams,
  CreateExtracurricularEventInput,
  UpdateExtracurricularEventInput,
} from '@/types/extracurricular'

const BASE_URL = '/api/v1/extracurricular/events'

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

// useExtracurricularEvents fetches a paginated, audience-filtered list.
// Server applies CanViewEvent matrix per actor role — pass filter
// params to narrow further (status/category/date range/organizer).
export function useExtracurricularEvents(params?: ExtracurricularEventFilterParams) {
  const searchParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null) return
      searchParams.append(key, String(value))
    })
  }

  const queryString = searchParams.toString()
  const url = queryString ? `${BASE_URL}?${queryString}` : BASE_URL

  const { data, error, isLoading, mutate } = useSWR<ExtracurricularEventListResponse>(
    url,
    fetcher,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.SHORT,
    }
  )

  return {
    events: data?.items || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate,
  }
}

// useExtracurricularEvent fetches one event by id. Pass null to skip.
export function useExtracurricularEvent(id: number | null) {
  const { data, error, isLoading, mutate } = useSWR<ExtracurricularEvent>(
    id ? `${BASE_URL}/${id}` : null,
    fetcher
  )

  return {
    event: data,
    isLoading,
    error,
    mutate,
  }
}

// === Mutations (Pair 2 — stubs until next GREEN) ===

const NOT_IMPL = new Error('extracurricular mutation stub — not implemented')

export async function createExtracurricularEvent(
  _input: CreateExtracurricularEventInput
): Promise<ExtracurricularEvent> {
  void _input
  throw NOT_IMPL
}

export async function updateExtracurricularEvent(
  _id: number,
  _input: UpdateExtracurricularEventInput
): Promise<ExtracurricularEvent> {
  void _id
  void _input
  throw NOT_IMPL
}

export async function deleteExtracurricularEvent(_id: number): Promise<void> {
  void _id
  throw NOT_IMPL
}

export async function registerForExtracurricularEvent(_id: number): Promise<void> {
  void _id
  throw NOT_IMPL
}

export async function unregisterFromExtracurricularEvent(_id: number): Promise<void> {
  void _id
  throw NOT_IMPL
}

// Pair 2 stub — table-driven mapping in next GREEN.
export function pickExtracurricularErrorKey(_err: unknown): string {
  void _err
  return 'unstubbed'
}
