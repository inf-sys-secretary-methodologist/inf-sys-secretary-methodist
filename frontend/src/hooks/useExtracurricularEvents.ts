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

// === Mutations ===

export async function createExtracurricularEvent(
  input: CreateExtracurricularEventInput
): Promise<ExtracurricularEvent> {
  const wrapper = await apiClient.post<ApiResponse<ExtracurricularEvent>>(BASE_URL, input)
  return wrapper.data
}

export async function updateExtracurricularEvent(
  id: number,
  input: UpdateExtracurricularEventInput
): Promise<ExtracurricularEvent> {
  const wrapper = await apiClient.put<ApiResponse<ExtracurricularEvent>>(`${BASE_URL}/${id}`, input)
  return wrapper.data
}

export async function deleteExtracurricularEvent(id: number): Promise<void> {
  await apiClient.delete(`${BASE_URL}/${id}`)
}

export async function registerForExtracurricularEvent(id: number): Promise<void> {
  await apiClient.post(`${BASE_URL}/${id}/register`)
}

export async function unregisterFromExtracurricularEvent(id: number): Promise<void> {
  await apiClient.delete(`${BASE_URL}/${id}/register`)
}

// === Error mapping ===
//
// Translates backend sentinel codes (see mapEventError in
// event_handler.go) to camelCase i18n keys under the
// extracurricular.errors.* namespace. Status-aware fallback for
// codes the backend omits (plain 403 / 404 / 5xx).

const ERROR_CODE_MAP: Record<string, string> = {
  VERSION_CONFLICT: 'versionConflict',
  ALREADY_REGISTERED: 'alreadyRegistered',
  EVENT_FULL: 'eventFull',
  REGISTRATION_CLOSED: 'registrationClosed',
  CANNOT_EDIT: 'cannotEdit',
  INVALID_EVENT: 'invalidEvent',
}

export function pickExtracurricularErrorKey(err: unknown): string {
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
