import { apiClient } from '../api'
import type {
  Notification,
  NotificationListInput,
  NotificationListOutput,
  UnreadCountOutput,
  NotificationStatsOutput,
  NotificationPreferences,
  PreferencesInput,
  ChannelToggleInput,
  QuietHoursInput,
} from '@/types/notification'

export const notificationsApi = {
  // Notification operations
  list: async (input?: NotificationListInput): Promise<NotificationListOutput> => {
    const params = new URLSearchParams()
    /* c8 ignore start - Optional filter params */
    if (input?.type) params.append('type', input.type)
    if (input?.priority) params.append('priority', input.priority)
    if (input?.is_read !== undefined) params.append('is_read', String(input.is_read))
    if (input?.limit) params.append('limit', String(input.limit))
    if (input?.offset) params.append('offset', String(input.offset))
    /* c8 ignore stop */

    const query = params.toString()
    return apiClient.get<NotificationListOutput>(`/api/notifications${query ? `?${query}` : ''}`)
  },

  getById: async (id: number): Promise<Notification> => {
    return apiClient.get<Notification>(`/api/notifications/${id}`)
  },

  markAsRead: async (id: number): Promise<{ message: string }> => {
    return apiClient.put<{ message: string }>(`/api/notifications/${id}/read`)
  },

  markAllAsRead: async (): Promise<{ message: string }> => {
    return apiClient.put<{ message: string }>('/api/notifications/read-all')
  },

  delete: async (id: number): Promise<{ message: string }> => {
    return apiClient.delete<{ message: string }>(`/api/notifications/${id}`)
  },

  deleteAll: async (): Promise<{ message: string }> => {
    return apiClient.delete<{ message: string }>('/api/notifications')
  },

  getUnreadCount: async (): Promise<UnreadCountOutput> => {
    return apiClient.get<UnreadCountOutput>('/api/notifications/unread-count')
  },

  getStats: async (): Promise<NotificationStatsOutput> => {
    return apiClient.get<NotificationStatsOutput>('/api/notifications/stats')
  },

  // Preferences operations
  getPreferences: async (): Promise<NotificationPreferences> => {
    return apiClient.get<NotificationPreferences>('/api/notifications/preferences')
  },

  updatePreferences: async (input: PreferencesInput): Promise<NotificationPreferences> => {
    return apiClient.put<NotificationPreferences>('/api/notifications/preferences', input)
  },

  toggleChannel: async (input: ChannelToggleInput): Promise<NotificationPreferences> => {
    return apiClient.put<NotificationPreferences>('/api/notifications/preferences/channel', input)
  },

  updateQuietHours: async (input: QuietHoursInput): Promise<NotificationPreferences> => {
    return apiClient.put<NotificationPreferences>(
      '/api/notifications/preferences/quiet-hours',
      input
    )
  },

  resetPreferences: async (): Promise<NotificationPreferences> => {
    return apiClient.post<NotificationPreferences>('/api/notifications/preferences/reset')
  },

  getTimezones: async (): Promise<{ timezones: string[] }> => {
    return apiClient.get<{ timezones: string[] }>('/api/notifications/timezones')
  },
}
