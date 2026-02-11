// AI Message types
export type AIMessageRole = 'user' | 'assistant' | 'system'
export type AIMessageStatus = 'pending' | 'streaming' | 'complete' | 'error'

// Document source for RAG citations
export interface DocumentSource {
  id: number
  document_id: number
  document_title: string
  chunk_text: string
  similarity_score: number
  page_number?: number
  metadata?: Record<string, unknown>
}

// AI Message
export interface AIMessage {
  id: number
  conversation_id: number
  role: AIMessageRole
  content: string
  sources?: DocumentSource[]
  tokens_used?: number
  model?: string
  status: AIMessageStatus
  error_message?: string
  created_at: string
}

// AI Conversation
export interface AIConversation {
  id: number
  user_id: number
  title: string
  last_message?: string
  message_count: number
  model: string
  created_at: string
  updated_at: string
}

// Search result
export interface AISearchResult {
  document_id: number
  document_title: string
  chunk_text: string
  similarity_score: number
  highlights?: string[]
  metadata?: Record<string, unknown>
}

// Quick action for suggestions
export interface AIQuickAction {
  id: string
  label: string
  prompt: string
  icon?: string
  category?: string
}

// Input types
export interface SendAIMessageInput {
  content: string
  conversation_id?: number
  include_sources?: boolean
  max_sources?: number
}

export interface AISearchInput {
  query: string
  limit?: number
  threshold?: number
  document_types?: string[]
}

export interface CreateAIConversationInput {
  title?: string
  model?: string
}

export interface UpdateAIConversationInput {
  title: string
}

// Output types
export interface AIConversationListOutput {
  conversations: AIConversation[]
  total: number
  limit: number
  offset: number
}

export interface AIMessageListOutput {
  messages: AIMessage[]
  has_more: boolean
}

export interface AISearchOutput {
  results: AISearchResult[]
  query: string
  total: number
}

export interface AIChatResponse {
  message: AIMessage
  conversation_id: number
  sources?: DocumentSource[]
}

// Streaming event types
export interface AIStreamEvent {
  type: 'content' | 'source' | 'done' | 'error'
  content?: string
  source?: DocumentSource
  error?: string
  message_id?: number
}

// Filter types
export interface AIConversationFilterInput {
  search?: string
  limit?: number
  offset?: number
}

export interface AIMessageFilterInput {
  before_id?: number
  after_id?: number
  limit?: number
}

// Index document request
export interface IndexDocumentInput {
  document_id: number
  force_reindex?: boolean
}

export interface IndexDocumentOutput {
  document_id: number
  chunks_created: number
  status: 'indexed' | 'already_indexed' | 'error'
  message?: string
}

// AI Settings
export interface AISettings {
  model: string
  temperature: number
  max_tokens: number
  include_sources: boolean
  max_sources: number
}

export const DEFAULT_AI_SETTINGS: AISettings = {
  model: 'gpt-4o-mini',
  temperature: 0.7,
  max_tokens: 2048,
  include_sources: true,
  max_sources: 5,
}
