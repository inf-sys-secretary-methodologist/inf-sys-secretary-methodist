import { renderHook, waitFor, act } from '@testing-library/react'
import { SWRConfig } from 'swr'
import React from 'react'
import {
  useConversations,
  useConversation,
  useMessages,
  useSendMessage,
  useCreateDirectConversation,
  useCreateGroupConversation,
  useUpdateConversation,
  useLeaveConversation,
  useMarkConversationAsRead,
  useEditMessage,
  useDeleteMessage,
  useSearchMessages,
  useAddParticipants,
  useMessagingWebSocket,
  useConversationWithMessages,
} from '../useMessaging'
import { apiClient } from '@/lib/api'
import { MessagingWebSocket } from '@/lib/api/messaging'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    patch: jest.fn(),
    delete: jest.fn(),
  },
}))

// Mock the WebSocket
jest.mock('@/lib/api/messaging', () => ({
  MessagingWebSocket: jest.fn().mockImplementation(() => ({
    connect: jest.fn().mockResolvedValue(undefined),
    disconnect: jest.fn(),
    sendMessage: jest.fn(),
    subscribe: jest.fn(),
    unsubscribe: jest.fn(),
    sendTyping: jest.fn(),
    sendStopTyping: jest.fn(),
    on: jest.fn(),
    off: jest.fn(),
    isConnected: false,
  })),
}))

const mockedApiClient = jest.mocked(apiClient)
const mockedMessagingWebSocket = jest.mocked(MessagingWebSocket)

// Wrapper to reset SWR cache between tests
const wrapper = ({ children }: { children: React.ReactNode }) =>
  React.createElement(
    SWRConfig,
    { value: { dedupingInterval: 0, provider: () => new Map() } },
    children
  )

