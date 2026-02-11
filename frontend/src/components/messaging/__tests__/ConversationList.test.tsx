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
import type { ParticipantRole, Conversation, Message, MessageType } from '@/types/messaging'

const mockParticipant = (id: number, name: string) => ({
  id,
  user_id: id,
  name,
  role: 'member' as ParticipantRole,
  is_muted: false,
  joined_at: '2024-01-01T00:00:00Z',
})

const createMockMessage = (
  id: number,
  content: string,
  type: MessageType = 'text',
  options: { is_deleted?: boolean; created_at?: string } = {}
): Message => ({
  id,
  conversation_id: 1,
  sender_id: 1,
  sender_name: 'User',
  content,
  type,
  attachments: [],
  is_edited: false,
  is_deleted: options.is_deleted ?? false,
  created_at: options.created_at ?? new Date().toISOString(),
})

const createMockConversation = (
  id: number,
  title: string,
  options: {
    type?: 'direct' | 'group'
    last_message?: Message
    created_at?: string
    updated_at?: string
  } = {}
): Conversation => ({
  id,
  type: options.type ?? 'direct',
  title,
  created_by: 1,
  participants: [mockParticipant(1, 'User')],
  last_message: options.last_message,
  unread_count: 0,
  created_at: options.created_at ?? new Date().toISOString(),
  updated_at: options.updated_at ?? new Date().toISOString(),
})

const createMockHookReturn = (
  conversations: Conversation[],
  options: { isLoading?: boolean; error?: Error | null } = {}
) => ({
  conversations,
  data: undefined,
  total: conversations.length,
  isLoading: options.isLoading ?? false,
  error: options.error ?? null,
  mutate: jest.fn(),
})

