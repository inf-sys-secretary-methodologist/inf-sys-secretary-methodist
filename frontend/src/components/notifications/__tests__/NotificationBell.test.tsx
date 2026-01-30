import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { NotificationBell } from '../NotificationBell'
import { useNotifications, useUnreadCount } from '@/hooks/useNotifications'

// Mock ResizeObserver for Radix UI ScrollArea
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserverMock

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: (namespace: string) => (key: string, params?: { count?: number }) => {
    const translations: Record<string, Record<string, string>> = {
      notificationBell: {
        ariaLabel: `Notifications (${params?.count || 0} unread)`,
        'tabs.all': 'All',
        'tabs.tasks': 'Tasks',
        'tabs.documents': 'Documents',
        'tabs.events': 'Events',
        showAll: 'Show all',
        settingsAriaLabel: 'Notification settings',
      },
      notifications: {
        title: 'Notifications',
        markAllRead: 'Mark all as read',
        markError: 'Failed to mark as read',
        allMarkedRead: 'All marked as read',
        markAllError: 'Failed to mark all as read',
        empty: 'No notifications',
      },
    }
    return translations[namespace]?.[key] || key
  },
}))

// Mock next/link
jest.mock('next/link', () => {
  const MockLink = ({
    children,
    href,
    onClick,
  }: {
    children: React.ReactNode
    href: string
    onClick?: () => void
  }) => (
    <a href={href} onClick={onClick}>
      {children}
    </a>
  )
  MockLink.displayName = 'MockLink'
  return MockLink
})

// Mock sonner toast
jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockNotifications = [
  {
    id: 1,
    type: 'task' as const,
    priority: 'normal' as const,
    title: 'Task notification',
    message: 'You have a new task',
    is_read: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  },
  {
    id: 2,
    type: 'document' as const,
    priority: 'high' as const,
    title: 'Document notification',
    message: 'A document was shared with you',
    is_read: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  },
  {
    id: 3,
    type: 'reminder' as const,
    priority: 'normal' as const,
    title: 'Reminder',
    message: 'Event starting soon',
    is_read: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  },
]

const mockMarkAsRead = jest.fn()
const mockMarkAllAsRead = jest.fn()

// Mock notification hooks
jest.mock('@/hooks/useNotifications', () => ({
  useNotifications: jest.fn(() => ({
    notifications: mockNotifications,
    isLoading: false,
  })),
  useUnreadCount: jest.fn(() => ({
    count: 2,
  })),
  useMarkAsRead: jest.fn(() => ({
    mutateAsync: mockMarkAsRead,
    isPending: false,
  })),
  useMarkAllAsRead: jest.fn(() => ({
    mutateAsync: mockMarkAllAsRead,
    isPending: false,
  })),
}))

const mockedUseNotifications = jest.mocked(useNotifications)
const mockedUseUnreadCount = jest.mocked(useUnreadCount)

// Mock NotificationItem
jest.mock('../NotificationItem', () => ({
  NotificationItem: ({ notification }: { notification: { title: string } }) => (
    <div data-testid="notification-item">{notification.title}</div>
  ),
}))

