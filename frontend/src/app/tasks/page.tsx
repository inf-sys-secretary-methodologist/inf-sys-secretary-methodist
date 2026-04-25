'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Plus, ListTodo, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { TaskCard } from '@/components/tasks/TaskCard'
import { TaskFilters } from '@/components/tasks/TaskFilters'
import { TaskForm } from '@/components/tasks/TaskForm'
import {
  useTasks,
  createTask,
  updateTask,
  deleteTask,
} from '@/hooks/useTasks'
import type { CreateTaskInput, Task, TaskFilterParams } from '@/types/tasks'
import { useAuthCheck } from '@/hooks/useAuth'

export default function TasksPage() {
  const t = useTranslations('tasks')
  useAuthCheck()

  const [filters, setFilters] = useState<TaskFilterParams>({})
  const [editingTask, setEditingTask] = useState<Task | undefined>(undefined)
  const [isFormOpen, setIsFormOpen] = useState(false)

  const { tasks, total, isLoading, error, mutate } = useTasks({ ...filters, limit: 100 })

  const openCreate = () => {
    setEditingTask(undefined)
    setIsFormOpen(true)
  }

  const openEdit = (task: Task) => {
    setEditingTask(task)
    setIsFormOpen(true)
  }

  const handleSubmit = async (input: CreateTaskInput) => {
    try {
      if (editingTask) {
        await updateTask(editingTask.id, input)
      } else {
        await createTask(input)
      }
      setIsFormOpen(false)
      setEditingTask(undefined)
      await mutate()
    } catch {
      toast.error(t(editingTask ? 'errors.updateFailed' : 'errors.createFailed'))
    }
  }

  const handleDelete = async (task: Task) => {
    try {
      await deleteTask(task.id)
      await mutate()
    } catch {
      toast.error(t('errors.deleteFailed'))
    }
  }

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-muted-foreground">{t('description')}</p>
          </div>
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" />
            {t('create')}
          </Button>
        </div>

        <TaskFilters value={filters} onChange={setFilters} />

        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('loadFailed')}</p>
          </div>
        ) : tasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <ListTodo className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('noTasks')}</h3>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onClick={() => openEdit(task)}
                onEdit={() => openEdit(task)}
                onDelete={() => handleDelete(task)}
              />
            ))}
          </div>
        )}

        {tasks.length > 0 && (
          <p className="text-sm text-muted-foreground text-right">
            {tasks.length} / {total}
          </p>
        )}

        <Dialog open={isFormOpen} onOpenChange={setIsFormOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {editingTask ? t('form.editTitle') : t('form.createTitle')}
              </DialogTitle>
            </DialogHeader>
            <TaskForm
              task={editingTask}
              onSubmit={handleSubmit}
              onCancel={() => setIsFormOpen(false)}
            />
          </DialogContent>
        </Dialog>
      </div>
    </AppLayout>
  )
}
