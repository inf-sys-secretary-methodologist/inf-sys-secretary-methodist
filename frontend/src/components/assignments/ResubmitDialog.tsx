'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, RotateCcw } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { resubmitSubmission } from '@/hooks/useMyAssignments'

interface ResubmitDialogProps {
  assignmentId: number
  open: boolean
  onClose: () => void
  onResubmitted?: () => void
}

// ResubmitDialog — confirmation modal for the student-side
// status='returned' → 'pending' transition. Mirrors ReturnDialog's
// shape but carries no input (no reason / no textarea) — the backend
// resubmit endpoint accepts an empty body and identifies the row by
// (path id, JWT subject). On success fires onResubmitted (caller
// refreshes SWR) and onClose. On failure stays open so the student
// can retry without re-opening the dialog.
export function ResubmitDialog({
  assignmentId,
  open,
  onClose,
  onResubmitted,
}: ResubmitDialogProps) {
  const t = useTranslations()
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) {
      onClose()
    }
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await resubmitSubmission(assignmentId)
      toast.success(t('myAssignments.resubmitDialog.successToast'))
      onResubmitted?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'myAssignments.resubmitDialog.errors.notReturned'
          break
        case 403:
          key = 'myAssignments.resubmitDialog.errors.forbidden'
          break
        default:
          key = 'myAssignments.resubmitDialog.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('myAssignments.resubmitDialog.title')}</DialogTitle>
          <DialogDescription>{t('myAssignments.resubmitDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('myAssignments.resubmitDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <RotateCcw className="h-4 w-4 mr-2" />
            )}
            {submitting
              ? t('myAssignments.resubmitDialog.submitting')
              : t('myAssignments.resubmitDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
