'use client'

import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Activity, Calendar as CalendarIcon } from 'lucide-react'

import { useExtracurricularEvents } from '@/hooks/useExtracurricularEvents'

const UPCOMING_LIMIT = 5

export function UpcomingEventsWidget() {
  const t = useTranslations('extracurricular')
  const router = useRouter()
  const { events, isLoading } = useExtracurricularEvents({
    status: 'published',
    limit: UPCOMING_LIMIT * 4,
  })

  // Render-time `now` snapshot — React Compiler's purity rule
  // flags Date.now(), but this is the canonical "give me an upper
  // bound for past-event filtering" pattern. Re-renders refresh
  // the bound naturally; the small list is recomputed cheaply.
  // eslint-disable-next-line react-hooks/purity
  const now = Date.now()
  const upcoming = events
    .filter((e) => new Date(e.start_at).getTime() > now)
    .sort((a, b) => a.start_at.localeCompare(b.start_at))
    .slice(0, UPCOMING_LIMIT)

  return (
    <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 p-4 sm:p-6">
      <div className="flex items-center gap-2 mb-4">
        <Activity className="h-5 w-5 text-violet-500" />
        <h3 className="text-base font-semibold">{t('upcoming')}</h3>
      </div>

      {isLoading ? (
        <div className="space-y-2 animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4" />
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3" />
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2" />
        </div>
      ) : upcoming.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t('empty')}</p>
      ) : (
        <ul className="space-y-2">
          {upcoming.map((event) => (
            <li key={event.id}>
              <button
                onClick={() => router.push(`/extracurricular/${event.id}`)}
                className="w-full text-left flex items-center justify-between gap-3 px-2 py-1.5 rounded-md hover:bg-muted/50 transition-colors"
              >
                <span className="text-sm font-medium truncate">{event.title}</span>
                <span className="text-xs text-muted-foreground inline-flex items-center gap-1 shrink-0">
                  <CalendarIcon className="h-3.5 w-3.5" />
                  {new Date(event.start_at).toLocaleDateString()}
                </span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
