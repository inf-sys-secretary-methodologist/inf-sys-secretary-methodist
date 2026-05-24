'use client'

import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import {
  EVENT_STATUSES,
  EVENT_CATEGORIES,
  type ExtracurricularEventFilterParams,
  type EventStatus,
  type EventCategory,
} from '@/types/extracurricular'

export interface ExtracurricularEventFiltersProps {
  value: ExtracurricularEventFilterParams
  onChange: (next: ExtracurricularEventFilterParams) => void
  className?: string
}

const SELECT_CLASS =
  'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring'

export function ExtracurricularEventFilters({
  value,
  onChange,
  className,
}: ExtracurricularEventFiltersProps) {
  const t = useTranslations('extracurricular')

  // Apply a patch then drop any key whose value is empty/undefined —
  // mirrors backend behavior (empty filter = no narrowing).
  const update = (patch: Partial<ExtracurricularEventFilterParams>) => {
    const next: ExtracurricularEventFilterParams = { ...value, ...patch }
    ;(Object.keys(next) as (keyof ExtracurricularEventFilterParams)[]).forEach((key) => {
      const v = next[key]
      if (v === undefined || v === '' || v === null) delete next[key]
    })
    onChange(next)
  }

  return (
    <div className={cn('flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap', className)}>
      <div className="w-full sm:w-44">
        <label
          htmlFor="ext-filter-status"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('status.label')}
        </label>
        <select
          id="ext-filter-status"
          aria-label={t('status.label')}
          value={value.status ?? ''}
          onChange={(e) =>
            update({ status: (e.target.value || undefined) as EventStatus | undefined })
          }
          className={SELECT_CLASS}
        >
          <option value="">{t('status.all')}</option>
          {EVENT_STATUSES.map((s) => (
            <option key={s} value={s}>
              {t(`status.${s}`)}
            </option>
          ))}
        </select>
      </div>

      <div className="w-full sm:w-44">
        <label
          htmlFor="ext-filter-category"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('category.label')}
        </label>
        <select
          id="ext-filter-category"
          aria-label={t('category.label')}
          value={value.category ?? ''}
          onChange={(e) =>
            update({ category: (e.target.value || undefined) as EventCategory | undefined })
          }
          className={SELECT_CLASS}
        >
          <option value="">{t('category.all')}</option>
          {EVENT_CATEGORIES.map((c) => (
            <option key={c} value={c}>
              {t(`category.${c}`)}
            </option>
          ))}
        </select>
      </div>
    </div>
  )
}
