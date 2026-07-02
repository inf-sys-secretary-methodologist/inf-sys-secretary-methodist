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

// createCalendarSubscription creates the subscription if none exists. Callers
// revalidate via the hook's mutate; the response body is intentionally not
// returned (apiClient does not unwrap the {success,data} envelope for POST).
export async function createCalendarSubscription(): Promise<void> {
  await apiClient.post(CALENDAR_SUBSCRIPTION_URL)
}

// rotateCalendarSubscription issues a new secret URL, invalidating the old one.
export async function rotateCalendarSubscription(): Promise<void> {
  await apiClient.post(`${CALENDAR_SUBSCRIPTION_URL}/rotate`)
}

// deleteCalendarSubscription disables the feed.
export async function deleteCalendarSubscription(): Promise<void> {
  await apiClient.delete(CALENDAR_SUBSCRIPTION_URL)
}
