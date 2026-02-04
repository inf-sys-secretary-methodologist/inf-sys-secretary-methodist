'use client'

import { ReactNode } from 'react'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { getAvailableNavEntries } from '@/config/navigation'
import { useRouteAnnouncer } from '@/hooks/useRouteAnnouncer'
import { AppHeader } from './AppHeader'
import { InstallPrompt } from '@/components/pwa/install-prompt'
import { SkipToContent } from '@/components/ui/skip-to-content'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const t = useTranslations('common')
  const { user, isLoading } = useAuthCheck()

  // Announce route changes for screen readers
  useRouteAnnouncer()

  // Get navigation entries (items and groups) filtered by user role
  const navEntries = getAvailableNavEntries(user?.role)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">{t('loading')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      <SkipToContent />
      <AppHeader entries={navEntries} />
      <main
        id="main-content"
        tabIndex={-1}
        className="max-w-7xl mx-auto px-4 py-6 sm:px-6 sm:py-8 lg:px-8 pb-16 focus:outline-none"
      >
        {children}
      </main>
      <InstallPrompt />
    </div>
  )
}
