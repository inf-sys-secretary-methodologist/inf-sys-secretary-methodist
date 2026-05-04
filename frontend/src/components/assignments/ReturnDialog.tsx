'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, RotateCcw } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { returnSubmission } from '@/hooks/useAssignments'
import type { SubmissionView } from '@/types/assignments'

const MAX_REASON = 4096

interface ReturnDialogProps {
  assignmentId: number
  submission: SubmissionView
  open: boolean
  onClose: () => void
  onReturned?: () => void
}

// ReturnDialog — modal for returning a submission for revision. The
// reason field is required (trim-non-empty, ≤ 4096 chars). On success
// fires onReturned (caller refreshes the SWR list) and onClose. On
// failure stays open so the teacher can adjust and retry.
export function ReturnDialog({
  assignmentId,
  submission,
  open,
  onClose,
  onReturned,
}: ReturnDialogProps) {
  const t = useTranslations()
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const trimmed = reason.trim()
  const tooLong = reason.length > MAX_REASON
  const canConfirm = trimmed.length > 0 && !tooLong && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) {
      onClose()
    }
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await returnSubmission(assignmentId, {
        student_id: submission.student_id,
        reason: trimmed,
      })
      toast.success(
        t('assignments.returnDialog.successToast', { name: submission.student_name })
      )
      onReturned?.()
      setReason('')
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'assignments.returnDialog.errors.alreadyReturned'
          break
        case 422:
          key = 'assignments.returnDialog.errors.invalidReason'
          break
        case 403:
          key = 'assignments.returnDialog.errors.forbidden'
          break
        default:
          key = 'assignments.returnDialog.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('assignments.returnDialog.title')}</DialogTitle>
          <DialogDescription>
            {t('assignments.returnDialog.description', { name: submission.student_name })}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor={`return-reason-${submission.id}`}>
            {t('assignments.returnDialog.reasonLabel')}
          </Label>
          <Textarea
            id={`return-reason-${submission.id}`}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t('assignments.returnDialog.reasonPlaceholder')}
            rows={5}
            disabled={submitting}
          />
          <p className={`text-xs ${tooLong ? 'text-destructive' : 'text-muted-foreground'}`}>
            {reason.length} / {MAX_REASON}
          </p>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('assignments.returnDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <RotateCcw className="h-4 w-4 mr-2" />
            )}
            {submitting
              ? t('assignments.returnDialog.saving')
              : t('assignments.returnDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
