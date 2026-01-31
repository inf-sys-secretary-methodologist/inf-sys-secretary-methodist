import { apiClient } from './api'

// Types for Push Notifications
export interface WebPushSubscription {
  id: number
  device_name?: string
  user_agent?: string
  is_active: boolean
  last_used_at?: string
  created_at: string
}

export interface WebPushStatus {
  is_enabled: boolean
  subscriptions: WebPushSubscription[]
  total_devices: number
}

export interface WebPushVAPIDKey {
  public_key: string
}

// Check if push notifications are supported
export function isPushSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    'serviceWorker' in navigator &&
    'PushManager' in window &&
    'Notification' in window
  )
}

// Get current permission status
export function getPermissionStatus(): NotificationPermission | 'unsupported' {
  if (!isPushSupported()) {
    return 'unsupported'
  }
  return Notification.permission
}

// Request notification permission
export async function requestPermission(): Promise<NotificationPermission> {
  if (!isPushSupported()) {
    throw new Error('Push notifications are not supported')
  }

  const permission = await Notification.requestPermission()
  return permission
}

// Get the service worker registration
async function getServiceWorkerRegistration(): Promise<ServiceWorkerRegistration> {
  if (!('serviceWorker' in navigator)) {
    throw new Error('Service Worker is not supported')
  }

  // Wait for the service worker to be ready
  const registration = await navigator.serviceWorker.ready
  return registration
}

// Get VAPID public key from the server
async function getVAPIDKey(): Promise<string> {
  const response = await apiClient.get<WebPushVAPIDKey>('/api/notifications/push/vapid-key')
  return response.public_key
}

// Convert VAPID key from base64 to Uint8Array
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')

  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}

// Get device name from user agent
function getDeviceName(): string {
  const ua = navigator.userAgent
  if (/iPhone/.test(ua)) return 'iPhone'
  if (/iPad/.test(ua)) return 'iPad'
  if (/Android/.test(ua)) return 'Android'
  if (/Windows/.test(ua)) return 'Windows'
  if (/Mac/.test(ua)) return 'Mac'
  if (/Linux/.test(ua)) return 'Linux'
  return 'Unknown Device'
}

// Subscribe to push notifications
export async function subscribeToPush(): Promise<WebPushSubscription> {
  if (!isPushSupported()) {
    throw new Error('Push notifications are not supported')
  }

  // Check permission
  const permission = await requestPermission()
  if (permission !== 'granted') {
    throw new Error('Notification permission denied')
  }

  // Get service worker registration
  const registration = await getServiceWorkerRegistration()

  // Get VAPID public key
  const vapidKey = await getVAPIDKey()

  // Subscribe to push
  const applicationServerKey = urlBase64ToUint8Array(vapidKey)
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: applicationServerKey.buffer as ArrayBuffer,
  })

  // Extract keys from subscription
  const key = subscription.getKey('p256dh')
  const auth = subscription.getKey('auth')

  if (!key || !auth) {
    throw new Error('Failed to get subscription keys')
  }

  // Convert keys to base64
  const p256dh = btoa(String.fromCharCode(...new Uint8Array(key)))
  const authKey = btoa(String.fromCharCode(...new Uint8Array(auth)))

  // Send subscription to server
  const response = await apiClient.post<WebPushSubscription>('/api/notifications/push/subscribe', {
    endpoint: subscription.endpoint,
    p256dh: p256dh,
    auth: authKey,
    user_agent: navigator.userAgent,
    device_name: getDeviceName(),
  })

  return response
}

// Unsubscribe from push notifications
export async function unsubscribeFromPush(): Promise<void> {
  if (!isPushSupported()) {
    throw new Error('Push notifications are not supported')
  }

  const registration = await getServiceWorkerRegistration()
  const subscription = await registration.pushManager.getSubscription()

  if (subscription) {
    // Unsubscribe locally
    await subscription.unsubscribe()

    // Notify server
    await apiClient.post('/api/notifications/push/unsubscribe', {
      endpoint: subscription.endpoint,
    })
  }
}

// Get current subscription status
export async function getCurrentSubscription(): Promise<PushSubscription | null> {
  if (!isPushSupported()) {
    return null
  }

  try {
    const registration = await getServiceWorkerRegistration()
    return await registration.pushManager.getSubscription()
  } catch {
    return null
  }
}

// Check if currently subscribed
export async function isSubscribed(): Promise<boolean> {
  const subscription = await getCurrentSubscription()
  return subscription !== null
}

// Get push status from server
export async function getPushStatus(): Promise<WebPushStatus> {
  return await apiClient.get<WebPushStatus>('/api/notifications/push/status')
}

// Delete a specific subscription by ID
export async function deleteSubscription(subscriptionId: number): Promise<void> {
  await apiClient.delete(`/api/notifications/push/subscriptions/${subscriptionId}`)
}

// Send a test notification
export async function sendTestNotification(title?: string, message?: string): Promise<void> {
  await apiClient.post('/api/notifications/push/test', {
    title: title || 'Test Notification',
    message: message || 'This is a test push notification.',
  })
}