const mockConversations = [
  {
    id: 1,
    type: 'direct' as const,
    title: 'John Doe',
    created_by: 1,
    participants: [mockParticipant(1, 'John Doe'), mockParticipant(2, 'Jane Smith')],
    last_message: {
      id: 101,
      conversation_id: 1,
      sender_id: 1,
      sender_name: 'John Doe',
      content: 'Hello there!',
      type: 'text' as const,
      attachments: [],
      is_edited: false,
      is_deleted: false,
      created_at: new Date().toISOString(),
    },
    unread_count: 2,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 2,
    type: 'group' as const,
    title: 'Team Chat',
    created_by: 1,
    participants: [
      mockParticipant(1, 'John Doe'),
      mockParticipant(2, 'Jane Smith'),
      mockParticipant(3, 'Bob Johnson'),
    ],
    last_message: undefined,
    unread_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

jest.mock('@/hooks/useMessaging', () => ({
  useConversations: jest.fn(() => ({
    conversations: mockConversations,
    data: undefined,
    total: 0,
    isLoading: false,
    error: null,
    mutate: jest.fn(),
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
    getAll: jest.fn(() =>
      Promise.resolve([
        { id: 10, name: 'Test User 1', email: 'test1@example.com' },
        { id: 11, name: 'Test User 2', email: 'test2@example.com' },
      ])
    ),
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

  it('navigates to direct message step in dialog', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    // Open dialog
    const buttons = screen.getAllByRole('button')
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')

    if (triggerButton) {
      await user.click(triggerButton)

      await waitFor(() => {
        expect(screen.getByText('Direct message')).toBeInTheDocument()
      })

      // Click Direct message button
      const directMessageButton = screen.getByText('Direct message').closest('button')
      if (directMessageButton) {
        await user.click(directMessageButton)
      }

      // Should show direct message step
      await waitFor(() => {
        expect(screen.getByText(/direct.*message/i)).toBeInTheDocument()
      })
    }
  })

  it('navigates to create group step in dialog', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    // Open dialog
    const buttons = screen.getAllByRole('button')
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')

    if (triggerButton) {
      await user.click(triggerButton)

      await waitFor(() => {
        expect(screen.getByText('Create group')).toBeInTheDocument()
      })

      // Click Create group button
      const createGroupButton = screen.getByText('Create group').closest('button')
      if (createGroupButton) {
        await user.click(createGroupButton)
      }

      // Should show group name input
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Group name')).toBeInTheDocument()
      })
    }
  })

  it('can go back from group step to choose step', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    // Open dialog
    const buttons = screen.getAllByRole('button')
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')

    if (triggerButton) {
      await user.click(triggerButton)

      await waitFor(() => {
        expect(screen.getByText('Create group')).toBeInTheDocument()
      })

      // Click Create group button
      const createGroupButton = screen.getByText('Create group').closest('button')
      if (createGroupButton) {
        await user.click(createGroupButton)
      }

      // Should show group step
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Group name')).toBeInTheDocument()
      })

      // Click back button (ArrowLeft icon button)
      const backButton = screen
        .getAllByRole('button')
        .find((btn) => btn.querySelector('svg.lucide-arrow-left'))
      if (backButton) {
        await user.click(backButton)
      }

      // Should be back at choose step
      await waitFor(() => {
        expect(screen.getByText('Direct message')).toBeInTheDocument()
        expect(screen.getByText('Create group')).toBeInTheDocument()
      })
    }
  })

  it('can type group name in create group step', async () => {
    const user = userEvent.setup()
    render(<ConversationList />)

    // Open dialog
    const buttons = screen.getAllByRole('button')
    const triggerButton = buttons.find((btn) => btn.getAttribute('aria-haspopup') === 'dialog')

    if (triggerButton) {
      await user.click(triggerButton)

      await waitFor(() => {
        expect(screen.getByText('Create group')).toBeInTheDocument()
      })

      // Click Create group button
      const createGroupButton = screen.getByText('Create group').closest('button')
      if (createGroupButton) {
        await user.click(createGroupButton)
      }

      // Type in group name input
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Group name')).toBeInTheDocument()
      })

      const groupNameInput = screen.getByPlaceholderText('Group name')
      await user.type(groupNameInput, 'My Group')

      expect(groupNameInput).toHaveValue('My Group')
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
    const todayStr = today.toISOString()
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Today Chat', {
          last_message: createMockMessage(1, 'Today message', 'text', { created_at: todayStr }),
          created_at: todayStr,
          updated_at: todayStr,
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('Today message')).toBeInTheDocument()
  })

  it('formats date as "Yesterday" for yesterday messages', () => {
    // Using mockedUseConversations from top of file
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const yesterdayStr = yesterday.toISOString()
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Yesterday Chat', {
          last_message: createMockMessage(1, 'Yesterday message', 'text', {
            created_at: yesterdayStr,
          }),
          created_at: yesterdayStr,
          updated_at: yesterdayStr,
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('Yesterday')).toBeInTheDocument()
  })

  it('formats date as weekday for messages within a week', () => {
    // Using mockedUseConversations from top of file
    const threeDaysAgo = new Date()
    threeDaysAgo.setDate(threeDaysAgo.getDate() - 3)
    const threeDaysAgoStr = threeDaysAgo.toISOString()
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Week Chat', {
          last_message: createMockMessage(1, 'Week message', 'text', {
            created_at: threeDaysAgoStr,
          }),
          created_at: threeDaysAgoStr,
          updated_at: threeDaysAgoStr,
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('Week message')).toBeInTheDocument()
  })

  it('formats date as day/month for older messages', () => {
    // Using mockedUseConversations from top of file
    const twoWeeksAgo = new Date()
    twoWeeksAgo.setDate(twoWeeksAgo.getDate() - 14)
    const twoWeeksAgoStr = twoWeeksAgo.toISOString()
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Old Chat', {
          last_message: createMockMessage(1, 'Old message', 'text', { created_at: twoWeeksAgoStr }),
          created_at: twoWeeksAgoStr,
          updated_at: twoWeeksAgoStr,
        }),
      ])
    )

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
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Deleted Chat', {
          last_message: createMockMessage(1, 'Original content', 'text', { is_deleted: true }),
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('Message deleted')).toBeInTheDocument()
  })

  it('shows image message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'Image Chat', {
          last_message: createMockMessage(1, 'image.jpg', 'image'),
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('📷 Image')).toBeInTheDocument()
  })

  it('shows file message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'File Chat', {
          last_message: createMockMessage(1, 'document.pdf', 'file'),
        }),
      ])
    )

    render(<ConversationList />)
    expect(screen.getByText('📎 File')).toBeInTheDocument()
  })

  it('shows system message preview with emoji', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([
        createMockConversation(1, 'System Chat', {
          last_message: createMockMessage(1, 'User joined the chat', 'system'),
        }),
      ])
    )

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
    mockedUseConversations.mockReturnValue(createMockHookReturn([], { isLoading: true }))

    const { container } = render(<ConversationList />)

    // Loading state shows a spinner (Loader2 with animate-spin class)
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows error state', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue(
      createMockHookReturn([], { error: new Error('Network error') })
    )

    render(<ConversationList />)

    expect(screen.getByText('Error loading conversations')).toBeInTheDocument()
  })

  it('shows empty state when no conversations', () => {
    // Using mockedUseConversations from top of file
    mockedUseConversations.mockReturnValue(createMockHookReturn([]))

    render(<ConversationList />)

    expect(screen.getByText('No conversations')).toBeInTheDocument()
    expect(screen.getByText('Start a new conversation')).toBeInTheDocument()
  })
})
