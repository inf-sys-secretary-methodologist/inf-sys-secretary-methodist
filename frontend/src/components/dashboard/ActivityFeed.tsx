'use client'

import { FileText, ClipboardList, Calendar, Megaphone, User } from 'lucide-react'
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

const typeLabels: Record<string, string> = {
  document: 'Документ',
  report: 'Отчет',
  task: 'Задача',
  event: 'Мероприятие',
  announcement: 'Объявление',
}

const actionLabels: Record<string, string> = {
  created: 'создан(о)',
  updated: 'обновлен(о)',
  deleted: 'удален(о)',
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'только что'
  if (diffMins < 60) return `${diffMins} мин. назад`
  if (diffHours < 24) return `${diffHours} ч. назад`
  if (diffDays < 7) return `${diffDays} дн. назад`

  return date.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

function ActivityItemCard({ activity }: { activity: ActivityItem }) {
  const Icon = typeIcons[activity.type] || FileText
  const typeLabel = typeLabels[activity.type] || activity.type
  const actionLabel = actionLabels[activity.action] || activity.action

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
          <User className="h-3 w-3 text-gray-400" />
          <span className="text-xs text-gray-500 dark:text-gray-500">{activity.user_name}</span>
          <span className="text-xs text-gray-400">•</span>
          <span className="text-xs text-gray-500 dark:text-gray-500">
            {formatRelativeTime(activity.created_at)}
          </span>
        </div>
      </div>
    </div>
  )
}

export function ActivityFeed({
  activities,
  title = 'Последние действия',
  className,
}: ActivityFeedProps) {
  return (
    <div
      className={`relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 ${className}`}
    >
      <div className="relative z-10">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">{title}</h3>
        <div className="space-y-3 max-h-96 overflow-y-auto">
          {activities.length > 0 ? (
            activities.map((activity) => (
              <ActivityItemCard key={`${activity.type}-${activity.id}`} activity={activity} />
            ))
          ) : (
            <p className="text-center text-gray-500 dark:text-gray-400 py-8">
              Нет активности для отображения
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
