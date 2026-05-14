'use client'

import useSWR from 'swr'
import { SWR_DEDUPING } from '@/config/swr'
import type { TaskReminder, CreateTaskReminderInput } from '@/types/taskReminders'

const REMINDERS_NOT_IMPLEMENTED = 'useTaskReminders: deferred-runtime stub — Pair 1 GREEN required'

// Stub fetcher — throws at runtime so RED tests fail with a clear
// sentinel. Replaced by the real fetcher в Pair 1 GREEN.
const fetcher = async (_url: string): Promise<TaskReminder[]> => {
  throw new Error(REMINDERS_NOT_IMPLEMENTED)
}

// useTaskReminders subscribes к /api/tasks/:id/reminders. RED stub —
// short-circuits to null SWR key so renderHook does not call fetcher
// и the test asserts on the not-yet-implemented payload shape.
export function useTaskReminders(_taskID: number | null) {
  const { data, error, isLoading, mutate } = useSWR<TaskReminder[]>(null, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return {
    reminders: (data || []) as TaskReminder[],
    isLoading,
    error,
    mutate,
  }
}

// createTaskReminder RED stub. GREEN импл POST'ит /api/tasks/:id/reminders.
export async function createTaskReminder(
  _taskID: number,
  _input: CreateTaskReminderInput
): Promise<TaskReminder> {
  throw new Error(REMINDERS_NOT_IMPLEMENTED)
}

// deleteTaskReminder RED stub. GREEN импл DELETE'ит
// /api/tasks/:id/reminders/:reminderID.
export async function deleteTaskReminder(_taskID: number, _reminderID: number): Promise<void> {
  throw new Error(REMINDERS_NOT_IMPLEMENTED)
}
