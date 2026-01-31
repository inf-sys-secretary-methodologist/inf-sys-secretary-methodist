import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PushNotificationSettings } from '../PushNotificationSettings'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { count?: number; date?: string }) => {
    const translations: Record<string, string> = {
      title: 'Push Notifications',
      description: 'Receive notifications in your browser',
      notSupported: 'Push notifications are not supported',
      notSupportedDescription:
        'Your browser does not support push notifications. Please use a modern browser.',
      permissionBlocked: 'Permission blocked',
      permissionBlockedTitle: 'Notifications blocked',
      permissionBlockedDescription: 'Enable notifications in your browser settings.',
      enabledDescription: 'Push notifications are enabled',
      enabled: 'Enabled',
      browserNotifications: 'Browser Notifications',
      browserNotificationsDesc: 'Receive push notifications in this browser',
      devices: `${params?.count || 0} registered devices`,
      unknownDevice: 'Unknown Device',
      addedOn: `Added on ${params?.date || ''}`,
      removeDeviceTitle: 'Remove device',
      removeDeviceDescription: 'This device will no longer receive push notifications.',
      cancel: 'Cancel',
      remove: 'Remove',
      sendTest: 'Send Test Notification',
      sendingTest: 'Sending...',
      disableAll: 'Disable Push Notifications',
      disableTitle: 'Disable push notifications',
      disableDescription: 'You will no longer receive push notifications.',
      confirmDisable: 'Disable',
      enableInstructions: 'Click the button below to enable push notifications.',
      enable: 'Enable Push Notifications',
      enabling: 'Enabling...',
      enabledSuccess: 'Push notifications enabled',
      permissionDenied: 'Permission denied',
      enableError: 'Failed to enable push notifications',
      disabledSuccess: 'Push notifications disabled',
      disableError: 'Failed to disable push notifications',
      deviceRemoved: 'Device removed',
      removeError: 'Failed to remove device',
      testSent: 'Test notification sent',
      testError: 'Failed to send test notification',
    }
    return translations[key] || key
  },
}))

// Mock sonner
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

// Mock hooks state
const mockHookState = {
  isSupported: true,
  permission: 'default' as 'default' | 'granted' | 'denied' | 'unsupported',
  isEnabled: false,
  isLocallySubscribed: false,
  subscriptions: [] as Array<{
    id: number
    device_name: string
    is_active: boolean
    created_at: string
  }>,
  totalDevices: 0,
  isLoading: false,
  isSubscribing: false,
  isUnsubscribing: false,
  error: null as Error | null,
  subscribe: jest.fn().mockResolvedValue({ id: 1 }),
  unsubscribe: jest.fn().mockResolvedValue(undefined),
  removeSubscription: jest.fn().mockResolvedValue(undefined),
  testNotification: jest.fn().mockResolvedValue(undefined),
  refreshStatus: jest.fn(),
}

// Mock usePushNotifications hook
jest.mock('@/hooks/usePushNotifications', () => ({
  usePushNotifications: () => mockHookState,
}))

