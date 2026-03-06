'use client'

import useSWR, { mutate } from 'swr'
import { apiClient } from '@/lib/api'
import { swrFetcher } from '@/lib/api/fetchers'
import { SWR_DEDUPING, SWR_REFRESH } from '@/config/swr'
import { useState, useCallback } from 'react'
import type {
  Notification,
  NotificationListInput,
  NotificationListOutput,
  UnreadCountOutput,
  NotificationStatsOutput,
  NotificationPreferences,
  PreferencesInput,
  QuietHoursInput,
  NotificationChannel,
} from '@/types/notification'

const NOTIFICATIONS_BASE_URL = '/api/notifications'

// Build URL with query params
function buildNotificationsUrl(input?: NotificationListInput): string {
  const params = new URLSearchParams()
  if (input?.type) params.append('type', input.type)
  if (input?.priority) params.append('priority', input.priority)
  if (input?.is_read !== undefined) params.append('is_read', String(input.is_read))
  if (input?.limit) params.append('limit', String(input.limit))
  if (input?.offset) params.append('offset', String(input.offset))

  const query = params.toString()
  return `${NOTIFICATIONS_BASE_URL}${query ? `?${query}` : ''}`
}

// Notification list hook
export function useNotifications(input?: NotificationListInput) {
  const url = buildNotificationsUrl(input)

  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<NotificationListOutput>(url, swrFetcher<NotificationListOutput>, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.LONG,
    refreshInterval: SWR_REFRESH.STANDARD,
  })

  return {
    data,
    notifications: data?.notifications || [],
    totalCount: data?.total_count || 0,
    unreadCount: data?.unread_count || 0,
    isLoading,
    error,
    mutate: revalidate,
  }
}

// Single notification hook
export function useNotification(id: number) {
  const { data, error, isLoading } = useSWR<Notification>(
    id ? `${NOTIFICATIONS_BASE_URL}/${id}` : null,
    swrFetcher<Notification>
  )

  return { notification: data, isLoading, error }
}

// Unread count hook
export function useUnreadCount() {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<UnreadCountOutput>(
    `${NOTIFICATIONS_BASE_URL}/unread-count`,
    swrFetcher<UnreadCountOutput>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.NOTIFICATIONS,
      refreshInterval: SWR_REFRESH.REALTIME,
    }
  )

  return {
    data,
    count: data?.count || 0,
    isLoading,
    error,
    mutate: revalidate,
  }
}

// Notification stats hook
export function useNotificationStats() {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<NotificationStatsOutput>(
    `${NOTIFICATIONS_BASE_URL}/stats`,
    swrFetcher<NotificationStatsOutput>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.EXTRA_LONG,
    }
  )

  return { stats: data, isLoading, error, mutate: revalidate }
}

// Mark as read mutation hook
export function useMarkAsRead() {
  const [isPending, setIsPending] = useState(false)

  const markAsRead = useCallback(async (id: number) => {
    setIsPending(true)
    try {
      await apiClient.put(`${NOTIFICATIONS_BASE_URL}/${id}/read`)
      // Revalidate all notification-related caches
      /* c8 ignore next 3 -- SWR cache key matcher callback */
      mutate((key) => typeof key === 'string' && key.includes('/notifications'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: markAsRead, isPending }
}

// Mark all as read mutation hook
export function useMarkAllAsRead() {
  const [isPending, setIsPending] = useState(false)

  const markAllAsRead = useCallback(async () => {
    setIsPending(true)
    try {
      await apiClient.put(`${NOTIFICATIONS_BASE_URL}/read-all`)
      /* c8 ignore next 3 -- SWR cache key matcher callback */
      mutate((key) => typeof key === 'string' && key.includes('/notifications'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: markAllAsRead, isPending }
}

// Delete notification mutation hook
export function useDeleteNotification() {
  const [isPending, setIsPending] = useState(false)

  const deleteNotification = useCallback(async (id: number) => {
    setIsPending(true)
    try {
      await apiClient.delete(`${NOTIFICATIONS_BASE_URL}/${id}`)
      /* c8 ignore next 3 -- SWR cache key matcher callback */
      mutate((key) => typeof key === 'string' && key.includes('/notifications'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: deleteNotification, isPending }
}

// Delete all notifications mutation hook
export function useDeleteAllNotifications() {
  const [isPending, setIsPending] = useState(false)

  const deleteAll = useCallback(async () => {
    setIsPending(true)
    try {
      await apiClient.delete(NOTIFICATIONS_BASE_URL)
      /* c8 ignore next 3 -- SWR cache key matcher callback */
      mutate((key) => typeof key === 'string' && key.includes('/notifications'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: deleteAll, isPending }
}

// Preferences hooks
export function useNotificationPreferences() {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<NotificationPreferences>(
    `${NOTIFICATIONS_BASE_URL}/preferences`,
    swrFetcher<NotificationPreferences>
  )

  return { data, isLoading, error, mutate: revalidate }
}

export function useUpdatePreferences() {
  const [isPending, setIsPending] = useState(false)

  const updatePreferences = useCallback(async (input: PreferencesInput) => {
    setIsPending(true)
    try {
      await apiClient.put(`${NOTIFICATIONS_BASE_URL}/preferences`, input)
      mutate(`${NOTIFICATIONS_BASE_URL}/preferences`)
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: updatePreferences, isPending }
}

export function useToggleChannel() {
  const [isPending, setIsPending] = useState(false)

  const toggleChannel = useCallback(
    async ({ channel, enabled }: { channel: string; enabled: boolean }) => {
      setIsPending(true)
      try {
        await apiClient.put(`${NOTIFICATIONS_BASE_URL}/preferences/channel`, {
          channel: channel as NotificationChannel,
          enabled,
        })
        mutate(`${NOTIFICATIONS_BASE_URL}/preferences`)
      } finally {
        setIsPending(false)
      }
    },
    []
  )

  return { mutateAsync: toggleChannel, isPending }
}

export function useUpdateQuietHours() {
  const [isPending, setIsPending] = useState(false)

  const updateQuietHours = useCallback(async (input: QuietHoursInput) => {
    setIsPending(true)
    try {
      await apiClient.put(`${NOTIFICATIONS_BASE_URL}/preferences/quiet-hours`, input)
      mutate(`${NOTIFICATIONS_BASE_URL}/preferences`)
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: updateQuietHours, isPending }
}

export function useResetPreferences() {
  const [isPending, setIsPending] = useState(false)

  const resetPreferences = useCallback(async () => {
    setIsPending(true)
    try {
      await apiClient.post(`${NOTIFICATIONS_BASE_URL}/preferences/reset`)
      mutate(`${NOTIFICATIONS_BASE_URL}/preferences`)
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: resetPreferences, isPending }
}

export function useTimezones() {
  const { data, error, isLoading } = useSWR<{ timezones: string[] }>(
    `${NOTIFICATIONS_BASE_URL}/timezones`,
    swrFetcher<{ timezones: string[] }>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.NONE,
    }
  )

  return { data, isLoading, error }
}
