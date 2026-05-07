'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { useLocale, useTranslations } from 'next-intl'
import { CheckCircle2, Clock, RotateCcw } from 'lucide-react'

import type { SubmissionView } from '@/types/assignments'
import { cn } from '@/lib/utils'
import { parseLocalDate } from '@/lib/assignments/dates'
import { Button } from '@/components/ui/button'
import { GradeForm } from './GradeForm'
import { ReturnDialog } from './ReturnDialog'

const localeMap = { ru, en: enUS, fr, ar }

interface SubmissionRowProps {
  assignmentId: number
  maxScore: number
  submission: SubmissionView
  onGraded?: () => void
}

const STATUS_STYLES: Record<
  SubmissionView['status'],
  { bg: string; text: string; Icon: typeof Clock }
> = {
  pending: {
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    text: 'text-amber-700 dark:text-amber-300',
    Icon: Clock,
  },
  graded: {
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    text: 'text-emerald-700 dark:text-emerald-300',
    Icon: CheckCircle2,
  },
  returned: {
    bg: 'bg-sky-50 dark:bg-sky-950/30',
    text: 'text-sky-700 dark:text-sky-300',
    Icon: RotateCcw,
  },
}

export function SubmissionRow({
  assignmentId,
  maxScore,
  submission,
  onGraded,
}: SubmissionRowProps) {
  const t = useTranslations('assignments')
  const locale = useLocale() as keyof typeof localeMap
  const dateLocale = localeMap[locale] ?? enUS
  const styles = STATUS_STYLES[submission.status]
  const Icon = styles.Icon
  const [dialogOpen, setDialogOpen] = useState(false)
  const canReturn = submission.status === 'pending' || submission.status === 'graded'

  // Parse with parseLocalDate so the date portion matches the teacher's
  // wall clock; for graded_at the time-of-day display is intentionally
  // truncated — the row's purpose is "did this get graded today" not
  // a precision timestamp (CLAUDE.md #9).
  const gradedAt = submission.graded_at
    ? format(parseLocalDate(submission.graded_at), 'd MMM yyyy', { locale: dateLocale })
    : null

  return (
    <article className="rounded-xl border border-border bg-card p-4 sm:p-5">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold">{submission.student_name}</h3>
          <p className="text-xs text-muted-foreground">
            {t('submissionRow.studentIdLabel', { id: submission.student_id })}
          </p>
        </div>

        <div
          className={cn(
            'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
            styles.bg,
            styles.text
          )}
        >
          <Icon className="h-3.5 w-3.5" />
          {t(`status.${submission.status}`)}
        </div>
      </header>

      {submission.status === 'graded' && submission.grade_value != null && (
        <p className="mt-3 text-sm">
          <span className="font-semibold text-emerald-600 dark:text-emerald-400">
            {submission.grade_value} / {maxScore}
          </span>
          {submission.feedback && (
            <span className="ml-2 text-muted-foreground">{submission.feedback}</span>
          )}
          {gradedAt && <span className="ml-2 text-xs text-muted-foreground">{gradedAt}</span>}
        </p>
      )}

      {submission.status === 'returned' && submission.return_reason && (
        <div className="mt-3 rounded-lg bg-sky-50 dark:bg-sky-950/30 p-3 text-sm">
          <p className="font-semibold text-sky-700 dark:text-sky-300">
            {t('submissionRow.returnedReasonLabel')}
          </p>
          <p className="mt-1 text-sky-900 dark:text-sky-100">{submission.return_reason}</p>
          {submission.returned_at && (
            <p className="mt-2 text-xs text-sky-700 dark:text-sky-400">
              {t('submissionRow.returnedAtLabel')}{' '}
              {format(parseLocalDate(submission.returned_at), 'd MMM yyyy', { locale: dateLocale })}
            </p>
          )}
        </div>
      )}

      <div className="mt-4">
        {/*
          Key includes status + grade so React remounts GradeForm when
          a submission flips pending → graded. Without remount the form
          keeps its previous useState(initialValue) and would display
          stale user input after a successful save (the inputs would
          be disabled, but the visible value would not match the
          persisted normalisation).
        */}
        <GradeForm
          key={`${submission.id}:${submission.status}:${submission.grade_value ?? ''}`}
          assignmentId={assignmentId}
          maxScore={maxScore}
          submission={submission}
          onSaved={onGraded}
        />
      </div>

      {canReturn && (
        <div className="mt-3 flex justify-end">
          <Button type="button" variant="outline" size="sm" onClick={() => setDialogOpen(true)}>
            <RotateCcw className="h-4 w-4 mr-2" />
            {t('returnButton')}
          </Button>
        </div>
      )}

      {canReturn && (
        <ReturnDialog
          assignmentId={assignmentId}
          submission={submission}
          open={dialogOpen}
          onClose={() => setDialogOpen(false)}
          onReturned={onGraded}
        />
      )}
    </article>
  )
}
