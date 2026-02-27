'use client'

import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { getAvailableNavEntries } from '@/config/navigation'
import { AppHeader } from '@/components/layout/AppHeader'
import { SkipToContent } from '@/components/ui/skip-to-content'
import { AIAssistantCard } from '@/components/ai'

export default function AIPage() {
  const { user, isLoading } = useAuthCheck()
  const t = useTranslations('common')

  // Get navigation entries filtered by user role
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
    <div className="h-screen flex flex-col overflow-hidden">
      <SkipToContent />
      <AppHeader entries={navEntries} />
      <main
        id="main-content"
        tabIndex={-1}
        className="flex-1 flex overflow-hidden focus:outline-none"
      >
        <AIAssistantCard className="flex-1" showHistory={true} />
      </main>
    </div>
  )
}
