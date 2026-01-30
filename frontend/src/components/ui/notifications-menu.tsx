'use client'

import React from 'react'
import { CheckCheck, Settings, Bell, Download } from 'lucide-react'
import { useTranslations } from 'next-intl'

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

export type NotificationMenuType = {
  id: number
  type: string
  user: {
    name: string
    avatar: string
    fallback: string
  }
  action: string
  target?: string
  content?: string
  timestamp: string
  timeAgo: string
  isRead: boolean
  hasActions?: boolean
  file?: {
    name: string
    size: string
    type: string
  }
}

interface NotificationItemProps {
  notification: NotificationMenuType
  onAccept?: (id: number) => void
  onDecline?: (id: number) => void
  onDownload?: (id: number) => void
}

function NotificationItem({
  notification,
  onAccept,
  onDecline,
  onDownload,
}: NotificationItemProps) {
  const t = useTranslations('notificationsMenu')
  return (
    <div className="w-full py-4 first:pt-0 last:pb-0">
      <div className="flex gap-3">
        <Avatar className="size-11">
          <AvatarImage
            /* c8 ignore next - Fallback avatar */
            src={notification.user.avatar || '/placeholder.svg'}
            alt={`${notification.user.name}`}
            className="object-cover ring-1 ring-border"
          />
          <AvatarFallback>{notification.user.fallback}</AvatarFallback>
        </Avatar>

        <div className="flex flex-1 flex-col space-y-2">
          <div className="w-full items-start">
            <div>
              <div className="flex items-center justify-between gap-2">
                <div className="text-sm">
                  <span className="font-medium">{notification.user.name}</span>
                  <span className="text-muted-foreground"> {notification.action} </span>
                  {notification.target && (
                    <span className="font-medium">{notification.target}</span>
                  )}
                </div>
                {!notification.isRead && (
                  <div className="size-1.5 rounded-full bg-emerald-500"></div>
                )}
              </div>
              <div className="flex items-center justify-between gap-2">
                <div className="mt-0.5 text-xs text-muted-foreground">{notification.timestamp}</div>
                <div className="text-xs text-muted-foreground">{notification.timeAgo}</div>
              </div>
            </div>
          </div>

          {notification.content && (
            <div className="rounded-lg bg-muted p-2.5 text-sm tracking-[-0.006em]">
              {notification.content}
            </div>
          )}

          {notification.file && (
            <div className="flex items-center gap-2 rounded-lg bg-muted p-2">
              <svg
                width="34"
                height="34"
                viewBox="0 0 40 40"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                className="relative shrink-0"
              >
                <path
                  d="M30 39.25H10C7.10051 39.25 4.75 36.8995 4.75 34V6C4.75 3.10051 7.10051 0.75 10 0.75H20.5147C21.9071 0.75 23.2425 1.30312 24.227 2.28769L33.7123 11.773C34.6969 12.7575 35.25 14.0929 35.25 15.4853V34C35.25 36.8995 32.8995 39.25 30 39.25Z"
                  className="fill-white stroke-border dark:fill-card/70"
                  strokeWidth="1.5"
                />
                <path
                  d="M23 1V9C23 11.2091 24.7909 13 27 13H35"
                  className="stroke-border dark:fill-muted-foreground"
                  strokeWidth="1.5"
                />
                <foreignObject x="0" y="0" width="40" height="40">
                  <div className="absolute bottom-1.5 left-0 flex h-4 items-center rounded bg-primary px-[3px] py-0.5 text-[11px] font-semibold leading-none text-white dark:bg-muted">
                    {notification.file.type}
                  </div>
                </foreignObject>
              </svg>
              <div className="flex-1">
                <div className="text-sm font-medium">{notification.file.name}</div>
                <div className="text-xs text-muted-foreground">
                  {notification.file.type} • {notification.file.size}
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="size-8"
                onClick={() => onDownload?.(notification.id)}
                aria-label={t('downloadFile')}
              >
                <Download className="size-4" />
              </Button>
            </div>
          )}

          {notification.hasActions && (
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="h-7 text-xs"
                onClick={() => onDecline?.(notification.id)}
              >
                {t('reject')}
              </Button>
              <Button size="sm" className="h-7 text-xs" onClick={() => onAccept?.(notification.id)}>
                {t('accept')}
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export interface NotificationsMenuProps {
  notifications: NotificationMenuType[]
  onMarkAllRead?: () => void
  onOpenSettings?: () => void
  onAccept?: (id: number) => void
  onDecline?: (id: number) => void
  onDownload?: (id: number) => void
  className?: string
}

export function NotificationsMenu({
  notifications,
  onMarkAllRead,
  onOpenSettings,
  onAccept,
  onDecline,
  onDownload,
  className,
}: NotificationsMenuProps) {
  const t = useTranslations('notificationsMenu')
  const [activeTab, setActiveTab] = React.useState<string>('all')

  const unreadCount = notifications.filter((n) => !n.isRead).length
  const mentionCount = notifications.filter((n) => n.type === 'mention').length

  const getFilteredNotifications = () => {
    switch (activeTab) {
      case 'unread':
        return notifications.filter((n) => !n.isRead)
      case 'mentions':
        return notifications.filter((n) => n.type === 'mention')
      default:
        return notifications
    }
  }

  const filteredNotifications = getFilteredNotifications()

  return (
    <Card
      className={`flex w-full max-w-[520px] flex-col gap-6 p-4 shadow-none md:p-8 ${className || ''}`}
    >
      <CardHeader className="p-0">
        <div className="flex items-center justify-between">
          <h3 className="text-base font-semibold leading-none tracking-[-0.006em]">{t('title')}</h3>
          <div className="flex items-center gap-2">
            <Button
              className="size-8"
              variant="ghost"
              size="icon"
              onClick={onMarkAllRead}
              aria-label={t('markAllRead')}
            >
              <CheckCheck className="size-4 text-muted-foreground" />
            </Button>
            <Button
              className="size-8"
              variant="ghost"
              size="icon"
              onClick={onOpenSettings}
              aria-label={t('settings')}
            >
              <Settings className="size-4 text-muted-foreground" />
            </Button>
          </div>
        </div>

        <Tabs
          value={activeTab}
          onValueChange={setActiveTab}
          className="w-full flex-col justify-start"
        >
          <div className="flex items-center justify-between">
            <TabsList className="[&_button]:gap-1.5">
              <TabsTrigger value="all">
                {t('tabs.all')}
                <Badge variant="secondary" className="size-5 rounded-full p-0 text-xs">
                  {notifications.length}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="unread">
                {t('tabs.unread')}
                <Badge variant="secondary" className="size-5 rounded-full p-0 text-xs">
                  {unreadCount}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value="mentions">
                {t('tabs.mentions')}
                <Badge variant="secondary" className="size-5 rounded-full p-0 text-xs">
                  {mentionCount}
                </Badge>
              </TabsTrigger>
            </TabsList>
          </div>
        </Tabs>
      </CardHeader>

      <CardContent className="h-full p-0">
        <div className="space-y-0 divide-y divide-dashed divide-border">
          {filteredNotifications.length > 0 ? (
            filteredNotifications.map((notification) => (
              <NotificationItem
                key={notification.id}
                notification={notification}
                onAccept={onAccept}
                onDecline={onDecline}
                onDownload={onDownload}
              />
            ))
          ) : (
            <div className="flex flex-col items-center justify-center space-y-2.5 py-12 text-center">
              <div className="rounded-full bg-muted p-4">
                <Bell className="size-6 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium tracking-[-0.006em] text-muted-foreground">
                {t('empty')}
              </p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export { NotificationItem }
