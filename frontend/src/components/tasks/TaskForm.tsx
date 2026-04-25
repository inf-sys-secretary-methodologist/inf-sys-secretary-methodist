'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  TASK_PRIORITIES,
  type Task,
  type CreateTaskInput,
  type TaskPriority,
} from '@/types/tasks'

interface TaskFormProps {
  task?: Task
  onSubmit: (input: CreateTaskInput) => Promise<void> | void
  onCancel: () => void
  className?: string
}

export function TaskForm({ task, onSubmit, onCancel, className }: TaskFormProps) {
  const t = useTranslations('tasks')

  const [title, setTitle] = useState(task?.title ?? '')
  const [description, setDescription] = useState(task?.description ?? '')
  const [priority, setPriority] = useState<TaskPriority>(task?.priority ?? 'normal')
  const [dueDate, setDueDate] = useState(task?.due_date ? task.due_date.slice(0, 10) : '')
  const [titleError, setTitleError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: { preventDefault: () => void }) => {
    e.preventDefault()
    if (!title.trim()) {
      setTitleError(t('errors.titleRequired'))
      return
    }
    setTitleError(null)
    setSubmitting(true)
    try {
      const input: CreateTaskInput = {
        title: title.trim(),
        priority,
      }
      if (description.trim()) input.description = description.trim()
      if (dueDate) input.due_date = new Date(dueDate).toISOString()
      await onSubmit(input)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className={cn('flex flex-col gap-4', className)}>
      <div>
        <label
          htmlFor="task-form-title"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.titleLabel')}
        </label>
        <Input
          id="task-form-title"
          value={title}
          onChange={(e) => {
            setTitle(e.target.value)
            if (titleError) setTitleError(null)
          }}
          placeholder={t('form.titlePlaceholder')}
          aria-invalid={!!titleError}
          aria-describedby={titleError ? 'task-form-title-error' : undefined}
        />
        {titleError && (
          <p id="task-form-title-error" className="mt-1 text-sm text-red-600 dark:text-red-400">
            {titleError}
          </p>
        )}
      </div>

      <div>
        <label
          htmlFor="task-form-description"
          className="block text-sm font-medium text-foreground mb-1"
        >
          {t('form.descriptionLabel')}
        </label>
        <textarea
          id="task-form-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder={t('form.descriptionPlaceholder')}
          rows={4}
          className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
        />
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label
            htmlFor="task-form-priority"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.priorityLabel')}
          </label>
          <select
            id="task-form-priority"
            value={priority}
            onChange={(e) => setPriority(e.target.value as TaskPriority)}
            className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {TASK_PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {t(`priority.${p}`)}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="task-form-due-date"
            className="block text-sm font-medium text-foreground mb-1"
          >
            {t('form.dueDateLabel')}
          </label>
          <Input
            id="task-form-due-date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
          />
        </div>
      </div>

      <div className="flex gap-2 justify-end">
        <Button type="button" variant="outline" onClick={onCancel} disabled={submitting}>
          {t('form.cancel')}
        </Button>
        <Button type="submit" disabled={submitting}>
          {t('form.save')}
        </Button>
      </div>
    </form>
  )
}
