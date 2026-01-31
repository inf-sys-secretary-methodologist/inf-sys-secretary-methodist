'use client'

import { useState, useCallback, useEffect } from 'react'
import useSWR, { mutate } from 'swr'
import { apiClient } from '@/lib/api'
import {
  isPushSupported,
  getPermissionStatus,
  subscribeToPush,
  unsubscribeFromPush,
  isSubscribed as checkIsSubscribed,
  sendTestNotification,
  deleteSubscription,
  WebPushStatus,
  WebPushSubscription,
} from '@/lib/push-notifications'

const PUSH_STATUS_URL = '/api/notifications/push/status'

// Fetcher for SWR
const fetcher = async <T>(url: string): Promise<T> => {
  return await apiClient.get<T>(url)
}

/**
 * Hook to manage Web Push notification subscription status
 */
export function usePushNotifications() {
  const [isSupported] = useState(() => isPushSupported())
  const [permission, setPermission] = useState<NotificationPermission | 'unsupported'>(() =>
    getPermissionStatus()
  )
  const [isLocallySubscribed, setIsLocallySubscribed] = useState<boolean | null>(null)
  const [isSubscribing, setIsSubscribing] = useState(false)
  const [isUnsubscribing, setIsUnsubscribing] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  // Fetch server-side push status
  const {
    data: serverStatus,
    error: fetchError,
    isLoading,
    mutate: revalidate,
  } = useSWR<WebPushStatus>(isSupported ? PUSH_STATUS_URL : null, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000,
  })

  // Check local subscription status on mount
  useEffect(() => {
    if (isSupported) {
      checkIsSubscribed().then(setIsLocallySubscribed)
    }
  }, [isSupported])

  // Update permission when it changes
  useEffect(() => {
    if (isSupported && 'permissions' in navigator) {
      navigator.permissions
        .query({ name: 'notifications' as PermissionName })
        .then((status) => {
          status.onchange = () => {
            setPermission(Notification.permission)
          }
        })
        .catch(() => {
          // Permissions API not supported for notifications
        })
    }
  }, [isSupported])

  // Subscribe to push notifications
  const subscribe = useCallback(async (): Promise<WebPushSubscription | null> => {
    if (!isSupported) {
      setError(new Error('Push notifications are not supported'))
      return null
    }

    setIsSubscribing(true)
    setError(null)

    try {
      const subscription = await subscribeToPush()
      setPermission(Notification.permission)
      setIsLocallySubscribed(true)
      // Revalidate server status
      await revalidate()
      // Also revalidate notification preferences
      mutate('/api/notifications/preferences')
      return subscription
    } catch (err) {
      const e = err instanceof Error ? err : new Error('Failed to subscribe')
      setError(e)
      setPermission(Notification.permission)
      return null
    } finally {
      setIsSubscribing(false)
    }
  }, [isSupported, revalidate])

  // Unsubscribe from push notifications
  const unsubscribe = useCallback(async (): Promise<void> => {
    if (!isSupported) {
      return
    }

    setIsUnsubscribing(true)
    setError(null)

    try {
      await unsubscribeFromPush()
      setIsLocallySubscribed(false)
      // Revalidate server status
      await revalidate()
      // Also revalidate notification preferences
      mutate('/api/notifications/preferences')
    } catch (err) {
      const e = err instanceof Error ? err : new Error('Failed to unsubscribe')
      setError(e)
      throw e
    } finally {
      setIsUnsubscribing(false)
    }
  }, [isSupported, revalidate])

  // Remove a specific subscription by ID
  const removeSubscription = useCallback(
    async (subscriptionId: number): Promise<void> => {
      try {
        await deleteSubscription(subscriptionId)
        await revalidate()
      } catch (err) {
        const e = err instanceof Error ? err : new Error('Failed to remove subscription')
        setError(e)
        throw e
      }
    },
    [revalidate]
  )

  // Send a test notification
  const testNotification = useCallback(async (title?: string, message?: string): Promise<void> => {
    try {
      await sendTestNotification(title, message)
    } catch (err) {
      const e = err instanceof Error ? err : new Error('Failed to send test notification')
      setError(e)
      throw e
    }
  }, [])

  return {
    // Status
    isSupported,
    permission,
    isEnabled: serverStatus?.is_enabled ?? false,
    isLocallySubscribed,
    subscriptions: serverStatus?.subscriptions ?? [],
    totalDevices: serverStatus?.total_devices ?? 0,

    // Loading states
    isLoading,
    isSubscribing,
    isUnsubscribing,

    // Error handling
    error: error || fetchError,

    // Actions
    subscribe,
    unsubscribe,
    removeSubscription,
    testNotification,
    revalidate,
  }
}

/**
 * Hook to check if push notifications can be enabled
 */
export function useCanEnablePush(): boolean {
  const [canEnable, setCanEnable] = useState(false)

  useEffect(() => {
    const supported = isPushSupported()
    const permission = getPermissionStatus()
    setCanEnable(supported && permission !== 'denied')
  }, [])

  return canEnable
}
