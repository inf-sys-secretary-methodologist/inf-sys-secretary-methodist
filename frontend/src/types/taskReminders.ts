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
