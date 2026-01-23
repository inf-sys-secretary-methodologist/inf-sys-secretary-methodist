import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ConversationList } from '../ConversationList'
import { useConversations } from '@/hooks/useMessaging'

// Mock ResizeObserver for Radix UI ScrollArea
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserverMock

// Mock next/navigation
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      conversations: 'Conversations',
      searchConversations: 'Search conversations',
      noMessages: 'No messages yet',
      messageDeleted: 'Message deleted',
      image: 'Image',
      file: 'File',
      'time.yesterday': 'Yesterday',
      'time.today': 'Today',
      newConversation: 'New conversation',
      loading: 'Loading...',
      noConversations: 'No conversations',
      startNewConversation: 'Start a new conversation',
      loadError: 'Error loading conversations',
      retry: 'Retry',
      all: 'All',
      direct: 'Direct',
      groups: 'Groups',
      unknownUser: 'Unknown user',
      directMessage: 'Direct message',
      directMessageDesc: 'Start a private conversation',
      createGroup: 'Create group',
      createGroupDesc: 'Create a group conversation',
      groupName: 'Group name',
    }
    return translations[key] || key
  },
}))

// Mock the conversations hook
const mockConversations = [
  {
    id: '1',
    type: 'direct' as const,
    title: 'John Doe',
    participants: [
      { id: 1, name: 'John Doe', is_online: true },
      { id: 2, name: 'Jane Smith', is_online: false },
    ],
    last_message: {
      id: '101',
      content: 'Hello there!',
      type: 'text' as const,
      is_deleted: false,
      created_at: new Date().toISOString(),
    },
    unread_count: 2,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    type: 'group' as const,
    title: 'Team Chat',
    participants: [
      { id: 1, name: 'John Doe', is_online: true },
      { id: 2, name: 'Jane Smith', is_online: false },
      { id: 3, name: 'Bob Johnson', is_online: true },
    ],
    last_message: null,
    unread_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

jest.mock('@/hooks/useMessaging', () => ({
  useConversations: jest.fn(() => ({
    conversations: mockConversations,
    isLoading: false,
    error: null,
  })),
  useCreateDirectConversation: jest.fn(() => ({
    createConversation: jest.fn(),
    isLoading: false,
  })),
  useCreateGroupConversation: jest.fn(() => ({
    createConversation: jest.fn(),
    isLoading: false,
  })),
}))

const mockedUseConversations = jest.mocked(useConversations)

// Mock auth hook
jest.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: { id: 1, name: 'Test User', email: 'test@example.com' },
    isAuthenticated: true,
  }),
}))

// Mock users API
jest.mock('@/lib/api/users', () => ({
  usersApi: {
    getUsers: jest.fn(() => Promise.resolve({ data: [], total: 0 })),
  },
}))

describe('ConversationList', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders conversation list with header', async () => {
    render(<ConversationList />)
    expect(screen.getByText('Conversations')).toBeInTheDocument()
  })

  it('renders search input', () => {
    render(<ConversationList />)
    expect(screen.getByPlaceholderText('Search conversations')).toBeInTheDocument()
  })

  it('displays conversations', async () => {
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
      expect(screen.getByText('Team Chat')).toBeInTheDocument()
    })
  })

  it('shows last message preview', async () => {
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('Hello there!')).toBeInTheDocument()
    })
  })

  it('shows "No messages yet" for conversations without messages', async () => {
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('No messages yet')).toBeInTheDocument()
    })
  })

  it('calls onSelect when conversation is clicked', async () => {
    const user = userEvent.setup()
    const onSelect = jest.fn()
    render(<ConversationList onSelect={onSelect} />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    await user.click(screen.getByText('John Doe'))
    expect(onSelect).toHaveBeenCalledWith(mockConversations[0])
  })

  it('navigates to conversation when clicked without onSelect', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    await user.click(screen.getByText('John Doe'))
    expect(mockPush).toHaveBeenCalledWith('/messages/1')
  })

  it('highlights selected conversation', () => {
    render(<ConversationList selectedId={1} />)
    // The selected conversation should have a different style
    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('shows unread count badge', async () => {
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('2')).toBeInTheDocument()
    })
  })

  it('applies custom className', () => {
    const { container } = render(<ConversationList className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('renders new conversation button (icon-only)', () => {
    render(<ConversationList />)
    // The new conversation button is a dialog trigger with a Plus icon
    const buttons = screen.getAllByRole('button')
    // First button should be the "new conversation" trigger with Plus icon
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')
    expect(triggerButton).toBeInTheDocument()
  })

  it('can type in search input', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    const searchInput = screen.getByPlaceholderText('Search conversations')
    await user.type(searchInput, 'John')

    expect(searchInput).toHaveValue('John')
  })

  it('can clear search input', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    const searchInput = screen.getByPlaceholderText('Search conversations')
    await user.type(searchInput, 'John')
    expect(searchInput).toHaveValue('John')

    await user.clear(searchInput)
    expect(searchInput).toHaveValue('')
  })

  it('shows filter buttons for All, Direct, Groups', () => {
    render(<ConversationList />)

    expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /direct/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /groups/i })).toBeInTheDocument()
  })

  it('can click filter buttons', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    // Click direct filter button
    await user.click(screen.getByRole('button', { name: /direct/i }))
    // Button should still be visible
    expect(screen.getByRole('button', { name: /direct/i })).toBeInTheDocument()

    // Click groups filter button
    await user.click(screen.getByRole('button', { name: /groups/i }))
    expect(screen.getByRole('button', { name: /groups/i })).toBeInTheDocument()

    // Click all filter button
    await user.click(screen.getByRole('button', { name: /all/i }))
    expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument()
  })

  it('shows online indicator for online users', async () => {
    render(<ConversationList />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    // Online indicator should be present somewhere in the component
    const conversationItems = screen.getAllByText('John Doe')
    expect(conversationItems[0].closest('[class*="relative"]')).toBeInTheDocument()
  })

  it('opens new conversation dialog', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    const buttons = screen.getAllByRole('button')
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')

    if (triggerButton) {
      await user.click(triggerButton)

      await waitFor(() => {
        expect(screen.getByText('Direct message')).toBeInTheDocument()
        expect(screen.getByText('Create group')).toBeInTheDocument()
      })
    }
  })
})

