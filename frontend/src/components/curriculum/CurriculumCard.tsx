'use client'

import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { ArrowRight, BookMarked, Calendar, GraduationCap } from 'lucide-react'

import type { Curriculum } from '@/types/curriculum'
import { cn } from '@/lib/utils'
import { STATUS_STYLES, statusKey } from './status'

interface CurriculumCardProps {
  curriculum: Curriculum
  className?: string
}

// CurriculumCard — list-item summary linking to the detail / edit page
// (v0.119.0). Pure presentation. The status pill colour-codes the
// lifecycle stage so a reader scanning the list (academic secretary
// looking for own drafts, methodist scanning the approval queue) can
// spot pending items without reading the label.
export function CurriculumCard({ curriculum, className }: CurriculumCardProps) {
  const t = useTranslations('curriculum')
  const styles = STATUS_STYLES[curriculum.status]
  const Icon = styles.Icon
  const sKey = statusKey(curriculum.status)

  return (
    <Link
      href={`/curriculum/${curriculum.id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:shadow-md hover:border-primary/40',
        className
      )}
      aria-label={t('card.openAria', { title: curriculum.title })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight group-hover:text-primary line-clamp-2">
          {curriculum.title}
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

      {curriculum.description && (
        <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{curriculum.description}</p>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <BookMarked className="h-3.5 w-3.5" />
          {curriculum.code}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <GraduationCap className="h-3.5 w-3.5" />
          {curriculum.specialty}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Calendar className="h-3.5 w-3.5" />
          {curriculum.year}
        </span>
        <ArrowRight className="ml-auto h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
      </div>
    </Link>
  )
}
