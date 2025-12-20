'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
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
import type { NotificationType, NotificationPriority } from '@/types/notification'

const NOTIFICATION_TYPE_KEYS = [
  'system',
  'document',
  'calendar',
  'task',
  'integration',
  'user',
] as const
const NOTIFICATION_PRIORITY_KEYS = ['low', 'normal', 'high', 'urgent'] as const

export default function NotificationsPage() {
  const t = useTranslations('notifications')
  const tTypes = useTranslations('notificationTypes')
  const tPriorities = useTranslations('notificationPriorities')
  const tSettings = useTranslations('settings')
  const tCommon = useTranslations('common')
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
      toast.error(t('markError'))
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead.mutateAsync()
      toast.success(t('allMarkedRead'))
    } catch {
      toast.error(t('markAllError'))
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteNotification.mutateAsync(id)
      toast.success(t('deleted'))
    } catch {
      toast.error(t('deleteError'))
    }
  }

  const handleDeleteAll = async () => {
    try {
      await deleteAll.mutateAsync()
      toast.success(t('deleteAllSuccess'))
    } catch {
      toast.error(t('deleteAllError'))
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
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-muted-foreground">{t('subtitle')}</p>
          </div>
          <Button variant="outline" asChild>
            <Link href="/settings/notifications">
              <Settings className="h-4 w-4 mr-2" />
              {tSettings('title')}
            </Link>
          </Button>
        </div>

        {/* Stats Cards */}
        {!isLoadingStats && stats && (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('stats.total')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{stats.total_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('stats.unread')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-primary">{stats.unread_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('stats.today')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{stats.today_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('stats.urgent')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold text-red-500">{stats.urgent_count}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('stats.expired')}
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
              <span className="text-sm text-muted-foreground">{t('filters.title')}:</span>
            </div>

            <Select
              value={typeFilter}
              onValueChange={(v) => setTypeFilter(v as NotificationType | 'all')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder={t('filters.type')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('filters.allTypes')}</SelectItem>
                {NOTIFICATION_TYPE_KEYS.map((key) => (
                  <SelectItem key={key} value={key}>
                    {tTypes(key)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select
              value={priorityFilter}
              onValueChange={(v) => setPriorityFilter(v as NotificationPriority | 'all')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder={t('filters.priority')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('filters.allPriorities')}</SelectItem>
                {NOTIFICATION_PRIORITY_KEYS.map((key) => (
                  <SelectItem key={key} value={key}>
                    {tPriorities(key)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select
              value={readFilter}
              onValueChange={(v) => setReadFilter(v as 'all' | 'read' | 'unread')}
            >
              <SelectTrigger className="w-auto min-w-[130px] h-9">
                <SelectValue placeholder={t('filters.status')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('filters.all')}</SelectItem>
                <SelectItem value="unread">{t('filters.unread')}</SelectItem>
                <SelectItem value="read">{t('filters.read')}</SelectItem>
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
                {t('markAllRead')}
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
                  {t('clearAll')}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{t('deleteAll')}</AlertDialogTitle>
                  <AlertDialogDescription>{t('deleteAllDesc')}</AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>{tCommon('cancel')}</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={handleDeleteAll}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  >
                    {tCommon('delete')}
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
              <h3 className="text-lg font-medium">{t('empty')}</h3>
              <p className="text-sm text-muted-foreground mt-1">
                {typeFilter !== 'all' || priorityFilter !== 'all' || readFilter !== 'all'
                  ? t('tryChangeFilters')
                  : t('noNotifications')}
              </p>
            </div>
          )}
        </div>

        {/* Pagination info */}
        {notificationsData && notificationsData.total_count > 0 && (
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <p>
              {t('pagination.showing', {
                count: notifications.length,
                total: notificationsData.total_count,
              })}
            </p>
            {notificationsData.total_count > notifications.length && (
              <Badge variant="secondary">
                {t('pagination.more', {
                  count: notificationsData.total_count - notifications.length,
                })}
              </Badge>
            )}
          </div>
        )}
      </div>
    </AppLayout>
  )
}
