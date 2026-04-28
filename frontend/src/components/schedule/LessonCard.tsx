'use client'

import { useTranslations } from 'next-intl'
import { MapPin, User, BookOpen } from 'lucide-react'

import { cn } from '@/lib/utils'
import type { Lesson } from '@/types/schedule'

interface LessonCardProps {
  lesson: Lesson
  onClick?: () => void
  className?: string
}

const DEFAULT_TYPE_COLOR = '#6366f1'

export function LessonCard({ lesson, onClick, className }: LessonCardProps) {
  const t = useTranslations('schedule')
  const typeColor = lesson.lesson_type?.color || DEFAULT_TYPE_COLOR
  const disciplineName = lesson.discipline?.name || `#${lesson.discipline_id}`
  const teacherName = lesson.teacher?.name || ''
  const classroomLabel = lesson.classroom
    ? `${lesson.classroom.building}-${lesson.classroom.number}`
    : ''
  const typeName = lesson.lesson_type?.short_name || ''

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full text-left rounded-lg p-2 text-xs',
        'transition-all duration-150 ease-out',
        'hover:shadow-md hover:scale-[1.02]',
        'border border-gray-100 dark:border-gray-700',
        'bg-white dark:bg-gray-900',
        lesson.is_cancelled && 'opacity-50 line-through',
        className
      )}
      style={{ borderLeftColor: typeColor, borderLeftWidth: 3 }}
      data-testid="lesson-card"
    >
      <div className="flex items-start justify-between gap-1 mb-1">
        <span className="font-semibold truncate leading-tight" data-testid="lesson-discipline">
          {disciplineName}
        </span>
        {typeName && (
          <span
            className="shrink-0 text-[10px] font-bold uppercase px-1.5 py-0.5 rounded"
            style={{ backgroundColor: typeColor + '20', color: typeColor }}
            data-testid="lesson-type-badge"
          >
            {typeName}
          </span>
        )}
      </div>

      {teacherName && (
        <div className="flex items-center gap-1 text-muted-foreground mt-0.5" data-testid="lesson-teacher">
          <User className="h-3 w-3 shrink-0" />
          <span className="truncate">{teacherName}</span>
        </div>
      )}

      {classroomLabel && (
        <div className="flex items-center gap-1 text-muted-foreground mt-0.5" data-testid="lesson-classroom">
          <MapPin className="h-3 w-3 shrink-0" />
          <span className="truncate">{classroomLabel}</span>
        </div>
      )}

      {lesson.is_cancelled && (
        <span className="text-[10px] text-destructive font-medium">{t('lesson.cancelled')}</span>
      )}
    </button>
  )
}
