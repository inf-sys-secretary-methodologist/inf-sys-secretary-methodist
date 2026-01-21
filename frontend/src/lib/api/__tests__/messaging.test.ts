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

      await messagingApi.sendMessage(1, {
        content: 'With file',
        attachments: [{ file_id: 'abc123', filename: 'doc.pdf' }],
      })

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/conversations/1/messages', {
        content: 'With file',
        attachments: [{ file_id: 'abc123', filename: 'doc.pdf' }],
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
  let mockWebSocket: {
    readyState: number
    send: jest.Mock
    close: jest.Mock
    onopen: (() => void) | null
    onmessage: ((event: { data: string }) => void) | null
    onclose: (() => void) | null
    onerror: ((error: Error) => void) | null
  }

  beforeEach(() => {
    mockWebSocket = {
      readyState: WebSocket.CONNECTING,
      send: jest.fn(),
      close: jest.fn(),
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ;(global as any).WebSocket = jest.fn(() => mockWebSocket)
  })

  it('creates websocket instance', () => {
    const ws = new MessagingWebSocket(() => 'test-token')
    expect(ws).toBeDefined()
  })

  it('connects to websocket server', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()

    // Simulate successful connection
    mockWebSocket.readyState = WebSocket.OPEN
    if (mockWebSocket.onopen) {
      mockWebSocket.onopen()
    }

    await connectPromise

    expect(ws.isConnected).toBe(true)
  })

  it('disconnects from websocket server', async () => {
    const ws = new MessagingWebSocket(() => 'test-token')

    const connectPromise = ws.connect()
    mockWebSocket.readyState = WebSocket.OPEN
    if (mockWebSocket.onopen) {
      mockWebSocket.onopen()
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
})
