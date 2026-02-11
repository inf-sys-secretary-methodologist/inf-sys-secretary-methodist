import { aiApi, handleAIStream } from '../ai'
import { apiClient } from '../../api'
import type {
  AIConversation,
  AIConversationListOutput,
  CreateAIConversationInput,
  UpdateAIConversationInput,
  AIMessageListOutput,
  SendAIMessageInput,
  AIChatResponse,
  AISearchInput,
  AISearchOutput,
  IndexDocumentOutput,
  DocumentSource,
} from '@/types/ai'

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

describe('aiApi', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('listConversations', () => {
    it('fetches conversations without filters', async () => {
      const mockResponse = {
        data: {
          conversations: [] as AIConversation[],
          total: 0,
        } as AIConversationListOutput,
      }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      const result = await aiApi.listConversations()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ai/conversations')
      expect(result).toEqual(mockResponse.data)
    })

    it('fetches conversations with filters', async () => {
      const mockResponse = {
        data: {
          conversations: [] as AIConversation[],
          total: 0,
        } as AIConversationListOutput,
      }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await aiApi.listConversations({ search: 'test', limit: 10, offset: 5 })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/ai/conversations?search=test&limit=10&offset=5'
      )
    })
  })

  describe('getConversation', () => {
    it('fetches single conversation', async () => {
      const mockConversation: AIConversation = {
        id: 1,
        user_id: 1,
        title: 'Test Conversation',
        last_message: 'Last message text',
        message_count: 0,
        model: 'gpt-4o-mini',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      const mockResponse = { data: mockConversation }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      const result = await aiApi.getConversation(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ai/conversations/1')
      expect(result).toEqual(mockConversation)
    })
  })

  describe('createConversation', () => {
    it('creates new conversation', async () => {
      const input: CreateAIConversationInput = { title: 'New Chat' }
      const mockConversation: AIConversation = {
        id: 1,
        user_id: 1,
        title: 'New Chat',
        last_message: '2024-01-01T00:00:00Z',
        message_count: 0,
        model: 'gpt-4o-mini',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      const mockResponse = { data: mockConversation }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const result = await aiApi.createConversation(input)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/conversations', input)
      expect(result).toEqual(mockConversation)
    })

    it('creates conversation without input', async () => {
      const mockConversation: AIConversation = {
        id: 1,
        user_id: 1,
        title: '',
        last_message: '2024-01-01T00:00:00Z',
        message_count: 0,
        model: 'gpt-4o-mini',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      const mockResponse = { data: mockConversation }
      mockedApiClient.post.mockResolvedValue(mockResponse)

      const result = await aiApi.createConversation()

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/conversations', {})
      expect(result).toEqual(mockConversation)
    })
  })

  describe('updateConversation', () => {
    it('updates conversation title', async () => {
      const input: UpdateAIConversationInput = { title: 'Updated Title' }
      const mockConversation: AIConversation = {
        id: 1,
        user_id: 1,
        title: 'Updated Title',
        last_message: '2024-01-01T00:00:00Z',
        message_count: 0,
        model: 'gpt-4o-mini',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      const mockResponse = { data: mockConversation }
      mockedApiClient.patch.mockResolvedValue(mockResponse)

      const result = await aiApi.updateConversation(1, input)

      expect(mockedApiClient.patch).toHaveBeenCalledWith('/api/ai/conversations/1', input)
      expect(result).toEqual(mockConversation)
    })
  })

  describe('deleteConversation', () => {
    it('deletes conversation', async () => {
      const mockResponse = { data: { message: 'Conversation deleted' } }
      mockedApiClient.delete.mockResolvedValue(mockResponse)

      const result = await aiApi.deleteConversation(1)

      expect(mockedApiClient.delete).toHaveBeenCalledWith('/api/ai/conversations/1')
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('getMessages', () => {
    it('fetches messages without filters', async () => {
      const mockResponse = {
        data: {
          messages: [],
          has_more: false,
        } as AIMessageListOutput,
      }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      const result = await aiApi.getMessages(1)

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ai/conversations/1/messages')
      expect(result).toEqual(mockResponse.data)
    })

    it('fetches messages with filters', async () => {
      const mockResponse = {
        data: {
          messages: [],
          has_more: false,
        } as AIMessageListOutput,
      }
      mockedApiClient.get.mockResolvedValue(mockResponse)

      await aiApi.getMessages(1, { before_id: 10, limit: 20 })

      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/api/ai/conversations/1/messages?before_id=10&limit=20'
      )
    })
  })

  describe('chat', () => {
    it('sends message and gets response', async () => {
      const input: SendAIMessageInput = {
        content: 'Hello AI',
        conversation_id: 1,
      }
      const mockResponse: AIChatResponse = {
        message: {
          id: 1,
          conversation_id: 1,
          role: 'assistant',
          content: 'Hello! How can I help?',
          status: 'complete',
          created_at: '2024-01-01T00:00:00Z',
        },
        conversation_id: 1,
        sources: [],
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.post.mockResolvedValue(apiResponse)

      const result = await aiApi.chat(input)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/chat', input)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('chatStream', () => {
    it('creates EventSource with correct URL', () => {
      const input: SendAIMessageInput = {
        content: 'Hello',
        conversation_id: 1,
        include_sources: true,
        max_sources: 3,
      }

      // Mock localStorage
      Object.defineProperty(window, 'localStorage', {
        value: {
          getItem: jest.fn(() => 'test-token'),
        },
        writable: true,
      })

      // Mock EventSource constructor
      const mockEventSource = {
        close: jest.fn(),
      }
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      global.EventSource = jest.fn(() => mockEventSource) as any

      aiApi.chatStream(input)

      expect(EventSource).toHaveBeenCalledWith(
        expect.stringContaining('/api/ai/chat/stream?content=Hello&conversation_id=1')
      )
      expect(EventSource).toHaveBeenCalledWith(
        expect.stringContaining('include_sources=true&max_sources=3&token=test-token')
      )
    })
  })

  describe('search', () => {
    it('performs semantic search', async () => {
      const input: AISearchInput = {
        query: 'test query',
        limit: 5,
      }
      const mockResponse: AISearchOutput = {
        query: 'test query',
        results: [],
        total: 0,
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.post.mockResolvedValue(apiResponse)

      const result = await aiApi.search(input)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/search', input)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('indexDocument', () => {
    it('indexes single document', async () => {
      const mockResponse: IndexDocumentOutput = {
        document_id: 123,
        status: 'indexed',
        chunks_created: 5,
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.post.mockResolvedValue(apiResponse)

      const result = await aiApi.indexDocument(123)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/index/123', {
        document_id: 123,
        force_reindex: undefined,
      })
      expect(result).toEqual(mockResponse)
    })

    it('indexes document with force flag', async () => {
      const mockResponse: IndexDocumentOutput = {
        document_id: 123,
        status: 'indexed',
        chunks_created: 5,
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.post.mockResolvedValue(apiResponse)

      await aiApi.indexDocument(123, true)

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/index/123', {
        document_id: 123,
        force_reindex: true,
      })
    })
  })

  describe('indexDocuments', () => {
    it('batch indexes multiple documents', async () => {
      const mockResponse = {
        results: [
          { document_id: 1, status: 'indexed', chunks_created: 5 },
          { document_id: 2, status: 'indexed', chunks_created: 3 },
        ],
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.post.mockResolvedValue(apiResponse)

      const result = await aiApi.indexDocuments([1, 2])

      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/ai/index/batch', {
        document_ids: [1, 2],
        force_reindex: undefined,
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getIndexingStatus', () => {
    it('fetches indexing status', async () => {
      const mockResponse = {
        total_documents: 100,
        indexed_documents: 80,
        pending_documents: 20,
        last_indexed_at: '2024-01-01T00:00:00Z',
      }
      const apiResponse = { data: mockResponse }
      mockedApiClient.get.mockResolvedValue(apiResponse)

      const result = await aiApi.getIndexingStatus()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/ai/index/status')
      expect(result).toEqual(mockResponse)
    })
  })
})

describe('handleAIStream', () => {
  let mockEventSource: // eslint-disable-next-line @typescript-eslint/no-explicit-any
  any

  beforeEach(() => {
    mockEventSource = {
      onmessage: null,
      onerror: null,
      close: jest.fn(),
    }
  })

  it('handles content events', async () => {
    const onContent = jest.fn()
    const _promise = handleAIStream(mockEventSource, { onContent })

    // Simulate content event
    mockEventSource.onmessage({
      data: JSON.stringify({ type: 'content', content: 'Hello' }),
    })

    expect(onContent).toHaveBeenCalledWith('Hello')
  })

  it('handles source events', async () => {
    const onSource = jest.fn()
    const _promise = handleAIStream(mockEventSource, { onSource })

    const source: DocumentSource = {
      id: 1,
      document_id: 1,
      document_title: 'Test Doc',
      chunk_text: 'Sample text',
      similarity_score: 0.95,
    }

    // Simulate source event
    mockEventSource.onmessage({
      data: JSON.stringify({ type: 'source', source }),
    })

    expect(onSource).toHaveBeenCalledWith(source)
  })

  it('handles done event and closes stream', async () => {
    const onDone = jest.fn()
    const _promise = handleAIStream(mockEventSource, { onDone })

    // Simulate done event
    mockEventSource.onmessage({
      data: JSON.stringify({ type: 'done', message_id: 42 }),
    })

    await _promise

    expect(onDone).toHaveBeenCalledWith(42)
    expect(mockEventSource.close).toHaveBeenCalled()
  })

  it('handles error event and rejects promise', async () => {
    const onError = jest.fn()
    const _promise = handleAIStream(mockEventSource, { onError })

    // Simulate error event
    mockEventSource.onmessage({
      data: JSON.stringify({ type: 'error', error: 'Something went wrong' }),
    })

    await expect(_promise).rejects.toThrow('Something went wrong')
    expect(onError).toHaveBeenCalledWith('Something went wrong')
    expect(mockEventSource.close).toHaveBeenCalled()
  })

  it('handles connection error', async () => {
    const onError = jest.fn()
    const _promise = handleAIStream(mockEventSource, { onError })

    // Simulate connection error
    mockEventSource.onerror()

    await expect(_promise).rejects.toThrow('EventSource connection error')
    expect(onError).toHaveBeenCalledWith('Connection error')
    expect(mockEventSource.close).toHaveBeenCalled()
  })

  it('handles malformed JSON gracefully', async () => {
    const consoleError = jest.spyOn(console, 'error').mockImplementation()
    const _promise = handleAIStream(mockEventSource, {})

    // Simulate malformed JSON
    mockEventSource.onmessage({
      data: 'invalid json',
    })

    expect(consoleError).toHaveBeenCalledWith('Failed to parse SSE event:', expect.any(SyntaxError))

    consoleError.mockRestore()
  })
})
