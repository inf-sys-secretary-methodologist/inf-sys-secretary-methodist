'use client'

import * as React from 'react'
import { format, addMonths, subMonths, addWeeks, subWeeks, addDays, subDays } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, Plus } from 'lucide-react'
import { useTranslations, useLocale } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { CalendarView } from '@/types/calendar'

const localeMap = { ru, en: enUS, fr, ar }

interface CalendarHeaderProps {
  currentDate: Date
  view: CalendarView
  onDateChange: (date: Date) => void
  onViewChange: (view: CalendarView) => void
  onAddEvent?: () => void
  className?: string
}

export function CalendarHeader({
  currentDate,
  view,
  onDateChange,
  onViewChange,
  onAddEvent,
  className,
}: CalendarHeaderProps) {
  const t = useTranslations('calendarView')
  const locale = useLocale()
  /* c8 ignore next */
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS

  const handlePrevious = () => {
    switch (view) {
      case 'month':
        onDateChange(subMonths(currentDate, 1))
        break
      case 'week':
        onDateChange(subWeeks(currentDate, 1))
        break
      case 'day':
        onDateChange(subDays(currentDate, 1))
        break
    }
  }

  const handleNext = () => {
    switch (view) {
      case 'month':
        onDateChange(addMonths(currentDate, 1))
        break
      case 'week':
        onDateChange(addWeeks(currentDate, 1))
        break
      case 'day':
        onDateChange(addDays(currentDate, 1))
        break
    }
  }

  const handleToday = () => {
    onDateChange(new Date())
  }

  const getHeaderTitle = () => {
    switch (view) {
      case 'month':
        return format(currentDate, 'LLLL yyyy', { locale: dateLocale })
      case 'week':
        return `${t('weekNumber')} ${format(currentDate, 'w')}, ${format(currentDate, 'LLLL yyyy', { locale: dateLocale })}`
      case 'day':
        return format(currentDate, 'd MMMM yyyy', { locale: dateLocale })
    }
  }

  return (
    <div
      className={cn(
        'flex flex-col space-y-4 p-4 md:flex-row md:items-center md:justify-between md:space-y-0',
        className
      )}
    >
      {/* Left side - Date display */}
      <div className="flex items-center gap-4">
        <div className="flex w-14 md:w-16 flex-col items-center justify-center rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-100 dark:bg-white/10 p-0.5">
          <span className="p-1 text-xs uppercase text-gray-500 dark:text-gray-400">
            {format(currentDate, 'MMM', { locale: dateLocale })}
          </span>
          <div className="flex w-full items-center justify-center rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95 p-0.5 text-lg font-bold text-gray-900 dark:text-white">
            <span>{format(currentDate, 'd')}</span>
          </div>
        </div>
        <div className="flex flex-col">
          <h2 className="text-lg font-semibold capitalize text-gray-900 dark:text-white">
            {getHeaderTitle()}
          </h2>
          <p className="hidden md:block text-sm text-gray-600 dark:text-gray-400">
            {view === 'month' && format(currentDate, 'd MMMM - ', { locale: dateLocale })}
            {view === 'month' &&
              format(addMonths(currentDate, 1), 'd MMMM yyyy', { locale: dateLocale })}
          </p>
        </div>
      </div>

      {/* Right side - Controls */}
      <div className="flex flex-col items-center gap-4 md:flex-row md:gap-6">
        {/* View Tabs */}
        <Tabs
          value={view}
          onValueChange={(v) => onViewChange(v as CalendarView)}
          className="hidden sm:block"
        >
          <TabsList>
            <TabsTrigger value="month">{t('month')}</TabsTrigger>
            <TabsTrigger value="week">{t('week')}</TabsTrigger>
            <TabsTrigger value="day">{t('day')}</TabsTrigger>
          </TabsList>
        </Tabs>

        <Separator orientation="vertical" className="hidden h-6 md:block" />

        {/* Navigation */}
        <div className="inline-flex w-full -space-x-px rounded-lg shadow-sm shadow-black/5 md:w-auto rtl:space-x-reverse">
          <Button
            onClick={handlePrevious}
            className="rounded-none shadow-none first:rounded-s-lg last:rounded-e-lg focus-visible:z-10"
            variant="outline"
            size="icon"
            aria-label={t('previous')}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button
            onClick={handleToday}
            className="w-full rounded-none shadow-none first:rounded-s-lg last:rounded-e-lg focus-visible:z-10 md:w-auto"
            variant="outline"
          >
            <CalendarIcon className="mr-2 h-4 w-4 md:hidden" />
            {t('today')}
          </Button>
          <Button
            onClick={handleNext}
            className="rounded-none shadow-none first:rounded-s-lg last:rounded-e-lg focus-visible:z-10"
            variant="outline"
            size="icon"
            aria-label={t('next')}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>

        <Separator orientation="vertical" className="hidden h-6 md:block" />
        <Separator orientation="horizontal" className="block w-full md:hidden" />

        {/* Add Event Button */}
        {onAddEvent && (
          <Button onClick={onAddEvent} className="w-full gap-2 md:w-auto">
            <Plus className="h-4 w-4" />
            <span>{t('newEvent')}</span>
          </Button>
        )}
      </div>
    </div>
  )
}
