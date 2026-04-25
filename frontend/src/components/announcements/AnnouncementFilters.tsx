'use client'

import { Search, X } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  ANNOUNCEMENT_STATUSES,
  ANNOUNCEMENT_PRIORITIES,
  TARGET_AUDIENCES,
  type AnnouncementFilterParams,
  type AnnouncementStatus,
  type AnnouncementPriority,
  type TargetAudience,
} from '@/types/announcements'

interface AnnouncementFiltersProps {
  value: AnnouncementFilterParams
  onChange: (next: AnnouncementFilterParams) => void
  className?: string
}

export function AnnouncementFilters({ value, onChange, className }: AnnouncementFiltersProps) {
  const t = useTranslations('announcements')

  const update = (patch: Partial<AnnouncementFilterParams>) => {
    const next: AnnouncementFilterParams = { ...value, ...patch }
    ;(Object.keys(next) as (keyof AnnouncementFilterParams)[]).forEach((key) => {
      const v = next[key]
      if (v === undefined || v === '' || v === null || v === false) delete next[key]
    })
    onChange(next)
  }

  const selectClass =
    'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring'

  return (
    <div
      className={cn(
        'flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap',
        className
      )}
    >
      <div className="flex-1 min-w-[200px]">
        <label
          htmlFor="ann-filter-search"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('search')}
        </label>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            id="ann-filter-search"
            type="search"
            placeholder={t('searchPlaceholder')}
            value={value.search ?? ''}
            onChange={(e) => update({ search: e.target.value || undefined })}
            className="pl-9"
          />
        </div>
      </div>

      <div className="w-full sm:w-40">
        <label
          htmlFor="ann-filter-status"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('status.label')}
        </label>
        <select
          id="ann-filter-status"
          value={value.status ?? ''}
          onChange={(e) =>
            update({ status: (e.target.value || undefined) as AnnouncementStatus | undefined })
          }
          className={selectClass}
        >
          <option value="">{t('status.all')}</option>
          {ANNOUNCEMENT_STATUSES.map((s) => (
            <option key={s} value={s}>
              {t(`status.${s}`)}
            </option>
          ))}
        </select>
      </div>

      <div className="w-full sm:w-40">
        <label
          htmlFor="ann-filter-priority"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('priority.label')}
        </label>
        <select
          id="ann-filter-priority"
          value={value.priority ?? ''}
          onChange={(e) =>
            update({ priority: (e.target.value || undefined) as AnnouncementPriority | undefined })
          }
          className={selectClass}
        >
          <option value="">{t('priority.all')}</option>
          {ANNOUNCEMENT_PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {t(`priority.${p}`)}
            </option>
          ))}
        </select>
      </div>

      <div className="w-full sm:w-40">
        <label
          htmlFor="ann-filter-audience"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('audience.label')}
        </label>
        <select
          id="ann-filter-audience"
          value={value.target_audience ?? ''}
          onChange={(e) =>
            update({
              target_audience: (e.target.value || undefined) as TargetAudience | undefined,
            })
          }
          className={selectClass}
        >
          <option value="">{t('audience.all')}</option>
          {TARGET_AUDIENCES.map((a) => (
            <option key={a} value={a}>
              {t(`audience.${a}`)}
            </option>
          ))}
        </select>
      </div>

      <label className="flex items-center gap-2 text-sm text-foreground sm:self-end h-10">
        <input
          type="checkbox"
          checked={value.is_pinned ?? false}
          onChange={(e) => update({ is_pinned: e.target.checked || undefined })}
          className="h-4 w-4 rounded border-input"
        />
        <span>{t('pinnedOnly')}</span>
      </label>

      <Button
        type="button"
        variant="outline"
        size="default"
        onClick={() => onChange({})}
        className="sm:self-end"
      >
        <X className="h-4 w-4 mr-1" />
        {t('reset')}
      </Button>
    </div>
  )
}
