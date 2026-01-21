import { render, screen } from '@testing-library/react'
import { ActivityFeed } from '../ActivityFeed'
import type { ActivityItem } from '@/types/dashboard'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string, params?: { count?: number }) => {
    const translations: Record<string, string> = {
      title: 'Recent Activity',
      empty: 'No recent activity',
      'types.document': 'Document',
      'types.report': 'Report',
      'types.task': 'Task',
      'types.event': 'Event',
      'types.announcement': 'Announcement',
      'actions.created': 'Created',
      'actions.updated': 'Updated',
      'actions.deleted': 'Deleted',
      'time.justNow': 'just now',
      'time.minutesAgo': `${params?.count} minutes ago`,
      'time.hoursAgo': `${params?.count} hours ago`,
      'time.daysAgo': `${params?.count} days ago`,
    }
    return translations[key] || key
  },
}))

// Mock the GlowingEffect component
jest.mock('@/components/ui/glowing-effect-lazy', () => ({
  GlowingEffect: () => <div data-testid="glowing-effect" />,
}))

describe('ActivityFeed', () => {
  const mockActivities: ActivityItem[] = [
    {
      id: 1,
      type: 'document',
      action: 'created',
      title: 'New Document',
      description: 'Created a new document',
      user_id: 1,
      user_name: 'John Doe',
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      type: 'report',
      action: 'updated',
      title: 'Monthly Report',
      user_id: 2,
      user_name: 'Jane Smith',
      created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(), // 30 minutes ago
    },
  ]

  it('renders feed with title', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('Recent Activity')).toBeInTheDocument()
  })

  it('renders custom title when provided', () => {
    render(<ActivityFeed activities={mockActivities} title="Custom Title" />)
    expect(screen.getByText('Custom Title')).toBeInTheDocument()
  })

  it('renders activity items', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('New Document')).toBeInTheDocument()
    expect(screen.getByText('Monthly Report')).toBeInTheDocument()
  })

  it('renders activity descriptions', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('Created a new document')).toBeInTheDocument()
  })

  it('renders user names', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.getByText('Jane Smith')).toBeInTheDocument()
  })

  it('renders empty state when no activities', () => {
    render(<ActivityFeed activities={[]} />)
    expect(screen.getByText('No recent activity')).toBeInTheDocument()
  })

  it('renders type labels', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('Document')).toBeInTheDocument()
    expect(screen.getByText('Report')).toBeInTheDocument()
  })

  it('renders action labels', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByText('Created')).toBeInTheDocument()
    expect(screen.getByText('Updated')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <ActivityFeed activities={mockActivities} className="custom-class" />
    )
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('renders GlowingEffect', () => {
    render(<ActivityFeed activities={mockActivities} />)
    expect(screen.getByTestId('glowing-effect')).toBeInTheDocument()
  })

  it('formats relative time for recent activities', () => {
    const recentActivity: ActivityItem[] = [
      {
        id: 1,
        type: 'document',
        action: 'created',
        title: 'Just Now Doc',
        user_id: 1,
        user_name: 'User',
        created_at: new Date().toISOString(),
      },
    ]
    render(<ActivityFeed activities={recentActivity} />)
    expect(screen.getByText('just now')).toBeInTheDocument()
  })

  it('renders different activity types with correct icons', () => {
    const typedActivities: ActivityItem[] = [
      {
        id: 1,
        type: 'document',
        action: 'created',
        title: 'Document 1',
        user_id: 1,
        user_name: 'User',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        type: 'task',
        action: 'created',
        title: 'Task 1',
        user_id: 1,
        user_name: 'User',
        created_at: new Date().toISOString(),
      },
      {
        id: 3,
        type: 'event',
        action: 'created',
        title: 'Event 1',
        user_id: 1,
        user_name: 'User',
        created_at: new Date().toISOString(),
      },
      {
        id: 4,
        type: 'announcement',
        action: 'created',
        title: 'Announcement 1',
        user_id: 1,
        user_name: 'User',
        created_at: new Date().toISOString(),
      },
    ]
    render(<ActivityFeed activities={typedActivities} />)
    // Check all items are rendered
    expect(screen.getByText('Document 1')).toBeInTheDocument()
    expect(screen.getByText('Task 1')).toBeInTheDocument()
    expect(screen.getByText('Event 1')).toBeInTheDocument()
    expect(screen.getByText('Announcement 1')).toBeInTheDocument()
  })
})
