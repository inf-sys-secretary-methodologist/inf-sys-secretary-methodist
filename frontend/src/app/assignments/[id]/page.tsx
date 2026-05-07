'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { ArrowLeft, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAssignment, useSubmissions } from '@/hooks/useAssignments'
import { useAuthCheck } from '@/hooks/useAuth'
import { SubmissionRow } from '@/components/assignments/SubmissionRow'
import type { SubmissionStatus } from '@/types/assignments'
import { SUBMISSION_STATUSES } from '@/types/assignments'
import { cn } from '@/lib/utils'

// AssignmentDetailPage — single-assignment read with grading list.
// Same student-redirect guard as the index page.
export default function AssignmentDetailPage() {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const parsed = Number(params?.id)
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('assignments')

  const [statusFilter, setStatusFilter] = useState<SubmissionStatus | 'all'>('all')

  const { assignment, isLoading: assignmentLoading, error: assignmentError } = useAssignment(id)

  const {
    items,
    isLoading: submissionsLoading,
    mutate: mutateSubmissions,
  } = useSubmissions(id, statusFilter === 'all' ? undefined : statusFilter)

  useEffect(() => {
    if (!authLoading && isAuthenticated && user?.role === 'student') {
      router.replace('/forbidden')
    }
  }, [authLoading, isAuthenticated, user, router])

  if (authLoading || !isAuthenticated || id == null) {
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
      <div className="max-w-5xl mx-auto space-y-6">
        <Link
          href="/assignments"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.backToList')}
        </Link>

        {assignmentLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : assignmentError || !assignment ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.loadFailed')}</p>
          </div>
        ) : (
          <>
            <header className="space-y-2">
              <h1 className="text-2xl font-bold">{assignment.title}</h1>
              {assignment.description && (
                <p className="text-muted-foreground">{assignment.description}</p>
              )}
              <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted-foreground">
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.subject')}:</dt>
                  <dd>{assignment.subject}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.group')}:</dt>
                  <dd>{assignment.group_name}</dd>
                </div>
                <div className="flex gap-2">
                  <dt className="font-medium">{t('detail.maxScore')}:</dt>
                  <dd>{assignment.max_score}</dd>
                </div>
              </dl>
            </header>

            <section>
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2 className="text-lg font-semibold">{t('detail.submissionsTitle')}</h2>
                <div
                  className="flex flex-wrap gap-1.5"
                  role="tablist"
                  aria-label={t('detail.statusFilterAria')}
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
                        {s === 'all' ? t('detail.statusAll') : t(`status.${s}`)}
                      </button>
                    )
                  })}
                </div>
              </div>

              {submissionsLoading ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : items.length === 0 ? (
                <div className="rounded-xl border border-border bg-card p-8 text-center text-sm text-muted-foreground">
                  {t('detail.submissionsEmpty')}
                </div>
              ) : (
                <div className="space-y-3">
                  {items.map((s) => (
                    <SubmissionRow
                      key={s.id}
                      assignmentId={assignment.id}
                      maxScore={assignment.max_score}
                      submission={s}
                      onGraded={mutateSubmissions}
                    />
                  ))}
                </div>
              )}
            </section>
          </>
        )}
      </div>
    </AppLayout>
  )
}
