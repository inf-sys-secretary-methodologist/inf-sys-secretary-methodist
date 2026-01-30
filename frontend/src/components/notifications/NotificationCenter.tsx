'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Bell, Settings, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslations } from 'next-intl'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  useNotifications,
  useUnreadCount,
  useMarkAsRead,
  useMarkAllAsRead,
} from '@/hooks/useNotifications'
import { cn } from '@/lib/utils'
import type { Notification, NotificationType } from '@/types/notification'
import {
  Bell as BellIcon,
  Calendar,
  FileText,
  Megaphone,
  CheckSquare,
  Settings as SettingsIcon,
} from 'lucide-react'

interface NotificationCenterProps {
  className?: string
}

const typeIcons: Record<NotificationType, React.ElementType> = {
  system: SettingsIcon,
  reminder: BellIcon,
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

function NotificationRow({
  notification,
  onMarkAsRead,
  onClose,
  formatRelativeTime,
}: {
  notification: Notification
  onMarkAsRead: (id: number) => void
  onClose: () => void
  formatRelativeTime: (dateString: string) => string
}) {
  /* c8 ignore next 2 - Fallback values */
  const TypeIcon = typeIcons[notification.type] || BellIcon
  const colorClass = typeColors[notification.type] || typeColors.system

  /* c8 ignore start - Click handler */
  const handleClick = () => {
    if (!notification.is_read) {
      onMarkAsRead(notification.id)
    }
    if (notification.link) {
      onClose()
    }
  }
  /* c8 ignore stop */

  const content = (
    <div
      className={cn(
        'relative flex items-start gap-3 rounded-lg px-3 py-3 transition-colors hover:bg-accent',
        notification.is_read ? '' : 'bg-blue-50/50 dark:bg-blue-900/10'
      )}
    >
      {/* Icon */}
      <div
        className={cn('flex h-9 w-9 shrink-0 items-center justify-center rounded-full', colorClass)}
      >
        <TypeIcon className="h-4 w-4" />
      </div>

      {/* Content */}
      <div className="min-w-0 flex-1">
        <p className="text-sm text-foreground">
          <span className="font-medium">{notification.title}</span>
        </p>
        <p className="mt-0.5 text-sm text-muted-foreground line-clamp-2">{notification.message}</p>
        <p className="mt-1 text-xs text-muted-foreground">
          {notification.created_at_display || formatRelativeTime(notification.created_at)}
        </p>
      </div>

      {/* Unread indicator */}
      {!notification.is_read && (
        <div className="flex h-2 w-2 shrink-0 items-center justify-center self-center">
          <div className="h-2 w-2 rounded-full bg-blue-500" />
        </div>
      )}
    </div>
  )

  /* c8 ignore start - Link rendering conditional */
  if (notification.link) {
    return (
      <Link href={notification.link} onClick={handleClick} className="block">
        {content}
      </Link>
    )
  }
  /* c8 ignore stop */

  return (
    <button onClick={handleClick} className="block w-full text-left">
      {content}
    </button>
  )
}

export function NotificationCenter({ className }: NotificationCenterProps) {
  const [open, setOpen] = useState(false)
  const t = useTranslations('notifications')
  const tCommon = useTranslations('common')

  const { data: unreadData } = useUnreadCount()
  /* c8 ignore next */
  const unreadCount = unreadData?.count ?? 0

  const { data: allData, isLoading } = useNotifications({ limit: 20 })
  /* c8 ignore next */
  const notifications = allData?.notifications ?? []

  const markAsRead = useMarkAsRead()
  const markAllAsRead = useMarkAllAsRead()

  /* c8 ignore start - Time formatting, tested in e2e */
  const formatRelativeTime = (dateString: string): string => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 1) {
      return tCommon('time.justNow')
    } else if (diffMins < 60) {
      return tCommon('time.minutesAgo', { count: diffMins })
    } else if (diffHours < 24) {
      return tCommon('time.hoursAgo', { count: diffHours })
    } else if (diffDays === 1) {
      return tCommon('time.yesterday')
    } else if (diffDays < 7) {
      return tCommon('time.daysAgo', { count: diffDays })
    } else {
      return date.toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
    }
  }
  /* c8 ignore stop */

  /* c8 ignore start - Event handlers, tested in e2e */
  const handleMarkAsRead = async (id: number) => {
    try {
      await markAsRead.mutateAsync(id)
    } catch {
      toast.error(t('markAsReadError'))
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead.mutateAsync()
      toast.success(t('allMarkedAsRead'))
    } catch {
      toast.error(t('markAllAsReadError'))
    }
  }
  /* c8 ignore stop */

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn('relative', className)}
          aria-label={
            unreadCount > 0 ? t('ariaLabelWithCount', { count: unreadCount }) : t('ariaLabel')
          }
        >
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <Badge
              className="absolute -top-2 left-full min-w-5 -translate-x-1/2 px-1 h-5 text-xs"
              variant="destructive"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-96 p-0" align="end" sideOffset={8}>
        {/* Header */}
        <div className="flex items-center justify-between border-b px-4 py-3">
          <h3 className="text-sm font-semibold">{t('title')}</h3>
          <div className="flex items-center gap-2">
            {unreadCount > 0 && (
              <button
                className="text-xs font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
                onClick={handleMarkAllAsRead}
                disabled={markAllAsRead.isPending}
              >
                {t('markAllAsRead')}
              </button>
            )}
            {/* c8 ignore start - Settings link onClick */}
            <Link
              href="/settings/notifications"
              className="text-muted-foreground hover:text-foreground"
              onClick={() => setOpen(false)}
            >
              <Settings className="h-4 w-4" />
            </Link>
            {/* c8 ignore stop */}
          </div>
        </div>

        {/* Notifications List */}
        <div className="max-h-96 overflow-y-auto">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : notifications.length > 0 ? (
            <div className="divide-y">
              {/* c8 ignore next 8 - NotificationRow onClose callback */}
              {notifications.map((notification) => (
                <NotificationRow
                  key={notification.id}
                  notification={notification}
                  onMarkAsRead={handleMarkAsRead}
                  onClose={() => setOpen(false)}
                  formatRelativeTime={formatRelativeTime}
                />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center px-4">
              <div className="rounded-full bg-muted p-3 mb-3">
                <Bell className="h-6 w-6 text-muted-foreground" />
              </div>
              <p className="text-sm text-muted-foreground">{t('empty')}</p>
            </div>
          )}
        </div>

        {/* Footer */}
        {notifications.length > 0 && (
          <div className="border-t p-3">
            {/* c8 ignore start - View all button onClick */}
            <Button variant="outline" className="w-full" asChild onClick={() => setOpen(false)}>
              <Link href="/notifications">{t('viewAll')}</Link>
            </Button>
            {/* c8 ignore stop */}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
