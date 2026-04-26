'use client'

import { Search, X } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export interface FileFilterValues {
  search?: string
  fileType?: string
}

const FILE_TYPE_OPTIONS = ['image', 'documents', 'spreadsheets', 'presentations', 'archives', 'other'] as const

interface FileFiltersProps {
  value: FileFilterValues
  onChange: (next: FileFilterValues) => void
  className?: string
}

export function FileFilters({ value, onChange, className }: FileFiltersProps) {
  const t = useTranslations('files')

  const update = (patch: Partial<FileFilterValues>) => {
    const next: FileFilterValues = { ...value, ...patch }
    ;(Object.keys(next) as (keyof FileFilterValues)[]).forEach((key) => {
      const v = next[key]
      if (v === undefined || v === '') delete next[key]
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
          htmlFor="file-filter-search"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('search')}
        </label>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            id="file-filter-search"
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
          htmlFor="file-filter-type"
          className="block text-xs font-medium text-muted-foreground mb-1"
        >
          {t('filters.type')}
        </label>
        <select
          id="file-filter-type"
          aria-label={t('filters.type')}
          value={value.fileType ?? ''}
          onChange={(e) => update({ fileType: e.target.value || undefined })}
          className={selectClass}
        >
          <option value="">{t('filters.allTypes')}</option>
          {FILE_TYPE_OPTIONS.map((type) => (
            <option key={type} value={type}>
              {t(`filters.${type}`)}
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
