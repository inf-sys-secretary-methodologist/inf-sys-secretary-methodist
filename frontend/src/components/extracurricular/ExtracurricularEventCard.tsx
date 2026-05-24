'use client'

import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import {
  MapPin,
  Calendar as CalendarIcon,
  Users,
  MoreHorizontal,
  Send,
  X as XIcon,
  CheckCircle,
  Edit,
  Trash2,
} from 'lucide-react'
import { useTranslations, useLocale } from 'next-intl'

import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type {
  EventCategory,
  EventStatus,
  EventTargetAudience,
  ExtracurricularEventSummary,
} from '@/types/extracurricular'

const localeMap = { ru, en: enUS, fr, ar }

export interface ExtracurricularEventCardProps {
  event: ExtracurricularEventSummary
  onClick?: () => void
  onEdit?: () => void
  onDelete?: () => void
  onRegister?: () => void
  onUnregister?: () => void
  onPublish?: () => void
  onCancel?: () => void
  onComplete?: () => void
  isRegistered?: boolean
  className?: string
}

const STATUS_COLORS: Record<EventStatus, string> = {
  draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  published: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  canceled: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
  completed: 'bg-gray-100 text-gray-500 dark:bg-gray-900/40 dark:text-gray-400',
}

const CATEGORY_COLORS: Record<EventCategory, string> = {
  academic: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  cultural: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  sports: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  volunteer: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  professional: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
}

const AUDIENCE_COLORS: Record<EventTargetAudience, string> = {
  all: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  students: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  teachers: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  staff: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
}

export function ExtracurricularEventCard({
  event,
  onClick,
  onEdit,
  onDelete,
  onRegister,
  onUnregister,
  onPublish,
  onCancel,
  onComplete,
  isRegistered,
  className,
}: ExtracurricularEventCardProps) {
  const t = useTranslations('extracurricular')
  const locale = useLocale()
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS

  const hasMenu = Boolean(onEdit || onDelete || onPublish || onCancel || onComplete)

  return (
    <div
      className={cn(
        'group relative rounded-xl border border-gray-100 dark:border-gray-800 p-4',
        'transition-all duration-200 ease-out hover:shadow-md',
        'bg-white dark:bg-gray-950/40',
        className
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2 flex-wrap">
            <span
              className={cn(
                'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                STATUS_COLORS[event.status]
              )}
            >
              {t(`status.${event.status}`)}
            </span>
            <span
              className={cn(
                'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                CATEGORY_COLORS[event.category]
              )}
            >
              {t(`category.${event.category}`)}
            </span>
            <span
              className={cn(
                'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                AUDIENCE_COLORS[event.target_audience]
              )}
            >
              {t(`audience.${event.target_audience}`)}
            </span>
          </div>

          <button
            onClick={onClick}
            className="text-base font-semibold hover:underline text-left w-full truncate text-gray-900 dark:text-gray-100"
          >
            {event.title}
          </button>

          <div className="mt-2 flex items-center gap-3 flex-wrap text-xs text-muted-foreground">
            {event.location && (
              <div className="flex items-center gap-1.5">
                <MapPin className="h-3.5 w-3.5" />
                <span className="truncate">{event.location}</span>
              </div>
            )}
            <div className="flex items-center gap-1.5">
              <CalendarIcon className="h-3.5 w-3.5" />
              <span>
                {format(new Date(event.start_at), 'd MMM yyyy HH:mm', { locale: dateLocale })}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <Users className="h-3.5 w-3.5" />
              <span>
                {event.max_capacity != null
                  ? `${event.participant_count} / ${event.max_capacity}`
                  : `${event.participant_count}`}
              </span>
            </div>
          </div>
        </div>

        {hasMenu && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                data-testid="event-card-menu-trigger"
                variant="ghost"
                size="icon"
                className="h-8 w-8 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {onEdit && (
                <DropdownMenuItem onClick={onEdit}>
                  <Edit className="h-4 w-4 mr-2" />
                  {t('edit')}
                </DropdownMenuItem>
              )}
              {event.status === 'draft' && onPublish && (
                <DropdownMenuItem onClick={onPublish}>
                  <Send className="h-4 w-4 mr-2" />
                  {t('actions.publish')}
                </DropdownMenuItem>
              )}
              {event.status === 'published' && onCancel && (
                <DropdownMenuItem onClick={onCancel}>
                  <XIcon className="h-4 w-4 mr-2" />
                  {t('actions.cancel')}
                </DropdownMenuItem>
              )}
              {event.status === 'published' && onComplete && (
                <DropdownMenuItem onClick={onComplete}>
                  <CheckCircle className="h-4 w-4 mr-2" />
                  {t('actions.complete')}
                </DropdownMenuItem>
              )}
              {onDelete && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={onDelete} className="text-destructive">
                    <Trash2 className="h-4 w-4 mr-2" />
                    {t('delete')}
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>

      {(onRegister || onUnregister) && (
        <div className="mt-3 flex justify-end">
          {isRegistered
            ? onUnregister && (
                <Button variant="outline" size="sm" onClick={onUnregister}>
                  {t('unregister')}
                </Button>
              )
            : onRegister && (
                <Button size="sm" onClick={onRegister}>
                  {t('register')}
                </Button>
              )}
        </div>
      )}
    </div>
  )
}
