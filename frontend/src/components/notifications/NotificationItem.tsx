'use client'

import Link from 'next/link'
import { Bell, Calendar, FileText, Megaphone, CheckSquare, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Notification, NotificationType } from '@/types/notification'

interface NotificationItemProps {
  notification: Notification
  onMarkAsRead?: (id: number) => void
  onDelete?: (id: number) => void
  compact?: boolean
}

const typeIcons: Record<NotificationType, React.ElementType> = {
  system: Settings,
  reminder: Bell,
  task: CheckSquare,
  document: FileText,
  announcement: Megaphone,
  event: Calendar,
}

const typeColors: Record<NotificationType, string> = {
  system: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
  reminder: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-400',
  task: 'bg-green-100 text-green-600 dark:bg-green-900/40 dark:text-green-400',
  document: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-400',
  announcement: 'bg-purple-100 text-purple-600 dark:bg-purple-900/40 dark:text-purple-400',
  event: 'bg-rose-100 text-rose-600 dark:bg-rose-900/40 dark:text-rose-400',
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) {
    return 'Только что'
  } else if (diffMins < 60) {
    return `${diffMins} мин. назад`
  } else if (diffHours < 24) {
    return `${diffHours} ч. назад`
  } else if (diffDays === 1) {
    return 'Вчера'
  } else if (diffDays < 7) {
    return `${diffDays} дн. назад`
  } else {
    return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
  }
}

export function NotificationItem({
  notification,
  onMarkAsRead,
  compact = false,
}: NotificationItemProps) {
  const TypeIcon = typeIcons[notification.type] || Bell
  const colorClass = typeColors[notification.type] || typeColors.system

  const handleClick = () => {
    if (!notification.is_read && onMarkAsRead) {
      onMarkAsRead(notification.id)
    }
  }

  const content = (
    <div className="w-full p-3">
      <div className="flex gap-3">
        {/* Icon */}
        <div
          className={cn(
            'flex size-11 shrink-0 items-center justify-center rounded-full ring-1 ring-border',
            colorClass
          )}
        >
          <TypeIcon className="h-5 w-5" />
        </div>

        {/* Content */}
        <div className="flex flex-1 flex-col space-y-2">
          <div className="w-full items-start">
            <div>
              <div className="flex items-center justify-between gap-2">
                <div className="text-sm">
                  <span
                    className={cn('font-medium', notification.is_read && 'text-muted-foreground')}
                  >
                    {notification.title}
                  </span>
                </div>
                {!notification.is_read && <div className="size-1.5 rounded-full bg-emerald-500" />}
              </div>
              <div className="flex items-center justify-between gap-2 mt-0.5">
                <div className="text-xs text-muted-foreground">
                  {notification.created_at_display || formatRelativeTime(notification.created_at)}
                </div>
              </div>
            </div>
          </div>

          {/* Message content */}
          <div
            className={cn(
              'rounded-lg p-2.5 text-sm tracking-[-0.006em]',
              notification.is_read
                ? 'bg-muted text-muted-foreground'
                : 'bg-blue-50 dark:bg-blue-900/20 text-foreground/80'
            )}
          >
            <p className={compact ? 'line-clamp-2' : undefined}>{notification.message}</p>
          </div>
        </div>
      </div>
    </div>
  )

  if (notification.link) {
    return (
      <Link
        href={notification.link}
        onClick={handleClick}
        className="block rounded-xl hover:bg-muted/50 transition-colors"
      >
        {content}
      </Link>
    )
  }

  return (
    <button
      onClick={handleClick}
      className="block w-full text-left rounded-xl hover:bg-muted/50 transition-colors"
    >
      {content}
    </button>
  )
}
