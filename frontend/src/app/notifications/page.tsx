'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Bell, CheckCheck, Settings, Trash2, Loader2, Filter } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { NotificationItem } from '@/components/notifications/NotificationItem'
import {
  useNotifications,
  useNotificationStats,
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
  useDeleteAllNotifications,
} from '@/hooks/useNotifications'
import { notificationTypeLabels, notificationPriorityLabels } from '@/types/notification'
import type { NotificationType, NotificationPriority } from '@/types/notification'

export default function NotificationsPage() {
  const [typeFilter, setTypeFilter] = useState<NotificationType | 'all'>('all')
  const [priorityFilter, setPriorityFilter] = useState<NotificationPriority | 'all'>('all')
  const [readFilter, setReadFilter] = useState<'all' | 'read' | 'unread'>('all')

  const { stats, isLoading: isLoadingStats } = useNotificationStats()
  const { data: notificationsData, isLoading: isLoadingNotifications } = useNotifications({
    type: typeFilter !== 'all' ? typeFilter : undefined,
    priority: priorityFilter !== 'all' ? priorityFilter : undefined,
    is_read: readFilter === 'all' ? undefined : readFilter === 'read',
    limit: 100,
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

  const notifications = notificationsData?.notifications ?? []
  const isLoading = isLoadingNotifications

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">Уведомления</h1>
            <p className="text-muted-foreground">Просмотр и управление всеми уведомлениями</p>
          </div>
          <Button variant="outline" asChild>
            <Link href="/settings/notifications">
              <Settings className="h-4 w-4 mr-2" />
              Настройки
            </Link>
          </Button>
        </div>

        {/* Stats Cards */}
        {!isLoadingStats && stats && (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Всего</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{stats.total_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Непрочитанные
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-primary">{stats.unread_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Сегодня</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{stats.today_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">Срочные</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-red-500">{stats.urgent_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  Истёкшие
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-muted-foreground">{stats.expired_count}</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Filters and Actions */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Фильтры:</span>
            </div>

            <Select
              value={typeFilter}
              onValueChange={(v) => setTypeFilter(v as NotificationType | 'all')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder="Тип" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все типы</SelectItem>
                {Object.entries(notificationTypeLabels).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select
              value={priorityFilter}
              onValueChange={(v) => setPriorityFilter(v as NotificationPriority | 'all')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder="Приоритет" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все приоритеты</SelectItem>
                {Object.entries(notificationPriorityLabels).map(([value, label]) => (
                  <SelectItem key={value} value={value}>
                    {label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select
              value={readFilter}
              onValueChange={(v) => setReadFilter(v as 'all' | 'read' | 'unread')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder="Статус" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Все</SelectItem>
                <SelectItem value="unread">Непрочитанные</SelectItem>
                <SelectItem value="read">Прочитанные</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            {(stats?.unread_count ?? 0) > 0 && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleMarkAllAsRead}
                disabled={markAllAsRead.isPending}
              >
                <CheckCheck className="h-4 w-4 mr-2" />
                Прочитать все
              </Button>
            )}

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="text-muted-foreground hover:text-destructive"
                  disabled={notifications.length === 0}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Очистить все
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Удалить все уведомления?</AlertDialogTitle>
                  <AlertDialogDescription>
                    Все уведомления будут безвозвратно удалены. Это действие нельзя отменить.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Отмена</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDeleteAll}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    Удалить все
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>

        {/* Notifications List */}
        <div className="rounded-xl bg-card border border-border p-4">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : notifications.length > 0 ? (
            <div className="space-y-0">
              {notifications.map((notification, index) => (
                <div key={notification.id} className={index > 0 ? 'border-t border-border' : ''}>
                  <NotificationItem
                    notification={notification}
                    onMarkAsRead={handleMarkAsRead}
                    onDelete={handleDelete}
                  />
                </div>
              ))}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <Bell className="h-16 w-16 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium">Нет уведомлений</h3>
              <p className="text-sm text-muted-foreground mt-1">
                {typeFilter !== 'all' || priorityFilter !== 'all' || readFilter !== 'all'
                  ? 'Попробуйте изменить фильтры'
                  : 'У вас пока нет уведомлений'}
              </p>
            </div>
          )}
        </div>

        {/* Pagination info */}
        {notificationsData && notificationsData.total_count > 0 && (
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <p>
              Показано {notifications.length} из {notificationsData.total_count} уведомлений
            </p>
            {notificationsData.total_count > notifications.length && (
              <Badge variant="secondary">
                +{notificationsData.total_count - notifications.length} ещё
              </Badge>
            )}
          </div>
        )}
      </div>
    </AppLayout>
  )
}
