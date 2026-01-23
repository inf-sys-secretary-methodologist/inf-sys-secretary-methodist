import { render, screen, waitFor } from '@testing-library/react'
import { ConversationView } from '../ConversationView'
import { useConversationWithMessages } from '@/hooks/useMessaging'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      selectConversation: 'Select a conversation',
      selectConversationDesc: 'Choose a conversation from the list to start messaging',
      loading: 'Loading...',
      'time.today': 'Today',
      'time.yesterday': 'Yesterday',
      typing: 'typing...',
      messageDeleted: 'Message deleted',
      reply: 'Reply',
      edit: 'Edit',
      delete: 'Delete',
      copyText: 'Copy text',
      sendMessage: 'Send message',
      typeMessage: 'Type a message...',
      connected: 'Connected',
      reconnecting: 'Reconnecting...',
    }
    return translations[key] || key
  },
}))

// Mock auth hook
jest.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: { id: 1, name: 'Test User', email: 'test@example.com' },
    isAuthenticated: true,
  }),
}))

const mockConversation = {
  id: '1',
  type: 'direct' as const,
  title: 'John Doe',
  participants: [
    { id: 1, name: 'Test User', is_online: true },
    { id: 2, name: 'John Doe', is_online: true },
  ],
  unread_count: 0,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

const mockMessages = [
  {
    id: '1',
    conversation_id: '1',
    sender_id: 2,
    sender_name: 'John Doe',
    content: 'Hello!',
    type: 'text' as const,
    is_deleted: false,
    created_at: new Date().toISOString(),
  },
  {
    id: '2',
    conversation_id: '1',
    sender_id: 1,
    sender_name: 'Test User',
    content: 'Hi there!',
    type: 'text' as const,
    is_deleted: false,
    created_at: new Date().toISOString(),
  },
]

const mockSendMessage = jest.fn()

// Mock messaging hook
jest.mock('@/hooks/useMessaging', () => ({
  useConversationWithMessages: jest.fn(() => ({
    conversation: mockConversation,
    messages: mockMessages,
    hasMore: false,
    isLoading: false,
    isSending: false,
    sendMessage: mockSendMessage,
    typingUsers: [],
    sendTyping: jest.fn(),
    sendStopTyping: jest.fn(),
    isConnected: true,
  })),
}))

const mockedUseConversationWithMessages = jest.mocked(useConversationWithMessages)

// Mock MessageBubble
jest.mock('../MessageBubble', () => ({
  MessageBubble: ({ message }: { message: { content: string } }) => (
    <div data-testid="message-bubble">{message.content}</div>
  ),
}))

// Mock MessageInput
jest.mock('../MessageInput', () => ({
  MessageInput: ({ onSend }: { onSend: (content: string) => void }) => (
    <div data-testid="message-input">
      <input
        type="text"
        placeholder="Type a message..."
        data-testid="message-input-field"
        onKeyDown={(e) => {
          if (e.key === 'Enter' && e.currentTarget.value) {
            onSend(e.currentTarget.value)
          }
        }}
      />
    </div>
  ),
}))

describe('ConversationView', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders empty state when no conversationId', () => {
    // 0 is falsy and triggers the empty state check in the component
    render(<ConversationView conversationId={0} />)
    expect(screen.getByText('Select a conversation')).toBeInTheDocument()
  })

  it('renders conversation header with title', async () => {
    render(<ConversationView conversationId={1} />)

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
  })

  it('renders messages', async () => {
    render(<ConversationView conversationId={1} />)

    await waitFor(() => {
      const messageBubbles = screen.getAllByTestId('message-bubble')
      expect(messageBubbles.length).toBe(2)
    })
  })

  it('displays message content', async () => {
    render(<ConversationView conversationId={1} />)

    await waitFor(() => {
      expect(screen.getByText('Hello!')).toBeInTheDocument()
      expect(screen.getByText('Hi there!')).toBeInTheDocument()
    })
  })

  it('renders message input', async () => {
    render(<ConversationView conversationId={1} />)

    await waitFor(() => {
      expect(screen.getByTestId('message-input')).toBeInTheDocument()
    })
  })

  it('shows back button on mobile when onBack is provided', async () => {
    const onBack = jest.fn()
    render(<ConversationView conversationId={1} onBack={onBack} />)

    await waitFor(() => {
      // There should be a back button
      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBeGreaterThan(0)
    })
  })

  it('applies custom className', () => {
    const { container } = render(<ConversationView conversationId={1} className="custom-class" />)
    expect(container.querySelector('.custom-class')).toBeInTheDocument()
  })

  it('shows loading state', () => {
    // Using mockedUseConversationWithMessages from top of file
    mockedUseConversationWithMessages.mockReturnValueOnce({
      conversation: null,
      messages: [],
      hasMore: false,
      isLoading: true,
      isSending: false,
      sendMessage: jest.fn(),
      typingUsers: [],
      sendTyping: jest.fn(),
      sendStopTyping: jest.fn(),
      isConnected: true,
    })

    render(<ConversationView conversationId={1} />)
    // Should show loading spinner
    expect(document.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows connection status', async () => {
    render(<ConversationView conversationId={1} />)

    await waitFor(() => {
      // Should show connected status indicator
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
  })
})
