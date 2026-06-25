'use client'

import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { AlertTriangle, ChevronRight } from 'lucide-react'

import { useMyStudentDebts } from '@/hooks/useStudentDebts'
import { STATUS_STYLES, statusKey } from '@/components/student-debts/status'

const WIDGET_LIMIT = 5

// MyStudentDebtsWidget — dashboard panel surfacing the caller's active
// academic debts (open / resit_scheduled / commission — anything not yet
// closed). Mirrors UpcomingEventsWidget. Meant for the student dashboard;
// a staff user has no own-debts, so the widget simply renders its empty
// state for them. Each row links to the debt detail.
export function MyStudentDebtsWidget() {
  const t = useTranslations('studentDebts')
  const router = useRouter()
  const { items, isLoading } = useMyStudentDebts({ limit: WIDGET_LIMIT })

  // Closed debts are resolved — only the active ones are worth surfacing.
  const active = items
    .filter((d) => d.status !== 'closed_passed' && d.status !== 'closed_failed')
    .slice(0, WIDGET_LIMIT)

  return (
    <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 p-4 sm:p-6">
      <div className="flex items-center gap-2 mb-4">
        <AlertTriangle className="h-5 w-5 text-amber-500" />
        <h3 className="text-base font-semibold">{t('widget.title')}</h3>
      </div>

      {isLoading ? (
        <div className="space-y-2 animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4" />
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3" />
        </div>
      ) : active.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t('widget.empty')}</p>
      ) : (
        <ul className="space-y-2">
          {active.map((debt) => {
            const styles = STATUS_STYLES[debt.status]
            const Icon = styles.Icon
            return (
              <li key={debt.id}>
                <button
                  onClick={() => router.push(`/student-debts/${debt.id}`)}
                  className="w-full text-left flex items-center justify-between gap-3 px-2 py-1.5 rounded-md hover:bg-muted/50 transition-colors"
                >
                  <span className="text-sm font-medium truncate">{debt.discipline_name}</span>
                  <span
                    className={`text-xs inline-flex items-center gap-1 shrink-0 rounded-full px-2 py-0.5 ${styles.bg} ${styles.text}`}
                  >
                    <Icon className="h-3 w-3" />
                    {t(`card.status.${statusKey(debt.status)}`)}
                  </span>
                </button>
              </li>
            )
          })}
        </ul>
      )}

      {active.length > 0 && (
        <button
          onClick={() => router.push('/student-debts/my')}
          className="mt-3 inline-flex items-center gap-1 text-sm text-primary hover:underline"
        >
          {t('widget.more')}
          <ChevronRight className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
