'use client'

import Link from 'next/link'
import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { useLocale, useTranslations } from 'next-intl'
import { Calendar, Users, BookOpen, ArrowRight } from 'lucide-react'

import type { Assignment } from '@/types/assignments'
import { cn } from '@/lib/utils'
import { parseLocalDate } from '@/lib/assignments/dates'

const localeMap = { ru, en: enUS, fr, ar }

interface AssignmentCardProps {
  assignment: Assignment
  className?: string
}

// AssignmentCard — list-item summary linking to the detail / grading
// page. Pure presentation; date parsing keeps to local midnight semantics
// so the rendered string matches the teacher's wall clock regardless of
// browser timezone.
export function AssignmentCard({ assignment, className }: AssignmentCardProps) {
  const t = useTranslations('assignments')
  const locale = useLocale() as keyof typeof localeMap
  const dateLocale = localeMap[locale] ?? enUS

  // Parse with parseLocalDate so the calendar date matches the
  // teacher's wall clock regardless of browser timezone (CLAUDE.md #9).
  const dueLabel = assignment.due_date
    ? format(parseLocalDate(assignment.due_date), 'd MMM yyyy', { locale: dateLocale })
    : null

  return (
    <Link
      href={`/assignments/${assignment.id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:shadow-md hover:border-primary/40',
        className
      )}
      aria-label={t('card.openAria', { title: assignment.title })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight group-hover:text-primary">
          {assignment.title}
        </h3>
        <ArrowRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
      </div>

      {assignment.description && (
        <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{assignment.description}</p>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <BookOpen className="h-3.5 w-3.5" />
          {assignment.subject}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Users className="h-3.5 w-3.5" />
          {assignment.group_name}
        </span>
        {dueLabel && (
          <span className="inline-flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            {dueLabel}
          </span>
        )}
        <span className="ml-auto inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 font-medium text-primary">
          {t('card.maxScoreLabel', { max: assignment.max_score })}
        </span>
      </div>
    </Link>
  )
}
