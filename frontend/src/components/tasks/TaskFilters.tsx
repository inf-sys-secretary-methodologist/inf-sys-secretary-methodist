'use client'

import { Search, X } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  TASK_STATUSES,
  TASK_PRIORITIES,
  type TaskFilterParams,
  type TaskStatus,
  type TaskPriority,
} from '@/types/tasks'

interface TaskFiltersProps {
  value: TaskFilterParams
  onChange: (next: TaskFilterParams) => void
  className?: string
}

export function TaskFilters({ value, onChange, className }: TaskFiltersProps) {
  const t = useTranslations('tasks')

  const update = (patch: Partial<TaskFilterParams>) => {
    const next: TaskFilterParams = { ...value, ...patch }
    // Strip empty values so the URL stays clean.
    ;(Object.keys(next) as (keyof TaskFilterParams)[]).forEach((key) => {
      const v = next[key]
      if (v === undefined || v === '' || v === null) delete next[key]
    })
    onChange(next)
  }

  return (
    <div className={cn('flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap', className)}>
      <div className="flex-1 min-w-[200px]">
        <label
          htmlFor="task-filter-search"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('search')}
        </label>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            id="task-filter-search"
            type="search"
            placeholder={t('searchPlaceholder')}
            value={value.search ?? ''}
            onChange={(e) => update({ search: e.target.value || undefined })}
            className="pl-9"
          />
        </div>
      </div>

      <div className="w-full sm:w-44">
        <label
          htmlFor="task-filter-status"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('status.label')}
        </label>
        <select
          id="task-filter-status"
          value={value.status ?? ''}
          onChange={(e) =>
            update({ status: (e.target.value || undefined) as TaskStatus | undefined })
          }
          className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <option value="">{t('status.all')}</option>
          {TASK_STATUSES.map((s) => (
            <option key={s} value={s}>
              {t(`status.${s}`)}
            </option>
          ))}
        </select>
      </div>

      <div className="w-full sm:w-44">
        <label
          htmlFor="task-filter-priority"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('priority.label')}
        </label>
        <select
          id="task-filter-priority"
          value={value.priority ?? ''}
          onChange={(e) =>
            update({ priority: (e.target.value || undefined) as TaskPriority | undefined })
          }
          className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <option value="">{t('priority.all')}</option>
          {TASK_PRIORITIES.map((p) => (
            <option key={p} value={p}>
              {t(`priority.${p}`)}
            </option>
          ))}
        </select>
      </div>

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
