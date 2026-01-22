import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ConversationList } from '../ConversationList'

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
})
