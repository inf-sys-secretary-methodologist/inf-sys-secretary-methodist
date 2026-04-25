'use client'

import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { Calendar, User, AlertTriangle, MoreHorizontal } from 'lucide-react'
import { useTranslations, useLocale } from 'next-intl'

const localeMap = { ru, en: enUS, fr, ar }

import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { Task, TaskPriority, TaskStatus } from '@/types/tasks'

interface TaskCardProps {
  task: Task
  onClick?: () => void
  onEdit?: () => void
  onDelete?: () => void
  className?: string
}

const PRIORITY_COLORS: Record<TaskPriority, { bg: string; border: string; text: string }> = {
  low: {
    bg: 'bg-slate-50 dark:bg-slate-900/40',
    border: 'bg-slate-400',
    text: 'text-slate-700 dark:text-slate-300',
  },
  normal: {
    bg: 'bg-blue-50 dark:bg-blue-950/40',
    border: 'bg-blue-500',
    text: 'text-blue-700 dark:text-blue-300',
  },
  high: {
    bg: 'bg-amber-50 dark:bg-amber-950/40',
    border: 'bg-amber-500',
    text: 'text-amber-700 dark:text-amber-300',
  },
  urgent: {
    bg: 'bg-rose-50 dark:bg-rose-950/40',
    border: 'bg-rose-500',
    text: 'text-rose-700 dark:text-rose-300',
  },
}

const STATUS_COLORS: Record<TaskStatus, string> = {
  new: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  assigned: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  in_progress: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  review: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  completed: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  canceled: 'bg-gray-100 text-gray-500 dark:bg-gray-900/40 dark:text-gray-400',
  deferred: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300',
}

export function TaskCard({ task, onClick, onEdit, onDelete, className }: TaskCardProps) {
  const t = useTranslations('tasks')
  const locale = useLocale()
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS
  const priorityColors = PRIORITY_COLORS[task.priority]

  return (
    <div
      className={cn(
        'group relative rounded-xl p-4',
        'transition-all duration-200 ease-out',
        'hover:shadow-md',
        'border border-gray-100 dark:border-gray-800',
        priorityColors.bg,
        className
      )}
    >
      <div
        className={cn('absolute left-0 top-3 bottom-3 w-1 rounded-full', priorityColors.border)}
      />

      <div className="pl-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1 flex-wrap">
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  STATUS_COLORS[task.status]
                )}
              >
                {t(`status.${task.status}`)}
              </span>
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  priorityColors.bg,
                  priorityColors.text
                )}
              >
                {t(`priority.${task.priority}`)}
              </span>
              {task.is_overdue && (
                <span
                  data-testid="task-overdue-indicator"
                  className="inline-flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300"
                >
                  <AlertTriangle className="h-3 w-3" />
                  {t('overdue')}
                </span>
              )}
            </div>

            <button
              onClick={onClick}
              className="text-sm font-semibold hover:underline text-left w-full truncate text-gray-900 dark:text-gray-100"
            >
              {task.title}
            </button>

            {task.description && (
              <p className="mt-1.5 text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                {task.description}
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
          {task.assignee && (
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <User className="h-3.5 w-3.5" />
              <span className="truncate">{task.assignee.name}</span>
            </div>
          )}

          {task.due_date && (
            <div
              className={cn(
                'flex items-center gap-2 text-xs',
                task.is_overdue ? 'text-rose-600 dark:text-rose-400' : 'text-muted-foreground'
              )}
            >
              <Calendar className="h-3.5 w-3.5" />
              <span>{format(new Date(task.due_date), 'd MMM yyyy', { locale: dateLocale })}</span>
            </div>
          )}

          <div className="flex items-center gap-2 mt-2">
            <div className="flex-1 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
              <div
                className={cn('h-full transition-all', priorityColors.border)}
                style={{ width: `${task.progress}%` }}
              />
            </div>
            <span className="text-[10px] font-medium text-muted-foreground tabular-nums">
              {task.progress}%
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
