'use client'

import useSWR, { mutate } from 'swr'
import { useState, useCallback, useEffect, useRef } from 'react'
import { apiClient } from '@/lib/api'
import { MessagingWebSocket } from '@/lib/api/messaging'
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
  SearchMessagesOutput,
  WebSocketEvent,
} from '@/types/messaging'

const CONVERSATIONS_BASE_URL = '/api/conversations'

// API Response wrapper type from backend
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}

// Fetcher for SWR - extracts data from wrapped response
const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T> | T>(url)

  // Check if response is the API wrapper format
  if (response && typeof response === 'object' && 'success' in response) {
    const wrappedResponse = response as ApiResponse<T>
    if (wrappedResponse.success && wrappedResponse.data !== undefined) {
      return wrappedResponse.data
    } else {
      throw new Error(wrappedResponse.error?.message || 'API returned error')
    }
  }

  // Response is already the data
  return response as T
}

// Build URL with query params
function buildConversationsUrl(input?: ConversationFilterInput): string {
  const params = new URLSearchParams()
  if (input?.type) params.append('type', input.type)
  if (input?.search) params.append('search', input.search)
  if (input?.limit) params.append('limit', String(input.limit))
  if (input?.offset) params.append('offset', String(input.offset))

  const query = params.toString()
  return `${CONVERSATIONS_BASE_URL}${query ? `?${query}` : ''}`
}

function buildMessagesUrl(conversationId: number, input?: MessageFilterInput): string {
  const params = new URLSearchParams()
  if (input?.before_id) params.append('before_id', String(input.before_id))
  if (input?.after_id) params.append('after_id', String(input.after_id))
  if (input?.search) params.append('search', input.search)
  if (input?.limit) params.append('limit', String(input.limit))

  const query = params.toString()
  return `${CONVERSATIONS_BASE_URL}/${conversationId}/messages${query ? `?${query}` : ''}`
}

// Conversation list hook
export function useConversations(input?: ConversationFilterInput) {
  const url = buildConversationsUrl(input)

  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<ConversationListOutput>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 10000,
    refreshInterval: 30000,
  })

  return {
    data,
    conversations: data?.conversations || [],
    total: data?.total || 0,
    isLoading,
    error,
    mutate: revalidate,
  }
}

// Single conversation hook
export function useConversation(id: number | null) {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<Conversation>(id ? `${CONVERSATIONS_BASE_URL}/${id}` : null, fetcher, {
    revalidateOnFocus: false,
  })

  return { conversation: data, isLoading, error, mutate: revalidate }
}

// Messages hook with pagination
export function useMessages(conversationId: number | null, input?: MessageFilterInput) {
  const url = conversationId ? buildMessagesUrl(conversationId, input) : null

  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<MessageListOutput>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 5000,
  })

  return {
    data,
    messages: data?.messages || [],
    hasMore: data?.has_more || false,
    isLoading,
    error,
    mutate: revalidate,
  }
}

