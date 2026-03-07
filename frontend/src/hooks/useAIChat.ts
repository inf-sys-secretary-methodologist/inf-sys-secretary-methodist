'use client'

import useSWR, { mutate } from 'swr'
import { useState, useCallback, useRef, useEffect } from 'react'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import { aiApi, handleAIStream } from '@/lib/api/ai'
import type {
  AIConversation,
  AIConversationListOutput,
  AIConversationFilterInput,
  CreateAIConversationInput,
  UpdateAIConversationInput,
  AIMessage,
  AIMessageListOutput,
  AIMessageFilterInput,
  SendAIMessageInput,
  AISearchInput,
  AISearchOutput,
  DocumentSource,
  AIMessageStatus,
} from '@/types/ai'

const AI_CONVERSATIONS_URL = '/api/ai/conversations'

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

  return response as T
}

// Build URL with query params
function buildConversationsUrl(input?: AIConversationFilterInput): string {
  const params = new URLSearchParams()
  if (input?.search) params.append('search', input.search)
  if (input?.limit) params.append('limit', String(input.limit))
  if (input?.offset) params.append('offset', String(input.offset))

  const query = params.toString()
  return `${AI_CONVERSATIONS_URL}${query ? `?${query}` : ''}`
}

function buildMessagesUrl(conversationId: number, input?: AIMessageFilterInput): string {
  const params = new URLSearchParams()
  if (input?.before_id) params.append('before_id', String(input.before_id))
  if (input?.after_id) params.append('after_id', String(input.after_id))
  if (input?.limit) params.append('limit', String(input.limit))

  const query = params.toString()
  return `${AI_CONVERSATIONS_URL}/${conversationId}/messages${query ? `?${query}` : ''}`
}