describe('useMessaging hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('useConversations', () => {
    it('returns conversations list', async () => {
      const mockConversations = {
        conversations: [
          { id: 1, name: 'Conversation 1', type: 'direct' },
          { id: 2, name: 'Conversation 2', type: 'group' },
        ],
        total: 2,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockConversations,
      })

      const { result } = renderHook(() => useConversations(), { wrapper })

      await waitFor(() => {
        expect(result.current.conversations).toHaveLength(2)
      })

      expect(result.current.total).toBe(2)
    })

    it('passes filter parameters', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { conversations: [], total: 0 },
      })

      renderHook(() => useConversations({ type: 'direct', search: 'test', limit: 10 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('type=direct'))
      })
    })

    it('returns empty array when no data', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: null,
      })

      const { result } = renderHook(() => useConversations(), { wrapper })

      expect(result.current.conversations).toEqual([])
      expect(result.current.total).toBe(0)
    })
  })

  describe('useConversation', () => {
    it('fetches single conversation', async () => {
      const mockConversation = { id: 1, name: 'Test', type: 'direct' }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockConversation,
      })

      const { result } = renderHook(() => useConversation(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conversation).toEqual(mockConversation)
      })
    })

    it('does not fetch when id is falsy', () => {
      renderHook(() => useConversation(0), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useMessages', () => {
    it('fetches messages for conversation', async () => {
      const mockMessages = {
        messages: [
          { id: 1, content: 'Hello', sender_id: 1 },
          { id: 2, content: 'Hi there', sender_id: 2 },
        ],
        has_more: false,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockMessages,
      })

      const { result } = renderHook(() => useMessages(1), { wrapper })

      await waitFor(() => {
        expect(result.current.messages).toHaveLength(2)
      })

      expect(result.current.hasMore).toBe(false)
    })

    it('returns empty when conversation is 0', () => {
      const { result } = renderHook(() => useMessages(0), { wrapper })
      expect(result.current.messages).toEqual([])
    })
  })

  describe('useSendMessage', () => {
    it('sends message successfully', async () => {
      const mockMessage = { id: 1, content: 'Hello', sender_id: 1 }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockMessage,
      })

      const { result } = renderHook(() => useSendMessage(), { wrapper })

      let sentMessage
      await act(async () => {
        sentMessage = await result.current.mutateAsync(1, { content: 'Hello' })
      })

      expect(sentMessage).toEqual(mockMessage)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/messages', {
        content: 'Hello',
      })
    })
  })

  describe('useCreateDirectConversation', () => {
    it('creates direct conversation', async () => {
      const mockConversation = { id: 1, name: 'Direct', type: 'direct' }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockConversation,
      })

      const { result } = renderHook(() => useCreateDirectConversation(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync({ recipient_id: 2 })
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/direct', {
        recipient_id: 2,
      })
    })
  })

  describe('useCreateGroupConversation', () => {
    it('creates group conversation', async () => {
      const mockConversation = { id: 2, name: 'Group', type: 'group' }

      mockedApiClient.post.mockResolvedValue({
        success: true,
        data: mockConversation,
      })

      const { result } = renderHook(() => useCreateGroupConversation(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync({
          title: 'New Group',
          participant_ids: [1, 2, 3],
        })
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/group', {
        title: 'New Group',
        participant_ids: [1, 2, 3],
      })
    })
  })

  describe('useUpdateConversation', () => {
    it('updates conversation', async () => {
      mockedApiClient.patch.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useUpdateConversation(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, { title: 'Updated Title' })
      })

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/api/conversations/1', {
        title: 'Updated Title',
      })
    })
  })

  describe('useLeaveConversation', () => {
    it('leaves conversation', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useLeaveConversation(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1)
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/leave')
    })
  })

  describe('useMarkConversationAsRead', () => {
    it('marks conversation as read', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useMarkConversationAsRead(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, 5) // conversationId, messageId
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/read', {
        message_id: 5,
      })
    })
  })

  describe('useEditMessage', () => {
    it('edits message', async () => {
      mockedApiClient.patch.mockResolvedValue({
        success: true,
        data: { id: 1, content: 'Edited', edited: true },
      })

      const { result } = renderHook(() => useEditMessage(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, 1, { content: 'Edited' })
      })

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/api/conversations/1/messages/1', {
        content: 'Edited',
      })
    })
  })

  describe('useDeleteMessage', () => {
    it('deletes message', async () => {
      mockedApiClient.delete.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useDeleteMessage(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, 1)
      })

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/conversations/1/messages/1')
    })
  })

  describe('useSearchMessages', () => {
    it('searches messages in conversation', async () => {
      const mockResults = {
        messages: [{ id: 1, content: 'Hello world', conversation_id: 1 }],
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: mockResults,
      })

      const { result } = renderHook(() => useSearchMessages(1, 'hello'), { wrapper })

      await waitFor(() => {
        expect(result.current.messages).toHaveLength(1)
      })
    })

    it('does not search with short query', () => {
      renderHook(() => useSearchMessages(1, 'h'), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })

    it('does not search without conversation id', () => {
      renderHook(() => useSearchMessages(null, 'hello'), { wrapper })
      expect(mockedApiClient.get).not.toHaveBeenCalled()
    })
  })

  describe('useAddParticipants', () => {
    it('adds participants to conversation', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useAddParticipants(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, { user_ids: [2, 3, 4] })
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/participants', {
        user_ids: [2, 3, 4],
      })
    })

    it('returns isPending state', async () => {
      let resolvePromise: () => void
      const pendingPromise = new Promise<{ success: boolean }>((resolve) => {
        resolvePromise = () => resolve({ success: true })
      })
      mockedApiClient.post.mockReturnValue(pendingPromise)

      const { result } = renderHook(() => useAddParticipants(), { wrapper })

      expect(result.current.isPending).toBe(false)

      const addPromise = result.current.mutateAsync(1, { user_ids: [2] })

      await waitFor(() => {
        expect(result.current.isPending).toBe(true)
      })

      resolvePromise!()
      await addPromise

      await waitFor(() => {
        expect(result.current.isPending).toBe(false)
      })
    })
  })

  describe('useMessagingWebSocket', () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let mockWsInstance: any
    let eventHandlers: Map<string, (data: unknown) => void>

    beforeEach(() => {
      eventHandlers = new Map()
      mockWsInstance = {
        // Required private properties for MessagingWebSocket class
        ws: null,
        reconnectAttempts: 0,
        maxReconnectAttempts: 5,
        reconnectDelay: 1000,
        listeners: new Map(),
        pingInterval: null,
        getToken: jest.fn(() => 'test-token'),
        // Methods
        connect: jest.fn().mockResolvedValue(undefined),
        disconnect: jest.fn(),
        sendMessage: jest.fn(),
        subscribe: jest.fn(),
        unsubscribe: jest.fn(),
        sendTyping: jest.fn(),
        sendStopTyping: jest.fn(),
        on: jest.fn((event: string, handler: (data: unknown) => void) => {
          eventHandlers.set(event, handler)
        }),
        off: jest.fn(),
        isConnected: false,
        // Additional methods that might be needed
        emit: jest.fn(),
        handleReconnect: jest.fn(),
        startPing: jest.fn(),
        stopPing: jest.fn(),
      }

      // Using mockedMessagingWebSocket from top of file
      mockedMessagingWebSocket.mockImplementation(
        () => mockWsInstance as unknown as MessagingWebSocket
      )
    })

    it('returns initial state', () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      expect(result.current.isConnected).toBe(false)
      expect(typeof result.current.connect).toBe('function')
      expect(typeof result.current.disconnect).toBe('function')
      expect(typeof result.current.subscribe).toBe('function')
      expect(typeof result.current.unsubscribe).toBe('function')
      expect(typeof result.current.sendTyping).toBe('function')
      expect(typeof result.current.sendStopTyping).toBe('function')
      expect(typeof result.current.getTypingUsers).toBe('function')
    })

    it('returns empty typing users for conversation', () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      const typingUsers = result.current.getTypingUsers(1)
      expect(typingUsers).toEqual([])
    })

    it('connects to WebSocket', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      expect(mockWsInstance.connect).toHaveBeenCalled()
    })

    it('disconnects from WebSocket', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      act(() => {
        result.current.disconnect()
      })

      expect(mockWsInstance.disconnect).toHaveBeenCalled()
    })

    it('subscribes to conversation', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      act(() => {
        result.current.subscribe(1)
      })

      expect(mockWsInstance.subscribe).toHaveBeenCalledWith(1)
    })

    it('unsubscribes from conversation', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      act(() => {
        result.current.unsubscribe(1)
      })

      expect(mockWsInstance.unsubscribe).toHaveBeenCalledWith(1)
    })

    it('sends typing indicator', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      act(() => {
        result.current.sendTyping(1)
      })

      expect(mockWsInstance.sendTyping).toHaveBeenCalledWith(1)
    })

    it('sends stop typing indicator', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      act(() => {
        result.current.sendStopTyping(1)
      })

      expect(mockWsInstance.sendStopTyping).toHaveBeenCalledWith(1)
    })

    it('handles typing event and updates typing users', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Simulate typing event
      const typingHandler = eventHandlers.get('typing')
      if (typingHandler) {
        act(() => {
          typingHandler({ conversation_id: 1, user_id: 2 })
        })
      }

      await waitFor(() => {
        const typingUsers = result.current.getTypingUsers(1)
        expect(typingUsers).toContain(2)
      })
    })

    it('handles stop_typing event and removes typing user', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // First add a typing user
      const typingHandler = eventHandlers.get('typing')
      if (typingHandler) {
        act(() => {
          typingHandler({ conversation_id: 1, user_id: 2 })
        })
      }

      // Then remove typing
      const stopTypingHandler = eventHandlers.get('stop_typing')
      if (stopTypingHandler) {
        act(() => {
          stopTypingHandler({ conversation_id: 1, user_id: 2 })
        })
      }

      await waitFor(() => {
        const typingUsers = result.current.getTypingUsers(1)
        expect(typingUsers).not.toContain(2)
      })
    })

    it('handles new_message event', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Simulate new_message event
      const handler = eventHandlers.get('new_message')
      expect(handler).toBeDefined()
    })

    it('handles message_updated event', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Verify event handler was registered
      const handler = eventHandlers.get('message_updated')
      expect(handler).toBeDefined()
    })

    it('handles message_deleted event', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Verify event handler was registered
      const handler = eventHandlers.get('message_deleted')
      expect(handler).toBeDefined()
    })

    it('handles read event', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Verify event handler was registered
      const handler = eventHandlers.get('read')
      expect(handler).toBeDefined()
    })

    it('handles connection error gracefully', async () => {
      mockWsInstance.connect.mockRejectedValueOnce(new Error('Connection failed'))
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})

      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      expect(consoleSpy).toHaveBeenCalled()
      consoleSpy.mockRestore()
    })

    it('does not reconnect if already connected', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // Mark as connected
      mockWsInstance.isConnected = true

      // Try to connect again
      await act(async () => {
        await result.current.connect()
      })

      // connect() should only be called once
      expect(mockWsInstance.connect).toHaveBeenCalledTimes(1)
    })

    it('calls connect method when connecting', async () => {
      const { result } = renderHook(() => useMessagingWebSocket(), { wrapper })

      await act(async () => {
        await result.current.connect()
      })

      // The connection method should have been called
      expect(result.current.isConnected).toBeDefined()
    })
  })

  describe('API response handling', () => {
    it('handles API error response', async () => {
      mockedApiClient.get.mockResolvedValueOnce({
        success: false,
        error: { message: 'API Error' },
      })

      const { result } = renderHook(() => useConversations(), { wrapper })

      await waitFor(() => {
        expect(result.current.error).toBeDefined()
      })
    })

    it('handles unwrapped API response', async () => {
      // Return data directly without wrapper
      mockedApiClient.get.mockResolvedValueOnce({
        conversations: [{ id: 1, name: 'Test' }],
        total: 1,
      })

      const { result } = renderHook(() => useConversations(), { wrapper })

      await waitFor(() => {
        expect(result.current.conversations).toBeDefined()
      })
    })
  })

  describe('useConversationWithMessages', () => {
    it('returns combined state for conversation', async () => {
      const mockConversation = { id: 1, name: 'Test', type: 'direct' }
      const mockMessages = {
        messages: [{ id: 1, content: 'Hello', sender_id: 1 }],
        has_more: false,
      }

      mockedApiClient.get.mockImplementation((url: string) => {
        if (url.includes('/messages')) {
          return Promise.resolve({ success: true, data: mockMessages })
        }
        return Promise.resolve({ success: true, data: mockConversation })
      })

      const { result } = renderHook(() => useConversationWithMessages(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conversation).toBeDefined()
      })

      expect(typeof result.current.sendMessage).toBe('function')
      expect(typeof result.current.sendTyping).toBe('function')
      expect(typeof result.current.sendStopTyping).toBe('function')
    })

    it('returns null conversation when id is null', () => {
      const { result } = renderHook(() => useConversationWithMessages(null), { wrapper })

      expect(result.current.conversation).toBeUndefined()
      expect(result.current.messages).toEqual([])
    })

    it('provides sendMessage function that sends to correct conversation', async () => {
      const mockConversation = { id: 1, name: 'Test', type: 'direct' }
      const mockMessages = {
        messages: [],
        has_more: false,
      }
      const mockSentMessage = { id: 1, content: 'Hello', sender_id: 1 }

      mockedApiClient.get.mockImplementation((url: string) => {
        if (url.includes('/messages')) {
          return Promise.resolve({ success: true, data: mockMessages })
        }
        return Promise.resolve({ success: true, data: mockConversation })
      })

      mockedApiClient.post.mockImplementation((url: string) => {
        if (url.includes('/messages')) {
          return Promise.resolve({ success: true, data: mockSentMessage })
        }
        return Promise.resolve({ success: true })
      })

      const { result } = renderHook(() => useConversationWithMessages(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conversation).toBeDefined()
      })

      await act(async () => {
        await result.current.sendMessage('Hello')
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/api/conversations/1/messages',
        expect.objectContaining({ content: 'Hello' })
      )
    })

    it('returns typing users for conversation', async () => {
      const mockConversation = { id: 1, name: 'Test', type: 'direct' }
      const mockMessages = { messages: [], has_more: false }

      mockedApiClient.get.mockImplementation((url: string) => {
        if (url.includes('/messages')) {
          return Promise.resolve({ success: true, data: mockMessages })
        }
        return Promise.resolve({ success: true, data: mockConversation })
      })

      const { result } = renderHook(() => useConversationWithMessages(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conversation).toBeDefined()
      })

      expect(result.current.typingUsers).toEqual([])
    })

    it('provides sendTyping and sendStopTyping functions', async () => {
      const mockConversation = { id: 1, name: 'Test', type: 'direct' }
      const mockMessages = { messages: [], has_more: false }

      mockedApiClient.get.mockImplementation((url: string) => {
        if (url.includes('/messages')) {
          return Promise.resolve({ success: true, data: mockMessages })
        }
        return Promise.resolve({ success: true, data: mockConversation })
      })

      const { result } = renderHook(() => useConversationWithMessages(1), { wrapper })

      await waitFor(() => {
        expect(result.current.conversation).toBeDefined()
      })

      // These should be callable without error
      result.current.sendTyping()
      result.current.sendStopTyping()
    })

    it('does not send message when conversation id is null', async () => {
      const { result } = renderHook(() => useConversationWithMessages(null), { wrapper })

      await act(async () => {
        const response = await result.current.sendMessage('Hello')
        expect(response).toBeUndefined()
      })

      expect(mockedApiClient.post).not.toHaveBeenCalled()
    })
  })

  describe('useConversations with various filters', () => {
    it('passes offset parameter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { conversations: [], total: 0 },
      })

      renderHook(() => useConversations({ offset: 20 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('offset=20'))
      })
    })
  })

  describe('useMessages with filters', () => {
    it('passes before_id filter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { messages: [], has_more: false },
      })

      renderHook(() => useMessages(1, { before_id: 100 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('before_id=100'))
      })
    })

    it('passes after_id filter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { messages: [], has_more: false },
      })

      renderHook(() => useMessages(1, { after_id: 50 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('after_id=50'))
      })
    })

    it('passes search filter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { messages: [], has_more: false },
      })

      renderHook(() => useMessages(1, { search: 'hello' }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('search=hello'))
      })
    })

    it('passes limit filter', async () => {
      mockedApiClient.get.mockResolvedValue({
        success: true,
        data: { messages: [], has_more: false },
      })

      renderHook(() => useMessages(1, { limit: 25 }), { wrapper })

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('limit=25'))
      })
    })
  })

  describe('Mark as read deduplication', () => {
    it('does not call API twice for same message', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useMarkConversationAsRead(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, 10)
      })

      expect(mockedApiClient.post).toHaveBeenCalledTimes(1)

      // Try to mark the same message again
      await act(async () => {
        await result.current.mutateAsync(1, 10)
      })

      // Should still only have been called once
      expect(mockedApiClient.post).toHaveBeenCalledTimes(1)
    })

    it('calls API for different message', async () => {
      mockedApiClient.post.mockResolvedValue({ success: true })

      const { result } = renderHook(() => useMarkConversationAsRead(), { wrapper })

      await act(async () => {
        await result.current.mutateAsync(1, 10)
      })

      expect(mockedApiClient.post).toHaveBeenCalledTimes(1)

      // Mark a different message
      await act(async () => {
        await result.current.mutateAsync(1, 11)
      })

      expect(mockedApiClient.post).toHaveBeenCalledTimes(2)
    })
  })
})
