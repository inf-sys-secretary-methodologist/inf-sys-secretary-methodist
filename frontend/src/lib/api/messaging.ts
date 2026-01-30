import { apiClient } from '../api'
import type {
  Conversation,
  ConversationListOutput,
  ConversationFilterInput,
  CreateDirectConversationInput,
  CreateGroupConversationInput,
  UpdateConversationInput,
  AddParticipantsInput,
  Message,
  MessageListOutput,
  MessageFilterInput,
  SendMessageInput,
  EditMessageInput,
  MarkReadInput,
  SearchMessagesInput,
  SearchMessagesOutput,
} from '@/types/messaging'

// API response wrapper type
interface ApiResponse<T> {
  data: T
}

export const messagingApi = {
  // Conversation operations
  createDirectConversation: async (input: CreateDirectConversationInput): Promise<Conversation> => {
    const response = await apiClient.post<ApiResponse<Conversation>>(
      '/api/conversations/direct',
      input
    )
    return response.data
  },

  createGroupConversation: async (input: CreateGroupConversationInput): Promise<Conversation> => {
    const response = await apiClient.post<ApiResponse<Conversation>>(
      '/api/conversations/group',
      input
    )
    return response.data
  },

  listConversations: async (input?: ConversationFilterInput): Promise<ConversationListOutput> => {
    const params = new URLSearchParams()
    /* c8 ignore start - Optional filter params */
    if (input?.type) params.append('type', input.type)
    if (input?.search) params.append('search', input.search)
    if (input?.limit) params.append('limit', String(input.limit))
    if (input?.offset) params.append('offset', String(input.offset))
    /* c8 ignore stop */

    const query = params.toString()
    const response = await apiClient.get<ApiResponse<ConversationListOutput>>(
      `/api/conversations${query ? `?${query}` : ''}`
    )
    return response.data
  },

  getConversation: async (id: number): Promise<Conversation> => {
    const response = await apiClient.get<ApiResponse<Conversation>>(`/api/conversations/${id}`)
    return response.data
  },

  updateConversation: async (id: number, input: UpdateConversationInput): Promise<Conversation> => {
    const response = await apiClient.patch<ApiResponse<Conversation>>(
      `/api/conversations/${id}`,
      input
    )
    return response.data
  },

  addParticipants: async (
    id: number,
    input: AddParticipantsInput
  ): Promise<{ message: string }> => {
    const response = await apiClient.post<ApiResponse<{ message: string }>>(
      `/api/conversations/${id}/participants`,
      input
    )
    return response.data
  },

  leaveConversation: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.post<ApiResponse<{ message: string }>>(
      `/api/conversations/${id}/leave`
    )
    return response.data
  },

  // Message operations
  sendMessage: async (conversationId: number, input: SendMessageInput): Promise<Message> => {
    const response = await apiClient.post<ApiResponse<Message>>(
      `/api/conversations/${conversationId}/messages`,
      input
    )
    return response.data
  },

  getMessages: async (
    conversationId: number,
    input?: MessageFilterInput
  ): Promise<MessageListOutput> => {
    const params = new URLSearchParams()
    /* c8 ignore start - Optional filter params */
    if (input?.before_id) params.append('before_id', String(input.before_id))
    if (input?.after_id) params.append('after_id', String(input.after_id))
    if (input?.search) params.append('search', input.search)
    if (input?.limit) params.append('limit', String(input.limit))
    /* c8 ignore stop */

    const query = params.toString()
    const response = await apiClient.get<ApiResponse<MessageListOutput>>(
      `/api/conversations/${conversationId}/messages${query ? `?${query}` : ''}`
    )
    return response.data
  },

  editMessage: async (
    conversationId: number,
    messageId: number,
    input: EditMessageInput
  ): Promise<Message> => {
    const response = await apiClient.patch<ApiResponse<Message>>(
      `/api/conversations/${conversationId}/messages/${messageId}`,
      input
    )
    return response.data
  },

  deleteMessage: async (
    conversationId: number,
    messageId: number
  ): Promise<{ message: string }> => {
    const response = await apiClient.delete<ApiResponse<{ message: string }>>(
      `/api/conversations/${conversationId}/messages/${messageId}`
    )
    return response.data
  },

  markAsRead: async (
    conversationId: number,
    input: MarkReadInput
  ): Promise<{ message: string }> => {
    const response = await apiClient.post<ApiResponse<{ message: string }>>(
      `/api/conversations/${conversationId}/read`,
      input
    )
    return response.data
  },

  searchMessages: async (
    conversationId: number,
    input: SearchMessagesInput
  ): Promise<SearchMessagesOutput> => {
    const params = new URLSearchParams()
    params.append('q', input.q)
    /* c8 ignore start - Optional params */
    if (input.limit) params.append('limit', String(input.limit))
    if (input.offset) params.append('offset', String(input.offset))
    /* c8 ignore stop */

    const response = await apiClient.get<ApiResponse<SearchMessagesOutput>>(
      `/api/conversations/${conversationId}/messages/search?${params.toString()}`
    )
    return response.data
  },
}

// WebSocket service for real-time messaging
export class MessagingWebSocket {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private listeners: Map<string, Set<(event: unknown) => void>> = new Map()
  private pingInterval: NodeJS.Timeout | null = null

  constructor(private getToken: () => string | null) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      const token = this.getToken()
      if (!token) {
        reject(new Error('No auth token available'))
        return
      }

      // Already connected
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const wsUrl = baseUrl.replace(/^http/, 'ws') + `/api/ws?token=${encodeURIComponent(token)}`

      console.log('🔌 Connecting to WebSocket:', wsUrl.replace(/token=[^&]+/, 'token=***'))
      this.ws = new WebSocket(wsUrl)
      let connected = false

      this.ws.onopen = () => {
        console.log('✅ WebSocket connected')
        connected = true
        this.reconnectAttempts = 0
        this.startPing()
        resolve()
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.emit(data.type, data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.ws.onerror = () => {
        // WebSocket error events don't contain useful info due to browser security
        console.warn('⚠️ WebSocket connection error')
      }

      this.ws.onclose = (event) => {
        console.log('🔌 WebSocket closed:', event.code, event.reason || 'No reason')
        this.stopPing()
        if (!connected) {
          // Failed to connect
          reject(new Error(`WebSocket connection failed: ${event.code}`))
        } else {
          // Was connected, now disconnected - try to reconnect
          this.attemptReconnect()
        }
      }
    })
  }

  private startPing() {
    this.pingInterval = setInterval(() => {
      this.send({ type: 'ping' })
    }, 30000)
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
      console.warn(`WebSocket reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
      setTimeout(() => this.connect(), delay)
    }
  }

  disconnect() {
    this.stopPing()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message: object) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  subscribe(conversationId: number) {
    this.send({ type: 'subscribe', conversation_id: conversationId })
  }

  unsubscribe(conversationId: number) {
    this.send({ type: 'unsubscribe', conversation_id: conversationId })
  }

  sendTyping(conversationId: number) {
    this.send({ type: 'typing', conversation_id: conversationId })
  }

  sendStopTyping(conversationId: number) {
    this.send({ type: 'stop_typing', conversation_id: conversationId })
  }

  on(event: string, callback: (data: unknown) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
    return () => this.off(event, callback)
  }

  off(event: string, callback: (data: unknown) => void) {
    this.listeners.get(event)?.delete(callback)
  }

  private emit(event: string, data: unknown) {
    this.listeners.get(event)?.forEach((callback) => callback(data))
    this.listeners.get('*')?.forEach((callback) => callback(data))
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}
