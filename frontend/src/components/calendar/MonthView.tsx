'use client'

import * as React from 'react'
import {
  format,
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  isSameMonth,
  isSameDay,
  isToday,
  getDay,
} from 'date-fns'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { EventCard } from './EventCard'
import type { CalendarEvent } from '@/types/calendar'

interface MonthViewProps {
  currentDate: Date
  selectedDate?: Date
  events: CalendarEvent[]
  onDateSelect?: (date: Date) => void
  onEventClick?: (event: CalendarEvent) => void
  className?: string
}

const WEEKDAY_KEYS = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'] as const

const colStartClasses = [
  '',
  'col-start-2',
  'col-start-3',
  'col-start-4',
  'col-start-5',
  'col-start-6',
  'col-start-7',
]

export function MonthView({
  currentDate,
  selectedDate,
  events,
  onDateSelect,
  onEventClick,
  className,
}: MonthViewProps) {
  const t = useTranslations('calendarView')
  const monthStart = startOfMonth(currentDate)
  const monthEnd = endOfMonth(currentDate)
  const calendarStart = startOfWeek(monthStart, { weekStartsOn: 0 })
  const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 0 })

  const days = eachDayOfInterval({
    start: calendarStart,
    end: calendarEnd,
  })

  const getEventsForDay = (day: Date) => {
    return events.filter((event) => {
      const eventDate = new Date(event.start_time)
      return isSameDay(eventDate, day)
    })
  }

  return (
    <div className={cn('flex flex-1 flex-col', className)}>
      {/* Weekday Headers */}
      <div className="grid grid-cols-7 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-white/5 text-center text-xs font-semibold">
        {WEEKDAY_KEYS.map((key, index) => (
          <div
            key={key}
            className={cn(
              'py-2.5 border-r border-gray-200 dark:border-gray-700 last:border-r-0',
              index === 0 || index === 6
                ? 'text-gray-500 dark:text-gray-400'
                : 'text-gray-900 dark:text-white'
            )}
          >
            {t(`weekdays.${key}`)}
          </div>
        ))}
      </div>

      {/* Calendar Grid */}
      <div className="flex-1 grid grid-cols-7 border-l border-gray-200 dark:border-gray-700">
        {days.map((day, dayIdx) => {
          const dayEvents = getEventsForDay(day)
          const isCurrentMonth = isSameMonth(day, currentDate)
          const isSelected = selectedDate && isSameDay(day, selectedDate)
          const isTodayDate = isToday(day)

          return (
            <div
              key={day.toISOString()}
              onClick={() => onDateSelect?.(day)}
              className={cn(
                'relative flex min-h-[100px] flex-col border-b border-r border-gray-200 dark:border-gray-700 p-1 cursor-pointer transition-colors',
                dayIdx === 0 && colStartClasses[getDay(day)],
                !isCurrentMonth && 'bg-gray-50 dark:bg-white/5 text-gray-400 dark:text-gray-500',
                isSelected && 'bg-gray-100 dark:bg-white/10',
                !isSelected && 'hover:bg-gray-50 dark:hover:bg-white/5'
              )}
            >
              {/* Day Number */}
              <div className="flex items-center justify-between p-1">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDateSelect?.(day)
                  }}
                  className={cn(
                    'flex h-7 w-7 items-center justify-center rounded-full text-sm font-medium transition-colors',
                    isTodayDate && 'bg-gray-900 dark:bg-white text-white dark:text-gray-900',
                    isSelected &&
                      !isTodayDate &&
                      'bg-gray-700 dark:bg-gray-300 text-white dark:text-gray-900',
                    !isTodayDate &&
                      !isSelected &&
                      'hover:bg-gray-100 dark:hover:bg-white/10 text-gray-900 dark:text-white'
                  )}
                >
                  {format(day, 'd')}
                </button>
              </div>

              {/* Events */}
              <div className="flex-1 space-y-0.5 overflow-hidden px-1">
                {dayEvents.slice(0, 3).map((event) => (
                  <EventCard
                    key={event.id}
                    event={event}
                    variant="compact"
                    onClick={() => onEventClick?.(event)}
                  />
                ))}
                {/* c8 ignore start - More events button click handler */}
                {dayEvents.length > 3 && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onDateSelect?.(day)
                    }}
                    className="w-full text-left text-xs text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-2"
                  >
                    {t('moreEvents', { count: dayEvents.length - 3 })}
                  </button>
                )}
                {/* c8 ignore stop */}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
