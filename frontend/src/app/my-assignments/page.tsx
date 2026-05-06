'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { GraduationCap, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { StudentAssignmentCard } from '@/components/assignments/StudentAssignmentCard'
import { useMyAssignments } from '@/hooks/useMyAssignments'
import { useAuthCheck } from '@/hooks/useAuth'
import type { SubmissionStatus } from '@/types/assignments'
import { SUBMISSION_STATUSES } from '@/types/assignments'
import { cn } from '@/lib/utils'

// MyAssignmentsPage — student-facing list view. The /assignments
// surface is gated to non-students by RequireNonStudent on the backend
// AND a /forbidden redirect on the client; this is its mirror — only
// students reach it. Non-student callers (e.g. an admin who navigated
// here on purpose) are bounced to /forbidden client-side; the backend
// itself returns 401 because the GET /api/assignments/my route sits
// behind RequireRole("student").
export default function MyAssignmentsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('myAssignments')

  const [statusFilter, setStatusFilter] = useState<SubmissionStatus | 'all'>('all')

  // Skip the SWR call entirely while role is unknown / not student so
  // the redirect-to-/forbidden window does not waste a 401 round-trip.
  const isStudent = user?.role === 'student'
  const {
    items,
    total,
    isLoading: listLoading,
    error,
  } = useMyAssignments(
    statusFilter === 'all' ? undefined : statusFilter,
    { enabled: isStudent }
  )

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role && user.role !== 'student') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  // Mirror /assignments shape: gate the page body on (auth ready AND
  // student role) so a logged-out user or a non-student bouncing
  // through here never sees the chrome or fires the SWR call. The
  // useEffect above schedules the /forbidden replace; this branch
  // keeps the UI quiet while it lands.
  if (isLoading || !isAuthenticated || (user?.role && user.role !== 'student')) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto space-y-6">
        <header>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground">{t('description')}</p>
        </header>

        <div
          className="flex flex-wrap gap-1.5"
          role="tablist"
          aria-label={t('statusFilterAria')}
        >
          {(['all', ...SUBMISSION_STATUSES] as Array<SubmissionStatus | 'all'>).map((s) => {
            const active = s === statusFilter
            return (
              <button
                key={s}
                type="button"
                role="tab"
                aria-selected={active}
                onClick={() => setStatusFilter(s)}
                className={cn(
                  'rounded-full px-3 py-1 text-xs font-medium transition',
                  active
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground hover:bg-muted/70'
                )}
              >
                {s === 'all' ? t('statusAll') : t(`status.${s}`)}
              </button>
            )
          })}
        </div>

        {listLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <GraduationCap className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((v) => (
              <StudentAssignmentCard key={v.assignment_id} view={v} />
            ))}
          </div>
        )}

        {items.length > 0 && (
          <p className="text-right text-sm text-muted-foreground">
            {t('countLabel', { shown: items.length, total })}
          </p>
        )}
      </div>
    </AppLayout>
  )
}
