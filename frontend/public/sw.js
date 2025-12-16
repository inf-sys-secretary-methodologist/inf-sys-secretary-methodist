const CACHE_NAME = 'sm-is-cache-v1'
const OFFLINE_URL = '/offline'

// Static assets to cache immediately on install
const STATIC_ASSETS = [
  '/',
  '/offline',
  '/manifest.webmanifest',
  '/icons/icon-192x192.png',
  '/icons/icon-512x512.png',
]

// Install event - cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    (async () => {
      const cache = await caches.open(CACHE_NAME)
      // Cache static assets, but don't fail install if some fail
      await Promise.allSettled(
        STATIC_ASSETS.map((url) =>
          cache.add(url).catch((err) => {
            console.warn(`Failed to cache ${url}:`, err)
          })
        )
      )
      // Activate immediately
      await self.skipWaiting()
    })()
  )
})

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    (async () => {
      // Clean up old caches
      const cacheNames = await caches.keys()
      await Promise.all(
        cacheNames.filter((name) => name !== CACHE_NAME).map((name) => caches.delete(name))
      )
      // Take control of all pages immediately
      await self.clients.claim()
    })()
  )
})

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
  const { request } = event

  // Skip non-GET requests
  if (request.method !== 'GET') {
    return
  }

  // Skip cross-origin requests
  if (!request.url.startsWith(self.location.origin)) {
    return
  }

  // Skip API requests - always fetch from network
  if (request.url.includes('/api/')) {
    return
  }

  // Skip Chrome extension requests
  if (request.url.startsWith('chrome-extension://')) {
    return
  }

  event.respondWith(
    (async () => {
      const cache = await caches.open(CACHE_NAME)

      // Try to get from cache first for navigation requests
      if (request.mode === 'navigate') {
        try {
          // Try network first for navigation
          const networkResponse = await fetch(request)
          // Cache successful responses
          if (networkResponse.ok) {
            cache.put(request, networkResponse.clone())
          }
          return networkResponse
        } catch (error) {
          // If offline, try cache
          const cachedResponse = await cache.match(request)
          if (cachedResponse) {
            return cachedResponse
          }
          // Return offline page as last resort
          const offlineResponse = await cache.match(OFFLINE_URL)
          if (offlineResponse) {
            return offlineResponse
          }
          throw error
        }
      }

      // For other requests (assets), try cache first
      const cachedResponse = await cache.match(request)
      if (cachedResponse) {
        // Refresh cache in background
        fetch(request)
          .then((response) => {
            if (response.ok) {
              cache.put(request, response)
            }
          })
          .catch(() => {})
        return cachedResponse
      }

      // If not in cache, fetch from network
      try {
        const networkResponse = await fetch(request)
        // Cache successful responses for static assets
        if (networkResponse.ok && isStaticAsset(request.url)) {
          cache.put(request, networkResponse.clone())
        }
        return networkResponse
      } catch (error) {
        // Return offline fallback for failed requests
        console.warn('Fetch failed:', error)
        throw error
      }
    })()
  )
})

// Handle push notifications
self.addEventListener('push', (event) => {
  if (!event.data) return

  try {
    const data = event.data.json()
    const options = {
      body: data.body || data.message,
      icon: data.icon || '/icons/icon-192x192.png',
      badge: '/icons/icon-72x72.png',
      vibrate: [100, 50, 100],
      tag: data.tag || 'notification',
      renotify: true,
      requireInteraction: data.requireInteraction || false,
      data: {
        url: data.url || '/',
        dateOfArrival: Date.now(),
        ...data.data,
      },
      actions: data.actions || [],
    }

    event.waitUntil(self.registration.showNotification(data.title, options))
  } catch (error) {
    console.error('Error showing notification:', error)
  }
})

// Handle notification clicks
self.addEventListener('notificationclick', (event) => {
  event.notification.close()

  const urlToOpen = event.notification.data?.url || '/'

  event.waitUntil(
    (async () => {
      // Check if there's already a window open
      const windowClients = await self.clients.matchAll({
        type: 'window',
        includeUncontrolled: true,
      })

      // Find existing window with our origin
      for (const client of windowClients) {
        if (client.url.startsWith(self.location.origin) && 'focus' in client) {
          await client.focus()
          if ('navigate' in client) {
            await client.navigate(urlToOpen)
          }
          return
        }
      }

      // Open new window if none exists
      await self.clients.openWindow(urlToOpen)
    })()
  )
})

// Handle notification close
self.addEventListener('notificationclose', () => {
  // Analytics or cleanup can be done here
})

// Helper function to check if URL is a static asset
function isStaticAsset(url) {
  const staticExtensions = [
    '.js',
    '.css',
    '.png',
    '.jpg',
    '.jpeg',
    '.gif',
    '.svg',
    '.ico',
    '.woff',
    '.woff2',
    '.ttf',
    '.eot',
  ]
  return staticExtensions.some((ext) => url.includes(ext))
}

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-pending-actions') {
    event.waitUntil(syncPendingActions())
  }
})

async function syncPendingActions() {
  // Get pending actions from IndexedDB and sync them
  // This can be implemented later for offline-first functionality
}

// Periodic background sync (if supported)
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'check-notifications') {
    event.waitUntil(checkForNewNotifications())
  }
})

async function checkForNewNotifications() {
  // Check for new notifications periodically
}
