'use client'

import * as React from 'react'
import { useState } from 'react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  REMINDER_TYPES,
  reminderTypeI18nKey,
  type CreateTaskReminderInput,
  type ReminderType,
} from '@/types/taskReminders'

export interface ReminderFormProps {
  onSubmit: (input: CreateTaskReminderInput) => Promise<void> | void
  onCancel: () => void
  submitting?: boolean
  className?: string
}

const DEFAULT_TYPE: ReminderType = 'telegram'
const DEFAULT_MINUTES = 60
const MIN_MINUTES = 1
const MAX_MINUTES = 10080

export function ReminderForm({ onSubmit, onCancel, submitting, className }: ReminderFormProps) {
  const t = useTranslations('taskReminders')

  const [reminderType, setReminderType] = useState<ReminderType>(DEFAULT_TYPE)
  const [minutes, setMinutes] = useState<string>(String(DEFAULT_MINUTES))

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const parsed = Number.parseInt(minutes, 10)
    if (!Number.isFinite(parsed) || parsed < MIN_MINUTES) {
      return
    }
    await onSubmit({ reminder_type: reminderType, minutes_before: parsed })
  }

  return (
    <form onSubmit={handleSubmit} className={cn('flex flex-col gap-4', className)}>
      <div>
        <label
          htmlFor="reminder-form-type"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('reminderTypeLabel')}
        </label>
        <select
          id="reminder-form-type"
          value={reminderType}
          onChange={(e) => setReminderType(e.target.value as ReminderType)}
          className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        >
          {REMINDER_TYPES.map((value) => (
            <option key={value} value={value}>
              {t(`type.${reminderTypeI18nKey(value)}`)}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label
          htmlFor="reminder-form-minutes"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('minutesBeforeLabel')}
        </label>
        <Input
          id="reminder-form-minutes"
          type="number"
          min={MIN_MINUTES}
          max={MAX_MINUTES}
          value={minutes}
          onChange={(e) => setMinutes(e.target.value)}
        />
      </div>

      <div className="flex gap-2 justify-end">
        <Button type="button" variant="outline" onClick={onCancel} disabled={submitting}>
          {t('cancel')}
        </Button>
        <Button type="submit" disabled={submitting}>
          {t('save')}
        </Button>
      </div>
    </form>
  )
}
