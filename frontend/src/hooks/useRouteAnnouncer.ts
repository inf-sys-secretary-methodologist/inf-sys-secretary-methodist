'use client'

import { useEffect, useRef } from 'react'
import { usePathname } from 'next/navigation'
import { useAnnouncer } from '@/components/ui/screen-reader-announcer'

/**
 * Map of routes to their display names for announcements
 */
const routeNames: Record<string, string> = {
  '/': 'Главная страница',
  '/dashboard': 'Панель управления',
  '/documents': 'Документы',
  '/documents/shared': 'Общие документы',
  '/calendar': 'Календарь',
  '/notifications': 'Уведомления',
  '/profile': 'Профиль',
  '/users': 'Пользователи',
  '/settings/appearance': 'Настройки внешнего вида',
  '/settings/notifications': 'Настройки уведомлений',
  '/integration': 'Интеграция 1С',
  '/login': 'Вход в систему',
  '/register': 'Регистрация',
}

/**
 * Hook that announces route changes to screen readers.
 * WCAG 4.1.3: Status Messages
 */
export function useRouteAnnouncer() {
  const pathname = usePathname()
  const { announce } = useAnnouncer()
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

    // Get the route name or generate from pathname
    const routeName = routeNames[pathname] || generateRouteName(pathname)

    // Announce the navigation
    announce(`Перешли на страницу: ${routeName}`)
  }, [pathname, announce])
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
    return 'Детали'
  }

  // Convert kebab-case to title
  return lastPart
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}
