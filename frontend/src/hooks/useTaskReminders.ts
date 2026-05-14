'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { TaskReminder, CreateTaskReminderInput } from '@/types/taskReminders'

// Backend reminder endpoints return plain JSON (no `{success, data}`
// envelope) — fetcher passes the body through directly.
const fetcher = (url: string) => apiClient.get<TaskReminder[]>(url)

// useTaskReminders subscribes к GET /api/tasks/:id/reminders. Passing
// null short-circuits the SWR key (mirror к useTask(null) precedent),
// so callers can opt out before a task ID is known.
export function useTaskReminders(taskID: number | null) {
  const key = taskID == null ? null : `/api/tasks/${taskID}/reminders`
  const { data, error, isLoading, mutate } = useSWR<TaskReminder[]>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return {
    reminders: data || [],
    isLoading,
    error,
    mutate,
  }
}

// createTaskReminder POSTs the body к /api/tasks/:id/reminders.
// Returns the created reminder DTO. Backend validates reminder_type +
// minutes_before; 422 propagates as an axios error.
export async function createTaskReminder(
  taskID: number,
  input: CreateTaskReminderInput
): Promise<TaskReminder> {
  return apiClient.post<TaskReminder>(`/api/tasks/${taskID}/reminders`, input)
}

// deleteTaskReminder removes the reminder by id under the supplied
// task. Three backend failure modes propagate as axios errors:
// 404 not-found, 404 wrong-task path mismatch, 403 wrong owner.
export async function deleteTaskReminder(taskID: number, reminderID: number): Promise<void> {
  await apiClient.delete(`/api/tasks/${taskID}/reminders/${reminderID}`)
}
