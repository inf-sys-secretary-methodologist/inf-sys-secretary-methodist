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
} from '../useMessaging'
import { apiClient } from '@/lib/api'

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
    connect: jest.fn(),
    disconnect: jest.fn(),
    sendMessage: jest.fn(),
    on: jest.fn(),
    off: jest.fn(),
  })),
}))

const mockedApiClient = jest.mocked(apiClient)

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
})
