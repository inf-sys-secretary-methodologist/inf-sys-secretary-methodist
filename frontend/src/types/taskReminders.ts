// Reminder channel — mirror к backend domain.ReminderType set
// (`internal/modules/tasks/domain/entities/task_reminder.go`).
export type ReminderType = 'email' | 'push' | 'in_app' | 'telegram'

export const REMINDER_TYPES: ReminderType[] = ['email', 'push', 'in_app', 'telegram']

// TaskReminder mirrors the backend dto.TaskReminderResponse shape
// returned by GET /api/tasks/:id/reminders + POST response.
export interface TaskReminder {
  id: number
  task_id: number
  user_id: number
  reminder_type: ReminderType
  minutes_before: number
  is_sent: boolean
  sent_at?: string | null
  created_at: string
}

// CreateTaskReminderInput is the POST body. ActorUserID is NOT here —
// backend derives it from the JWT subject.
export interface CreateTaskReminderInput {
  reminder_type: ReminderType
  minutes_before: number
}

// reminderTypeI18nKey maps the on-the-wire reminder_type к the
// camelCase i18n key fragment. Backend uses snake_case `in_app`; the
// i18n namespace stores `inApp` because nested JSON keys must remain
// valid identifiers in next-intl call signatures.
export function reminderTypeI18nKey(value: ReminderType | string): string {
  return value === 'in_app' ? 'inApp' : value
}
