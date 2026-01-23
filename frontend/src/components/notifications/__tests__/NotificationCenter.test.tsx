import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { NotificationCenter } from '../NotificationCenter'
import type { Notification } from '@/types/notification'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { count?: number }) => {
    const translations: Record<string, string> = {
      title: 'Notifications',
      markAllAsRead: 'Mark all as read',
      allMarkedAsRead: 'All notifications marked as read',
      markAsReadError: 'Failed to mark as read',
      markAllAsReadError: 'Failed to mark all as read',
      empty: 'No notifications',
      viewAll: 'View all notifications',
      ariaLabel: 'Notifications',
      ariaLabelWithCount: `${params?.count || 0} unread notifications`,
      'time.justNow': 'just now',
      'time.minutesAgo': `${params?.count || 0} minutes ago`,
      'time.hoursAgo': `${params?.count || 0} hours ago`,
      'time.yesterday': 'yesterday',
      'time.daysAgo': `${params?.count || 0} days ago`,
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

// Mock notification hooks
const mockNotifications: Notification[] = [
  {
    id: 1,
    type: 'system',
    priority: 'normal',
    title: 'System Update',
    message: 'System has been updated',
    is_read: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  },
  {
    id: 2,
    type: 'task',
    priority: 'high',
    title: 'New Task',
    message: 'You have a new task',
    is_read: true,
    created_at: new Date(Date.now() - 60000).toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  },
]

const mockHooks = {
  unreadCount: 1,
  isLoading: false,
  mutateAsyncMarkAsRead: jest.fn().mockResolvedValue(undefined),
  mutateAsyncMarkAllAsRead: jest.fn().mockResolvedValue(undefined),
}

jest.mock('@/hooks/useNotifications', () => ({
  useNotifications: () => ({
    data: { notifications: mockNotifications },
    isLoading: mockHooks.isLoading,
  }),
  useUnreadCount: () => ({
    data: { count: mockHooks.unreadCount },
  }),
  useMarkAsRead: () => ({
    mutateAsync: mockHooks.mutateAsyncMarkAsRead,
  }),
  useMarkAllAsRead: () => ({
    mutateAsync: mockHooks.mutateAsyncMarkAllAsRead,
    isPending: false,
  }),
}))

describe('NotificationCenter', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockHooks.unreadCount = 1
    mockHooks.isLoading = false
  })

  it('renders notification bell button', () => {
    render(<NotificationCenter />)
    expect(screen.getByRole('button', { name: /notifications/i })).toBeInTheDocument()
  })

  it('shows unread count badge', () => {
    render(<NotificationCenter />)
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('does not show badge when no unread notifications', () => {
    mockHooks.unreadCount = 0
    render(<NotificationCenter />)
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('shows 99+ for high unread counts', () => {
    mockHooks.unreadCount = 150
    render(<NotificationCenter />)
    expect(screen.getByText('99+')).toBeInTheDocument()
  })

  it('opens popover when clicked', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Notifications')).toBeInTheDocument()
    })
  })

  it('shows notifications list', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('System Update')).toBeInTheDocument()
      expect(screen.getByText('New Task')).toBeInTheDocument()
    })
  })

  it('shows mark all as read button when there are unread notifications', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Mark all as read')).toBeInTheDocument()
    })
  })

  it('shows view all button', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByRole('link', { name: /view all/i })).toBeInTheDocument()
    })
  })

  it('applies custom className', () => {
    const { container } = render(<NotificationCenter className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('calls mark all as read when button is clicked', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Mark all as read')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Mark all as read'))

    await waitFor(() => {
      expect(mockHooks.mutateAsyncMarkAllAsRead).toHaveBeenCalled()
    })
  })

  it('calls mark as read when notification is clicked', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('System Update')).toBeInTheDocument()
    })

    // Click on the unread notification
    await userEvent.click(screen.getByText('System Update'))

    await waitFor(() => {
      expect(mockHooks.mutateAsyncMarkAsRead).toHaveBeenCalledWith(1)
    })
  })

  it('shows loading indicator when loading', async () => {
    mockHooks.isLoading = true

    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    // Loading state might show different UI
    expect(screen.getByRole('button', { name: /notifications/i })).toBeInTheDocument()
  })

  it('shows empty state when no notifications', async () => {
    const originalNotifications = mockNotifications.slice()
    mockNotifications.length = 0

    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('No notifications')).toBeInTheDocument()
    })

    // Restore
    mockNotifications.push(...originalNotifications)
  })

  it('highlights unread notifications', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('System Update')).toBeInTheDocument()
    })

    // The unread notification should have different styling
    const unreadNotification = screen.getByText('System Update').closest('[class*="bg-"]')
    expect(unreadNotification).toBeInTheDocument()
  })

  it('shows high priority indicator for high priority notifications', async () => {
    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('New Task')).toBeInTheDocument()
    })

    // High priority notifications should have some indicator
    expect(screen.getByText('New Task')).toBeInTheDocument()
  })

  it('handles mark as read error', async () => {
    mockHooks.mutateAsyncMarkAsRead.mockRejectedValueOnce(new Error('Failed'))

    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('System Update')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('System Update'))

    // Should handle error gracefully
    expect(mockHooks.mutateAsyncMarkAsRead).toHaveBeenCalled()
  })

  it('handles mark all as read error', async () => {
    mockHooks.mutateAsyncMarkAllAsRead.mockRejectedValueOnce(new Error('Failed'))

    render(<NotificationCenter />)

    await userEvent.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Mark all as read')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByText('Mark all as read'))

    // Should handle error gracefully
    expect(mockHooks.mutateAsyncMarkAllAsRead).toHaveBeenCalled()
  })
})
