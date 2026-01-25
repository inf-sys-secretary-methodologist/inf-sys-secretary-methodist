'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import type {
  CalendarEvent,
  EventListResponse,
  EventFilterParams,
  CreateEventInput,
  UpdateEventInput,
} from '@/types/calendar'

const EVENTS_BASE_URL = '/api/events'

// API Response wrapper type from backend
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
  meta?: {
    request_id: string
    timestamp: string
    version: string
  }
}

// Fetcher for SWR - extracts data from wrapped response
const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

// Hook for fetching events with filters
export function useEvents(params?: EventFilterParams) {
  const searchParams = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value))
      }
    })
  }

  const queryString = searchParams.toString()
  const url = queryString ? `${EVENTS_BASE_URL}?${queryString}` : EVENTS_BASE_URL

  const { data, error, isLoading, mutate } = useSWR<EventListResponse>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
  })

  return {
    events: data?.events || [],
    total: data?.total || 0,
    page: data?.page || 1,
    pageSize: data?.page_size || 20,
    totalPages: data?.total_pages || 0,
    isLoading,
    error,
    mutate,
  }
}

// Hook for fetching events by date range
export function useEventsByDateRange(start: Date, end: Date) {
  const startISO = start.toISOString()
  const endISO = end.toISOString()
  const url = `${EVENTS_BASE_URL}/range?start=${encodeURIComponent(startISO)}&end=${encodeURIComponent(endISO)}`

  const { data, error, isLoading, mutate } = useSWR<CalendarEvent[]>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
  })

  return {
    events: data || [],
    isLoading,
    error,
    mutate,
  }
}

// Hook for fetching a single event
export function useEvent(id: number | null) {
  const { data, error, isLoading, mutate } = useSWR<CalendarEvent>(
    id ? `${EVENTS_BASE_URL}/${id}` : null,
    fetcher
  )

  return {
    event: data,
    isLoading,
    error,
    mutate,
  }
}

// Hook for fetching upcoming events
export function useUpcomingEvents(limit = 10) {
  const url = `${EVENTS_BASE_URL}/upcoming?limit=${limit}`

  const { data, error, isLoading, mutate } = useSWR<CalendarEvent[]>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
  })

  return {
    events: data || [],
    isLoading,
    error,
    mutate,
  }
}

// Mutation functions
export async function createEvent(input: CreateEventInput): Promise<CalendarEvent> {
  return apiClient.post<CalendarEvent>(EVENTS_BASE_URL, input)
}

export async function updateEvent(id: number, input: UpdateEventInput): Promise<CalendarEvent> {
  return apiClient.put<CalendarEvent>(`${EVENTS_BASE_URL}/${id}`, input)
}

export async function deleteEvent(id: number): Promise<void> {
  await apiClient.delete(`${EVENTS_BASE_URL}/${id}`)
}

export async function cancelEvent(id: number): Promise<CalendarEvent> {
  return apiClient.post<CalendarEvent>(`${EVENTS_BASE_URL}/${id}/cancel`)
}

export async function rescheduleEvent(
  id: number,
  startTime: string,
  endTime?: string
): Promise<CalendarEvent> {
  return apiClient.post<CalendarEvent>(`${EVENTS_BASE_URL}/${id}/reschedule`, {
    start_time: startTime,
    end_time: endTime,
  })
}

/* c8 ignore start - Calendar operations wrapper */
// Combined hook for calendar operations
export function useCalendarOperations() {
  const create = async (input: CreateEventInput) => {
    return createEvent(input)
  }

  const update = async (id: number, input: UpdateEventInput) => {
    return updateEvent(id, input)
  }

  const remove = async (id: number) => {
    return deleteEvent(id)
  }

  const cancel = async (id: number) => {
    return cancelEvent(id)
  }

  const reschedule = async (id: number, startTime: string, endTime?: string) => {
    return rescheduleEvent(id, startTime, endTime)
  }

  return {
    create,
    update,
    remove,
    cancel,
    reschedule,
  }
}
/* c8 ignore stop */
