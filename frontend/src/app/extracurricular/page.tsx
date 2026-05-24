'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Calendar, Loader2, Plus } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { ExtracurricularEventCard } from '@/components/extracurricular/ExtracurricularEventCard'
import { ExtracurricularEventFilters } from '@/components/extracurricular/ExtracurricularEventFilters'
import { useExtracurricularEvents } from '@/hooks/useExtracurricularEvents'
import type { ExtracurricularEventFilterParams } from '@/types/extracurricular'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'

// Edit roles per CLAUDE.md role matrix: all except student can author/edit
// events. Backend authz mirror — see ext authz.go AuthorizeEventCreate.
function canCreateEvents(role: string | undefined): boolean {
  return role !== undefined && role !== 'student'
}

export default function ExtracurricularEventsPage() {
  const t = useTranslations('extracurricular')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const userCanCreate = canCreateEvents(user?.role)
  const router = useRouter()

  const [filters, setFilters] = useState<ExtracurricularEventFilterParams>({})
  const { events, isLoading, error } = useExtracurricularEvents(filters)

  return (
    <AppLayout>
      <div className="mx-auto max-w-6xl space-y-6 p-4 md:p-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-2">
            <Calendar className="h-6 w-6" />
            <div>
              <h1 className="text-2xl font-bold">{t('title')}</h1>
              <p className="text-sm text-muted-foreground">{t('description')}</p>
            </div>
          </div>
          {userCanCreate && (
            <Button onClick={() => router.push('/extracurricular/new')}>
              <Plus className="h-4 w-4 mr-2" />
              {t('create')}
            </Button>
          )}
        </div>

        <ExtracurricularEventFilters value={filters} onChange={setFilters} />

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('loadFailed')}</p>
          </div>
        ) : events.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Calendar className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <p className="text-lg font-medium">{t('empty')}</p>
          </div>
        ) : (
          <div className="grid gap-3">
            {events.map((event) => (
              <ExtracurricularEventCard
                key={event.id}
                event={event}
                onClick={() => router.push(`/extracurricular/${event.id}`)}
              />
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  )
}
