import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { NotificationsMenu, NotificationItem, NotificationMenuType } from '../notifications-menu'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      title: 'Notifications',
      markAllRead: 'Mark all as read',
      settings: 'Settings',
      downloadFile: 'Download file',
      reject: 'Reject',
      accept: 'Accept',
      empty: 'No notifications',
      'tabs.all': 'All',
      'tabs.unread': 'Unread',
      'tabs.mentions': 'Mentions',
    }
    return translations[key] || key
  },
}))

const mockNotifications: NotificationMenuType[] = [
  {
    id: 1,
    type: 'comment',
    user: {
      name: 'John Doe',
      avatar: '/avatar1.jpg',
      fallback: 'JD',
    },
    action: 'commented on',
    target: 'Project X',
    content: 'This looks great!',
    timestamp: '2024-01-15 10:30',
    timeAgo: '5 min ago',
    isRead: false,
    hasActions: false,
  },
  {
    id: 2,
    type: 'mention',
    user: {
      name: 'Jane Smith',
      avatar: '/avatar2.jpg',
      fallback: 'JS',
    },
    action: 'mentioned you in',
    target: 'Document Y',
    timestamp: '2024-01-15 09:00',
    timeAgo: '2 hours ago',
    isRead: true,
    hasActions: true,
  },
  {
    id: 3,
    type: 'file',
    user: {
      name: 'Bob Wilson',
      avatar: '/avatar3.jpg',
      fallback: 'BW',
    },
    action: 'shared a file',
    timestamp: '2024-01-14 15:00',
    timeAgo: '1 day ago',
    isRead: false,
    file: {
      name: 'report.pdf',
      size: '2.5 MB',
      type: 'PDF',
    },
  },
]

describe('NotificationsMenu', () => {
  it('renders notifications menu with title', () => {
    render(<NotificationsMenu notifications={mockNotifications} />)
    expect(screen.getByText('Notifications')).toBeInTheDocument()
  })

  it('renders all notifications', () => {
    render(<NotificationsMenu notifications={mockNotifications} />)
    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.getByText('Jane Smith')).toBeInTheDocument()
    expect(screen.getByText('Bob Wilson')).toBeInTheDocument()
  })

  it('shows correct counts on tabs', () => {
    render(<NotificationsMenu notifications={mockNotifications} />)
    // All tab should show total count
    expect(screen.getByText('3')).toBeInTheDocument()
    // Unread tab should show unread count (2 unread)
    expect(screen.getByText('2')).toBeInTheDocument()
    // Mentions tab should show mention count (1 mention)
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('filters notifications when clicking unread tab', async () => {
    const user = userEvent.setup()
    render(<NotificationsMenu notifications={mockNotifications} />)

    await user.click(screen.getByRole('tab', { name: /unread/i }))

    // Should show only unread notifications
    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.getByText('Bob Wilson')).toBeInTheDocument()
    expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument()
  })

  it('filters notifications when clicking mentions tab', async () => {
    const user = userEvent.setup()
    render(<NotificationsMenu notifications={mockNotifications} />)

    await user.click(screen.getByRole('tab', { name: /mentions/i }))

    // Should show only mention notifications
    expect(screen.getByText('Jane Smith')).toBeInTheDocument()
    expect(screen.queryByText('John Doe')).not.toBeInTheDocument()
    expect(screen.queryByText('Bob Wilson')).not.toBeInTheDocument()
  })

  it('calls onMarkAllRead when mark all read button is clicked', async () => {
    const user = userEvent.setup()
    const onMarkAllRead = jest.fn()
    render(<NotificationsMenu notifications={mockNotifications} onMarkAllRead={onMarkAllRead} />)

    await user.click(screen.getByRole('button', { name: /mark all as read/i }))
    expect(onMarkAllRead).toHaveBeenCalled()
  })

  it('calls onOpenSettings when settings button is clicked', async () => {
    const user = userEvent.setup()
    const onOpenSettings = jest.fn()
    render(<NotificationsMenu notifications={mockNotifications} onOpenSettings={onOpenSettings} />)

    await user.click(screen.getByRole('button', { name: /settings/i }))
    expect(onOpenSettings).toHaveBeenCalled()
  })

  it('shows empty state when no notifications', () => {
    render(<NotificationsMenu notifications={[]} />)
    expect(screen.getByText('No notifications')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <NotificationsMenu notifications={mockNotifications} className="custom-menu" />
    )
    expect(container.querySelector('.custom-menu')).toBeInTheDocument()
  })

  it('shows unread indicator for unread notifications', () => {
    render(<NotificationsMenu notifications={mockNotifications} />)
    // Unread notifications should have green indicator
    const unreadIndicators = document.querySelectorAll('.bg-emerald-500')
    expect(unreadIndicators.length).toBe(2) // 2 unread notifications
  })
})

