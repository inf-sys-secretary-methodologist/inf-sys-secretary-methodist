// Set up WebSocket mock BEFORE importing the module
let lastCreatedWebSocket: {
  readyState: number
  send: jest.Mock
  close: jest.Mock
  onopen: (() => void) | null
  onmessage: ((event: { data: string }) => void) | null
  onclose: ((event: { code: number; reason: string }) => void) | null
  onerror: (() => void) | null
} | null = null

const MockWebSocketConstructor = jest.fn().mockImplementation(function () {
  const ws = {
    readyState: 0, // CONNECTING
    send: jest.fn(),
    close: jest.fn(),
    onopen: null as (() => void) | null,
    onmessage: null as ((event: { data: string }) => void) | null,
    onclose: null as ((event: { code: number; reason: string }) => void) | null,
    onerror: null as (() => void) | null,
  }
  lastCreatedWebSocket = ws
  return ws
})

// Set up WebSocket constants
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(MockWebSocketConstructor as any).CONNECTING = 0
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(MockWebSocketConstructor as any).OPEN = 1
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(MockWebSocketConstructor as any).CLOSING = 2
// eslint-disable-next-line @typescript-eslint/no-explicit-any
;(MockWebSocketConstructor as any).CLOSED = 3

Object.defineProperty(global, 'WebSocket', {
  value: MockWebSocketConstructor,
  writable: true,
  configurable: true,
})

import { messagingApi, MessagingWebSocket } from '../messaging'
import { apiClient } from '../../api'

// Mock the API client
jest.mock('../../api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn(),
    patch: jest.fn(),
    delete: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('messagingApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('createDirectConversation', () => {
    it('creates direct conversation', async () => {
      const mockConv = { id: 1, type: 'direct' }
      mockedApiClient.post.mockResolvedValue({ data: mockConv })

      const result = await messagingApi.createDirectConversation({ recipient_id: 2 })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/direct', {
        recipient_id: 2,
      })
      expect(result).toEqual(mockConv)
    })
  })

  describe('createGroupConversation', () => {
    it('creates group conversation', async () => {
      const mockConv = { id: 2, type: 'group', title: 'Test Group' }
      mockedApiClient.post.mockResolvedValue({ data: mockConv })

      const result = await messagingApi.createGroupConversation({
        title: 'Test Group',
        participant_ids: [1, 2, 3],
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/group', {
        title: 'Test Group',
        participant_ids: [1, 2, 3],
      })
      expect(result).toEqual(mockConv)
    })
  })

  describe('listConversations', () => {
    it('lists conversations without filters', async () => {
      const mockList = { conversations: [], total: 0 }
      mockedApiClient.get.mockResolvedValue({ data: mockList })

      const result = await messagingApi.listConversations()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/conversations')
      expect(result).toEqual(mockList)
    })

    it('lists conversations with filters', async () => {
      const mockList = { conversations: [], total: 0 }
      mockedApiClient.get.mockResolvedValue({ data: mockList })

      await messagingApi.listConversations({
        type: 'direct',
        search: 'test',
        limit: 10,
        offset: 0,
      })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('/api/conversations?')
      )
    })
  })

  describe('getConversation', () => {
    it('fetches single conversation', async () => {
      const mockConv = { id: 1, type: 'direct' }
      mockedApiClient.get.mockResolvedValue({ data: mockConv })

      const result = await messagingApi.getConversation(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/conversations/1')
      expect(result).toEqual(mockConv)
    })
  })

  describe('updateConversation', () => {
    it('updates conversation', async () => {
      const mockConv = { id: 1, title: 'Updated Title' }
      mockedApiClient.patch.mockResolvedValue({ data: mockConv })

      const result = await messagingApi.updateConversation(1, { title: 'Updated Title' })

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/api/conversations/1', {
        title: 'Updated Title',
      })
      expect(result).toEqual(mockConv)
    })
  })

  describe('addParticipants', () => {
    it('adds participants to conversation', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Added' } })

      const result = await messagingApi.addParticipants(1, { user_ids: [2, 3] })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/participants', {
        user_ids: [2, 3],
      })
      expect(result).toEqual({ message: 'Added' })
    })
  })

  describe('leaveConversation', () => {
    it('leaves conversation', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Left' } })

      const result = await messagingApi.leaveConversation(1)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/leave')
      expect(result).toEqual({ message: 'Left' })
    })
  })

  describe('sendMessage', () => {
    it('sends message to conversation', async () => {
      const mockMessage = { id: 1, content: 'Hello', sender_id: 1 }
      mockedApiClient.post.mockResolvedValue({ data: mockMessage })

      const result = await messagingApi.sendMessage(1, { content: 'Hello' })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/messages', {
        content: 'Hello',
      })
      expect(result).toEqual(mockMessage)
    })

    it('sends message with attachments', async () => {
      const mockMessage = { id: 1, content: 'With file', attachments: [] }
      mockedApiClient.post.mockResolvedValue({ data: mockMessage })

      const mockAttachment = {
        file_id: 123,
        file_name: 'doc.pdf',
        file_size: 1024,
        mime_type: 'application/pdf',
        url: '/files/doc.pdf',
      }

      await messagingApi.sendMessage(1, {
        content: 'With file',
        attachments: [mockAttachment],
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/messages', {
        content: 'With file',
        attachments: [mockAttachment],
      })
    })
  })

  describe('getMessages', () => {
    it('fetches messages from conversation', async () => {
      const mockMessages = { messages: [], has_more: false }
      mockedApiClient.get.mockResolvedValue({ data: mockMessages })

      const result = await messagingApi.getMessages(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/conversations/1/messages')
      expect(result).toEqual(mockMessages)
    })

    it('fetches messages with pagination', async () => {
      const mockMessages = { messages: [], has_more: true }
      mockedApiClient.get.mockResolvedValue({ data: mockMessages })

      await messagingApi.getMessages(1, { before_id: 100, limit: 50 })

      expect(mockedApiClient.get).toHaveBeenCalledWith(expect.stringContaining('before_id=100'))
    })
  })

  describe('editMessage', () => {
    it('edits message', async () => {
      const mockMessage = { id: 1, content: 'Edited', edited: true }
      mockedApiClient.patch.mockResolvedValue({ data: mockMessage })

      const result = await messagingApi.editMessage(1, 1, { content: 'Edited' })

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/api/conversations/1/messages/1', {
        content: 'Edited',
      })
      expect(result).toEqual(mockMessage)
    })
  })

  describe('deleteMessage', () => {
    it('deletes message', async () => {
      mockedApiClient.delete.mockResolvedValue({ data: { message: 'Deleted' } })

      await messagingApi.deleteMessage(1, 1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/conversations/1/messages/1')
    })
  })

  describe('markAsRead', () => {
    it('marks conversation as read', async () => {
      mockedApiClient.post.mockResolvedValue({ data: { message: 'Marked' } })

      await messagingApi.markAsRead(1, { message_id: 10 })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/read', {
        message_id: 10,
      })
    })
  })

  describe('searchMessages', () => {
    it('searches messages in conversation', async () => {
      const mockResults = { messages: [], total: 0 }
      mockedApiClient.get.mockResolvedValue({ data: mockResults })

      await messagingApi.searchMessages(1, { q: 'test' })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('/api/conversations/1/messages/search')
      )
    })
  })
})

