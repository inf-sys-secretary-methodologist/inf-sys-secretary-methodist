import { apiClient } from '../api'
import { getStoredToken } from '@/lib/auth/token'
import type {
  AIConversation,
  AIConversationListOutput,
  AIConversationFilterInput,
  CreateAIConversationInput,
  UpdateAIConversationInput,
  AIMessageListOutput,
  AIMessageFilterInput,
  SendAIMessageInput,
  AIChatResponse,
  AISearchInput,
  AISearchOutput,
  IndexDocumentInput,
  IndexDocumentOutput,
  DocumentSource,
} from '@/types/ai'

// API response wrapper type
interface ApiResponse<T> {
  data: T
}

export const aiApi = {
  // Conversation operations
  listConversations: async (
    input?: AIConversationFilterInput
  ): Promise<AIConversationListOutput> => {
    const params = new URLSearchParams()
    if (input?.search) params.append('search', input.search)
    if (input?.limit) params.append('limit', String(input.limit))
    if (input?.offset) params.append('offset', String(input.offset))

    const query = params.toString()
    const response = await apiClient.get<ApiResponse<AIConversationListOutput>>(
      `/api/ai/conversations${query ? `?${query}` : ''}`
    )
    return response.data
  },

  getConversation: async (id: number): Promise<AIConversation> => {
    const response = await apiClient.get<ApiResponse<AIConversation>>(`/api/ai/conversations/${id}`)
    return response.data
  },

  createConversation: async (input?: CreateAIConversationInput): Promise<AIConversation> => {
    const response = await apiClient.post<ApiResponse<AIConversation>>(
      '/api/ai/conversations',
      input || {}
    )
    return response.data
  },

  updateConversation: async (
    id: number,
    input: UpdateAIConversationInput
  ): Promise<AIConversation> => {
    const response = await apiClient.patch<ApiResponse<AIConversation>>(
      `/api/ai/conversations/${id}`,
      input
    )
    return response.data
  },

  deleteConversation: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.delete<ApiResponse<{ message: string }>>(
      `/api/ai/conversations/${id}`
    )
    return response.data
  },

  // Message operations
  getMessages: async (
    conversationId: number,
    input?: AIMessageFilterInput
  ): Promise<AIMessageListOutput> => {
    const params = new URLSearchParams()
    if (input?.before_id) params.append('before_id', String(input.before_id))
    if (input?.after_id) params.append('after_id', String(input.after_id))
    if (input?.limit) params.append('limit', String(input.limit))

    const query = params.toString()
    const response = await apiClient.get<ApiResponse<AIMessageListOutput>>(
      `/api/ai/conversations/${conversationId}/messages${query ? `?${query}` : ''}`
    )
    return response.data
  },

  // Chat - send message and get AI response
  chat: async (input: SendAIMessageInput): Promise<AIChatResponse> => {
    const response = await apiClient.post<ApiResponse<AIChatResponse>>('/api/ai/chat', input)
    return response.data
  },

  // Streaming chat - returns EventSource URL
  chatStream: (input: SendAIMessageInput): EventSource => {
    const token = getStoredToken()
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

    // Build URL with query params for GET request (SSE doesn't support POST body)
    const params = new URLSearchParams()
    params.append('content', input.content)
    if (input.conversation_id) params.append('conversation_id', String(input.conversation_id))
    if (input.include_sources !== undefined)
      params.append('include_sources', String(input.include_sources))
    if (input.max_sources) params.append('max_sources', String(input.max_sources))
    if (token) params.append('token', token)

    const url = `${baseUrl}/api/ai/chat/stream?${params.toString()}`
    return new EventSource(url)
  },

  // Semantic search
  search: async (input: AISearchInput): Promise<AISearchOutput> => {
    const response = await apiClient.post<ApiResponse<AISearchOutput>>('/api/ai/search', input)
    return response.data
  },

  // Document indexing
  indexDocument: async (documentId: number, force?: boolean): Promise<IndexDocumentOutput> => {
    const input: IndexDocumentInput = {
      document_id: documentId,
      force_reindex: force,
    }
    const response = await apiClient.post<ApiResponse<IndexDocumentOutput>>(
      `/api/ai/index/${documentId}`,
      input
    )
    return response.data
  },

  // Batch index documents
  indexDocuments: async (
    documentIds: number[],
    force?: boolean
  ): Promise<{ results: IndexDocumentOutput[] }> => {
    const response = await apiClient.post<ApiResponse<{ results: IndexDocumentOutput[] }>>(
      '/api/ai/index/batch',
      {
        document_ids: documentIds,
        force_reindex: force,
      }
    )
    return response.data
  },

  // Get indexing status
  getIndexingStatus: async (): Promise<{
    total_documents: number
    indexed_documents: number
    pending_documents: number
    last_indexed_at?: string
  }> => {
    const response = await apiClient.get<
      ApiResponse<{
        total_documents: number
        indexed_documents: number
        pending_documents: number
        last_indexed_at?: string
      }>
    >('/api/ai/index/status')
    return response.data
  },
}

// Helper to handle streaming responses
export async function handleAIStream(
  eventSource: EventSource,
  callbacks: {
    onContent?: (content: string) => void
    onSource?: (source: DocumentSource) => void
    onDone?: (messageId: number) => void
    onError?: (error: string) => void
  }
): Promise<void> {
  return new Promise((resolve, reject) => {
    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)

        switch (data.type) {
          case 'content':
            callbacks.onContent?.(data.content)
            break
          case 'source':
            callbacks.onSource?.(data.source)
            break
          case 'done':
            callbacks.onDone?.(data.message_id)
            eventSource.close()
            resolve()
            break
          case 'error':
            callbacks.onError?.(data.error)
            eventSource.close()
            reject(new Error(data.error))
            break
        }
      } catch (error) {
        console.error('Failed to parse SSE event:', error)
      }
    }

    eventSource.onerror = () => {
      callbacks.onError?.('Connection error')
      eventSource.close()
      reject(new Error('EventSource connection error'))
    }
  })
}
