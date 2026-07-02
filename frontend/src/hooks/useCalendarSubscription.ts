'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { swrFetcher } from '@/lib/api/fetchers'
import { SWR_DEDUPING } from '@/config/swr'
import type { CalendarSubscription } from '@/types/calendarFeed'

export const CALENDAR_SUBSCRIPTION_URL = '/api/schedule/calendar-subscription'

// useCalendarSubscription loads the current user's calendar feed subscription.
export function useCalendarSubscription() {
  const { data, error, isLoading, mutate } = useSWR<CalendarSubscription>(
    CALENDAR_SUBSCRIPTION_URL,
    swrFetcher<CalendarSubscription>,
    { revalidateOnFocus: false, dedupingInterval: SWR_DEDUPING.LONG }
  )
  return { subscription: data, isLoading, error, mutate }
}

// createCalendarSubscription creates (or returns the existing) subscription.
export async function createCalendarSubscription(): Promise<CalendarSubscription> {
  return apiClient.post<CalendarSubscription>(CALENDAR_SUBSCRIPTION_URL)
}

// rotateCalendarSubscription issues a new secret URL, invalidating the old one.
export async function rotateCalendarSubscription(): Promise<CalendarSubscription> {
  return apiClient.post<CalendarSubscription>(`${CALENDAR_SUBSCRIPTION_URL}/rotate`)
}

// deleteCalendarSubscription disables the feed.
export async function deleteCalendarSubscription(): Promise<void> {
  await apiClient.delete(CALENDAR_SUBSCRIPTION_URL)
}
