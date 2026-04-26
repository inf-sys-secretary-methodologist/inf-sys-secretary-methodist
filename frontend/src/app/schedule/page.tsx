'use client'

import { useTranslations } from 'next-intl'
import { Calendar, Construction } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

export default function SchedulePage() {
  const t = useTranslations('schedule')
  useAuthCheck()

  return (
    <AppLayout>
      <div className="mx-auto max-w-6xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <Calendar className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('title')}</h1>
        </div>

        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-muted-foreground/25 py-16">
          <Construction className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-lg font-medium">{t('comingSoon')}</p>
          <p className="text-sm text-muted-foreground mt-1">{t('comingSoonDescription')}</p>
        </div>
      </div>
    </AppLayout>
  )
}
