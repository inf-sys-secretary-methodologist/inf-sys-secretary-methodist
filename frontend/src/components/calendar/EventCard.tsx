'use client'

import * as React from 'react'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { Clock, MapPin, Users, MoreHorizontal } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { CalendarEvent, EventType } from '@/types/calendar'

interface EventCardProps {
  event: CalendarEvent
  variant?: 'compact' | 'full'
  onClick?: () => void
  onEdit?: () => void
  onDelete?: () => void
  className?: string
}

// Modern gradient-based colors for event types
const EVENT_TYPE_COLORS: Record<
  EventType,
  { bg: string; border: string; text: string; icon: string }
> = {
  meeting: {
    bg: 'bg-blue-50 dark:bg-blue-950/40',
    border: 'bg-blue-500',
    text: 'text-blue-700 dark:text-blue-300',
    icon: 'text-blue-500',
  },
  deadline: {
    bg: 'bg-rose-50 dark:bg-rose-950/40',
    border: 'bg-rose-500',
    text: 'text-rose-700 dark:text-rose-300',
    icon: 'text-rose-500',
  },
  task: {
    bg: 'bg-emerald-50 dark:bg-emerald-950/40',
    border: 'bg-emerald-500',
    text: 'text-emerald-700 dark:text-emerald-300',
    icon: 'text-emerald-500',
  },
  reminder: {
    bg: 'bg-amber-50 dark:bg-amber-950/40',
    border: 'bg-amber-500',
    text: 'text-amber-700 dark:text-amber-300',
    icon: 'text-amber-500',
  },
  holiday: {
    bg: 'bg-violet-50 dark:bg-violet-950/40',
    border: 'bg-violet-500',
    text: 'text-violet-700 dark:text-violet-300',
    icon: 'text-violet-500',
  },
  personal: {
    bg: 'bg-slate-50 dark:bg-slate-900/40',
    border: 'bg-slate-500',
    text: 'text-slate-700 dark:text-slate-300',
    icon: 'text-slate-500',
  },
}

// Custom color parsing
function getCustomColorStyles(color: string) {
  return {
    bg: 'bg-gray-50 dark:bg-gray-900/40',
    border: color,
    text: 'text-gray-700 dark:text-gray-300',
    icon: 'text-gray-500',
    customBorder: true,
  }
}

export function EventCard({
  event,
  variant = 'compact',
  onClick,
  onEdit,
  onDelete,
  className,
}: EventCardProps) {
  const t = useTranslations('calendar')
  const startTime = new Date(event.start_time)
  const endTime = event.end_time ? new Date(event.end_time) : null

  const colors = event.color
    ? getCustomColorStyles(event.color)
    : EVENT_TYPE_COLORS[event.event_type]

  const hasCustomColor = event.color && 'customBorder' in colors

  if (variant === 'compact') {
    return (
      <button
        onClick={onClick}
        className={cn(
          'group relative w-full text-left rounded-md px-2.5 py-1.5 text-xs',
          'transition-all duration-200 ease-out',
          'hover:shadow-sm hover:scale-[1.02]',
          colors.bg,
          className
        )}
      >
        {/* Left color indicator */}
        <div
          className={cn(
            'absolute left-0 top-1 bottom-1 w-1 rounded-full',
            !hasCustomColor && colors.border
          )}
          /* c8 ignore next - Conditional custom color style */
          style={hasCustomColor ? { backgroundColor: colors.border } : undefined}
        />

        <div className="pl-2">
          <p className={cn('font-medium truncate leading-tight', colors.text)}>{event.title}</p>
          {!event.all_day && (
            <p className="text-muted-foreground/80 truncate mt-0.5 text-[10px] font-medium">
              {format(startTime, 'HH:mm')}
            </p>
          )}
        </div>
      </button>
    )
  }

  return (
    <div
      className={cn(
        'group relative rounded-xl p-4',
        'transition-all duration-200 ease-out',
        'hover:shadow-md',
        'border border-gray-100 dark:border-gray-800',
        colors.bg,
        className
      )}
    >
      {/* Left color indicator */}
      <div
        className={cn(
          'absolute left-0 top-3 bottom-3 w-1 rounded-full',
          !hasCustomColor && colors.border
        )}
        style={hasCustomColor ? { backgroundColor: colors.border } : undefined}
      />

      <div className="pl-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  colors.bg,
                  colors.text
                )}
              >
                {t(`eventTypes.${event.event_type}`)}
              </span>
              {event.is_recurring && (
                <span className="text-[10px] text-muted-foreground font-medium">
                  {t('recurring')}
                </span>
              )}
            </div>
            <button
              onClick={onClick}
              className={cn(
                'text-sm font-semibold hover:underline text-left w-full truncate',
                'text-gray-900 dark:text-gray-100'
              )}
            >
              {event.title}
            </button>
            {event.description && (
              <p className="mt-1.5 text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                {event.description}
              </p>
            )}
          </div>

          {(onEdit || onDelete) && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <MoreHorizontal className="h-4 w-4" />
                  <span className="sr-only">{t('menu')}</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {onEdit && <DropdownMenuItem onClick={onEdit}>{t('edit')}</DropdownMenuItem>}
                {onDelete && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={onDelete} className="text-destructive">
                      {t('delete')}
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>

        <div className="mt-3 space-y-1.5">
          <div className={cn('flex items-center gap-2 text-xs', colors.icon)}>
            <Clock className="h-3.5 w-3.5" />
            <span className="text-muted-foreground font-medium">
              {event.all_day ? (
                t('allDay')
              ) : (
                <>
                  {format(startTime, 'HH:mm', { locale: ru })}
                  {endTime && ` — ${format(endTime, 'HH:mm', { locale: ru })}`}
                </>
              )}
            </span>
          </div>

          {event.location && (
            <div className={cn('flex items-center gap-2 text-xs', colors.icon)}>
              <MapPin className="h-3.5 w-3.5" />
              <span className="truncate text-muted-foreground">{event.location}</span>
            </div>
          )}

          {event.participants && event.participants.length > 0 && (
            <div className={cn('flex items-center gap-2 text-xs', colors.icon)}>
              <Users className="h-3.5 w-3.5" />
              <span className="text-muted-foreground">
                {t('participants', { count: event.participants.length })}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