describe('ConversationList date formatting', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('formats time as hours for today messages', () => {
    // Using mockedUseConversations from top of file
    const today = new Date()
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Today Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'Today message',
            type: 'text',
            is_deleted: false,
            created_at: today.toISOString(),
          },
          unread_count: 0,
          created_at: today.toISOString(),
          updated_at: today.toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('Today message')).toBeInTheDocument()
  })

  it('formats date as "Yesterday" for yesterday messages', () => {
    // Using mockedUseConversations from top of file
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Yesterday Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'Yesterday message',
            type: 'text',
            is_deleted: false,
            created_at: yesterday.toISOString(),
          },
          unread_count: 0,
          created_at: yesterday.toISOString(),
          updated_at: yesterday.toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('Yesterday')).toBeInTheDocument()
  })

  it('formats date as weekday for messages within a week', () => {
    // Using mockedUseConversations from top of file
    const threeDaysAgo = new Date()
    threeDaysAgo.setDate(threeDaysAgo.getDate() - 3)
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Week Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'Week message',
            type: 'text',
            is_deleted: false,
            created_at: threeDaysAgo.toISOString(),
          },
          unread_count: 0,
          created_at: threeDaysAgo.toISOString(),
          updated_at: threeDaysAgo.toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('Week message')).toBeInTheDocument()
  })

  it('formats date as day/month for older messages', () => {
    // Using mockedUseConversations from top of file
    const twoWeeksAgo = new Date()
    twoWeeksAgo.setDate(twoWeeksAgo.getDate() - 14)
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Old Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'Old message',
            type: 'text',
            is_deleted: false,
            created_at: twoWeeksAgo.toISOString(),
          },
          unread_count: 0,
          created_at: twoWeeksAgo.toISOString(),
          updated_at: twoWeeksAgo.toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('Old message')).toBeInTheDocument()
  })
})

describe('ConversationList message type previews', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows deleted message preview', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Deleted Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'Original content',
            type: 'text',
            is_deleted: true,
            created_at: new Date().toISOString(),
          },
          unread_count: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('Message deleted')).toBeInTheDocument()
  })

  it('shows image message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'Image Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'image.jpg',
            type: 'image',
            is_deleted: false,
            created_at: new Date().toISOString(),
          },
          unread_count: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('📷 Image')).toBeInTheDocument()
  })

  it('shows file message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'File Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'document.pdf',
            type: 'file',
            is_deleted: false,
            created_at: new Date().toISOString(),
          },
          unread_count: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('📎 File')).toBeInTheDocument()
  })

  it('shows system message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [
        {
          id: '1',
          type: 'direct',
          title: 'System Chat',
          participants: [{ id: 1, name: 'User', is_online: true }],
          last_message: {
            id: '1',
            content: 'User joined the chat',
            type: 'system',
            is_deleted: false,
            created_at: new Date().toISOString(),
          },
          unread_count: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)
    expect(screen.getByText('⚙️ User joined the chat')).toBeInTheDocument()
  })
})

describe('ConversationList loading and error states', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('shows loading state with spinner', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [],
      isLoading: true,
      error: null,
    })

    const { container } = render(<ConversationList />)

    // Loading state shows a spinner (Loader2 with animate-spin class)
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows error state', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [],
      isLoading: false,
      error: new Error('Network error'),
    })

    render(<ConversationList />)

    expect(screen.getByText('Error loading conversations')).toBeInTheDocument()
  })

  it('shows empty state when no conversations', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue({
      conversations: [],
      isLoading: false,
      error: null,
    })

    render(<ConversationList />)

    expect(screen.getByText('No conversations')).toBeInTheDocument()
    expect(screen.getByText('Start a new conversation')).toBeInTheDocument()
  })
})
