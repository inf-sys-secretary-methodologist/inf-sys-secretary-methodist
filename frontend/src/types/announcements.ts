// Announcement module types matching backend DTO at
// internal/modules/announcements/application/dto/announcement_dto.go
// and domain/types.go.

export type AnnouncementStatus = 'draft' | 'published' | 'archived'

export type AnnouncementPriority = 'low' | 'normal' | 'high' | 'urgent'

export type TargetAudience = 'all' | 'students' | 'teachers' | 'staff' | 'admins'

export const ANNOUNCEMENT_STATUSES: AnnouncementStatus[] = ['draft', 'published', 'archived']

export const ANNOUNCEMENT_PRIORITIES: AnnouncementPriority[] = [
  'low',
  'normal',
  'high',
  'urgent',
]

export const TARGET_AUDIENCES: TargetAudience[] = [
  'all',
  'students',
  'teachers',
  'staff',
  'admins',
]

export interface AnnouncementAuthor {
  id: number
  name: string
  email: string
}

export interface AnnouncementAttachment {
  id: number
  file_name: string
  file_size: number
  mime_type: string
  created_at: string
}

export interface Announcement {
  id: number
  title: string
  content: string
  summary?: string
  author_id: number
  author?: AnnouncementAuthor
  status: AnnouncementStatus
  priority: AnnouncementPriority
  target_audience: TargetAudience
  publish_at?: string
  expire_at?: string
  is_pinned: boolean
  view_count: number
  tags?: string[]
  attachments?: AnnouncementAttachment[]
  created_at: string
  updated_at: string
}

export interface AnnouncementListResponse {
  announcements: Announcement[]
  total: number
  limit: number
  offset: number
}

export interface AnnouncementFilterParams {
  author_id?: number
  status?: AnnouncementStatus
  priority?: AnnouncementPriority
  target_audience?: TargetAudience
  is_pinned?: boolean
  search?: string
  tags?: string[]
  limit?: number
  offset?: number
}

export interface CreateAnnouncementInput {
  title: string
  content: string
  summary?: string
  priority: AnnouncementPriority
  target_audience: TargetAudience
  publish_at?: string
  expire_at?: string
  is_pinned?: boolean
  tags?: string[]
}

export interface UpdateAnnouncementInput {
  title?: string
  content?: string
  summary?: string
  priority?: AnnouncementPriority
  target_audience?: TargetAudience
  publish_at?: string
  expire_at?: string
  is_pinned?: boolean
  tags?: string[]
}