describe('MessagingWebSocket', () => {
  beforeEach(() => {
    lastCreatedWebSocket = null
    MockWebSocketConstructor.mockClear()
  })

  it('creates websocket instance', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    expect(ws).toBeDefined()
  })

  it('connects to websocket server', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()

    // Simulate successful connection
    if (lastCreatedWebSocket) {
      lastCreatedWebSocket.readyState = WebSocket.OPEN
      lastCreatedWebSocket.onopen?.()
    }

    await connectPromise

    expect(ws.isConnected).toBe(true)
  })

  it('disconnects from websocket server', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    if (lastCreatedWebSocket) {
      lastCreatedWebSocket.readyState = WebSocket.OPEN
      lastCreatedWebSocket.onopen?.()
    }
    await connectPromise

    expect(ws.isConnected).toBe(true)

    // Disconnect should not throw
    expect(() => ws.disconnect()).not.toThrow()
  })

  it('handles event listeners', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    const callback = jest.fn()

    ws.on('new_message', callback)
    ws.off('new_message', callback)

    // No errors should be thrown
    expect(true).toBe(true)
  })

  it('rejects connect when no token available', async () => {
    const ws = new MessagingWebSocket(() => null)

    await expect(ws.connect()).rejects.toThrow('No auth token available')
  })

  it('resolves immediately when already connected', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    // First connect
    const connectPromise1 = ws.connect()
    if (lastCreatedWebSocket) {
      lastCreatedWebSocket.readyState = WebSocket.OPEN
      lastCreatedWebSocket.onopen?.()
    }
    await connectPromise1

    // Second connect should resolve immediately
    await expect(ws.connect()).resolves.toBeUndefined()
  })

  it('handles incoming messages', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    const callback = jest.fn()

    ws.on('new_message', callback)

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    // Simulate incoming message
    lastCreatedWebSocket!.onmessage?.({
      data: JSON.stringify({ type: 'new_message', content: 'hello' }),
    })

    expect(callback).toHaveBeenCalledWith({ type: 'new_message', content: 'hello' })
  })

  it('handles wildcard event listeners', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    const callback = jest.fn()

    ws.on('*', callback)

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    lastCreatedWebSocket!.onmessage?.({ data: JSON.stringify({ type: 'any_event', data: 'test' }) })

    expect(callback).toHaveBeenCalled()
  })

  it('handles message parse errors gracefully', async () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    // Send invalid JSON
    lastCreatedWebSocket!.onmessage?.({ data: 'not valid json' })

    expect(consoleSpy).toHaveBeenCalledWith('Failed to parse WebSocket message:', expect.any(Error))
    consoleSpy.mockRestore()
  })

  it('handles WebSocket errors', async () => {
    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation()
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    // Simulate error (onerror doesn't receive useful info)
    lastCreatedWebSocket!.onerror?.()

    expect(consoleSpy).toHaveBeenCalledWith('WebSocket connection error')
    consoleSpy.mockRestore()
  })

  it('rejects when connection fails before opening', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()

    // Simulate close before open (connection failed)
    lastCreatedWebSocket!.onclose?.({ code: 1006, reason: 'Connection failed' })

    await expect(connectPromise).rejects.toThrow('WebSocket connection failed: 1006')
  })

  it('attempts reconnect when connection closes after being open', async () => {
    jest.useFakeTimers()
    const consoleSpy = jest.spyOn(console, 'warn').mockImplementation()
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    // Simulate unexpected close
    lastCreatedWebSocket!.onclose?.({ code: 1006, reason: '' })

    expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('WebSocket reconnecting'))

    consoleSpy.mockRestore()
    jest.useRealTimers()
  })

  it('sends messages when connected', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.send({ type: 'test', data: 'hello' })

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'test', data: 'hello' })
    )
  })

  it('does not send messages when not connected', () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    // Trigger connect to create the WebSocket instance but don't complete the connection
    ws.connect()
    // Send before connection completes - WebSocket exists but is not OPEN
    ws.send({ type: 'test' })

    expect(lastCreatedWebSocket!.send).not.toHaveBeenCalled()
  })

  it('subscribes to conversation', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.subscribe(123)

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'subscribe', conversation_id: 123 })
    )
  })

  it('unsubscribes from conversation', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.unsubscribe(123)

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'unsubscribe', conversation_id: 123 })
    )
  })

  it('sends typing indicator', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.sendTyping(123)

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'typing', conversation_id: 123 })
    )
  })

  it('sends stop typing indicator', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.sendStopTyping(123)

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'stop_typing', conversation_id: 123 })
    )
  })

  it('closes WebSocket on disconnect', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.disconnect()

    expect(lastCreatedWebSocket!.close).toHaveBeenCalled()
  })

  it('returns unsubscribe function from on()', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    const callback = jest.fn()

    const unsubscribe = ws.on('test_event', callback)

    expect(typeof unsubscribe).toBe('function')

    // Should be able to unsubscribe
    unsubscribe()
  })

  it('starts ping interval on connect', async () => {
    jest.useFakeTimers()
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    // Advance by 30 seconds (ping interval)
    jest.advanceTimersByTime(30000)

    expect(lastCreatedWebSocket!.send).toHaveBeenCalledWith(JSON.stringify({ type: 'ping' }))

    jest.useRealTimers()
  })

  it('stops ping interval on disconnect', async () => {
    jest.useFakeTimers()
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    lastCreatedWebSocket!.readyState = WebSocket.OPEN
    lastCreatedWebSocket!.onopen?.()
    await connectPromise

    ws.disconnect()

    // Clear mocks after disconnect
    lastCreatedWebSocket!.send.mockClear()

    // Advance by 30 seconds
    jest.advanceTimersByTime(30000)

    // Should not have sent ping after disconnect
    expect(lastCreatedWebSocket!.send).not.toHaveBeenCalled()

    jest.useRealTimers()
  })

  it('handles off() for non-existing listener gracefully', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    const callback = jest.fn()

    // Should not throw when removing non-existing listener
    expect(() => ws.off('non_existing', callback)).not.toThrow()
  })

  it('isConnected returns false when WebSocket is null', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    expect(ws.isConnected).toBe(false)
  })

  it('isConnected returns false when WebSocket is not OPEN', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    ws.connect()
    // Still connecting
    lastCreatedWebSocket!.readyState = WebSocket.CONNECTING

    expect(ws.isConnected).toBe(false)
  })
})
