'use client'

import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { ArrowRight, BookMarked, GraduationCap, Layers } from 'lucide-react'

import type { StudentDebtListItem } from '@/types/studentDebts'
import { cn } from '@/lib/utils'
import { STATUS_STYLES, statusKey } from './status'

interface StudentDebtCardProps {
  debt: StudentDebtListItem
  className?: string
}

// StudentDebtCard — debt-registry list-item summary linking to the detail
// page. Pure presentation. The status pill colour-codes the lifecycle stage
// (open → amber, resit scheduled → sky, commission → violet, passed →
// emerald, failed → rose) so a methodist scanning the registry spots the
// state without reading the label. The list projection carries the resolved
// student / group / discipline names (unlike work programs, which only carry
// a discipline id), so the card shows them directly.
export function StudentDebtCard({ debt, className }: StudentDebtCardProps) {
  const t = useTranslations('studentDebts')
  const styles = STATUS_STYLES[debt.status]
  const Icon = styles.Icon
  const sKey = statusKey(debt.status)

  return (
    <Link
      href={`/student-debts/${debt.id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:shadow-md hover:border-primary/40',
        className
      )}
      aria-label={t('card.openAria', { student: debt.student_full_name })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight group-hover:text-primary line-clamp-2">
          {debt.student_full_name}
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
          <GraduationCap className="h-3.5 w-3.5" />
          {debt.group_name}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <BookMarked className="h-3.5 w-3.5" />
          {debt.discipline_name}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Layers className="h-3.5 w-3.5" />
          {t('card.semester', { n: debt.semester })}
        </span>
        <ArrowRight className="ml-auto h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
      </div>
    </Link>
  )
}
