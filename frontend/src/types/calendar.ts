// Calendar event types matching backend API

export type EventType = 'meeting' | 'deadline' | 'task' | 'reminder' | 'holiday' | 'personal'

export type EventStatus = 'scheduled' | 'ongoing' | 'completed' | 'cancelled' | 'postponed'

export type RecurrenceFrequency = 'daily' | 'weekly' | 'monthly' | 'yearly'

export type Weekday = 'MO' | 'TU' | 'WE' | 'TH' | 'FR' | 'SA' | 'SU'

export type ReminderType = 'email' | 'push' | 'in_app' | 'telegram'

export type ParticipantRole = 'required' | 'optional' | 'resource'

export type ResponseStatus = 'pending' | 'accepted' | 'declined' | 'tentative'

export interface RecurrenceRule {
  frequency: RecurrenceFrequency
  interval: number
  count?: number
  until?: string
  by_weekday?: Weekday[]
  by_monthday?: number[]
  by_month?: number[]
  week_start: Weekday
}

export interface Participant {
  user_id: number
  user_name?: string
  email?: string
  response_status: ResponseStatus
  role: ParticipantRole
  responded_at?: string
}

export interface Reminder {
  id: number
  reminder_type: ReminderType
  minutes_before: number
  is_sent: boolean
  sent_at?: string
}

export interface CalendarEvent {
  id: number
  title: string
  description?: string
  event_type: EventType
  status: EventStatus

  // Time
  start_time: string
  end_time?: string
  all_day: boolean
  timezone: string

  // Location
  location?: string

  // Organizer
  organizer_id: number
  organizer_name?: string

  // Participants
  participants?: Participant[]

  // Recurrence
  is_recurring: boolean
  recurrence_rule?: RecurrenceRule
  parent_event_id?: number

  // Display
  color?: string
  priority: number

  // Reminders
  reminders?: Reminder[]

  // Audit
  created_at: string
  updated_at: string
}

export interface CreateEventInput {
  title: string
  description?: string
  event_type: EventType
  start_time: string
  end_time?: string
  all_day: boolean
  timezone?: string
  location?: string
  participant_ids?: number[]
  color?: string
  priority?: number
  is_recurring: boolean
  recurrence_rule?: Omit<RecurrenceRule, 'week_start'> & { week_start?: Weekday }
  reminders?: Array<{
    reminder_type: ReminderType
    minutes_before: number
  }>
}

export interface UpdateEventInput {
  title?: string
  description?: string
  event_type?: EventType
  status?: EventStatus
  start_time?: string
  end_time?: string
  all_day?: boolean
  timezone?: string
  location?: string
  color?: string
  priority?: number
  is_recurring?: boolean
  recurrence_rule?: Omit<RecurrenceRule, 'week_start'> & { week_start?: Weekday }
}

export interface EventListResponse {
  events: CalendarEvent[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface EventFilterParams {
  organizer_id?: number
  participant_id?: number
  event_type?: EventType
  status?: EventStatus
  start_from?: string
  start_to?: string
  search?: string
  is_recurring?: boolean
  page?: number
  page_size?: number
  order_by?: string
}

export type CalendarView = 'month' | 'week' | 'day'

export interface CalendarDay {
  date: Date
  events: CalendarEvent[]
  isCurrentMonth: boolean
  isToday: boolean
  isSelected: boolean
}

// Color mapping for event types
export const EVENT_TYPE_COLORS: Record<EventType, string> = {
  meeting: 'bg-blue-500',
  deadline: 'bg-red-500',
  task: 'bg-green-500',
  reminder: 'bg-yellow-500',
  holiday: 'bg-purple-500',
  personal: 'bg-gray-500',
}

export const EVENT_TYPE_LABELS: Record<EventType, string> = {
  meeting: 'Встреча',
  deadline: 'Дедлайн',
  task: 'Задача',
  reminder: 'Напоминание',
  holiday: 'Праздник',
  personal: 'Личное',
}

export const EVENT_STATUS_LABELS: Record<EventStatus, string> = {
  scheduled: 'Запланировано',
  ongoing: 'В процессе',
  completed: 'Завершено',
  cancelled: 'Отменено',
  postponed: 'Отложено',
}