describe('PushNotificationSettings', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset to default state
    mockHookState.isSupported = true
    mockHookState.permission = 'default'
    mockHookState.isEnabled = false
    mockHookState.isLocallySubscribed = false
    mockHookState.subscriptions = []
    mockHookState.totalDevices = 0
    mockHookState.isLoading = false
    mockHookState.isSubscribing = false
    mockHookState.isUnsubscribing = false
    mockHookState.error = null
    mockHookState.subscribe = jest.fn().mockResolvedValue({ id: 1 })
    mockHookState.unsubscribe = jest.fn().mockResolvedValue(undefined)
    mockHookState.removeSubscription = jest.fn().mockResolvedValue(undefined)
    mockHookState.testNotification = jest.fn().mockResolvedValue(undefined)
  })

  describe('Not Supported State', () => {
    it('shows not supported message when push is not supported', () => {
      mockHookState.isSupported = false

      render(<PushNotificationSettings />)

      expect(screen.getByText('Push Notifications')).toBeInTheDocument()
      expect(screen.getByText('Push notifications are not supported')).toBeInTheDocument()
      expect(
        screen.getByText(
          'Your browser does not support push notifications. Please use a modern browser.'
        )
      ).toBeInTheDocument()
    })
  })

  describe('Permission Denied State', () => {
    it('shows permission blocked message when permission is denied', () => {
      mockHookState.permission = 'denied'

      render(<PushNotificationSettings />)

      expect(screen.getByText('Push Notifications')).toBeInTheDocument()
      expect(screen.getByText('Permission blocked')).toBeInTheDocument()
      expect(screen.getByText('Notifications blocked')).toBeInTheDocument()
      expect(screen.getByText('Enable notifications in your browser settings.')).toBeInTheDocument()
    })
  })

  describe('Disabled State (Not Enabled)', () => {
    it('shows enable button when push is not enabled', () => {
      render(<PushNotificationSettings />)

      expect(screen.getByText('Push Notifications')).toBeInTheDocument()
      expect(screen.getByText('Receive notifications in your browser')).toBeInTheDocument()
      expect(
        screen.getByText('Click the button below to enable push notifications.')
      ).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /enable push notifications/i })).toBeInTheDocument()
    })

    it('shows loading state when subscribing', () => {
      mockHookState.isSubscribing = true

      render(<PushNotificationSettings />)

      expect(screen.getByText('Enabling...')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /enabling/i })).toBeDisabled()
    })

    it('calls subscribe when enable button is clicked', async () => {
      const user = userEvent.setup()

      render(<PushNotificationSettings />)

      await user.click(screen.getByRole('button', { name: /enable push notifications/i }))

      await waitFor(() => {
        expect(mockHookState.subscribe).toHaveBeenCalled()
      })
    })
  })

  describe('Enabled State', () => {
    beforeEach(() => {
      mockHookState.isEnabled = true
      mockHookState.isLocallySubscribed = true
      mockHookState.subscriptions = [
        {
          id: 1,
          device_name: 'Chrome on Windows',
          is_active: true,
          created_at: '2024-01-15T10:30:00Z',
        },
        {
          id: 2,
          device_name: 'Firefox on Mac',
          is_active: true,
          created_at: '2024-01-10T08:00:00Z',
        },
      ]
      mockHookState.totalDevices = 2
    })

    it('shows enabled badge and subscriptions', () => {
      render(<PushNotificationSettings />)

      expect(screen.getByText('Push Notifications')).toBeInTheDocument()
      expect(screen.getByText('Enabled')).toBeInTheDocument()
      expect(screen.getByText('Push notifications are enabled')).toBeInTheDocument()
      expect(screen.getByText('Chrome on Windows')).toBeInTheDocument()
      expect(screen.getByText('Firefox on Mac')).toBeInTheDocument()
    })

    it('shows browser notifications toggle switch', () => {
      render(<PushNotificationSettings />)

      expect(screen.getByText('Browser Notifications')).toBeInTheDocument()
      expect(screen.getByText('Receive push notifications in this browser')).toBeInTheDocument()
      expect(screen.getByRole('switch')).toBeInTheDocument()
      expect(screen.getByRole('switch')).toBeChecked()
    })

    it('shows send test notification button', () => {
      render(<PushNotificationSettings />)

      expect(screen.getByRole('button', { name: /send test notification/i })).toBeInTheDocument()
    })

    it('shows disable all button', () => {
      render(<PushNotificationSettings />)

      expect(
        screen.getByRole('button', { name: /disable push notifications/i })
      ).toBeInTheDocument()
    })

    it('shows device remove buttons', () => {
      render(<PushNotificationSettings />)

      // Each device should have a delete button (trash icon)
      const deleteButtons = screen.getAllByRole('button', { name: '' })
      // Filter for icon buttons (they have no accessible name by default)
      expect(deleteButtons.length).toBeGreaterThanOrEqual(2)
    })

    it('calls testNotification when send test button is clicked', async () => {
      const user = userEvent.setup()

      render(<PushNotificationSettings />)

      await user.click(screen.getByRole('button', { name: /send test notification/i }))

      await waitFor(() => {
        expect(mockHookState.testNotification).toHaveBeenCalled()
      })
    })

    it('shows unknown device name for devices without name', () => {
      mockHookState.subscriptions = [
        {
          id: 1,
          device_name: '',
          is_active: true,
          created_at: '2024-01-15T10:30:00Z',
        },
      ]
      mockHookState.totalDevices = 1

      render(<PushNotificationSettings />)

      expect(screen.getByText('Unknown Device')).toBeInTheDocument()
    })

    it('shows error message when error exists', () => {
      mockHookState.error = new Error('Something went wrong')

      render(<PushNotificationSettings />)

      expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    })
  })

  describe('Alert Dialogs', () => {
    beforeEach(() => {
      mockHookState.isEnabled = true
      mockHookState.isLocallySubscribed = true
      mockHookState.subscriptions = [
        {
          id: 1,
          device_name: 'Chrome on Windows',
          is_active: true,
          created_at: '2024-01-15T10:30:00Z',
        },
      ]
      mockHookState.totalDevices = 1
    })

    it('shows disable confirmation dialog', async () => {
      const user = userEvent.setup()

      render(<PushNotificationSettings />)

      await user.click(screen.getByRole('button', { name: /disable push notifications/i }))

      await waitFor(() => {
        expect(screen.getByText('Disable push notifications')).toBeInTheDocument()
        expect(
          screen.getByText('You will no longer receive push notifications.')
        ).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /disable$/i })).toBeInTheDocument()
      })
    })

    it('calls unsubscribe when disable is confirmed', async () => {
      const user = userEvent.setup()

      render(<PushNotificationSettings />)

      await user.click(screen.getByRole('button', { name: /disable push notifications/i }))

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /disable$/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /disable$/i }))

      await waitFor(() => {
        expect(mockHookState.unsubscribe).toHaveBeenCalled()
      })
    })
  })

  describe('Inactive subscriptions', () => {
    it('shows inactive subscriptions with reduced opacity', () => {
      mockHookState.isEnabled = true
      mockHookState.isLocallySubscribed = true
      mockHookState.subscriptions = [
        {
          id: 1,
          device_name: 'Old Device',
          is_active: false,
          created_at: '2024-01-01T00:00:00Z',
        },
      ]
      mockHookState.totalDevices = 1

      render(<PushNotificationSettings />)

      expect(screen.getByText('Old Device')).toBeInTheDocument()
      // The inactive device should still be displayed
      const deviceElement = screen.getByText('Old Device').closest('[class*="opacity"]')
      expect(deviceElement).toBeInTheDocument()
    })
  })

  describe('Locally subscribed but not enabled on server', () => {
    it('shows enabled UI when locally subscribed', () => {
      mockHookState.isLocallySubscribed = true
      mockHookState.isEnabled = false
      mockHookState.subscriptions = []
      mockHookState.totalDevices = 0

      render(<PushNotificationSettings />)

      expect(screen.getByText('Enabled')).toBeInTheDocument()
      expect(screen.getByText('Push notifications are enabled')).toBeInTheDocument()
    })
  })
})
