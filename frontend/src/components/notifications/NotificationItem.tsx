'use client'

import Link from 'next/link'
import {
  Bell,
  Calendar,
  FileText,
  Megaphone,
  CheckSquare,
  Settings,
  Trash2,
  Check,
  AlertCircle,
  AlertTriangle,
  Info,
  Clock,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { Notification, NotificationType, NotificationPriority } from '@/types/notification'

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

const priorityIcons: Record<NotificationPriority, React.ElementType> = {
  low: Info,
  normal: Info,
  high: AlertTriangle,
  urgent: AlertCircle,
}

const priorityColors: Record<NotificationPriority, string> = {
  low: 'text-muted-foreground',
  normal: 'text-foreground',
  high: 'text-gold-500',
  urgent: 'text-red-500',
}

export function NotificationItem({
  notification,
  onMarkAsRead,
  onDelete,
  compact = false,
}: NotificationItemProps) {
  const TypeIcon = typeIcons[notification.type] || Bell
  const PriorityIcon = priorityIcons[notification.priority]
  const priorityColor = priorityColors[notification.priority]

  const content = (
    <div
      className={cn(
        'flex gap-3 py-4 first:pt-0 last:pb-0 transition-all duration-200',
        notification.link && 'cursor-pointer'
      )}
    >
      {/* Avatar/Icon container */}
      <div
        className={cn(
          'flex-shrink-0 size-11 rounded-full flex items-center justify-center ring-1 ring-border',
          notification.is_read
            ? 'bg-muted text-muted-foreground'
            : 'bg-blue-100 text-blue-500 dark:bg-blue-900/40 dark:text-blue-400'
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
                  className={cn(
                    'font-medium',
                    notification.is_read ? 'text-muted-foreground' : 'text-foreground'
                  )}
                >
                  {notification.title}
                </span>
                {notification.priority !== 'normal' && (
                  <PriorityIcon className={cn('inline-block ml-1.5 h-3.5 w-3.5', priorityColor)} />
                )}
              </div>
              {!notification.is_read && (
                <div className="size-2 rounded-full bg-blue-500 dark:bg-blue-400 flex-shrink-0" />
              )}
            </div>
            <div className="flex items-center justify-between gap-2 mt-0.5">
              <div className="flex items-center text-xs text-muted-foreground">
                <Clock className="h-3 w-3 mr-1" />
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

        {/* Actions */}
        {!compact && (
          <div className="flex gap-2">
            {!notification.is_read && onMarkAsRead && (
              <Button
                variant="outline"
                size="sm"
                className="h-7 text-xs"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  onMarkAsRead(notification.id)
                }}
              >
                <Check className="h-3.5 w-3.5 mr-1" />
                Прочитано
              </Button>
            )}
            {onDelete && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs text-muted-foreground hover:text-destructive"
                onClick={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                  onDelete(notification.id)
                }}
              >
                <Trash2 className="h-3.5 w-3.5 mr-1" />
                Удалить
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )

  if (notification.link) {
    return (
      <Link
        href={notification.link}
        onClick={() => !notification.is_read && onMarkAsRead?.(notification.id)}
        className="block hover:bg-muted/50 rounded-lg transition-colors px-3 -mx-3"
      >
        {content}
      </Link>
    )
  }

  return <div className="px-0">{content}</div>
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
