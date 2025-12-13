'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Bell, CheckCheck, Settings, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { NotificationItem } from './NotificationItem'
import {
  useNotifications,
  useUnreadCount,
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
} from '@/hooks/useNotifications'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface NotificationBellProps {
  className?: string
}

export function NotificationBell({ className }: NotificationBellProps) {
  const [open, setOpen] = useState(false)

  const { count: unreadCount } = useUnreadCount()
  const { notifications, isLoading } = useNotifications({ limit: 10 })

  const markAsRead = useMarkAsRead()
  const markAllAsRead = useMarkAllAsRead()
  const deleteNotification = useDeleteNotification()

  const handleMarkAsRead = async (id: number) => {
    try {
      await markAsRead.mutateAsync(id)
    } catch {
      toast.error('Не удалось отметить уведомление')
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead.mutateAsync()
      toast.success('Все уведомления отмечены как прочитанные')
    } catch {
      toast.error('Не удалось отметить уведомления')
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteNotification.mutateAsync(id)
      toast.success('Уведомление удалено')
    } catch {
      toast.error('Не удалось удалить уведомление')
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            'relative inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white text-gray-900 transition-all duration-200 hover:bg-gray-100 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            !open && 'hover:scale-105 hover:shadow-lg active:scale-95',
            className
          )}
          aria-label={`Уведомления${unreadCount > 0 ? ` (${unreadCount} непрочитанных)` : ''}`}
          type="button"
        >
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <Badge
              variant="destructive"
              className={cn(
                'absolute -top-1 -right-1 h-5 min-w-5 px-1 flex items-center justify-center',
                'text-[10px] font-bold rounded-full animate-in zoom-in-50 duration-300'
              )}
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </button>
      </PopoverTrigger>

      <PopoverContent className="w-[380px] p-0" align="end" sideOffset={8}>
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold">Уведомления</h3>
            {unreadCount > 0 && (
              <Badge variant="secondary" className="h-5 px-1.5 text-xs">
                {unreadCount} новых
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-1">
            {unreadCount > 0 && (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={handleMarkAllAsRead}
                disabled={markAllAsRead.isPending}
                title="Прочитать все"
              >
                <CheckCheck className="h-4 w-4" />
              </Button>
            )}
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              asChild
              title="Настройки"
              onClick={() => setOpen(false)}
            >
              <Link href="/settings/notifications">
                <Settings className="h-4 w-4" />
              </Link>
            </Button>
          </div>
        </div>

        {/* Notifications list */}
        <ScrollArea className="h-[400px]">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : notifications.length > 0 ? (
            <div>
              {notifications.map((notification, index) => (
                <div
                  key={notification.id}
                  className={cn('px-2', index > 0 && 'border-t border-border')}
                >
                  <NotificationItem
                    notification={notification}
                    onMarkAsRead={handleMarkAsRead}
                    onDelete={handleDelete}
                    compact
                  />
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center px-4">
              <div className="rounded-full bg-muted p-3 mb-3">
                <Bell className="h-6 w-6 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium text-muted-foreground">Нет уведомлений</p>
              <p className="text-xs text-muted-foreground mt-1">
                Здесь будут появляться ваши уведомления
              </p>
            </div>
          )}
        </ScrollArea>

        {/* Footer */}
        {notifications.length > 0 && (
          <div className="border-t p-2">
            <Button
              variant="ghost"
              className="w-full justify-center text-sm"
              asChild
              onClick={() => setOpen(false)}
            >
              <Link href="/notifications">Показать все уведомления</Link>
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
