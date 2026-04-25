'use client'

import { format } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import {
  Eye,
  Pin,
  Paperclip,
  User,
  Calendar,
  MoreHorizontal,
  Send,
  Archive,
  EyeOff,
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
  Announcement,
  AnnouncementPriority,
  AnnouncementStatus,
  TargetAudience,
} from '@/types/announcements'

const localeMap = { ru, en: enUS, fr, ar }

interface AnnouncementCardProps {
  announcement: Announcement
  onClick?: () => void
  onEdit?: () => void
  onDelete?: () => void
  onPublish?: () => void
  onUnpublish?: () => void
  onArchive?: () => void
  className?: string
}

const PRIORITY_COLORS: Record<AnnouncementPriority, { bg: string; border: string; text: string }> = {
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

const STATUS_COLORS: Record<AnnouncementStatus, string> = {
  draft: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
  published: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  archived: 'bg-gray-100 text-gray-500 dark:bg-gray-900/40 dark:text-gray-400',
}

const AUDIENCE_COLORS: Record<TargetAudience, string> = {
  all: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  students: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  teachers: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  staff: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
  admins: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
}

export function AnnouncementCard({
  announcement,
  onClick,
  onEdit,
  onDelete,
  onPublish,
  onUnpublish,
  onArchive,
  className,
}: AnnouncementCardProps) {
  const t = useTranslations('announcements')
  const locale = useLocale()
  const dateLocale = localeMap[locale as keyof typeof localeMap] || enUS
  const priorityColors = PRIORITY_COLORS[announcement.priority]
  const attachmentCount = announcement.attachments?.length ?? 0

  return (
    <div
      className={cn(
        'group relative rounded-xl p-4',
        'transition-all duration-200 ease-out',
        'hover:shadow-md',
        'border border-gray-100 dark:border-gray-800',
        priorityColors.bg,
        announcement.is_pinned && 'ring-2 ring-amber-400/40',
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
              {announcement.is_pinned && (
                <span
                  data-testid="announcement-pinned-indicator"
                  className="inline-flex items-center gap-1 text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300"
                >
                  <Pin className="h-3 w-3" />
                  {t('pinned')}
                </span>
              )}
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  STATUS_COLORS[announcement.status]
                )}
              >
                {t(`status.${announcement.status}`)}
              </span>
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  priorityColors.bg,
                  priorityColors.text
                )}
              >
                {t(`priority.${announcement.priority}`)}
              </span>
              <span
                className={cn(
                  'text-[10px] font-semibold uppercase tracking-wide px-2 py-0.5 rounded-full',
                  AUDIENCE_COLORS[announcement.target_audience]
                )}
              >
                {t(`audience.${announcement.target_audience}`)}
              </span>
            </div>

            <button
              onClick={onClick}
              className="text-base font-semibold hover:underline text-left w-full truncate text-gray-900 dark:text-gray-100"
            >
              {announcement.title}
            </button>

            {announcement.summary && (
              <p className="mt-1.5 text-sm text-muted-foreground line-clamp-2 leading-relaxed">
                {announcement.summary}
              </p>
            )}

            {announcement.tags && announcement.tags.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {announcement.tags.map((tag) => (
                  <span
                    key={tag}
                    className="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400"
                  >
                    #{tag}
                  </span>
                ))}
              </div>
            )}
          </div>

          {(onEdit || onDelete || onPublish || onUnpublish || onArchive) && (
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
                {announcement.status === 'draft' && onPublish && (
                  <DropdownMenuItem onClick={onPublish}>
                    <Send className="h-4 w-4 mr-2" />
                    {t('actions.publish')}
                  </DropdownMenuItem>
                )}
                {announcement.status === 'published' && onUnpublish && (
                  <DropdownMenuItem onClick={onUnpublish}>
                    <EyeOff className="h-4 w-4 mr-2" />
                    {t('actions.unpublish')}
                  </DropdownMenuItem>
                )}
                {announcement.status !== 'archived' && onArchive && (
                  <DropdownMenuItem onClick={onArchive}>
                    <Archive className="h-4 w-4 mr-2" />
                    {t('actions.archive')}
                  </DropdownMenuItem>
                )}
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

        <div className="mt-3 flex items-center gap-3 flex-wrap text-xs text-muted-foreground">
          {announcement.author && (
            <div className="flex items-center gap-1.5">
              <User className="h-3.5 w-3.5" />
              <span className="truncate">{announcement.author.name}</span>
            </div>
          )}

          <div className="flex items-center gap-1.5">
            <Calendar className="h-3.5 w-3.5" />
            <span>{format(new Date(announcement.created_at), 'd MMM yyyy', { locale: dateLocale })}</span>
          </div>

          <div className="flex items-center gap-1.5">
            <Eye className="h-3.5 w-3.5" />
            <span>{announcement.view_count}</span>
          </div>

          {attachmentCount > 0 && (
            <div className="flex items-center gap-1.5">
              <Paperclip className="h-3.5 w-3.5" />
              <span data-testid="announcement-attachment-count">{attachmentCount}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
