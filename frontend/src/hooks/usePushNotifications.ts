'use client'

import { useState, useCallback, useEffect } from 'react'
import useSWR, { mutate } from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
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

type PushStatusResult = { available: true; status: WebPushStatus } | { available: false }

// Web push routes are only mounted when the server has VAPID keys configured;
// without them the whole /push group 404s (or a guarded endpoint 503s). That is an
// expected "feature disabled" signal, not an error — degrade gracefully so the
// browser console stays clean and SWR does not retry. If VAPID is configured later
// the endpoint returns 200 and this transparently starts reporting real status.
const pushStatusFetcher = async (url: string): Promise<PushStatusResult> => {
  try {
    const status = await apiClient.get<WebPushStatus>(url)
    return { available: true, status }
  } catch (err) {
    const httpStatus = (err as { response?: { status?: number } })?.response?.status
    if (httpStatus === 404 || httpStatus === 503) {
      return { available: false }
    }
    throw err
  }
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
    data,
    error: fetchError,
    isLoading,
    mutate: revalidate,
  } = useSWR<PushStatusResult>(isSupported ? PUSH_STATUS_URL : null, pushStatusFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.LONG,
  })

  // Web push is "available" unless the server explicitly reported it unconfigured.
  // While loading (data undefined) assume available to avoid a flash of the
  // unavailable state.
  const isAvailable = data ? data.available : true
  const serverStatus = data?.available ? data.status : undefined

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
      setIsSubscribing(false)
      return subscription
    } catch (err) {
      const e = err instanceof Error ? err : new Error('Failed to subscribe')
      setError(e)
      setPermission(Notification.permission)
      setIsSubscribing(false)
      return null
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
    isAvailable,
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
