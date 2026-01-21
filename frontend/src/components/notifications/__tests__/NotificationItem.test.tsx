import { render, screen, fireEvent } from '@testing-library/react'
import { NotificationItem } from '../NotificationItem'
import type { Notification } from '@/types/notification'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { count?: number }) => {
    const translations: Record<string, string> = {
      'time.justNow': 'just now',
      'time.minutesAgo': `${params?.count} minutes ago`,
      'time.hoursAgo': `${params?.count} hours ago`,
      'time.yesterday': 'yesterday',
      'time.daysAgo': `${params?.count} days ago`,
    }
    return translations[key] || key
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

describe('NotificationItem', () => {
  const baseNotification: Notification = {
    id: 1,
    type: 'system',
    priority: 'normal',
    title: 'Test Notification',
    message: 'This is a test notification message',
    is_read: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: 1,
  }

  it('renders notification title and message', () => {
    render(<NotificationItem notification={baseNotification} />)
    expect(screen.getByText('Test Notification')).toBeInTheDocument()
    expect(screen.getByText('This is a test notification message')).toBeInTheDocument()
  })

  it('shows unread indicator for unread notifications', () => {
    render(<NotificationItem notification={{ ...baseNotification, is_read: false }} />)
    // Unread indicator is a small green dot
    const unreadDot = document.querySelector('.bg-emerald-500')
    expect(unreadDot).toBeInTheDocument()
  })

  it('does not show unread indicator for read notifications', () => {
    render(<NotificationItem notification={{ ...baseNotification, is_read: true }} />)
    const unreadDot = document.querySelector('.bg-emerald-500')
    expect(unreadDot).not.toBeInTheDocument()
  })

  it('renders as link when notification has link', () => {
    render(<NotificationItem notification={{ ...baseNotification, link: '/dashboard' }} />)
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/dashboard')
  })

  it('renders as button when notification has no link', () => {
    render(<NotificationItem notification={baseNotification} />)
    expect(screen.getByRole('button')).toBeInTheDocument()
  })

  it('calls onMarkAsRead when unread notification is clicked', () => {
    const onMarkAsRead = jest.fn()
    render(
      <NotificationItem
        notification={{ ...baseNotification, is_read: false }}
        onMarkAsRead={onMarkAsRead}
      />
    )
    fireEvent.click(screen.getByRole('button'))
    expect(onMarkAsRead).toHaveBeenCalledWith(1)
  })

  it('does not call onMarkAsRead when read notification is clicked', () => {
    const onMarkAsRead = jest.fn()
    render(
      <NotificationItem
        notification={{ ...baseNotification, is_read: true }}
        onMarkAsRead={onMarkAsRead}
      />
    )
    fireEvent.click(screen.getByRole('button'))
    expect(onMarkAsRead).not.toHaveBeenCalled()
  })

  it('renders correct icon for each notification type', () => {
    const types: Array<Notification['type']> = [
      'system',
      'reminder',
      'task',
      'document',
      'announcement',
      'event',
    ]

    types.forEach((type) => {
      const { unmount } = render(<NotificationItem notification={{ ...baseNotification, type }} />)
      // Icon should be rendered inside the colored container
      const iconContainer = document.querySelector('.flex.size-11')
      expect(iconContainer).toBeInTheDocument()
      unmount()
    })
  })

  it('applies correct color classes for system type', () => {
    render(<NotificationItem notification={{ ...baseNotification, type: 'system' }} />)
    const iconContainer = document.querySelector('.bg-slate-100')
    expect(iconContainer).toBeInTheDocument()
  })

  it('applies correct color classes for task type', () => {
    render(<NotificationItem notification={{ ...baseNotification, type: 'task' }} />)
    const iconContainer = document.querySelector('.bg-green-100')
    expect(iconContainer).toBeInTheDocument()
  })

  it('applies compact styling when compact prop is true', () => {
    render(<NotificationItem notification={baseNotification} compact />)
    const messageContainer = screen.getByText('This is a test notification message').closest('p')
    expect(messageContainer).toHaveClass('line-clamp-2')
  })

  it('does not apply compact styling when compact prop is false', () => {
    render(<NotificationItem notification={baseNotification} compact={false} />)
    const messageContainer = screen.getByText('This is a test notification message').closest('p')
    expect(messageContainer).not.toHaveClass('line-clamp-2')
  })

  it('uses created_at_display if provided', () => {
    render(
      <NotificationItem notification={{ ...baseNotification, created_at_display: '2 hours ago' }} />
    )
    expect(screen.getByText('2 hours ago')).toBeInTheDocument()
  })

  it('formats relative time for recent notifications', () => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString()
    render(<NotificationItem notification={{ ...baseNotification, created_at: fiveMinutesAgo }} />)
    expect(screen.getByText('5 minutes ago')).toBeInTheDocument()
  })

  it('applies different styling for read vs unread message content', () => {
    const { rerender } = render(
      <NotificationItem notification={{ ...baseNotification, is_read: false }} />
    )
    let messageBox = screen.getByText('This is a test notification message').closest('.rounded-lg')
    expect(messageBox).toHaveClass('bg-blue-50')

    rerender(<NotificationItem notification={{ ...baseNotification, is_read: true }} />)
    messageBox = screen.getByText('This is a test notification message').closest('.rounded-lg')
    expect(messageBox).toHaveClass('bg-muted')
  })
})
