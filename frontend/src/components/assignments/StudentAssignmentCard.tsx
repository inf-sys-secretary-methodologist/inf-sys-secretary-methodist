'use client'

import Link from 'next/link'
import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { useLocale, useTranslations } from 'next-intl'
import { Calendar, BookOpen, Users, ArrowRight, CheckCircle2, Clock, RotateCcw } from 'lucide-react'

import type { StudentAssignmentView, SubmissionStatus } from '@/types/assignments'
import { cn } from '@/lib/utils'
import { parseLocalDate } from '@/lib/assignments/dates'

const localeMap = { ru, en: enUS, fr, ar }

interface StudentAssignmentCardProps {
  view: StudentAssignmentView
  className?: string
}

// StudentAssignmentCard — list-item summary for the student "My
// Assignments" surface. Mirrors AssignmentCard's structure but renders
// a status pill (pending / graded / returned) and a status-specific
// secondary label (grade fraction or return-reason snippet) instead of
// the teacher max-score chip. Pure presentation; date parsing keeps to
// local midnight semantics (CLAUDE.md #9).
export function StudentAssignmentCard({ view, className }: StudentAssignmentCardProps) {
  const t = useTranslations('myAssignments')
  const locale = useLocale() as keyof typeof localeMap
  const dateLocale = localeMap[locale] ?? enUS

  const dueLabel = view.due_date
    ? format(parseLocalDate(view.due_date), 'd MMM yyyy', { locale: dateLocale })
    : null

  return (
    <Link
      href={`/my-assignments/${view.assignment_id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:shadow-md hover:border-primary/40',
        className
      )}
      aria-label={t('card.openAria', { title: view.title })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight group-hover:text-primary">
          {view.title}
        </h3>
        <ArrowRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
      </div>

      {view.description && (
        <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{view.description}</p>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <BookOpen className="h-3.5 w-3.5" />
          {view.subject}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Users className="h-3.5 w-3.5" />
          {view.group_name}
        </span>
        {dueLabel && (
          <span className="inline-flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            {dueLabel}
          </span>
        )}
        <StatusPill view={view} />
      </div>

      <SecondaryLine view={view} />
    </Link>
  )
}

const STATUS_STYLE: Record<SubmissionStatus, string> = {
  pending: 'bg-amber-500/10 text-amber-700 dark:text-amber-300',
  graded: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  returned: 'bg-sky-500/10 text-sky-700 dark:text-sky-300',
}

const STATUS_ICON: Record<SubmissionStatus, typeof Clock> = {
  pending: Clock,
  graded: CheckCircle2,
  returned: RotateCcw,
}

function StatusPill({ view }: { view: StudentAssignmentView }) {
  const t = useTranslations('myAssignments')
  const Icon = STATUS_ICON[view.status]
  return (
    <span
      className={cn(
        'ml-auto inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 font-medium',
        STATUS_STYLE[view.status]
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {t(`status.${view.status}`)}
    </span>
  )
}

function SecondaryLine({ view }: { view: StudentAssignmentView }) {
  if (view.status === 'graded' && view.grade_value != null) {
    return (
      <p className="mt-3 text-sm font-medium text-emerald-700 dark:text-emerald-300">
        {view.grade_value} / {view.max_score}
      </p>
    )
  }
  if (view.status === 'returned' && view.return_reason) {
    return (
      <p className="mt-3 line-clamp-1 text-sm text-sky-700 dark:text-sky-300">
        {view.return_reason}
      </p>
    )
  }
  return null
}
