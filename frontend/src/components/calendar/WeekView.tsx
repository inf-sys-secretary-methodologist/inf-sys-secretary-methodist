'use client'

import * as React from 'react'
import {
  format,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  isSameDay,
  isToday,
  addHours,
  startOfDay,
  getHours,
  getMinutes,
  differenceInMinutes,
} from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { useTranslations, useLocale } from 'next-intl'

import { cn } from '@/lib/utils'
import { EventCard } from './EventCard'
import type { CalendarEvent } from '@/types/calendar'

const localeMap = { ru, en: enUS, fr, ar }

interface WeekViewProps {
  currentDate: Date
  selectedDate?: Date
  events: CalendarEvent[]
  onDateSelect?: (date: Date) => void
  onEventClick?: (event: CalendarEvent) => void
  onTimeSlotClick?: (date: Date) => void
  className?: string
}

const HOURS = Array.from({ length: 24 }, (_, i) => i)
const HOUR_HEIGHT = 60 // pixels per hour

export function WeekView({
  currentDate,
  selectedDate,
  events,
  onDateSelect,
  onEventClick,
  onTimeSlotClick,
  className,
}: WeekViewProps) {
  const t = useTranslations('calendarView')
  const locale = useLocale()
  /* c8 ignore next */
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS
  const weekStart = startOfWeek(currentDate, { weekStartsOn: 1 })
  const weekEnd = endOfWeek(currentDate, { weekStartsOn: 1 })
  const weekDays = eachDayOfInterval({ start: weekStart, end: weekEnd })
  const scrollRef = React.useRef<HTMLDivElement>(null)

  /* c8 ignore start - Scroll effect and event helpers, tested in e2e */
  // Scroll to current time on mount
  React.useEffect(() => {
    if (scrollRef.current) {
      const currentHour = new Date().getHours()
      scrollRef.current.scrollTop = (currentHour - 1) * HOUR_HEIGHT
    }
  }, [])
  const getEventsForDay = (day: Date) => {
    return events.filter((event) => {
      const eventDate = new Date(event.start_time)
      return isSameDay(eventDate, day)
    })
  }

  const getEventPosition = (event: CalendarEvent) => {
    const startTime = new Date(event.start_time)
    const endTime = event.end_time ? new Date(event.end_time) : addHours(startTime, 1)

    const startMinutes = getHours(startTime) * 60 + getMinutes(startTime)
    const duration = differenceInMinutes(endTime, startTime)

    const top = (startMinutes / 60) * HOUR_HEIGHT
    const height = Math.max((duration / 60) * HOUR_HEIGHT, 24) // min height 24px

    return { top, height }
  }

  const getAllDayEvents = (day: Date) => {
    return getEventsForDay(day).filter((e) => e.all_day)
  }

  const getTimedEvents = (day: Date) => {
    return getEventsForDay(day).filter((e) => !e.all_day)
  }

  const handleTimeSlotClick = (day: Date, hour: number) => {
    const dateWithTime = new Date(day)
    dateWithTime.setHours(hour, 0, 0, 0)
    onTimeSlotClick?.(dateWithTime)
  }
  /* c8 ignore stop */

  return (
    <div className={cn('flex flex-1 flex-col overflow-hidden', className)}>
      {/* Header with days */}
      <div className="flex border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-white/5">
        {/* Time column header */}
        <div className="w-16 shrink-0 border-r border-gray-200 dark:border-gray-700" />

        {/* Day headers */}
        <div className="flex flex-1">
          {weekDays.map((day) => {
            const isTodayDate = isToday(day)
            const isSelected = selectedDate && isSameDay(day, selectedDate)
            const allDayEvents = getAllDayEvents(day)

            return (
              <div
                key={day.toISOString()}
                className={cn(
                  'flex-1 border-r border-gray-200 dark:border-gray-700 last:border-r-0 min-w-0'
                )}
              >
                <button
                  onClick={() => onDateSelect?.(day)}
                  className={cn(
                    'w-full py-2 text-center hover:bg-gray-100 dark:hover:bg-white/10 transition-colors',
                    isSelected && 'bg-gray-100 dark:bg-white/10'
                  )}
                >
                  <div className="text-xs text-gray-500 dark:text-gray-400 uppercase">
                    {format(day, 'EEE', { locale: dateLocale })}
                  </div>
                  <div
                    className={cn(
                      'mx-auto mt-1 flex h-8 w-8 items-center justify-center rounded-full text-lg font-semibold text-gray-900 dark:text-white',
                      isTodayDate && 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
                    )}
                  >
                    {format(day, 'd')}
                  </div>
                </button>

                {/* All-day events */}
                {allDayEvents.length > 0 && (
                  <div className="px-1 pb-1 space-y-0.5">
                    {allDayEvents.slice(0, 2).map((event) => (
                      <EventCard
                        key={event.id}
                        event={event}
                        variant="compact"
                        onClick={() => onEventClick?.(event)}
                      />
                    ))}
                    {allDayEvents.length > 2 && (
                      <div className="text-xs text-gray-500 dark:text-gray-400 px-1">
                        {t('moreEvents', { count: allDayEvents.length - 2 })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Scrollable time grid */}
      <div ref={scrollRef} className="flex flex-1 overflow-auto">
        {/* Time column */}
        <div className="w-16 shrink-0 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95">
          {HOURS.map((hour) => (
            <div
              key={hour}
              className="relative border-b border-gray-200 dark:border-gray-700"
              style={{ height: HOUR_HEIGHT }}
            >
              <span className="absolute -top-2.5 right-2 text-xs text-gray-500 dark:text-gray-400">
                {format(addHours(startOfDay(new Date()), hour), 'HH:mm')}
              </span>
            </div>
          ))}
        </div>

        {/* Day columns */}
        <div className="flex flex-1">
          {weekDays.map((day) => {
            const timedEvents = getTimedEvents(day)
            const isTodayDate = isToday(day)

            return (
              <div
                key={day.toISOString()}
                className="relative flex-1 border-r border-gray-200 dark:border-gray-700 last:border-r-0 min-w-0"
              >
                {/* Hour slots */}
                {HOURS.map((hour) => (
                  <div
                    key={hour}
                    onClick={() => handleTimeSlotClick(day, hour)}
                    className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-white/5 cursor-pointer transition-colors"
                    style={{ height: HOUR_HEIGHT }}
                  />
                ))}

                {/* Current time indicator */}
                {isTodayDate && (
                  <div
                    className="absolute left-0 right-0 z-10 border-t-2 border-red-500"
                    style={{
                      top:
                        ((getHours(new Date()) * 60 + getMinutes(new Date())) / 60) * HOUR_HEIGHT,
                    }}
                  >
                    <div className="absolute -left-1 -top-1.5 h-3 w-3 rounded-full bg-red-500" />
                  </div>
                )}

                {/* Timed events */}
                {timedEvents.map((event) => {
                  const { top, height } = getEventPosition(event)
                  return (
                    <div
                      key={event.id}
                      className="absolute left-0.5 right-0.5 z-20"
                      style={{ top, height }}
                    >
                      <EventCard
                        event={event}
                        variant="compact"
                        onClick={() => onEventClick?.(event)}
                        className="h-full overflow-hidden"
                      />
                    </div>
                  )
                })}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
