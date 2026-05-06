'use client'

import { useEffect, useMemo } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTranslations, useLocale } from 'next-intl'
import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { ArrowLeft, Loader2, CheckCircle2, RotateCcw, Clock } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useMyAssignment } from '@/hooks/useMyAssignments'
import { useAuthCheck } from '@/hooks/useAuth'
import { parseLocalDate } from '@/lib/assignments/dates'
import type { StudentAssignmentView } from '@/types/assignments'

const localeMap = { ru, en: enUS, fr, ar }

// MyAssignmentDetailPage — student-facing detail. Shows assignment
// metadata at the top and a status-specific panel below: pending
// (waiting for grading), graded (score + feedback), returned (return
// reason — UI v0.115.0 will add the resubmit button here).
export default function MyAssignmentDetailPage() {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const parsed = Number(params?.id)
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('myAssignments')
  const locale = useLocale() as keyof typeof localeMap
  const dateLocale = localeMap[locale] ?? enUS

  const { view, isLoading, error } = useMyAssignment(id)

  useEffect(() => {
    if (!authLoading && isAuthenticated && user?.role && user.role !== 'student') {
      router.replace('/forbidden')
    }
  }, [authLoading, isAuthenticated, user, router])

  if (authLoading) {
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
      <div className="max-w-4xl mx-auto space-y-6">
        <Link
          href="/my-assignments"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.backToList')}
        </Link>

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error || !view ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.loadFailed')}</p>
          </div>
        ) : (
          <>
            <header className="space-y-2">
              <h1 className="text-2xl font-bold">{view.title}</h1>
              {view.description && (
                <p className="text-muted-foreground">{view.description}</p>
              )}
              <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted-foreground">
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.subject')}:</dt>
                  <dd>{view.subject}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.group')}:</dt>
                  <dd>{view.group_name}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.maxScore')}:</dt>
                  <dd>{view.max_score}</dd>
                </div>
                {view.due_date && (
                  <div className="flex gap-2">
                    <dt className="font-medium">{t('detail.dueDate')}:</dt>
                    <dd>{format(parseLocalDate(view.due_date), 'd MMM yyyy', { locale: dateLocale })}</dd>
                  </div>
                )}
              </dl>
            </header>

            <StatusPanel view={view} dateLocale={dateLocale} />
          </>
        )}
      </div>
    </AppLayout>
  )
}

interface StatusPanelProps {
  view: StudentAssignmentView
  dateLocale: typeof enUS
}

function StatusPanel({ view, dateLocale }: StatusPanelProps) {
  const t = useTranslations('myAssignments')

  if (view.status === 'pending') {
    return (
      <section className="rounded-xl border border-amber-500/30 bg-amber-500/5 p-6">
        <div className="flex items-start gap-3">
          <Clock className="h-5 w-5 text-amber-600 dark:text-amber-400 mt-0.5" />
          <div>
            <h2 className="font-semibold text-amber-900 dark:text-amber-100">
              {t('detail.pendingTitle')}
            </h2>
            <p className="mt-1 text-sm text-amber-800 dark:text-amber-200">
              {t('detail.pendingDescription')}
            </p>
          </div>
        </div>
      </section>
    )
  }

  if (view.status === 'graded') {
    const gradedLabel = view.graded_at
      ? format(parseLocalDate(view.graded_at), 'd MMM yyyy', { locale: dateLocale })
      : null
    return (
      <section className="rounded-xl border border-emerald-500/30 bg-emerald-500/5 p-6 space-y-3">
        <div className="flex items-start gap-3">
          <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400 mt-0.5" />
          <div className="flex-1">
            <h2 className="font-semibold text-emerald-900 dark:text-emerald-100">
              {t('detail.gradedTitle')}
            </h2>
            <p className="mt-1 text-2xl font-bold text-emerald-900 dark:text-emerald-100">
              {view.grade_value} / {view.max_score}
            </p>
            {gradedLabel && (
              <p className="text-xs text-emerald-700 dark:text-emerald-300 mt-1">
                {t('detail.gradedAt', { date: gradedLabel })}
              </p>
            )}
          </div>
        </div>
        {view.feedback && (
          <div className="rounded-lg bg-emerald-500/10 p-4">
            <p className="text-sm font-medium text-emerald-900 dark:text-emerald-100 mb-1">
              {t('detail.feedbackTitle')}
            </p>
            <p className="text-sm whitespace-pre-line text-emerald-800 dark:text-emerald-200">
              {view.feedback}
            </p>
          </div>
        )}
      </section>
    )
  }

  // Returned.
  const returnedLabel = view.returned_at
    ? format(parseLocalDate(view.returned_at), 'd MMM yyyy', { locale: dateLocale })
    : null
  return (
    <section className="rounded-xl border border-sky-500/30 bg-sky-500/5 p-6 space-y-3">
      <div className="flex items-start gap-3">
        <RotateCcw className="h-5 w-5 text-sky-600 dark:text-sky-400 mt-0.5" />
        <div className="flex-1">
          <h2 className="font-semibold text-sky-900 dark:text-sky-100">
            {t('detail.returnedTitle')}
          </h2>
          {returnedLabel && (
            <p className="text-xs text-sky-700 dark:text-sky-300 mt-1">
              {t('detail.returnedAt', { date: returnedLabel })}
            </p>
          )}
        </div>
      </div>
      {view.return_reason && (
        <div className="rounded-lg bg-sky-500/10 p-4">
          <p className="text-sm font-medium text-sky-900 dark:text-sky-100 mb-1">
            {t('detail.returnReasonTitle')}
          </p>
          <p className="text-sm whitespace-pre-line text-sky-800 dark:text-sky-200">
            {view.return_reason}
          </p>
        </div>
      )}
      <p className="text-xs text-sky-700 dark:text-sky-300 italic">
        {t('detail.resubmitHint')}
      </p>
    </section>
  )
}
