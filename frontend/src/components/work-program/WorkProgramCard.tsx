'use client'

import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { ArrowRight, BookMarked, Calendar, GraduationCap } from 'lucide-react'

import type { WorkProgramSummary } from '@/types/workProgram'
import { cn } from '@/lib/utils'
import { STATUS_STYLES, statusKey } from './status'

interface WorkProgramCardProps {
  workProgram: WorkProgramSummary
  className?: string
}

// WorkProgramCard — РПД list-item summary linking to the detail page.
// Pure presentation. The status pill colour-codes the lifecycle stage so
// a reader scanning the list (teacher looking for own drafts, methodist
// scanning pending items, student browsing approved РПД) can spot state
// without reading the label. The discipline shows as an id reference
// because the list projection carries no discipline name (joined only on
// the detail view).
export function WorkProgramCard({ workProgram, className }: WorkProgramCardProps) {
  const t = useTranslations('workProgram')
  const styles = STATUS_STYLES[workProgram.status]
  const Icon = styles.Icon
  const sKey = statusKey(workProgram.status)

  return (
    <Link
      href={`/work-programs/${workProgram.id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:shadow-md hover:border-primary/40',
        className
      )}
      aria-label={t('card.openAria', { title: workProgram.title })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight group-hover:text-primary line-clamp-2">
          {workProgram.title}
        </h3>
        <div
          className={cn(
            'inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
            styles.bg,
            styles.text
          )}
        >
          <Icon className="h-3.5 w-3.5" />
          {t(`card.status.${sKey}`)}
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <BookMarked className="h-3.5 w-3.5" />
          {t('card.discipline', { id: workProgram.discipline_id })}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <GraduationCap className="h-3.5 w-3.5" />
          {workProgram.specialty_code}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Calendar className="h-3.5 w-3.5" />
          {workProgram.applicable_from_year}
        </span>
        <ArrowRight className="ml-auto h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
      </div>
    </Link>
  )
}
