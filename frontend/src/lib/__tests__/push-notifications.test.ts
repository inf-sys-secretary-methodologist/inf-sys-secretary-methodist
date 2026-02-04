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

  describe('getDeviceName via subscribeToPush', () => {
    const setupSubscribeMocks = (userAgent: string) => {
      Object.defineProperty(window, 'navigator', {
        value: {
          serviceWorker: {
            ready: Promise.resolve({
              pushManager: {
                subscribe: jest.fn().mockResolvedValue({
                  endpoint: 'https://push.example.com/sub1',
                  getKey: jest.fn((name: string) => {
                    if (name === 'p256dh') return new Uint8Array([1, 2, 3]).buffer
                    if (name === 'auth') return new Uint8Array([4, 5, 6]).buffer
                    return null
                  }),
                }),
              },
            }),
          },
          userAgent,
        },
        writable: true,
        configurable: true,
      })

      mockNotification.requestPermission = jest.fn().mockResolvedValue('granted')
      mockNotification.permission = 'granted'
      ;(apiClient.get as jest.Mock).mockResolvedValue({ public_key: 'BEl62iUYgUivxIkv69yViEuiBIa' })
      ;(apiClient.post as jest.Mock).mockResolvedValue({ id: 1 })
    }

    it('detects iPhone', async () => {
      setupSubscribeMocks('Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'iPhone' })
      )
    })

    it('detects iPad', async () => {
      setupSubscribeMocks('Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'iPad' })
      )
    })

    it('detects Android', async () => {
      setupSubscribeMocks('Mozilla/5.0 (Linux; Android 11; Pixel 5)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'Android' })
      )
    })

    it('detects Windows', async () => {
      setupSubscribeMocks('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'Windows' })
      )
    })

    it('detects Mac', async () => {
      setupSubscribeMocks('Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'Mac' })
      )
    })

    it('detects Linux', async () => {
      setupSubscribeMocks('Mozilla/5.0 (X11; Linux x86_64)')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'Linux' })
      )
    })

    it('returns Unknown Device for unrecognized user agent', async () => {
      setupSubscribeMocks('SomeUnknownBrowser/1.0')
      await subscribeToPush()
      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/notifications/push/subscribe',
        expect.objectContaining({ device_name: 'Unknown Device' })
      )
    })
  })

  describe('getCurrentSubscription error handling', () => {
    it('returns null when getSubscription throws', async () => {
      const mockPushManager = {
        getSubscription: jest.fn().mockRejectedValue(new Error('Failed')),
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
      expect(result).toBeNull()
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

    it('successfully subscribes to push notifications', async () => {
      // Mock permission granted
      mockNotification.requestPermission = jest.fn().mockResolvedValue('granted')
      mockNotification.permission = 'granted'

      // Mock VAPID key response
      const mockVapidKey =
        'BEl62iUYgUivxIkv69yViEuiBIa-Ib9-SkvMeAtA3LFgDzkrxZJjSgSnfckjBJuBkr3qBUYIHBQFLXYp5Nksh8U'
      ;(apiClient.get as jest.Mock).mockResolvedValue({ public_key: mockVapidKey })

      // Mock subscription keys
      const mockP256dhKey = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8]).buffer
      const mockAuthKey = new Uint8Array([9, 10, 11, 12]).buffer

      const mockSubscription = {
        endpoint: 'https://push.example.com/sub1',
        getKey: jest.fn((name: string) => {
          if (name === 'p256dh') return mockP256dhKey
          if (name === 'auth') return mockAuthKey
          return null
        }),
      }

      const mockPushManager = {
        subscribe: jest.fn().mockResolvedValue(mockSubscription),
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

      // Mock server response
      const mockServerResponse = {
        id: 1,
        device_name: 'Unknown Device',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
      }
      ;(apiClient.post as jest.Mock).mockResolvedValue(mockServerResponse)

      const result = await subscribeToPush()

      expect(result).toEqual(mockServerResponse)
      expect(mockPushManager.subscribe).toHaveBeenCalledWith({
        userVisibleOnly: true,
        applicationServerKey: expect.any(ArrayBuffer),
      })
      expect(apiClient.post).toHaveBeenCalledWith('/api/notifications/push/subscribe', {
        endpoint: 'https://push.example.com/sub1',
        p256dh: expect.any(String),
        auth: expect.any(String),
        user_agent: 'Mozilla/5.0 Test Browser',
        device_name: 'Unknown Device',
      })
    })

    it('throws when subscription keys are null', async () => {
      mockNotification.requestPermission = jest.fn().mockResolvedValue('granted')
      mockNotification.permission = 'granted'

      const mockVapidKey =
        'BEl62iUYgUivxIkv69yViEuiBIa-Ib9-SkvMeAtA3LFgDzkrxZJjSgSnfckjBJuBkr3qBUYIHBQFLXYp5Nksh8U'
      ;(apiClient.get as jest.Mock).mockResolvedValue({ public_key: mockVapidKey })

      const mockSubscription = {
        endpoint: 'https://push.example.com/sub1',
        getKey: jest.fn().mockReturnValue(null),
      }

      const mockPushManager = {
        subscribe: jest.fn().mockResolvedValue(mockSubscription),
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

      await expect(subscribeToPush()).rejects.toThrow('Failed to get subscription keys')
    })
  })

  describe('getServiceWorkerRegistration edge case', () => {
    it('throws when serviceWorker disappears after isPushSupported check', async () => {
      // Track 'in' operator checks vs property access
      let inCheckCount = 0
      const handler = {
        has(_target: object, prop: string): boolean {
          if (prop === 'serviceWorker') {
            inCheckCount++
            // First two 'in' checks (isPushSupported in subscribeToPush and requestPermission) pass
            // Third 'in' check (in getServiceWorkerRegistration) should fail
            return inCheckCount < 3
          }
          return true
        },
        get(target: Record<string, unknown>, prop: string) {
          if (prop === 'userAgent') return 'Test Browser'
          if (prop === 'serviceWorker') {
            return {
              ready: Promise.resolve({
                pushManager: {
                  subscribe: jest.fn(),
                  getSubscription: jest.fn(),
                },
              }),
            }
          }
          return target[prop]
        },
      }

      const navigatorProxy = new Proxy({} as Record<string, unknown>, handler)
      Object.defineProperty(window, 'navigator', {
        value: navigatorProxy,
        writable: true,
        configurable: true,
      })

      // Mock Notification to pass isPushSupported and requestPermission
      mockNotification.permission = 'granted'
      mockNotification.requestPermission = jest.fn().mockResolvedValue('granted')

      await expect(subscribeToPush()).rejects.toThrow('Service Worker is not supported')
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
