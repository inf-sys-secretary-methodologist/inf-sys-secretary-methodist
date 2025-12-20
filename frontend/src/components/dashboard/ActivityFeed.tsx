'use client'

import { FileText, ClipboardList, Calendar, Megaphone, User } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import type { ActivityItem } from '@/types/dashboard'

interface ActivityFeedProps {
  activities: ActivityItem[]
  title?: string
  className?: string
}

const typeIcons: Record<string, typeof FileText> = {
  document: FileText,
  report: ClipboardList,
  task: ClipboardList,
  event: Calendar,
  announcement: Megaphone,
}

function ActivityItemCard({
  activity,
  typeLabel,
  actionLabel,
  formatRelativeTime,
}: {
  activity: ActivityItem
  typeLabel: string
  actionLabel: string
  formatRelativeTime: (dateString: string) => string
}) {
  const Icon = typeIcons[activity.type] || FileText

  return (
    <div className="flex items-start gap-4 p-4 rounded-lg bg-gray-50 dark:bg-white/5 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gray-200 dark:bg-white/10 text-gray-700 dark:text-gray-300">
        <Icon className="h-5 w-5" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-xs px-2 py-0.5 rounded-full bg-gray-200 dark:bg-white/20 text-gray-600 dark:text-gray-300">
            {typeLabel}
          </span>
          <span className="text-xs text-gray-500 dark:text-gray-500">{actionLabel}</span>
        </div>
        <p className="font-medium text-gray-900 dark:text-white truncate">{activity.title}</p>
        {activity.description && (
          <p className="text-sm text-gray-600 dark:text-gray-400 truncate">
            {activity.description}
          </p>
        )}
        <div className="flex items-center gap-2 mt-2">
          <User className="h-3 w-3 text-gray-500" />
          <span className="text-xs text-gray-500 dark:text-gray-500">{activity.user_name}</span>
          <span className="text-xs text-gray-500">•</span>
          <span className="text-xs text-gray-500 dark:text-gray-500">
            {formatRelativeTime(activity.created_at)}
          </span>
        </div>
      </div>
    </div>
  )
}

export function ActivityFeed({ activities, title, className }: ActivityFeedProps) {
  const t = useTranslations('dashboard.activityFeed')
  const tCommon = useTranslations('common')

  const formatRelativeTime = (dateString: string): string => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return tCommon('time.justNow')
    if (diffMins < 60) return tCommon('time.minutesAgo', { count: diffMins })
    if (diffHours < 24) return tCommon('time.hoursAgo', { count: diffHours })
    if (diffDays < 7) return tCommon('time.daysAgo', { count: diffDays })

    return date.toLocaleDateString(undefined, {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
    })
  }

  const getTypeLabel = (type: string): string => {
    const key = `types.${type}` as const
    try {
      return t(key)
    } catch {
      return type
    }
  }

  const getActionLabel = (action: string): string => {
    const key = `actions.${action}` as const
    try {
      return t(key)
    } catch {
      return action
    }
  }

  return (
    <div
      className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
    >
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={3}
      />
      <div className="relative z-10">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">
          {title || t('title')}
        </h3>
        <div className="space-y-3 max-h-96 overflow-y-auto">
          {activities.length > 0 ? (
            activities.map((activity) => (
              <ActivityItemCard
                key={`${activity.type}-${activity.id}`}
                activity={activity}
                typeLabel={getTypeLabel(activity.type)}
                actionLabel={getActionLabel(activity.action)}
                formatRelativeTime={formatRelativeTime}
              />
            ))
          ) : (
            <p className="text-center text-gray-500 dark:text-gray-400 py-8">{t('empty')}</p>
          )}
        </div>
      </div>
    </div>
  )
}
