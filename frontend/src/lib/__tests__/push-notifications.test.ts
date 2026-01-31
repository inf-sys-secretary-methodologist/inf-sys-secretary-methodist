/**
 * @jest-environment jsdom
 */

// Mock apiClient
jest.mock('../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
  },
}))

import { apiClient } from '../api'
import {
  isPushSupported,
  getPermissionStatus,
  requestPermission,
  subscribeToPush,
  unsubscribeFromPush,
  getCurrentSubscription,
  isSubscribed,
  getPushStatus,
  deleteSubscription,
  sendTestNotification,
} from '../push-notifications'

// Mock navigator
const mockNavigator = {
  serviceWorker: {
    ready: Promise.resolve({
      pushManager: {
        subscribe: jest.fn(),
        getSubscription: jest.fn(),
      },
    }),
  },
  userAgent: 'Mozilla/5.0 Test Browser',
}

// Mock Notification
const mockNotification = {
  permission: 'default' as NotificationPermission,
  requestPermission: jest.fn(),
}

describe('push-notifications', () => {
  beforeEach(() => {
    jest.clearAllMocks()

    // Reset navigator mock
    Object.defineProperty(window, 'navigator', {
      value: mockNavigator,
      writable: true,
      configurable: true,
    })

    // Reset Notification mock
    Object.defineProperty(window, 'Notification', {
      value: mockNotification,
      writable: true,
      configurable: true,
    })

    // Reset PushManager
    Object.defineProperty(window, 'PushManager', {
      value: {},
      writable: true,
      configurable: true,
    })

    mockNotification.permission = 'default'
  })

  describe('isPushSupported', () => {
    it('returns true when all requirements are met', () => {
      expect(isPushSupported()).toBe(true)
    })

    it('returns false when serviceWorker is not available', () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })
      expect(isPushSupported()).toBe(false)
    })

    it('returns false when PushManager is not available', () => {
      // @ts-expect-error - Intentionally deleting for test
      delete window.PushManager
      expect(isPushSupported()).toBe(false)
    })

    it('returns false when Notification is not available', () => {
      // @ts-expect-error - Intentionally deleting for test
      delete window.Notification
      expect(isPushSupported()).toBe(false)
    })
  })

  describe('getPermissionStatus', () => {
    it('returns current permission status', () => {
      mockNotification.permission = 'granted'
      expect(getPermissionStatus()).toBe('granted')
    })

    it('returns unsupported when push is not supported', () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })
      expect(getPermissionStatus()).toBe('unsupported')
    })
  })

  describe('requestPermission', () => {
    it('requests permission and returns result', async () => {
      mockNotification.requestPermission = jest.fn().mockResolvedValue('granted')

      const result = await requestPermission()
      expect(result).toBe('granted')
      expect(mockNotification.requestPermission).toHaveBeenCalled()
    })

    it('throws when push is not supported', async () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })

      await expect(requestPermission()).rejects.toThrow('Push notifications are not supported')
    })
  })

  describe('getCurrentSubscription', () => {
    it('returns current subscription', async () => {
      const mockSubscription = { endpoint: 'https://push.example.com' }
      const mockPushManager = {
        getSubscription: jest.fn().mockResolvedValue(mockSubscription),
      }
      const mockRegistration = {
        pushManager: mockPushManager,
      }

      Object.defineProperty(window.navigator, 'serviceWorker', {
        value: {
          ready: Promise.resolve(mockRegistration),
        },
        writable: true,
        configurable: true,
      })

      const result = await getCurrentSubscription()
      expect(result).toBe(mockSubscription)
    })

    it('returns null when push is not supported', async () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })

      const result = await getCurrentSubscription()
      expect(result).toBeNull()
    })
  })

  describe('isSubscribed', () => {
    it('returns true when subscribed', async () => {
      const mockSubscription = { endpoint: 'https://push.example.com' }
      const mockPushManager = {
        getSubscription: jest.fn().mockResolvedValue(mockSubscription),
      }
      const mockRegistration = {
        pushManager: mockPushManager,
      }

      Object.defineProperty(window.navigator, 'serviceWorker', {
        value: {
          ready: Promise.resolve(mockRegistration),
        },
        writable: true,
        configurable: true,
      })

      const result = await isSubscribed()
      expect(result).toBe(true)
    })

    it('returns false when not subscribed', async () => {
      const mockPushManager = {
        getSubscription: jest.fn().mockResolvedValue(null),
      }
      const mockRegistration = {
        pushManager: mockPushManager,
      }

      Object.defineProperty(window.navigator, 'serviceWorker', {
        value: {
          ready: Promise.resolve(mockRegistration),
        },
        writable: true,
        configurable: true,
      })

      const result = await isSubscribed()
      expect(result).toBe(false)
    })
  })

  describe('getPushStatus', () => {
    it('fetches push status from API', async () => {
      const mockStatus = {
        is_enabled: true,
        subscriptions: [],
        total_devices: 1,
      }
      ;(apiClient.get as jest.Mock).mockResolvedValue(mockStatus)

      const result = await getPushStatus()
      expect(result).toEqual(mockStatus)
      expect(apiClient.get).toHaveBeenCalledWith('/api/notifications/push/status')
    })
  })

  describe('deleteSubscription', () => {
    it('deletes subscription via API', async () => {
      ;(apiClient.delete as jest.Mock).mockResolvedValue({})

      await deleteSubscription(123)
      expect(apiClient.delete).toHaveBeenCalledWith('/api/notifications/push/subscriptions/123')
    })
  })

  describe('sendTestNotification', () => {
    it('sends test notification with custom message', async () => {
      ;(apiClient.post as jest.Mock).mockResolvedValue({})

      await sendTestNotification('Custom Title', 'Custom Message')
      expect(apiClient.post).toHaveBeenCalledWith('/api/notifications/push/test', {
        title: 'Custom Title',
        message: 'Custom Message',
      })
    })

    it('sends test notification with default message', async () => {
      ;(apiClient.post as jest.Mock).mockResolvedValue({})

      await sendTestNotification()
      expect(apiClient.post).toHaveBeenCalledWith('/api/notifications/push/test', {
        title: 'Test Notification',
        message: 'This is a test push notification.',
      })
    })
  })

  describe('subscribeToPush', () => {
    it('throws when push is not supported', async () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })

      await expect(subscribeToPush()).rejects.toThrow('Push notifications are not supported')
    })

    it('throws when permission is denied', async () => {
      mockNotification.requestPermission = jest.fn().mockResolvedValue('denied')

      await expect(subscribeToPush()).rejects.toThrow('Notification permission denied')
    })
  })

  describe('unsubscribeFromPush', () => {
    it('throws when push is not supported', async () => {
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      })

      await expect(unsubscribeFromPush()).rejects.toThrow('Push notifications are not supported')
    })

    it('unsubscribes and notifies server', async () => {
      const mockSubscription = {
        endpoint: 'https://push.example.com/sub1',
        unsubscribe: jest.fn().mockResolvedValue(true),
      }
      const mockPushManager = {
        getSubscription: jest.fn().mockResolvedValue(mockSubscription),
      }
      const mockRegistration = {
        pushManager: mockPushManager,
      }

      Object.defineProperty(window.navigator, 'serviceWorker', {
        value: {
          ready: Promise.resolve(mockRegistration),
        },
        writable: true,
        configurable: true,
      })
      ;(apiClient.post as jest.Mock).mockResolvedValue({})

      await unsubscribeFromPush()

      expect(mockSubscription.unsubscribe).toHaveBeenCalled()
      expect(apiClient.post).toHaveBeenCalledWith('/api/notifications/push/unsubscribe', {
        endpoint: 'https://push.example.com/sub1',
      })
    })

    it('does nothing when no subscription exists', async () => {
      const mockPushManager = {
        getSubscription: jest.fn().mockResolvedValue(null),
      }
      const mockRegistration = {
        pushManager: mockPushManager,
      }

      Object.defineProperty(window.navigator, 'serviceWorker', {
        value: {
          ready: Promise.resolve(mockRegistration),
        },
        writable: true,
        configurable: true,
      })

      await unsubscribeFromPush()
      expect(apiClient.post).not.toHaveBeenCalled()
    })
  })
})