// AI Conversation list hook
export function useAIConversations(input?: AIConversationFilterInput) {
  const url = buildConversationsUrl(input)

  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<AIConversationListOutput>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.MEDIUM,
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

// Single AI conversation hook
export function useAIConversation(id: number | null) {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<AIConversation>(id ? `${AI_CONVERSATIONS_URL}/${id}` : null, fetcher, {
    revalidateOnFocus: false,
  })

  return { conversation: data, isLoading, error, mutate: revalidate }
}

// AI Messages hook with pagination
export function useAIMessages(conversationId: number | null, input?: AIMessageFilterInput) {
  const url = conversationId ? buildMessagesUrl(conversationId, input) : null

  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<AIMessageListOutput>(url, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
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

// Create AI conversation hook
export function useCreateAIConversation() {
  const [isPending, setIsPending] = useState(false)

  const createConversation = useCallback(async (input?: CreateAIConversationInput) => {
    setIsPending(true)
    try {
      const conversation = await aiApi.createConversation(input)
      mutate((key) => typeof key === 'string' && key.includes('/ai/conversations'), undefined, {
        revalidate: true,
      })
      return conversation
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: createConversation, isPending }
}

// Update AI conversation hook
export function useUpdateAIConversation() {
  const [isPending, setIsPending] = useState(false)

  const updateConversation = useCallback(async (id: number, input: UpdateAIConversationInput) => {
    setIsPending(true)
    try {
      const conversation = await aiApi.updateConversation(id, input)
      mutate((key) => typeof key === 'string' && key.includes('/ai/conversations'), undefined, {
        revalidate: true,
      })
      return conversation
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: updateConversation, isPending }
}

// Delete AI conversation hook
export function useDeleteAIConversation() {
  const [isPending, setIsPending] = useState(false)

  const deleteConversation = useCallback(async (id: number) => {
    setIsPending(true)
    try {
      await aiApi.deleteConversation(id)
      mutate((key) => typeof key === 'string' && key.includes('/ai/conversations'), undefined, {
        revalidate: true,
      })
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: deleteConversation, isPending }
}

// Send AI message hook (non-streaming)
export function useSendAIMessage() {
  const [isPending, setIsPending] = useState(false)

  const sendMessage = useCallback(async (input: SendAIMessageInput) => {
    setIsPending(true)
    try {
      const response = await aiApi.chat(input)
      // Revalidate messages and conversations
      if (input.conversation_id) {
        mutate(
          (key) =>
            typeof key === 'string' &&
            key.includes(`/ai/conversations/${input.conversation_id}/messages`),
          undefined,
          { revalidate: true }
        )
      }
      mutate((key) => typeof key === 'string' && key.includes('/ai/conversations'), undefined, {
        revalidate: true,
      })
      return response
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: sendMessage, isPending }
}

// Streaming AI message hook
export function useStreamAIMessage() {
  const [isStreaming, setIsStreaming] = useState(false)
  const [streamingContent, setStreamingContent] = useState('')
  const [streamingSources, setStreamingSources] = useState<DocumentSource[]>([])
  const eventSourceRef = useRef<EventSource | null>(null)
  const contentBufferRef = useRef('')
  const rafRef = useRef<number | null>(null)

  const startStream = useCallback(async (input: SendAIMessageInput) => {
    setIsStreaming(true)
    setStreamingContent('')
    setStreamingSources([])
    contentBufferRef.current = ''

    const eventSource = aiApi.chatStream(input)
    eventSourceRef.current = eventSource

    try {
      await handleAIStream(eventSource, {
        onContent: (content) => {
          contentBufferRef.current += content
          // Batch state updates to animation frames (~60fps max)
          if (rafRef.current === null) {
            rafRef.current = requestAnimationFrame(() => {
              rafRef.current = null
              setStreamingContent(contentBufferRef.current)
            })
          }
        },
        onSource: (source) => {
          setStreamingSources((prev) => [...prev, source])
        },
        onDone: () => {
          // Flush remaining content
          if (rafRef.current !== null) {
            cancelAnimationFrame(rafRef.current)
            rafRef.current = null
          }
          setStreamingContent(contentBufferRef.current)
        },
        onError: (error) => {
          console.error('Stream error:', error)
          if (rafRef.current !== null) {
            cancelAnimationFrame(rafRef.current)
            rafRef.current = null
          }
          setIsStreaming(false)
        },
      })

      // Stop cursor but keep content visible (no flash).
      // Content will be cleaned up by the caller after SWR revalidation.
      setIsStreaming(false)
    } catch {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current)
        rafRef.current = null
      }
      setIsStreaming(false)
    }
  }, [])

  const stopStream = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    // Flush remaining buffer before stopping
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
    if (contentBufferRef.current) {
      setStreamingContent(contentBufferRef.current)
    }
    setIsStreaming(false)
  }, [])

  const resetStream = useCallback(() => {
    setStreamingContent('')
    setStreamingSources([])
    contentBufferRef.current = ''
    if (rafRef.current !== null) {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = null
    }
  }, [])

  // Cleanup on unmount: close EventSource and cancel pending RAF
  useEffect(() => {
    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current)
        rafRef.current = null
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
        eventSourceRef.current = null
      }
    }
  }, [])

  return {
    startStream,
    stopStream,
    resetStream,
    isStreaming,
    streamingContent,
    streamingSources,
  }
}

// AI Search hook
export function useAISearch() {
  const [isPending, setIsPending] = useState(false)
  const [results, setResults] = useState<AISearchOutput | null>(null)

  const search = useCallback(async (input: AISearchInput) => {
    setIsPending(true)
    try {
      const response = await aiApi.search(input)
      setResults(response)
      return response
    } finally {
      setIsPending(false)
    }
  }, [])

  const clearResults = useCallback(() => {
    setResults(null)
  }, [])

  return { search, clearResults, results, isPending }
}

// Combined hook for AI chat with streaming support
export function useAIChat(conversationId: number | null) {
  const {
    conversation,
    isLoading: isLoadingConversation,
    mutate: mutateConversation,
  } = useAIConversation(conversationId)
  const {
    messages,
    hasMore,
    isLoading: isLoadingMessages,
    mutate: mutateMessages,
  } = useAIMessages(conversationId)
  const { mutateAsync: createConversation } = useCreateAIConversation()
  const stream = useStreamAIMessage()

  // Local messages state for optimistic updates during streaming
  const [localMessages, setLocalMessages] = useState<AIMessage[]>([])
  const [pendingMessage, setPendingMessage] = useState<AIMessage | null>(null)
  const [streamingStartTime, setStreamingStartTime] = useState('')

  // Combine server messages with local/streaming state
  const allMessages = [...messages, ...localMessages]
  if (pendingMessage) {
    allMessages.push(pendingMessage)
  }
  // Show streaming/completed assistant message while content exists.
  // After streaming ends, content stays visible (status: 'complete', no cursor)
  // until SWR revalidation brings the real message and caller clears the stream.
  if (stream.streamingContent) {
    allMessages.push({
      id: -1, // Temporary ID
      conversation_id: conversationId || 0,
      role: 'assistant',
      content: stream.streamingContent,
      sources: stream.streamingSources,
      status: (stream.isStreaming ? 'streaming' : 'complete') as AIMessageStatus,
      created_at: streamingStartTime,
    })
  }

  const sendMessage = useCallback(
    async (content: string, useStreaming = true) => {
      let activeConversationId = conversationId

      // Create conversation if needed
      if (!activeConversationId) {
        const newConversation = await createConversation({
          title: content.slice(0, 50) + (content.length > 50 ? '...' : ''),
        })
        activeConversationId = newConversation.id
      }

      // Add user message optimistically
      const userMessage: AIMessage = {
        id: Date.now(), // Temporary ID
        conversation_id: activeConversationId,
        role: 'user',
        content,
        status: 'complete',
        created_at: new Date().toISOString(),
      }
      setLocalMessages((prev) => [...prev, userMessage])

      // Add pending assistant message
      setPendingMessage({
        id: Date.now() + 1,
        conversation_id: activeConversationId,
        role: 'assistant',
        content: '',
        status: 'pending',
        created_at: new Date().toISOString(),
      })

      if (useStreaming) {
        // Remove pending message when streaming starts
        stream.resetStream()
        setPendingMessage(null)
        setStreamingStartTime(new Date().toISOString())

        await stream.startStream({
          content,
          conversation_id: activeConversationId,
          include_sources: true,
        })

        // Streaming done — text stays visible without cursor.
        // Now revalidate SWR and wait for real message from server.
        await mutateMessages()

        // Server data is in SWR cache. Clean up local state in one batch —
        // streaming message removed at same time as real message appears. No flash.
        stream.resetStream()
        setLocalMessages([])
      } else {
        // Non-streaming: use regular API
        try {
          await aiApi.chat({
            content,
            conversation_id: activeConversationId,
            include_sources: true,
          })
          setLocalMessages([])
          setPendingMessage(null)
          await mutateMessages()
          await mutateConversation()
        } catch (error) {
          setPendingMessage(null)
          throw error
        }
      }

      return activeConversationId
    },
    [conversationId, createConversation, stream, mutateMessages, mutateConversation]
  )

  const stopGeneration = useCallback(() => {
    stream.stopStream()
    setPendingMessage(null)
  }, [stream])

  return {
    conversation,
    messages: allMessages,
    hasMore,
    isLoading: isLoadingConversation || isLoadingMessages,
    isStreaming: stream.isStreaming,
    isPending: !!pendingMessage,
    sendMessage,
    stopGeneration,
    mutateConversation,
    mutateMessages,
  }
}
