export type NotificationType =
  | 'system'
  | 'reminder'
  | 'task'
  | 'document'
  | 'announcement'
  | 'event'

export type NotificationPriority = 'low' | 'normal' | 'high' | 'urgent'

export type NotificationChannel = 'email' | 'push' | 'in_app' | 'telegram' | 'slack'

export interface Notification {
  id: number
  user_id: number
  type: NotificationType
  priority: NotificationPriority
  title: string
  message: string
  link?: string
  image_url?: string
  is_read: boolean
  expires_at?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
  created_at_display?: string
}

export interface NotificationListInput {
  type?: NotificationType
  priority?: NotificationPriority
  is_read?: boolean
  limit?: number
  offset?: number
}

export interface NotificationListOutput {
  notifications: Notification[]
  total_count: number
  unread_count: number
  limit: number
  offset: number
}

export interface UnreadCountOutput {
  count: number
}

export interface NotificationStatsOutput {
  total_count: number
  unread_count: number
  today_count: number
  urgent_count: number
  expired_count: number
}

export interface CreateNotificationInput {
  user_id: number
  type: NotificationType
  priority?: NotificationPriority
  title: string
  message: string
  link?: string
  image_url?: string
  expires_at?: string
  metadata?: Record<string, unknown>
}

export interface CreateBulkNotificationInput {
  user_ids: number[]
  type: NotificationType
  priority?: NotificationPriority
  title: string
  message: string
  link?: string
  image_url?: string
  expires_at?: string
  metadata?: Record<string, unknown>
}

export interface NotificationPreferences {
  id: number
  user_id: number
  email_enabled: boolean
  push_enabled: boolean
  in_app_enabled: boolean
  telegram_enabled: boolean
  slack_enabled: boolean
  quiet_hours_enabled: boolean
  quiet_hours_start?: string
  quiet_hours_end?: string
  timezone: string
  created_at: string
  updated_at: string
}

export interface PreferencesInput {
  email_enabled?: boolean
  push_enabled?: boolean
  in_app_enabled?: boolean
  telegram_enabled?: boolean
  slack_enabled?: boolean
}

export interface ChannelToggleInput {
  channel: NotificationChannel
  enabled: boolean
}

export interface QuietHoursInput {
  enabled: boolean
  start_time?: string
  end_time?: string
  timezone?: string
}

// Labels are now provided via i18n (messages/*.json)
// Use t('notificationTypes.system'), t('notificationPriorities.low'), t('notificationChannels.email'), etc.
