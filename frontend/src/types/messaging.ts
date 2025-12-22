// Conversation types
export type ConversationType = 'direct' | 'group'
export type ParticipantRole = 'member' | 'admin'
export type MessageType = 'text' | 'image' | 'file' | 'system'

// Participant
export interface Participant {
  id: number
  user_id: number
  name: string
  avatar_url?: string
  role: ParticipantRole
  is_muted: boolean
  joined_at: string
  left_at?: string
}

// Attachment
export interface Attachment {
  id: number
  file_name: string
  file_size: number
  mime_type: string
  url: string
}

export interface AttachmentInput {
  file_id: number
  file_name: string
  file_size: number
  mime_type: string
  url: string
}

// Message
export interface Message {
  id: number
  conversation_id: number
  sender_id: number
  sender_name: string
  sender_avatar?: string
  type: MessageType
  content: string
  reply_to?: Message
  attachments: Attachment[]
  is_edited: boolean
  edited_at?: string
  is_deleted: boolean
  created_at: string
}

// Conversation
export interface Conversation {
  id: number
  type: ConversationType
  title?: string
  description?: string
  avatar_url?: string
  created_by: number
  last_message?: Message
  unread_count: number
  participants: Participant[]
  created_at: string
  updated_at: string
}

// Input types
export interface CreateDirectConversationInput {
  recipient_id: number
}

export interface CreateGroupConversationInput {
  title: string
  description?: string
  participant_ids: number[]
}

export interface UpdateConversationInput {
  title?: string
  description?: string
  avatar_url?: string
}

export interface AddParticipantsInput {
  user_ids: number[]
}

export interface SendMessageInput {
  content: string
  reply_to_id?: number
  type?: MessageType
  attachments?: AttachmentInput[]
}

export interface EditMessageInput {
  content: string
}

export interface MarkReadInput {
  message_id: number
}

// Filter types
export interface ConversationFilterInput {
  type?: ConversationType
  search?: string
  limit?: number
  offset?: number
}

export interface MessageFilterInput {
  before_id?: number
  after_id?: number
  search?: string
  limit?: number
}

export interface SearchMessagesInput {
  q: string
  limit?: number
  offset?: number
}

// Output types
export interface ConversationListOutput {
  conversations: Conversation[]
  total: number
  limit: number
  offset: number
}

export interface MessageListOutput {
  messages: Message[]
  has_more: boolean
}

export interface SearchMessagesOutput {
  messages: Message[]
  total: number
  limit: number
  offset: number
}

// WebSocket event types
export type WebSocketEventType =
  | 'new_message'
  | 'message_updated'
  | 'message_deleted'
  | 'typing'
  | 'stop_typing'
  | 'read'
  | 'user_online'
  | 'user_offline'
  | 'conversation_updated'

export interface WebSocketEvent {
  type: WebSocketEventType
  conversation_id?: number
  user_id?: number
  payload?: unknown
}

export interface WebSocketCommand {
  type: 'subscribe' | 'unsubscribe' | 'typing' | 'stop_typing' | 'ping'
  conversation_id?: number
}
