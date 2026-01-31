/**
 * @jest-environment jsdom
 */

import { renderHook, waitFor, act } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import { usePushNotifications, useCanEnablePush } from '../usePushNotifications'

// Mock the push-notifications module
jest.mock('@/lib/push-notifications', () => ({
  isPushSupported: jest.fn(),
  getPermissionStatus: jest.fn(),
  subscribeToPush: jest.fn(),
  unsubscribeFromPush: jest.fn(),
  isSubscribed: jest.fn(),
  sendTestNotification: jest.fn(),
  deleteSubscription: jest.fn(),
}))

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

import {
  isPushSupported,
  getPermissionStatus,
  subscribeToPush,
  unsubscribeFromPush,
  isSubscribed,
  sendTestNotification,
  deleteSubscription,
} from '@/lib/push-notifications'
import { apiClient } from '@/lib/api'

const mockedIsPushSupported = jest.mocked(isPushSupported)
const mockedGetPermissionStatus = jest.mocked(getPermissionStatus)
const mockedSubscribeToPush = jest.mocked(subscribeToPush)
const mockedUnsubscribeFromPush = jest.mocked(unsubscribeFromPush)
const mockedIsSubscribed = jest.mocked(isSubscribed)
const mockedSendTestNotification = jest.mocked(sendTestNotification)
const mockedDeleteSubscription = jest.mocked(deleteSubscription)
const mockedApiClient = jest.mocked(apiClient)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('usePushNotifications', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockedIsPushSupported.mockReturnValue(true)
    mockedGetPermissionStatus.mockReturnValue('default')
    mockedIsSubscribed.mockResolvedValue(false)

    // Mock Notification API for subscribe error handling
    Object.defineProperty(window, 'Notification', {
      value: { permission: 'default' },
      writable: true,
      configurable: true,
    })
  })

  describe('initial state', () => {
    it('returns correct initial state when supported', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: false,
        subscriptions: [],
        total_devices: 0,
      })

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      expect(result.current.isSupported).toBe(true)
      expect(result.current.permission).toBe('default')

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })
    })

    it('returns correct state when not supported', async () => {
      mockedIsPushSupported.mockReturnValue(false)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      expect(result.current.isSupported).toBe(false)
      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('server status', () => {
    it('fetches and displays server status', async () => {
      const mockStatus = {
        is_enabled: true,
        subscriptions: [
          { id: 1, device_name: 'Chrome', is_active: true, created_at: '2024-01-01' },
        ],
        total_devices: 1,
      }
      mockedApiClient.get.mockResolvedValue(mockStatus)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isEnabled).toBe(true)
      })

      expect(result.current.subscriptions).toHaveLength(1)
      expect(result.current.totalDevices).toBe(1)
    })

    it('handles fetch error', async () => {
      mockedApiClient.get.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeTruthy()
      })
    })
  })

  describe('subscribe', () => {
    it('subscribes successfully', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: false,
        subscriptions: [],
        total_devices: 0,
      })
      mockedSubscribeToPush.mockResolvedValue({
        id: 1,
        device_name: 'Test',
        is_active: true,
        created_at: '2024-01-01',
      })

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.subscribe()
      })

      expect(mockedSubscribeToPush).toHaveBeenCalled()
    })

    it('handles subscribe error', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: false,
        subscriptions: [],
        total_devices: 0,
      })
      mockedSubscribeToPush.mockRejectedValue(new Error('Permission denied'))

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.subscribe()
      })

      expect(result.current.error).toBeTruthy()
    })

    it('returns null when not supported', async () => {
      mockedIsPushSupported.mockReturnValue(false)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      const subscription = await result.current.subscribe()
      expect(subscription).toBeNull()
    })
  })

  describe('unsubscribe', () => {
    it('unsubscribes successfully', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: true,
        subscriptions: [{ id: 1, is_active: true }],
        total_devices: 1,
      })
      mockedUnsubscribeFromPush.mockResolvedValue(undefined)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.unsubscribe()
      })

      expect(mockedUnsubscribeFromPush).toHaveBeenCalled()
    })

    it('handles unsubscribe error', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: true,
        subscriptions: [],
        total_devices: 0,
      })
      mockedUnsubscribeFromPush.mockRejectedValue(new Error('Failed'))

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await expect(
        act(async () => {
          await result.current.unsubscribe()
        })
      ).rejects.toThrow()
    })
  })

  describe('removeSubscription', () => {
    it('removes subscription successfully', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: true,
        subscriptions: [{ id: 1, is_active: true }],
        total_devices: 1,
      })
      mockedDeleteSubscription.mockResolvedValue(undefined)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.removeSubscription(1)
      })

      expect(mockedDeleteSubscription).toHaveBeenCalledWith(1)
    })
  })

  describe('testNotification', () => {
    it('sends test notification successfully', async () => {
      mockedApiClient.get.mockResolvedValue({
        is_enabled: true,
        subscriptions: [],
        total_devices: 0,
      })
      mockedSendTestNotification.mockResolvedValue(undefined)

      const { result } = renderHook(() => usePushNotifications(), { wrapper })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      await act(async () => {
        await result.current.testNotification('Test', 'Message')
      })

      expect(mockedSendTestNotification).toHaveBeenCalledWith('Test', 'Message')
    })
  })
})

describe('useCanEnablePush', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('returns true when supported and not denied', () => {
    mockedIsPushSupported.mockReturnValue(true)
    mockedGetPermissionStatus.mockReturnValue('default')

    const { result } = renderHook(() => useCanEnablePush())

    expect(result.current).toBe(true)
  })

  it('returns true when supported and granted', () => {
    mockedIsPushSupported.mockReturnValue(true)
    mockedGetPermissionStatus.mockReturnValue('granted')

    const { result } = renderHook(() => useCanEnablePush())

    expect(result.current).toBe(true)
  })

  it('returns false when not supported', () => {
    mockedIsPushSupported.mockReturnValue(false)
    mockedGetPermissionStatus.mockReturnValue('unsupported')

    const { result } = renderHook(() => useCanEnablePush())

    expect(result.current).toBe(false)
  })

  it('returns false when permission is denied', () => {
    mockedIsPushSupported.mockReturnValue(true)
    mockedGetPermissionStatus.mockReturnValue('denied')

    const { result } = renderHook(() => useCanEnablePush())

    expect(result.current).toBe(false)
  })
})
