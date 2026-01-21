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
})
