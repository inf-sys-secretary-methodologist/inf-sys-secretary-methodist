'use client'

import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { useLocale, useTranslations } from 'next-intl'
import { CheckCircle2, Clock, RotateCcw } from 'lucide-react'

import type { SubmissionView } from '@/types/assignments'
import { cn } from '@/lib/utils'
import { parseLocalDate } from '@/lib/assignments/dates'
import { GradeForm } from './GradeForm'

const localeMap = { ru, en: enUS, fr, ar }

interface SubmissionRowProps {
  assignmentId: number
  maxScore: number
  submission: SubmissionView
  onGraded?: () => void
}

const STATUS_STYLES: Record<SubmissionView['status'], { bg: string; text: string; Icon: typeof Clock }> = {
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

export function SubmissionRow({ assignmentId, maxScore, submission, onGraded }: SubmissionRowProps) {
  const t = useTranslations('assignments')
  const locale = useLocale() as keyof typeof localeMap
  const dateLocale = localeMap[locale] ?? enUS
  const styles = STATUS_STYLES[submission.status]
  const Icon = styles.Icon

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
          {gradedAt && (
            <span className="ml-2 text-xs text-muted-foreground">{gradedAt}</span>
          )}
        </p>
      )}

      <div className="mt-4">
        <GradeForm
          assignmentId={assignmentId}
          maxScore={maxScore}
          submission={submission}
          onSaved={onGraded}
        />
      </div>
    </article>
  )
}
