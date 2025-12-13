'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Bell, CheckCheck, Settings, Trash2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { NotificationItem } from './NotificationItem'
import {
  useNotifications,
  useUnreadCount,
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
  useDeleteAllNotifications,
} from '@/hooks/useNotifications'
import { cn } from '@/lib/utils'

interface NotificationCenterProps {
  className?: string
}

export function NotificationCenter({ className }: NotificationCenterProps) {
  const [open, setOpen] = useState(false)
  const [activeTab, setActiveTab] = useState<'all' | 'unread'>('all')

  const { data: unreadData } = useUnreadCount()
  const unreadCount = unreadData?.count ?? 0

  const { data: allData, isLoading: isLoadingAll } = useNotifications({ limit: 50 })
  const { data: unreadOnlyData, isLoading: isLoadingUnread } = useNotifications({
    is_read: false,
    limit: 50,
  })

  const markAsRead = useMarkAsRead()
  const markAllAsRead = useMarkAllAsRead()
  const deleteNotification = useDeleteNotification()
  const deleteAll = useDeleteAllNotifications()

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

  const handleDeleteAll = async () => {
    try {
      await deleteAll.mutateAsync()
      toast.success('Все уведомления удалены')
    } catch {
      toast.error('Не удалось удалить уведомления')
    }
  }

  const notifications = activeTab === 'all' ? allData?.notifications : unreadOnlyData?.notifications
  const isLoading = activeTab === 'all' ? isLoadingAll : isLoadingUnread

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn('relative', className)}
          aria-label={`Уведомления${unreadCount > 0 ? ` (${unreadCount} непрочитанных)` : ''}`}
        >
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <Badge
              variant="destructive"
              className="absolute -top-1 -right-1 h-5 min-w-5 px-1 flex items-center justify-center text-xs"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent className="w-[420px] p-0" align="end" sideOffset={8}>
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-blue-500 dark:text-blue-400" />
            <h3 className="text-base font-semibold tracking-[-0.006em]">Уведомления</h3>
          </div>
          <div className="flex items-center gap-1">
            {unreadCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleMarkAllAsRead}
                disabled={markAllAsRead.isPending}
                className="h-8 px-2 text-xs text-blue-500 hover:text-blue-600 dark:text-blue-400 dark:hover:text-blue-300"
              >
                <CheckCheck className="h-4 w-4 mr-1" />
                Прочитать все
              </Button>
            )}
            <Button variant="ghost" size="icon" asChild className="h-8 w-8">
              <Link href="/settings/notifications">
                <Settings className="h-4 w-4 text-muted-foreground" />
              </Link>
            </Button>
          </div>
        </div>

        {/* Tabs */}
        <Tabs
          value={activeTab}
          onValueChange={(v) => setActiveTab(v as 'all' | 'unread')}
          className="w-full"
        >
          <div className="px-4 pt-3">
            <TabsList className="w-full grid grid-cols-2 [&_button]:gap-1.5">
              <TabsTrigger value="all" className="text-sm">
                Все
                {(allData?.total_count ?? 0) > 0 && (
                  <Badge variant="secondary" className="ml-1.5 h-5 px-1.5 rounded-full">
                    {allData?.total_count}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger value="unread" className="text-sm">
                Непрочитанные
                {unreadCount > 0 && (
                  <Badge variant="destructive" className="ml-1.5 h-5 px-1.5 rounded-full">
                    {unreadCount}
                  </Badge>
                )}
              </TabsTrigger>
            </TabsList>
          </div>

          <TabsContent value={activeTab} className="mt-0">
            <ScrollArea className="h-[400px]">
              {isLoading ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : notifications && notifications.length > 0 ? (
                <div className="space-y-0 divide-y divide-dashed divide-border px-4">
                  {notifications.map((notification) => (
                    <NotificationItem
                      key={notification.id}
                      notification={notification}
                      onMarkAsRead={handleMarkAsRead}
                      onDelete={handleDelete}
                    />
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center px-4">
                  <div className="rounded-full bg-muted p-4 mb-4">
                    <Bell className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <p className="text-sm font-medium tracking-[-0.006em] text-muted-foreground">
                    {activeTab === 'unread'
                      ? 'Нет непрочитанных уведомлений'
                      : 'У вас пока нет уведомлений'}
                  </p>
                </div>
              )}
            </ScrollArea>
          </TabsContent>
        </Tabs>

        {/* Footer */}
        {notifications && notifications.length > 0 && (
          <div className="flex items-center justify-between p-3 border-t bg-muted/30">
            <Button variant="ghost" size="sm" asChild className="text-xs">
              <Link href="/notifications" onClick={() => setOpen(false)}>
                Смотреть все
              </Link>
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleDeleteAll}
              disabled={deleteAll.isPending}
              className="text-xs text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-3.5 w-3.5 mr-1" />
              Очистить все
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
