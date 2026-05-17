'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, UserCheck } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { assignExecutorDocument } from '@/hooks/useDocumentWorkflow'
import { usersApi, User } from '@/lib/api/users'

interface AssignExecutorDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onAssigned?: () => void
}

// AssignExecutorDialog — admin-only modal для shaping executor assignment
// on execution-status document. Status stays execution — shape-only per
// ADR-1. Executor user-picker fetched via usersApi.getAll(); date input
// optional (no hard deadline когда empty).
//
// Issue: #232
export function AssignExecutorDialog({
  documentId,
  open,
  onClose,
  onAssigned,
}: AssignExecutorDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [executorID, setExecutorID] = useState<string>('')
  const [dueDate, setDueDate] = useState<string>('')
  const [users, setUsers] = useState<User[]>([])
  const [loadingUsers, setLoadingUsers] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return
    setExecutorID('')
    setDueDate('')
    let cancelled = false
    setLoadingUsers(true)
    usersApi
      .getAll()
      .then((list) => {
        if (!cancelled) setUsers(list)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('assignExecutorToast.errors.loadUsers'))
      })
      .finally(() => {
        if (!cancelled) setLoadingUsers(false)
      })
    return () => {
      cancelled = true
    }
  }, [open, t])

  const parsedExecutor = parseInt(executorID, 10)
  const canConfirm = parsedExecutor > 0 && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await assignExecutorDocument(documentId, {
        executor_id: parsedExecutor,
        due_date: dueDate || undefined,
      })
      toast.success(t('assignExecutorToast.success'))
      onAssigned?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'assignExecutorToast.errors.invalid'
          break
        case 409:
          key = 'assignExecutorToast.errors.notExecution'
          break
        case 403:
          key = 'assignExecutorToast.errors.forbidden'
          break
        case 404:
          key = 'assignExecutorToast.errors.notFound'
          break
        default:
          key = 'assignExecutorToast.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('assignExecutor.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('assignExecutor.dialogBody')}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="assign-executor-user">{t('assignExecutor.executorLabel')}</Label>
            <Select
              value={executorID}
              onValueChange={setExecutorID}
              disabled={submitting || loadingUsers}
            >
              <SelectTrigger id="assign-executor-user">
                <SelectValue placeholder={t('assignExecutor.executorPlaceholder')} />
              </SelectTrigger>
              <SelectContent>
                {users.map((u) => (
                  <SelectItem key={u.id} value={String(u.id)}>
                    {u.name} ({u.email})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="assign-executor-due">{t('assignExecutor.dueDateLabel')}</Label>
            <Input
              id="assign-executor-due"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              disabled={submitting}
            />
            <p className="text-muted-foreground text-xs">{t('assignExecutor.dueDateHint')}</p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('assignExecutor.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <UserCheck className="mr-2 h-4 w-4" />
            )}
            {t('assignExecutor.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
