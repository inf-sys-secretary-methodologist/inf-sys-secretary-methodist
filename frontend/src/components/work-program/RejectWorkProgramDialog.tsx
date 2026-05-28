'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, XCircle } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { rejectWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface RejectWorkProgramDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onRejected?: () => void
}

// RejectWorkProgramDialog — approver-side modal for the pending_approval
// → draft transition with a mandatory reason so the author knows what to
// revise. Mirrors RejectCurriculumDialog (textarea + reset-on-open) but
// imposes NO client-side max length: the РПД backend reject only enforces
// `binding:"required"` + domain non-empty-after-trim, so a length cap
// here would invent an invariant the domain does not have. The reason is
// trimmed before sending; errors route through pickWorkProgramErrorKey.
export function RejectWorkProgramDialog({
  workProgramId,
  open,
  onClose,
  onRejected,
}: RejectWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset on each fresh open so the approver starts blank for every
  // programme in the queue — including the error→cancel→reopen path.
  useEffect(() => {
    if (open) setReason('')
  }, [open])

  const trimmed = reason.trim()
  const canConfirm = trimmed.length > 0 && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await rejectWorkProgram(workProgramId, { reason: trimmed })
      toast.success(t('rejectDialog.successToast'))
      onRejected?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('rejectDialog.title')}</DialogTitle>
          <DialogDescription>{t('rejectDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor={`wp-reject-reason-${workProgramId}`}>
            {t('rejectDialog.reasonLabel')}
          </Label>
          <Textarea
            id={`wp-reject-reason-${workProgramId}`}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t('rejectDialog.reasonPlaceholder')}
            rows={5}
            disabled={submitting}
          />
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('rejectDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm} variant="destructive">
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <XCircle className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('rejectDialog.submitting') : t('rejectDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
