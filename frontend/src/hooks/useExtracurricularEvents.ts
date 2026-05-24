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

// Stub implementations — replaced in GREEN commit. RED commit ships
// these to keep eslint happy (resolvable imports) while tests fail
// on assertion (per project pre-commit-friendly RED pattern).
const NOT_IMPL = new Error('extracurricular hook stub — not implemented')

export function useExtracurricularEvents(_params?: ExtracurricularEventFilterParams) {
  void _params
  void useSWR
  void apiClient
  void SWR_DEDUPING
  return {
    events: [] as ExtracurricularEventListResponse['items'],
    total: 0,
    isLoading: false,
    error: undefined as Error | undefined,
    mutate: async () => undefined,
  }
}

export function useExtracurricularEvent(_id: number | null) {
  void _id
  return {
    event: undefined as ExtracurricularEvent | undefined,
    isLoading: false,
    error: undefined as Error | undefined,
    mutate: async () => undefined,
  }
}

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
