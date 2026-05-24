// Extracurricular event types mirroring backend DTO at
// internal/modules/extracurricular/interfaces/http/handlers/event_handler.go
// (EventDTO / EventSummaryDTO / CreateEventRequest / UpdateEventRequest).
// Sentinel error codes mirror handler mapEventError.

export type EventStatus = 'draft' | 'published' | 'canceled' | 'completed'

export type EventCategory = 'academic' | 'cultural' | 'sports' | 'volunteer' | 'professional'

export type EventTargetAudience = 'all' | 'students' | 'teachers' | 'staff'

export const EVENT_STATUSES: EventStatus[] = ['draft', 'published', 'canceled', 'completed']

export const EVENT_CATEGORIES: EventCategory[] = [
  'academic',
  'cultural',
  'sports',
  'volunteer',
  'professional',
]

export const EVENT_TARGET_AUDIENCES: EventTargetAudience[] = [
  'all',
  'students',
  'teachers',
  'staff',
]

export interface ExtracurricularParticipant {
  user_id: number
  registered_at: string
}

export interface ExtracurricularEvent {
  id: number
  title: string
  description: string
  category: EventCategory
  target_audience: EventTargetAudience
  status: EventStatus
  location: string
  start_at: string
  end_at: string
  max_capacity?: number | null
  organizer_id: number
  participants?: ExtracurricularParticipant[]
  participant_count: number
  version: number
  created_at: string
  updated_at: string
}

export interface ExtracurricularEventSummary {
  id: number
  title: string
  category: EventCategory
  target_audience: EventTargetAudience
  status: EventStatus
  location: string
  start_at: string
  end_at: string
  max_capacity?: number | null
  organizer_id: number
  participant_count: number
  version: number
  created_at: string
  updated_at: string
}

export interface ExtracurricularEventListResponse {
  items: ExtracurricularEventSummary[]
  total: number
}

export interface ExtracurricularEventFilterParams {
  status?: EventStatus
  category?: EventCategory
  organizer_id?: number
  from?: string
  to?: string
  limit?: number
  offset?: number
}

export interface CreateExtracurricularEventInput {
  title: string
  description?: string
  category: EventCategory
  target_audience: EventTargetAudience
  location?: string
  start_at: string
  end_at: string
  max_capacity?: number | null
}

export interface UpdateExtracurricularEventInput {
  title: string
  description?: string
  category: EventCategory
  target_audience: EventTargetAudience
  location?: string
  start_at: string
  end_at: string
  max_capacity?: number | null
}

// Sentinel error codes mirrored from backend handler mapEventError.
// Frontend maps these via pickErrorKey to i18n strings.
export type ExtracurricularErrorCode =
  | 'VERSION_CONFLICT'
  | 'INVALID_EVENT'
  | 'ALREADY_REGISTERED'
  | 'EVENT_FULL'
  | 'REGISTRATION_CLOSED'
  | 'CANNOT_EDIT'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'GENERIC'

export const EXTRACURRICULAR_ERROR_CODES: ExtracurricularErrorCode[] = [
  'VERSION_CONFLICT',
  'INVALID_EVENT',
  'ALREADY_REGISTERED',
  'EVENT_FULL',
  'REGISTRATION_CLOSED',
  'CANNOT_EDIT',
  'FORBIDDEN',
  'NOT_FOUND',
  'GENERIC',
]
