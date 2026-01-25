'use client'

import { useEffect } from 'react'
import { useTranslations } from 'next-intl'

export function ServiceWorkerRegistration() {
  const t = useTranslations('pwa')
  const updateMessage = t('updateAvailable')

  useEffect(() => {
    if (typeof window !== 'undefined' && 'serviceWorker' in navigator) {
      registerServiceWorker(updateMessage)
    }
  }, [updateMessage])

  return null
}

async function registerServiceWorker(updateMessage: string) {
  try {
    const registration = await navigator.serviceWorker.register('/sw.js', {
      scope: '/',
      updateViaCache: 'none',
    })

    // Check for updates periodically
    setInterval(
      () => {
        registration.update()
      },
      60 * 60 * 1000
    ) // Check every hour

    // Handle updates
    registration.addEventListener('updatefound', () => {
      const newWorker = registration.installing
      if (!newWorker) return

      newWorker.addEventListener('statechange', () => {
        if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
          // New content is available, show update prompt
          if (window.confirm(updateMessage)) {
            newWorker.postMessage({ type: 'SKIP_WAITING' })
            window.location.reload()
          }
        }
      })
    })

    /* c8 ignore start - Service Worker controller change handler */
    // Handle controller change (when new SW takes over)
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      // Optionally reload when new service worker takes control
    })
    /* c8 ignore stop */

    // Service Worker registered successfully
  } catch (error) {
    console.error('Service Worker registration failed:', error)
  }
}
