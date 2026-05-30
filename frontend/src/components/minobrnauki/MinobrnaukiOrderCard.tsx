'use client'

import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { Calendar, FileText, Hash } from 'lucide-react'

import type { MinobrnaukiOrderSummary } from '@/types/minobrnaukiOrder'
import { cn } from '@/lib/utils'

interface MinobrnaukiOrderCardProps {
  order: MinobrnaukiOrderSummary
  className?: string
}

// Colour-codes the change scope so a reader scanning the list can tell a
// minor edit from a major (ФГОС-level) one without reading the label.
const SCOPE_STYLES: Record<MinobrnaukiOrderSummary['change_scope'], { bg: string; text: string }> =
  {
    minor: { bg: 'bg-sky-100 dark:bg-sky-950/40', text: 'text-sky-700 dark:text-sky-300' },
    major: { bg: 'bg-amber-100 dark:bg-amber-950/40', text: 'text-amber-700 dark:text-amber-300' },
  }

// MinobrnaukiOrderCard — order list-item summary linking to the detail
// page. The change-scope pill flags regulatory weight so a reader scanning
// the list can tell a minor edit from a major (ФГОС-level) one at a glance.
export function MinobrnaukiOrderCard({ order, className }: MinobrnaukiOrderCardProps) {
  const t = useTranslations('minobrnaukiOrder')
  const scope = SCOPE_STYLES[order.change_scope]

  return (
    <Link
      href={`/minobrnauki-orders/${order.id}`}
      className={cn(
        'group relative block overflow-hidden rounded-xl border border-border bg-card p-5',
        'transition hover:border-primary/40 hover:shadow-md',
        className
      )}
      aria-label={t('card.openAria', { title: order.title })}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-base font-semibold leading-tight line-clamp-2">{order.title}</h3>
        <div
          className={cn(
            'inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
            scope.bg,
            scope.text
          )}
        >
          <FileText className="h-3.5 w-3.5" />
          {t(`card.changeScope.${order.change_scope}`)}
        </div>
      </div>

      {order.summary && (
        <p className="mt-2 text-sm text-muted-foreground line-clamp-2">{order.summary}</p>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1.5">
          <Hash className="h-3.5 w-3.5" />
          {order.order_number}
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Calendar className="h-3.5 w-3.5" />
          {t('card.published', { date: order.published_at })}
        </span>
      </div>
    </Link>
  )
}
