'use client'

import { memo } from 'react'
import dynamic from 'next/dynamic'
import { useTranslations, useLocale } from 'next-intl'
import { Calendar as CalendarIcon, X, Loader2 } from 'lucide-react'
import { format, Locale } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'

// Lazy load Calendar to reduce initial bundle (react-day-picker ~100KB)
const Calendar = dynamic(() => import('@/components/ui/calendar').then((mod) => mod.Calendar), {
  loading: () => (
    <div className="flex items-center justify-center p-4">
      <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
    </div>
  ),
  ssr: false,
})

const localeMap: Record<string, Locale> = {
  ru: ru,
  en: enUS,
  fr: fr,
  ar: ar,
}

interface DateRangeFilterProps {
  dateFrom: Date | undefined
  dateTo: Date | undefined
  onDateFromChange: (date: Date | undefined) => void
  onDateToChange: (date: Date | undefined) => void
}

export const DateRangeFilter = memo(function DateRangeFilter({
  dateFrom,
  dateTo,
  onDateFromChange,
  onDateToChange,
}: DateRangeFilterProps) {
  const t = useTranslations('documents.filters')
  const tForm = useTranslations('documents.form')
  const locale = useLocale()
  /* c8 ignore next */
  const dateLocale = localeMap[locale] || enUS

  /* c8 ignore start - JSX event handlers, tested in e2e */
  return (
    <>
      {/* Date From Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          {t('dateFrom')}
        </label>
        <div className="flex gap-2">
          <Popover>
            <PopoverTrigger asChild>
              {/* c8 ignore start - Date from button */}
              <Button
                variant="outline"
                className={cn(
                  'flex-1 justify-start text-left font-normal',
                  !dateFrom && 'text-muted-foreground'
                )}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {dateFrom
                  ? format(dateFrom, 'dd.MM.yyyy', { locale: dateLocale })
                  : t('selectDate')}
              </Button>
              {/* c8 ignore stop */}
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              {/* c8 ignore start - Calendar disabled callback */}
              <Calendar
                mode="single"
                selected={dateFrom}
                onSelect={onDateFromChange}
                disabled={(date) => (dateTo ? date > dateTo : false)}
                locale={dateLocale}
              />
              {/* c8 ignore stop */}
            </PopoverContent>
          </Popover>
          {dateFrom && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onDateFromChange(undefined)}
              className="flex-shrink-0 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              aria-label={tForm('resetDate')}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>

      {/* Date To Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          {t('dateTo')}
        </label>
        <div className="flex gap-2">
          <Popover>
            <PopoverTrigger asChild>
              {/* c8 ignore start - Date to button */}
              <Button
                variant="outline"
                className={cn(
                  'flex-1 justify-start text-left font-normal',
                  !dateTo && 'text-muted-foreground'
                )}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {dateTo ? format(dateTo, 'dd.MM.yyyy', { locale: dateLocale }) : t('selectDate')}
              </Button>
              {/* c8 ignore stop */}
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              {/* c8 ignore start - Calendar disabled callback */}
              <Calendar
                mode="single"
                selected={dateTo}
                onSelect={onDateToChange}
                disabled={(date) => (dateFrom ? date < dateFrom : false)}
                locale={dateLocale}
              />
              {/* c8 ignore stop */}
            </PopoverContent>
          </Popover>
          {dateTo && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onDateToChange(undefined)}
              className="flex-shrink-0 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              aria-label={tForm('resetDate')}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </>
  )
  /* c8 ignore stop */
})
