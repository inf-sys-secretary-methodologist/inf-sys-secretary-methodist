'use client'

import { useEffect, useRef } from 'react'
import { usePathname } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useAnnouncer } from '@/components/ui/screen-reader-announcer'

/**
 * Map of routes to their i18n keys
 */
const routeKeys: Record<string, string> = {
  '/': 'home',
  '/dashboard': 'dashboard',
  '/documents': 'documents',
  '/documents/shared': 'sharedDocuments',
  '/calendar': 'calendar',
  '/notifications': 'notifications',
  '/profile': 'profile',
  '/users': 'users',
  '/settings/appearance': 'appearance',
  '/settings/notifications': 'notificationSettings',
  '/integration': 'integration',
  '/login': 'login',
  '/register': 'register',
}

/**
 * Hook that announces route changes to screen readers.
 * WCAG 4.1.3: Status Messages
 */
export function useRouteAnnouncer() {
  const pathname = usePathname()
  const { announce } = useAnnouncer()
  const t = useTranslations('routeAnnouncer')
  const previousPathname = useRef<string | null>(null)

  useEffect(() => {
    // Skip announcement on initial mount
    if (previousPathname.current === null) {
      previousPathname.current = pathname
      return
    }

    // Don't announce if pathname hasn't changed
    if (previousPathname.current === pathname) {
      return
    }

    previousPathname.current = pathname

    // Get the route key or use fallback
    const routeKey = routeKeys[pathname]
    const routeName = routeKey ? t(`routes.${routeKey}`) : generateRouteName(pathname)

    // Announce the navigation
    announce(t('navigatedTo', { page: routeName }))
  }, [pathname, announce, t])
}

/**
 * Generate a human-readable route name from a pathname
 */
function generateRouteName(pathname: string): string {
  // Remove leading slash and split by /
  const parts = pathname.slice(1).split('/')

  // Take the last meaningful part
  const lastPart = parts[parts.length - 1]

  // Handle dynamic routes (e.g., [id])
  if (lastPart.startsWith('[') && lastPart.endsWith(']')) {
    return 'Details'
  }

  // Convert kebab-case to title
  return lastPart
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}
