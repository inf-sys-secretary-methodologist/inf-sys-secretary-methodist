'use client'

import { useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import {
  ArrowLeft,
  BookMarked,
  CalendarClock,
  ClipboardCheck,
  GraduationCap,
  Layers,
  Loader2,
  User,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { useStudentDebt } from '@/hooks/useStudentDebts'
import { useAuthCheck } from '@/hooks/useAuth'
import { canManageStudentDebts } from '@/lib/auth/permissions'
import { ScheduleResitDialog } from '@/components/student-debts/ScheduleResitDialog'
import { RecordResitResultDialog } from '@/components/student-debts/RecordResitResultDialog'
import {
  STATUS_STYLES,
  statusKey,
  controlFormKey,
  resitResultKey,
} from '@/components/student-debts/status'
import { cn } from '@/lib/utils'

// StudentDebtDetailPage — full debt view with the attempt timeline and the
// resit-lifecycle actions. A manager (admin/methodist/secretary) sees
// Schedule resit (from open / commission) or Record result (from
// resit_scheduled); a teacher reads it scoped to their disciplines and a
// student only ever reaches their own via /my. The backend remains the FSM
// source of truth — the UI gates the affordance, the server enforces it.
export default function StudentDebtDetailPage() {
  const params = useParams<{ id: string }>()
  const id = useMemo(() => {
    const raw = params?.id
    const parsed = typeof raw === 'string' ? Number(raw) : NaN
    return Number.isInteger(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const t = useTranslations('studentDebts')

  const [scheduleOpen, setScheduleOpen] = useState(false)
  const [recordOpen, setRecordOpen] = useState(false)

  const enabled = !authLoading && isAuthenticated && id !== null
  const { debt, isLoading: detailLoading, error, mutate } = useStudentDebt(id, { enabled })

  const canManage = canManageStudentDebts(user?.role)

  if (authLoading || !isAuthenticated) {
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
          href="/student-debts"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.back')}
        </Link>

        {detailLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error || !debt ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.notFound')}</p>
          </div>
        ) : (
          (() => {
            const styles = STATUS_STYLES[debt.status]
            const StatusIcon = styles.Icon
            const sKey = statusKey(debt.status)
            const canSchedule =
              canManage && (debt.status === 'open' || debt.status === 'commission')
            const canRecord = canManage && debt.status === 'resit_scheduled'
            // The scheduled (pending) attempt is the last one in the timeline.
            const pendingAttempt = debt.attempts[debt.attempts.length - 1]

            return (
              <>
                <header className="flex flex-wrap items-start justify-between gap-4">
                  <div className="space-y-1">
                    <h1 className="text-2xl font-bold">{debt.student_full_name}</h1>
                    <div
                      className={cn(
                        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
                        styles.bg,
                        styles.text
                      )}
                    >
                      <StatusIcon className="h-3.5 w-3.5" />
                      {t(`detail.statusHint.${sKey}`)}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {canSchedule && (
                      <Button onClick={() => setScheduleOpen(true)}>
                        <CalendarClock className="h-4 w-4 mr-2" />
                        {t('detail.actions.scheduleResit')}
                      </Button>
                    )}
                    {canRecord && pendingAttempt && (
                      <Button onClick={() => setRecordOpen(true)}>
                        <ClipboardCheck className="h-4 w-4 mr-2" />
                        {t('detail.actions.recordResult')}
                      </Button>
                    )}
                  </div>
                </header>

                <section className="grid gap-3 sm:grid-cols-2 rounded-xl border border-border bg-card p-5">
                  <Field
                    icon={<GraduationCap className="h-4 w-4" />}
                    label={t('detail.fields.group')}
                  >
                    {debt.group_name}
                  </Field>
                  <Field
                    icon={<BookMarked className="h-4 w-4" />}
                    label={t('detail.fields.discipline')}
                  >
                    {debt.discipline_name}
                  </Field>
                  <Field icon={<Layers className="h-4 w-4" />} label={t('detail.fields.semester')}>
                    {debt.semester}
                  </Field>
                  <Field icon={<User className="h-4 w-4" />} label={t('detail.fields.controlForm')}>
                    {t(`card.controlForm.${controlFormKey(debt.control_form)}`)}
                  </Field>
                  {debt.source_ref && (
                    <Field
                      icon={<BookMarked className="h-4 w-4" />}
                      label={t('detail.fields.sourceRef')}
                    >
                      {debt.source_ref}
                    </Field>
                  )}
                </section>

                <section className="space-y-3">
                  <h2 className="text-lg font-semibold">{t('detail.attempts.title')}</h2>
                  {debt.attempts.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t('detail.attempts.empty')}</p>
                  ) : (
                    <ul className="space-y-2">
                      {debt.attempts.map((a) => (
                        <li
                          key={a.id}
                          className="rounded-lg border border-border bg-card p-4 text-sm"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <span className="font-medium">
                              {t('detail.attempts.attemptNo', { n: a.attempt_no })}
                              {a.is_commission && (
                                <span className="ml-2 rounded bg-violet-100 dark:bg-violet-950/40 px-1.5 py-0.5 text-xs text-violet-700 dark:text-violet-300">
                                  {t('detail.attempts.commission')}
                                </span>
                              )}
                            </span>
                            <span className="text-muted-foreground">
                              {t('detail.resitResult.' + resitResultKey(a.result))}
                            </span>
                          </div>
                          <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                            <span>
                              {t('detail.attempts.scheduledDate')}:{' '}
                              {new Date(a.scheduled_date).toLocaleDateString()}
                            </span>
                            <span>
                              {t('detail.attempts.examiner')}: {a.examiner}
                            </span>
                            {typeof a.grade === 'number' && (
                              <span>
                                {t('detail.attempts.grade')}: {a.grade}
                              </span>
                            )}
                          </div>
                        </li>
                      ))}
                    </ul>
                  )}
                </section>

                <ScheduleResitDialog
                  debtId={debt.id}
                  open={scheduleOpen}
                  onClose={() => setScheduleOpen(false)}
                  onScheduled={() => mutate()}
                />
                {pendingAttempt && (
                  <RecordResitResultDialog
                    debtId={debt.id}
                    attemptNo={pendingAttempt.attempt_no}
                    open={recordOpen}
                    onClose={() => setRecordOpen(false)}
                    onRecorded={() => mutate()}
                  />
                )}
              </>
            )
          })()
        )}
      </div>
    </AppLayout>
  )
}

function Field({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-0.5">
      <p className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
        {icon}
        {label}
      </p>
      <p className="text-sm font-medium">{children}</p>
    </div>
  )
}
