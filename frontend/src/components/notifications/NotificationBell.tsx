'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { Bell, CheckCheck, Settings, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { NotificationItem } from './NotificationItem'
import {
  useNotifications,
  useUnreadCount,
  useMarkAsRead,
  useMarkAllAsRead,
} from '@/hooks/useNotifications'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface NotificationBellProps {
  className?: string
}

export function NotificationBell({ className }: NotificationBellProps) {
  const [open, setOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<string>('all')
  const t = useTranslations('notificationBell')
  const tNotifications = useTranslations('notifications')

  const { count: unreadCount } = useUnreadCount()
  const { notifications, isLoading } = useNotifications({ limit: 20 })

  const markAsRead = useMarkAsRead()
  const markAllAsRead = useMarkAllAsRead()

  /* c8 ignore start - Notification handlers and filters */
  const handleMarkAsRead = async (id: number) => {
    try {
      await markAsRead.mutateAsync(id)
    } catch {
      toast.error(tNotifications('markError'))
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead.mutateAsync()
      toast.success(tNotifications('allMarkedRead'))
    } catch {
      toast.error(tNotifications('markAllError'))
    }
  }

  // Filter notifications by type
  const getFilteredNotifications = () => {
    switch (activeTab) {
      case 'tasks':
        return notifications.filter((n) => n.type === 'task')
      case 'documents':
        return notifications.filter((n) => n.type === 'document')
      case 'reminders':
        return notifications.filter((n) => n.type === 'reminder' || n.type === 'event')
      default:
        return notifications
    }
  }
  /* c8 ignore stop */

  const filteredNotifications = getFilteredNotifications()

  // Count by type
  const taskCount = notifications.filter((n) => n.type === 'task').length
  const documentCount = notifications.filter((n) => n.type === 'document').length
  const reminderCount = notifications.filter(
    (n) => n.type === 'reminder' || n.type === 'event'
  ).length

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            'relative inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 border-gray-300 bg-white text-gray-900 transition-all duration-200 hover:bg-gray-100 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800 shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            !open && 'hover:scale-105 hover:shadow-lg active:scale-95',
            className
          )}
          aria-label={t('ariaLabel', { count: unreadCount })}
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

      <PopoverContent className="w-[420px] p-0" align="end" sideOffset={8}>
        {/* Header */}
        <div className="flex flex-col gap-4 p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-base leading-none font-semibold tracking-[-0.006em]">
              {tNotifications('title')}
            </h3>
            <div className="flex items-center gap-2">
              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8"
                  onClick={handleMarkAllAsRead}
                  disabled={markAllAsRead.isPending}
                  aria-label={tNotifications('markAllRead')}
                >
                  <CheckCheck className="size-4 text-muted-foreground" />
                </Button>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="size-8"
                asChild
                aria-label={t('settingsAriaLabel')}
                onClick={() => setOpen(false)}
              >
                <Link href="/settings/notifications">
                  <Settings className="size-4 text-muted-foreground" />
                </Link>
              </Button>
            </div>
          </div>

          {/* Tabs */}
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="w-full justify-start gap-0.5 bg-transparent p-0 h-auto flex-wrap">
              <TabsTrigger
                value="all"
                className="gap-1 data-[state=active]:bg-muted data-[state=active]:shadow-none rounded-full px-2.5 py-1 text-xs"
              >
                {t('tabs.all')}
                <Badge
                  variant="secondary"
                  className="size-4 rounded-full p-0 justify-center text-[10px]"
                >
                  {notifications.length}
                </Badge>
              </TabsTrigger>
              <TabsTrigger
                value="tasks"
                className="gap-1 data-[state=active]:bg-muted data-[state=active]:shadow-none rounded-full px-2.5 py-1 text-xs"
              >
                {t('tabs.tasks')}
                <Badge
                  variant="secondary"
                  className="size-4 rounded-full p-0 justify-center text-[10px]"
                >
                  {taskCount}
                </Badge>
              </TabsTrigger>
              <TabsTrigger
                value="documents"
                className="gap-1 data-[state=active]:bg-muted data-[state=active]:shadow-none rounded-full px-2.5 py-1 text-xs"
              >
                {t('tabs.documents')}
                <Badge
                  variant="secondary"
                  className="size-4 rounded-full p-0 justify-center text-[10px]"
                >
                  {documentCount}
                </Badge>
              </TabsTrigger>
              <TabsTrigger
                value="reminders"
                className="gap-1 data-[state=active]:bg-muted data-[state=active]:shadow-none rounded-full px-2.5 py-1 text-xs"
              >
                {t('tabs.events')}
                <Badge
                  variant="secondary"
                  className="size-4 rounded-full p-0 justify-center text-[10px]"
                >
                  {reminderCount}
                </Badge>
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        {/* Notifications list */}
        <ScrollArea className="h-[350px]">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : filteredNotifications.length > 0 ? (
            <div className="space-y-1 px-2">
              {filteredNotifications.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onMarkAsRead={handleMarkAsRead}
                  compact
                />
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center space-y-2.5 py-12 text-center px-4">
              <div className="rounded-full bg-muted p-4">
                <Bell className="h-6 w-6 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium tracking-[-0.006em] text-muted-foreground">
                {tNotifications('empty')}
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
              <Link href="/notifications">{t('showAll')}</Link>
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
