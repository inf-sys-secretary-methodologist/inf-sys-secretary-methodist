'use client'

import { useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations, useLocale } from 'next-intl'
import { Calendar as CalendarIcon, Loader2 } from 'lucide-react'
import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'

import { AppLayout } from '@/components/layout'
import { ExtracurricularEventCard } from '@/components/extracurricular/ExtracurricularEventCard'
import { useExtracurricularEvents } from '@/hooks/useExtracurricularEvents'
import { useAuthCheck } from '@/hooks/useAuth'
import type { ExtracurricularEventSummary } from '@/types/extracurricular'

const localeMap = { ru, en: enUS, fr, ar }

// Calendar view shows published events grouped by start-month for
// easy chronological scan. Drafts / canceled / completed live in
// the list view (extracurricular/page.tsx).
function groupByMonth(events: ExtracurricularEventSummary[]) {
  const map = new Map<string, ExtracurricularEventSummary[]>()
  for (const e of events) {
    const key = e.start_at.slice(0, 7) // "YYYY-MM"
    const bucket = map.get(key) ?? []
    bucket.push(e)
    map.set(key, bucket)
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, items]) => ({
      key,
      // Sort events inside month by start_at ascending.
      items: items.slice().sort((a, b) => a.start_at.localeCompare(b.start_at)),
    }))
}

export default function ExtracurricularCalendarPage() {
  const t = useTranslations('extracurricular')
  const locale = useLocale()
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS
  useAuthCheck()
  const router = useRouter()
  const { events, isLoading, error } = useExtracurricularEvents({ status: 'published' })
  const groups = useMemo(() => groupByMonth(events), [events])

  return (
    <AppLayout>
      <div className="mx-auto max-w-5xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <CalendarIcon className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('calendar')}</h1>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('loadFailed')}</p>
          </div>
        ) : groups.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <CalendarIcon className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <p className="text-lg font-medium">{t('empty')}</p>
          </div>
        ) : (
          <div className="space-y-6">
            {groups.map((group) => (
              <section key={group.key}>
                <h2 className="text-lg font-semibold mb-3 capitalize">
                  {format(new Date(`${group.key}-01T00:00:00`), 'LLLL yyyy', {
                    locale: dateLocale,
                  })}
                </h2>
                <div className="grid gap-3">
                  {group.items.map((event) => (
                    <ExtracurricularEventCard
                      key={event.id}
                      event={event}
                      onClick={() => router.push(`/extracurricular/${event.id}`)}
                    />
                  ))}
                </div>
              </section>
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  )
}
