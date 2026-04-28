'use client'

import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { LessonCard } from './LessonCard'
import type { Lesson } from '@/types/schedule'
import { DAY_NAMES, TIME_SLOTS } from '@/types/schedule'

interface TimetableGridProps {
  lessons: Lesson[]
  onLessonClick?: (lesson: Lesson) => void
  canEdit: boolean
}

/**
 * Timetable grid: 6 columns (Mon-Sat) x 5 standard time slots.
 * Each cell displays LessonCards matching that day + time slot.
 */
export function TimetableGrid({ lessons, onLessonClick }: TimetableGridProps) {
  const t = useTranslations('schedule')

  const getLessonsForCell = (dayOfWeek: number, timeStart: string): Lesson[] => {
    return lessons.filter(
      (l) => l.day_of_week === dayOfWeek && l.time_start === timeStart
    )
  }

  return (
    <div className="overflow-x-auto" data-testid="timetable-grid">
      <table className="w-full border-collapse min-w-[800px]">
        <thead>
          <tr>
            <th className="sticky left-0 z-10 bg-background p-2 text-xs font-medium text-muted-foreground w-[100px] border-b border-r">
              {/* Time column header */}
            </th>
            {DAY_NAMES.map((day) => (
              <th
                key={day}
                className="p-2 text-xs font-semibold text-center border-b min-w-[140px]"
              >
                {t(`days.${day}`)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {TIME_SLOTS.map((slot) => (
            <tr key={slot.start} data-testid={`time-row-${slot.start}`}>
              <td className="sticky left-0 z-10 bg-background p-2 text-xs text-muted-foreground text-center border-r whitespace-nowrap align-top">
                <div className="font-medium">{slot.start}</div>
                <div className="text-[10px]">{slot.end}</div>
              </td>
              {DAY_NAMES.map((day, dayIndex) => {
                const dayOfWeek = dayIndex + 1 // 1=Monday
                const cellLessons = getLessonsForCell(dayOfWeek, slot.start)

                return (
                  <td
                    key={day}
                    className={cn(
                      'p-1 border border-gray-100 dark:border-gray-800 align-top min-h-[80px]',
                      cellLessons.length === 0 && 'bg-gray-50/50 dark:bg-gray-900/50'
                    )}
                    data-testid={`cell-${dayOfWeek}-${slot.start}`}
                  >
                    <div className="space-y-1">
                      {cellLessons.map((lesson) => (
                        <LessonCard
                          key={lesson.id}
                          lesson={lesson}
                          onClick={
                            onLessonClick ? () => onLessonClick(lesson) : undefined
                          }
                        />
                      ))}
                    </div>
                  </td>
                )
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