describe('NotificationItem', () => {
  const baseNotification: NotificationMenuType = {
    id: 1,
    type: 'comment',
    user: {
      name: 'Test User',
      avatar: '/test-avatar.jpg',
      fallback: 'TU',
    },
    action: 'commented on',
    target: 'Test Target',
    timestamp: '2024-01-15 10:30',
    timeAgo: '5 min ago',
    isRead: false,
  }

  it('renders notification with user info', () => {
    render(<NotificationItem notification={baseNotification} />)
    expect(screen.getByText('Test User')).toBeInTheDocument()
    expect(screen.getByText(/commented on/)).toBeInTheDocument()
    expect(screen.getByText('Test Target')).toBeInTheDocument()
  })

  it('renders notification content when provided', () => {
    const notification = {
      ...baseNotification,
      content: 'This is a comment',
    }
    render(<NotificationItem notification={notification} />)
    expect(screen.getByText('This is a comment')).toBeInTheDocument()
  })

  it('renders file info when file is provided', () => {
    const notification: NotificationMenuType = {
      ...baseNotification,
      file: {
        name: 'document.pdf',
        size: '1.5 MB',
        type: 'PDF',
      },
    }
    render(<NotificationItem notification={notification} />)
    expect(screen.getByText('document.pdf')).toBeInTheDocument()
    expect(screen.getByText(/PDF.*1.5 MB/)).toBeInTheDocument()
  })

  it('calls onDownload when download button is clicked', async () => {
    const user = userEvent.setup()
    const onDownload = jest.fn()
    const notification: NotificationMenuType = {
      ...baseNotification,
      file: {
        name: 'document.pdf',
        size: '1.5 MB',
        type: 'PDF',
      },
    }
    render(<NotificationItem notification={notification} onDownload={onDownload} />)

    await user.click(screen.getByRole('button', { name: /download/i }))
    expect(onDownload).toHaveBeenCalledWith(1)
  })

  it('renders action buttons when hasActions is true', () => {
    const notification: NotificationMenuType = {
      ...baseNotification,
      hasActions: true,
    }
    render(<NotificationItem notification={notification} />)
    expect(screen.getByRole('button', { name: /reject/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /accept/i })).toBeInTheDocument()
  })

  it('calls onAccept when accept button is clicked', async () => {
    const user = userEvent.setup()
    const onAccept = jest.fn()
    const notification: NotificationMenuType = {
      ...baseNotification,
      hasActions: true,
    }
    render(<NotificationItem notification={notification} onAccept={onAccept} />)

    await user.click(screen.getByRole('button', { name: /accept/i }))
    expect(onAccept).toHaveBeenCalledWith(1)
  })

  it('calls onDecline when decline button is clicked', async () => {
    const user = userEvent.setup()
    const onDecline = jest.fn()
    const notification: NotificationMenuType = {
      ...baseNotification,
      hasActions: true,
    }
    render(<NotificationItem notification={notification} onDecline={onDecline} />)

    await user.click(screen.getByRole('button', { name: /reject/i }))
    expect(onDecline).toHaveBeenCalledWith(1)
  })

  it('does not show unread indicator for read notifications', () => {
    const notification: NotificationMenuType = {
      ...baseNotification,
      isRead: true,
    }
    const { container } = render(<NotificationItem notification={notification} />)
    expect(container.querySelector('.bg-emerald-500')).not.toBeInTheDocument()
  })

  it('shows timestamp and time ago', () => {
    render(<NotificationItem notification={baseNotification} />)
    expect(screen.getByText('2024-01-15 10:30')).toBeInTheDocument()
    expect(screen.getByText('5 min ago')).toBeInTheDocument()
  })

  it('renders avatar with fallback', () => {
    render(<NotificationItem notification={baseNotification} />)
    expect(screen.getByText('TU')).toBeInTheDocument()
  })
})
