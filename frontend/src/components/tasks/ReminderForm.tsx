'use client'

import * as React from 'react'
import type { CreateTaskReminderInput } from '@/types/taskReminders'

export interface ReminderFormProps {
  onSubmit: (input: CreateTaskReminderInput) => Promise<void> | void
  onCancel: () => void
  submitting?: boolean
  className?: string
}

// Pair 2 RED — deferred-runtime stub. GREEN replaces with form
// rendering reminder_type select (4 options) + minutes_before input
// + save/cancel buttons.
export function ReminderForm(_props: ReminderFormProps): React.ReactElement {
  throw new Error('ReminderForm: deferred-runtime stub — Pair 2 GREEN required')
}