// Create direct conversation hook
export function useCreateDirectConversation() {
  const [isPending, setIsPending] = useState(false)

  const createConversation = useCallback(async (input: CreateDirectConversationInput) => {
    setIsPending(true)
    try {
      const response = await apiClient.post<ApiResponse<Conversation>>(
        `${CONVERSATIONS_BASE_URL}/direct`,
        input
      )
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
      return response.data
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: createConversation, isPending }
}

// Create group conversation hook
export function useCreateGroupConversation() {
  const [isPending, setIsPending] = useState(false)

  const createConversation = useCallback(async (input: CreateGroupConversationInput) => {
    setIsPending(true)
    try {
      const response = await apiClient.post<ApiResponse<Conversation>>(
        `${CONVERSATIONS_BASE_URL}/group`,
        input
      )
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
      return response.data
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: createConversation, isPending }
}

// Update conversation hook
export function useUpdateConversation() {
  const [isPending, setIsPending] = useState(false)

  const updateConversation = useCallback(async (id: number, input: UpdateConversationInput) => {
    setIsPending(true)
    try {
      const response = await apiClient.patch<ApiResponse<Conversation>>(
        `${CONVERSATIONS_BASE_URL}/${id}`,
        input
      )
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
      return response.data
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: updateConversation, isPending }
}

// Add participants hook
export function useAddParticipants() {
  const [isPending, setIsPending] = useState(false)

  const addParticipants = useCallback(async (id: number, input: AddParticipantsInput) => {
    setIsPending(true)
    try {
      await apiClient.post(`${CONVERSATIONS_BASE_URL}/${id}/participants`, input)
      mutate(`${CONVERSATIONS_BASE_URL}/${id}`)
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: addParticipants, isPending }
}

// Leave conversation hook
export function useLeaveConversation() {
  const [isPending, setIsPending] = useState(false)

  const leaveConversation = useCallback(async (id: number) => {
    setIsPending(true)
    try {
      await apiClient.post(`${CONVERSATIONS_BASE_URL}/${id}/leave`)
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: leaveConversation, isPending }
}

// Send message hook
export function useSendMessage() {
  const [isPending, setIsPending] = useState(false)

  const sendMessage = useCallback(async (conversationId: number, input: SendMessageInput) => {
    setIsPending(true)
    try {
      const response = await apiClient.post<ApiResponse<Message>>(
        `${CONVERSATIONS_BASE_URL}/${conversationId}/messages`,
        input
      )
      mutate(
        (key) =>
          typeof key === 'string' && key.includes(`/conversations/${conversationId}/messages`),
        undefined,
        { revalidate: true }
      )
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
      return response.data
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: sendMessage, isPending }
}

// Edit message hook
export function useEditMessage() {
  const [isPending, setIsPending] = useState(false)

  const editMessage = useCallback(
    async (conversationId: number, messageId: number, input: EditMessageInput) => {
      setIsPending(true)
      try {
        const response = await apiClient.patch<ApiResponse<Message>>(
          `${CONVERSATIONS_BASE_URL}/${conversationId}/messages/${messageId}`,
          input
        )
        mutate(
          (key) =>
            typeof key === 'string' && key.includes(`/conversations/${conversationId}/messages`),
          undefined,
          { revalidate: true }
        )
        return response.data
      } finally {
        setIsPending(false)
      }
    },
    []
  )

  return { mutateAsync: editMessage, isPending }
}

// Delete message hook
export function useDeleteMessage() {
  const [isPending, setIsPending] = useState(false)

  const deleteMessage = useCallback(async (conversationId: number, messageId: number) => {
    setIsPending(true)
    try {
      await apiClient.delete(`${CONVERSATIONS_BASE_URL}/${conversationId}/messages/${messageId}`)
      mutate(
        (key) =>
          typeof key === 'string' && key.includes(`/conversations/${conversationId}/messages`),
        undefined,
        { revalidate: true }
      )
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: deleteMessage, isPending }
}

// Mark as read hook
export function useMarkConversationAsRead() {
  const [isPending, setIsPending] = useState(false)
  const lastMarkedRef = useRef<string | null>(null)

  const markAsRead = useCallback(async (conversationId: number, messageId: number) => {
    // Prevent duplicate calls for the same message
    const key = `${conversationId}:${messageId}`
    if (lastMarkedRef.current === key) {
      return
    }
    lastMarkedRef.current = key

    setIsPending(true)
    try {
      await apiClient.post(`${CONVERSATIONS_BASE_URL}/${conversationId}/read`, {
        message_id: messageId,
      })
      // Only revalidate the conversations list, not messages
      mutate(`${CONVERSATIONS_BASE_URL}?limit=50&offset=0`)
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: markAsRead, isPending }
}

// Search messages hook
export function useSearchMessages(conversationId: number | null, query: string) {
  const params = new URLSearchParams()
  params.append('q', query)

  const url =
    conversationId && query.length >= 2
      ? `${CONVERSATIONS_BASE_URL}/${conversationId}/messages/search?${params.toString()}`
      : null

  const { data, error, isLoading } = useSWR<SearchMessagesOutput>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 1000,
  })

  return {
    messages: data?.messages || [],
    total: data?.total || 0,
    isLoading,
    error,
  }
}

// Global singleton for WebSocket connection
let globalWsInstance: MessagingWebSocket | null = null
let globalConnecting = false

// WebSocket hook for real-time messaging
export function useMessagingWebSocket() {
  const [isConnected, setIsConnected] = useState(false)
  const [typingUsers, setTypingUsers] = useState<Map<number, Set<number>>>(new Map())

  const getToken = useCallback(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('authToken')
    }
    return null
  }, [])

  const connect = useCallback(async () => {
    // Already connected
    if (globalWsInstance?.isConnected) {
      setIsConnected(true)
      return
    }
    // Already connecting
    if (globalConnecting) return

    globalConnecting = true
    globalWsInstance = new MessagingWebSocket(getToken)

    // Set up event listeners
    globalWsInstance.on('new_message', (data) => {
      const event = data as WebSocketEvent
      // Revalidate messages for the conversation
      if (event.conversation_id) {
        mutate(
          (key) =>
            typeof key === 'string' &&
            key.includes(`/conversations/${event.conversation_id}/messages`),
          undefined,
          { revalidate: true }
        )
        mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
          revalidate: true,
        })
      }
    })

    globalWsInstance.on('message_updated', (data) => {
      const event = data as WebSocketEvent
      if (event.conversation_id) {
        mutate(
          (key) =>
            typeof key === 'string' &&
            key.includes(`/conversations/${event.conversation_id}/messages`),
          undefined,
          { revalidate: true }
        )
      }
    })

    globalWsInstance.on('message_deleted', (data) => {
      const event = data as WebSocketEvent
      if (event.conversation_id) {
        mutate(
          (key) =>
            typeof key === 'string' &&
            key.includes(`/conversations/${event.conversation_id}/messages`),
          undefined,
          { revalidate: true }
        )
      }
    })

    globalWsInstance.on('typing', (data) => {
      const event = data as WebSocketEvent
      if (event.conversation_id && event.user_id) {
        setTypingUsers((prev) => {
          const newMap = new Map(prev)
          const users = newMap.get(event.conversation_id!) || new Set()
          users.add(event.user_id!)
          newMap.set(event.conversation_id!, users)
          return newMap
        })
      }
    })

    globalWsInstance.on('stop_typing', (data) => {
      const event = data as WebSocketEvent
      if (event.conversation_id && event.user_id) {
        setTypingUsers((prev) => {
          const newMap = new Map(prev)
          const users = newMap.get(event.conversation_id!)
          if (users) {
            users.delete(event.user_id!)
            if (users.size === 0) {
              newMap.delete(event.conversation_id!)
            } else {
              newMap.set(event.conversation_id!, users)
            }
          }
          return newMap
        })
      }
    })

    globalWsInstance.on('read', () => {
      mutate((key) => typeof key === 'string' && key.includes('/conversations'), undefined, {
        revalidate: true,
      })
    })

    try {
      await globalWsInstance.connect()
      setIsConnected(true)
    } catch (error) {
      console.error('Failed to connect WebSocket:', error)
      setIsConnected(false)
    } finally {
      globalConnecting = false
    }
  }, [getToken])

  const disconnect = useCallback(() => {
    globalWsInstance?.disconnect()
    globalWsInstance = null
    setIsConnected(false)
  }, [])

  const subscribe = useCallback((conversationId: number) => {
    globalWsInstance?.subscribe(conversationId)
  }, [])

  const unsubscribe = useCallback((conversationId: number) => {
    globalWsInstance?.unsubscribe(conversationId)
  }, [])

  const sendTyping = useCallback((conversationId: number) => {
    globalWsInstance?.sendTyping(conversationId)
  }, [])

  const sendStopTyping = useCallback((conversationId: number) => {
    globalWsInstance?.sendStopTyping(conversationId)
  }, [])

  const getTypingUsers = useCallback(
    (conversationId: number): number[] => {
      return Array.from(typingUsers.get(conversationId) || [])
    },
    [typingUsers]
  )

  // NOTE: We intentionally don't disconnect on unmount because
  // the WebSocket should stay alive for the duration of the user session.
  // The connection will be closed when the user logs out or closes the browser.

  return {
    isConnected,
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    sendTyping,
    sendStopTyping,
    getTypingUsers,
  }
}

// Combined hook for a single conversation with real-time updates
export function useConversationWithMessages(conversationId: number | null) {
  const {
    conversation,
    isLoading: isLoadingConversation,
    mutate: mutateConversation,
  } = useConversation(conversationId)
  const {
    messages,
    hasMore,
    isLoading: isLoadingMessages,
    mutate: mutateMessages,
  } = useMessages(conversationId)
  const { mutateAsync: sendMessage, isPending: isSending } = useSendMessage()
  const { mutateAsync: markAsRead } = useMarkConversationAsRead()
  const ws = useMessagingWebSocket()

  // Connect to WebSocket and subscribe to conversation
  useEffect(() => {
    if (conversationId) {
      ws.connect().then(() => {
        ws.subscribe(conversationId)
      })
    }

    return () => {
      if (conversationId) {
        ws.unsubscribe(conversationId)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [conversationId])

  // Mark as read when viewing conversation
  useEffect(() => {
    if (conversationId && messages.length > 0) {
      const lastMessage = messages[0] // Assuming messages are sorted newest first
      if (lastMessage) {
        markAsRead(conversationId, lastMessage.id).catch(console.error)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [conversationId, messages.length])

  const handleSendMessage = useCallback(
    async (content: string, attachments?: SendMessageInput['attachments']) => {
      if (!conversationId) return
      return sendMessage(conversationId, { content, attachments })
    },
    [conversationId, sendMessage]
  )

  return {
    conversation,
    messages,
    hasMore,
    isLoading: isLoadingConversation || isLoadingMessages,
    isSending,
    sendMessage: handleSendMessage,
    mutateConversation,
    mutateMessages,
    typingUsers: conversationId ? ws.getTypingUsers(conversationId) : [],
    sendTyping: () => conversationId && ws.sendTyping(conversationId),
    sendStopTyping: () => conversationId && ws.sendStopTyping(conversationId),
    isConnected: ws.isConnected,
  }
}