describe('NotificationBell', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders bell button', () => {
    render(<NotificationBell />)
    const button = screen.getByRole('button', { name: /notifications/i })
    expect(button).toBeInTheDocument()
  })

  it('shows unread count badge when there are unread notifications', () => {
    render(<NotificationBell />)
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('opens popover when clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Notifications')).toBeInTheDocument()
    })
  })

  it('displays notification items in popover', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      const items = screen.getAllByTestId('notification-item')
      expect(items.length).toBe(3)
    })
  })

  it('renders tabs for filtering notifications', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('All')).toBeInTheDocument()
      expect(screen.getByText('Tasks')).toBeInTheDocument()
      expect(screen.getByText('Documents')).toBeInTheDocument()
      expect(screen.getByText('Events')).toBeInTheDocument()
    })
  })

  it('filters notifications by task type when Tasks tab is clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))
    await waitFor(() => {
      expect(screen.getByText('Tasks')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Tasks'))

    await waitFor(() => {
      const items = screen.getAllByTestId('notification-item')
      expect(items.length).toBe(1)
      expect(screen.getByText('Task notification')).toBeInTheDocument()
    })
  })

  it('filters notifications by document type when Documents tab is clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))
    await waitFor(() => {
      expect(screen.getByText('Documents')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Documents'))

    await waitFor(() => {
      const items = screen.getAllByTestId('notification-item')
      expect(items.length).toBe(1)
      expect(screen.getByText('Document notification')).toBeInTheDocument()
    })
  })

  it('filters notifications by events/reminders when Events tab is clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))
    await waitFor(() => {
      expect(screen.getByText('Events')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Events'))

    await waitFor(() => {
      const items = screen.getAllByTestId('notification-item')
      expect(items.length).toBe(1)
      expect(screen.getByText('Reminder')).toBeInTheDocument()
    })
  })

  it('shows "Show all" link in footer', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Show all')).toBeInTheDocument()
    })
  })

  it('shows settings link', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      // Find link to settings by href
      const links = screen.getAllByRole('link')
      const settingsLink = links.find(
        (link) => link.getAttribute('href') === '/settings/notifications'
      )
      expect(settingsLink).toBeInTheDocument()
    })
  })

  it('shows mark all as read button when there are unread notifications', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /mark all as read/i })).toBeInTheDocument()
    })
  })

  it('shows loading state', async () => {
    // Set up mock before render - need to persist for multiple calls
    mockedUseNotifications.mockImplementation(
      () =>
        ({
          notifications: [],
          isLoading: true,
        }) as never
    )

    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })

    // Restore original mock
    mockedUseNotifications.mockImplementation(
      () =>
        ({
          notifications: mockNotifications,
          isLoading: false,
        }) as never
    )
  })

  it('shows empty state when no notifications', async () => {
    mockedUseNotifications.mockImplementation(
      () =>
        ({
          notifications: [],
          isLoading: false,
        }) as never
    )

    const user = userEvent.setup()
    render(<NotificationBell />)

    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('No notifications')).toBeInTheDocument()
    })

    // Restore original mock
    mockedUseNotifications.mockImplementation(
      () =>
        ({
          notifications: mockNotifications,
          isLoading: false,
        }) as never
    )
  })

  it('does not show unread badge when count is 0', () => {
    mockedUseUnreadCount.mockReturnValueOnce({ count: 0 } as never)

    render(<NotificationBell />)
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })

  it('shows 99+ when unread count exceeds 99', () => {
    mockedUseUnreadCount.mockReturnValueOnce({ count: 150 } as never)

    render(<NotificationBell />)
    expect(screen.getByText('99+')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(<NotificationBell className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('closes popover when settings link is clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    // Open popover
    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Notifications')).toBeInTheDocument()
    })

    // Find and click settings link by href
    const links = screen.getAllByRole('link')
    const settingsLink = links.find(
      (link) => link.getAttribute('href') === '/settings/notifications'
    )
    expect(settingsLink).toBeInTheDocument()
    await user.click(settingsLink!)

    // onClick was triggered - popover should close
    await waitFor(() => {
      expect(screen.queryByText('Notifications')).not.toBeInTheDocument()
    })
  })

  it('closes popover when show all link is clicked', async () => {
    const user = userEvent.setup()
    render(<NotificationBell />)

    // Open popover
    await user.click(screen.getByRole('button', { name: /notifications/i }))

    await waitFor(() => {
      expect(screen.getByText('Show all')).toBeInTheDocument()
    })

    // Click show all link by finding link to /notifications
    const links = screen.getAllByRole('link')
    const showAllLink = links.find((link) => link.getAttribute('href') === '/notifications')
    expect(showAllLink).toBeInTheDocument()
    await user.click(showAllLink!)

    // onClick was triggered - popover should close
    await waitFor(() => {
      expect(screen.queryByText('Show all')).not.toBeInTheDocument()
    })
  })
})
