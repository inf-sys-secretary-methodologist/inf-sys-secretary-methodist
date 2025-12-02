'use client'

import * as React from 'react'
import {
  format,
  addHours,
  startOfDay,
  getHours,
  getMinutes,
  differenceInMinutes,
  isToday,
} from 'date-fns'
import { ru } from 'date-fns/locale'

import { cn } from '@/lib/utils'
import { EventCard } from './EventCard'
import type { CalendarEvent } from '@/types/calendar'

interface DayViewProps {
  currentDate: Date
  events: CalendarEvent[]
  onEventClick?: (event: CalendarEvent) => void
  onTimeSlotClick?: (date: Date) => void
  className?: string
}

const HOURS = Array.from({ length: 24 }, (_, i) => i)
const HOUR_HEIGHT = 64 // pixels per hour

export function DayView({
  currentDate,
  events,
  onEventClick,
  onTimeSlotClick,
  className,
}: DayViewProps) {
  const scrollRef = React.useRef<HTMLDivElement>(null)
  const isTodayDate = isToday(currentDate)

  // Scroll to current time on mount
  React.useEffect(() => {
    if (scrollRef.current) {
      const currentHour = new Date().getHours()
      scrollRef.current.scrollTop = (currentHour - 1) * HOUR_HEIGHT
    }
  }, [])

  const allDayEvents = events.filter((e) => e.all_day)
  const timedEvents = events.filter((e) => !e.all_day)

  const getEventPosition = (event: CalendarEvent) => {
    const startTime = new Date(event.start_time)
    const endTime = event.end_time ? new Date(event.end_time) : addHours(startTime, 1)

    const startMinutes = getHours(startTime) * 60 + getMinutes(startTime)
    const duration = differenceInMinutes(endTime, startTime)

    const top = (startMinutes / 60) * HOUR_HEIGHT
    const height = Math.max((duration / 60) * HOUR_HEIGHT, 32)

    return { top, height }
  }

  const handleTimeSlotClick = (hour: number) => {
    const dateWithTime = new Date(currentDate)
    dateWithTime.setHours(hour, 0, 0, 0)
    onTimeSlotClick?.(dateWithTime)
  }

  return (
    <div className={cn('flex flex-1 flex-col overflow-hidden', className)}>
      {/* Day Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-white/5 p-4">
        <div className="flex items-center gap-4">
          <div
            className={cn(
              'flex flex-col items-center justify-center rounded-lg p-2 text-gray-900 dark:text-white',
              isTodayDate && 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
            )}
          >
            <span className="text-xs uppercase">{format(currentDate, 'EEE', { locale: ru })}</span>
            <span className="text-2xl font-bold">{format(currentDate, 'd')}</span>
          </div>
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
              {format(currentDate, 'd MMMM yyyy', { locale: ru })}
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400">{events.length} событий</p>
          </div>
        </div>

        {/* All-day events */}
        {allDayEvents.length > 0 && (
          <div className="mt-4">
            <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">
              На весь день
            </h3>
            <div className="space-y-2">
              {allDayEvents.map((event) => (
                <EventCard
                  key={event.id}
                  event={event}
                  variant="full"
                  onClick={() => onEventClick?.(event)}
                />
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Time Grid */}
      <div ref={scrollRef} className="flex flex-1 overflow-auto">
        {/* Time column */}
        <div className="w-20 shrink-0 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-black/95">
          {HOURS.map((hour) => (
            <div
              key={hour}
              className="relative border-b border-gray-200 dark:border-gray-700"
              style={{ height: HOUR_HEIGHT }}
            >
              <span className="absolute -top-2.5 right-3 text-sm text-gray-500 dark:text-gray-400">
                {format(addHours(startOfDay(new Date()), hour), 'HH:mm')}
              </span>
            </div>
          ))}
        </div>

        {/* Events area */}
        <div className="relative flex-1">
          {/* Hour slots */}
          {HOURS.map((hour) => (
            <div
              key={hour}
              onClick={() => handleTimeSlotClick(hour)}
              className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-white/5 cursor-pointer transition-colors"
              style={{ height: HOUR_HEIGHT }}
            >
              {/* Half hour line */}
              <div className="h-1/2 border-b border-dashed border-gray-100 dark:border-gray-800" />
            </div>
          ))}

          {/* Current time indicator */}
          {isTodayDate && (
            <div
              className="absolute left-0 right-0 z-10 border-t-2 border-red-500"
              style={{
                top: ((getHours(new Date()) * 60 + getMinutes(new Date())) / 60) * HOUR_HEIGHT,
              }}
            >
              <div className="absolute -left-1 -top-1.5 h-3 w-3 rounded-full bg-red-500" />
            </div>
          )}

          {/* Timed events */}
          <div className="absolute inset-0">
            {timedEvents.map((event) => {
              const { top, height } = getEventPosition(event)
              return (
                <div key={event.id} className="absolute left-2 right-2" style={{ top, height }}>
                  <EventCard
                    event={event}
                    variant="full"
                    onClick={() => onEventClick?.(event)}
                    className="h-full overflow-hidden"
                  />
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
