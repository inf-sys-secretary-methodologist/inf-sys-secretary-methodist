import { notificationsApi } from '../notifications'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('notificationsApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('list', () => {
    it('fetches notifications without filters', async () => {
      const mockResponse = { notifications: [], total: 0 }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await notificationsApi.list()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications')
    })

    it('fetches notifications with filters', async () => {
      const mockResponse = { notifications: [], total: 0 }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await notificationsApi.list({
        type: 'message',
        priority: 'high',
        is_read: false,
        limit: 10,
        offset: 0,
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('/api/notifications?')
      )
    })

    it('fetches only unread notifications', async () => {
      const mockResponse = { notifications: [], total: 0 }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await notificationsApi.list({ is_read: false })

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications?is_read=false')
    })
  })

  describe('getById', () => {
    it('fetches single notification', async () => {
      const mockNotification = { id: 1, title: 'Test', message: 'Test message' }
      mockedApiClient.get.mockResolvedValue(mockNotification)

      const result = await notificationsApi.getById(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications/1')
      expect(result).toEqual(mockNotification)
    })
  })

  describe('markAsRead', () => {
    it('marks notification as read', async () => {
      mockedApiClient.put.mockResolvedValue({ message: 'Marked as read' })

      await notificationsApi.markAsRead(1)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/1/read')
    })
  })

  describe('markAllAsRead', () => {
    it('marks all notifications as read', async () => {
      mockedApiClient.put.mockResolvedValue({ message: 'All marked as read' })

      await notificationsApi.markAllAsRead()

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/read-all')
    })
  })

  describe('delete', () => {
    it('deletes notification', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'Deleted' })

      await notificationsApi.delete(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/notifications/1')
    })
  })

  describe('deleteAll', () => {
    it('deletes all notifications', async () => {
      mockedApiClient.delete.mockResolvedValue({ message: 'All deleted' })

      await notificationsApi.deleteAll()

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/notifications')
    })
  })

  describe('getUnreadCount', () => {
    it('fetches unread count', async () => {
      mockedApiClient.get.mockResolvedValue({ count: 5 })

      const result = await notificationsApi.getUnreadCount()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications/unread-count')
      expect(result).toEqual({ count: 5 })
    })
  })

  describe('getStats', () => {
    it('fetches notification stats', async () => {
      const mockStats = { total: 100, unread: 5, by_type: {} }
      mockedApiClient.get.mockResolvedValue(mockStats)

      const result = await notificationsApi.getStats()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications/stats')
      expect(result).toEqual(mockStats)
    })
  })

  describe('getPreferences', () => {
    it('fetches notification preferences', async () => {
      const mockPrefs = { email_enabled: true, in_app_enabled: true }
      mockedApiClient.get.mockResolvedValue(mockPrefs)

      const result = await notificationsApi.getPreferences()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications/preferences')
      expect(result).toEqual(mockPrefs)
    })
  })

  describe('updatePreferences', () => {
    it('updates notification preferences', async () => {
      const input = { email_enabled: false }
      mockedApiClient.put.mockResolvedValue({ ...input })

      await notificationsApi.updatePreferences(input)

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/preferences', input)
    })
  })

  describe('toggleChannel', () => {
    it('toggles notification channel', async () => {
      const input = { channel: 'email', enabled: false }
      mockedApiClient.put.mockResolvedValue({ email_enabled: false })

      await notificationsApi.toggleChannel(input)

      expect(mockedApiClient.put).toHaveBeenCalledWith(
        '/api/notifications/preferences/channel',
        input
      )
    })
  })

  describe('updateQuietHours', () => {
    it('updates quiet hours', async () => {
      const input = { start: '22:00', end: '08:00', enabled: true }
      mockedApiClient.put.mockResolvedValue({ quiet_hours: input })

      await notificationsApi.updateQuietHours(input)

      expect(mockedApiClient.put).toHaveBeenCalledWith(
        '/api/notifications/preferences/quiet-hours',
        input
      )
    })
  })

  describe('resetPreferences', () => {
    it('resets preferences to defaults', async () => {
      const defaultPrefs = { email_enabled: true, in_app_enabled: true }
      mockedApiClient.post.mockResolvedValue(defaultPrefs)

      const result = await notificationsApi.resetPreferences()

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/notifications/preferences/reset')
      expect(result).toEqual(defaultPrefs)
    })
  })

  describe('getTimezones', () => {
    it('fetches available timezones', async () => {
      const mockTimezones = { timezones: ['UTC', 'America/New_York', 'Europe/Moscow'] }
      mockedApiClient.get.mockResolvedValue(mockTimezones)

      const result = await notificationsApi.getTimezones()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/notifications/timezones')
      expect(result).toEqual(mockTimezones)
    })
  })
})
