import { renderHook, act, waitFor } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useNotifications,
  useNotification,
  useUnreadCount,
  useNotificationStats,
  useMarkAsRead,
  useMarkAllAsRead,
  useDeleteNotification,
  useDeleteAllNotifications,
  useNotificationPreferences,
  useUpdatePreferences,
  useToggleChannel,
  useUpdateQuietHours,
  useResetPreferences,
  useTimezones,
} from '../useNotifications'
import { apiClient } from '@/lib/api'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    put: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useNotifications', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useNotifications hook', () => {
    it('returns empty notifications initially', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: {
          notifications: [],
          total_count: 0,
          unread_count: 0,
        },
      })

      const { result } = renderHook(() => useNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.notifications).toEqual([])
      expect(result.current.totalCount).toBe(0)
      expect(result.current.unreadCount).toBe(0)
    })

    it('returns notifications from API', async () => {
      const mockNotifications = [
        { id: 1, title: 'Test 1', type: 'info', is_read: false },
        { id: 2, title: 'Test 2', type: 'warning', is_read: true },
      ]

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: {
          notifications: mockNotifications,
          total_count: 2,
          unread_count: 1,
        },
      })

      const { result } = renderHook(() => useNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.notifications).toHaveLength(2)
      })

      expect(result.current.totalCount).toBe(2)
      expect(result.current.unreadCount).toBe(1)
    })

    it('handles non-wrapped API response', async () => {
      mockedApiClient.get.mockResolvedValue({
        notifications: [{ id: 1, title: 'Direct' }],
        total_count: 1,
        unread_count: 0,
      })

      const { result } = renderHook(() => useNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.notifications).toHaveLength(1)
      })
    })

    it('passes filter parameters to API', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { notifications: [], total_count: 0, unread_count: 0 },
      })

      renderHook(
        () =>
          useNotifications({
            type: 'system',
            priority: 'high',
            is_read: false,
            limit: 10,
            offset: 5,
          }),
        { wrapper }
      )

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('type=system'))
      })
    })
  })

  describe('useNotification hook', () => {
    it('fetches single notification by ID', async () => {
      const mockNotification = { id: 1, title: 'Single', type: 'info' }
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockNotification,
      })

      const { result } = renderHook(() => useNotification(1), { wrapper })

      await waitFor(() => {
        expect(result.current.notification).toEqual(mockNotification)
      })
    })

    it('does not fetch when ID is 0', () => {
      renderHook(() => useNotification(0), { wrapper })

      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useUnreadCount hook', () => {
    it('returns unread count', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { count: 5 },
      })

      const { result } = renderHook(() => useUnreadCount(), { wrapper })

      await waitFor(() => {
        expect(result.current.count).toBe(5)
      })
    })

    it('returns 0 when no data', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: null,
      })

      const { result } = renderHook(() => useUnreadCount(), { wrapper })

      expect(result.current.count).toBe(0)
    })
  })

  describe('useNotificationStats hook', () => {
    it('returns notification stats', async () => {
      const mockStats = {
        total: 100,
        unread: 10,
        by_type: { info: 50, warning: 30, error: 20 },
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockStats,
      })

      const { result } = renderHook(() => useNotificationStats(), { wrapper })

      await waitFor(() => {
        expect(result.current.stats).toEqual(mockStats)
      })
    })
  })

  describe('useMarkAsRead hook', () => {
    it('marks notification as read', async () => {
      mockedApiClient.put.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useMarkAsRead(), { wrapper })

      expect(result.current.isPending).toBe(false)

      await act(async () => {
        await result.current.mutateAsync(1)
      })

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/1/read')
    })
  })

  describe('useMarkAllAsRead hook', () => {
    it('marks all notifications as read', async () => {
      mockedApiClient.put.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useMarkAllAsRead(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/read-all')
    })
  })

  describe('useDeleteNotification hook', () => {
    it('deletes a notification', async () => {
      mockedApiClient.delete.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useDeleteNotification(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1)
      })

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/notifications/1')
    })
  })

  describe('useDeleteAllNotifications hook', () => {
    it('deletes all notifications', async () => {
      mockedApiClient.delete.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useDeleteAllNotifications(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/notifications')
    })
  })

  describe('useNotificationPreferences hook', () => {
    it('returns notification preferences', async () => {
      const mockPrefs = {
        email_enabled: true,
        push_enabled: false,
        quiet_hours_enabled: true,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockPrefs,
      })

      const { result } = renderHook(() => useNotificationPreferences(), { wrapper })

      await waitFor(() => {
        expect(result.current.data).toEqual(mockPrefs)
      })
    })
  })

  describe('useUpdatePreferences hook', () => {
    it('updates notification preferences', async () => {
      mockedApiClient.put.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useUpdatePreferences(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync({ email_enabled: false })
      })

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/preferences', {
        email_enabled: false,
      })
    })
  })

  describe('useToggleChannel hook', () => {
    it('toggles notification channel', async () => {
      mockedApiClient.put.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useToggleChannel(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync({ channel: 'email', enabled: true })
      })

      expect(mockedApiClient.put).toHaveBeenCalledWith('/api/notifications/preferences/channel', {
        channel: 'email',
        enabled: true,
      })
    })
  })

  describe('useUpdateQuietHours hook', () => {
    it('updates quiet hours', async () => {
      mockedApiClient.put.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useUpdateQuietHours(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync({
          enabled: true,
          start_time: '22:00',
          end_time: '08:00',
        })
      })

      expect(mockedApiClient.put).toHaveBeenCalledWith(
        '/api/notifications/preferences/quiet-hours',
        { enabled: true, start_time: '22:00', end_time: '08:00' }
      )
    })
  })

  describe('useResetPreferences hook', () => {
    it('resets preferences to defaults', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useResetPreferences(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync()
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/notifications/preferences/reset')
    })
  })

  describe('useTimezones hook', () => {
    it('returns available timezones', async () => {
      const mockTimezones = {
        timezones: ['UTC', 'Europe/Moscow', 'America/New_York'],
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockTimezones,
      })

      const { result } = renderHook(() => useTimezones(), { wrapper })

      await waitFor(() => {
        expect(result.current.data).toEqual(mockTimezones)
      })
    })
  })
})
