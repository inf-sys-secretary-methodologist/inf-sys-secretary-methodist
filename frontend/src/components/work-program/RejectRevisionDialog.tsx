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
import { rejectRevision, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface RejectRevisionDialogProps {
  workProgramId: number
  revisionId: number
  open: boolean
  onClose: () => void
  onRejected?: () => void
}

// RejectRevisionDialog — approver-side modal for the pending_approval →
// draft transition of a лист актуализации (revision) with a mandatory
// reason so the author knows what to revise. Mirrors RejectWorkProgramDialog
// (textarea + reset-on-open, no client-side max length — the backend reject
// only enforces binding:"required" + domain non-empty-after-trim). The
// reason is trimmed before sending; errors route through pickWorkProgramErrorKey.
export function RejectRevisionDialog({
  workProgramId,
  revisionId,
  open,
  onClose,
  onRejected,
}: RejectRevisionDialogProps) {
  const t = useTranslations('workProgram')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset on each fresh open so the approver starts blank for every
  // revision — including the error → cancel → reopen path.
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
      await rejectRevision(workProgramId, revisionId, { reason: trimmed })
      toast.success(t('rejectRevisionDialog.successToast'))
      onRejected?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  const reasonId = `wp-reject-revision-reason-${workProgramId}-${revisionId}`

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('rejectRevisionDialog.title')}</DialogTitle>
          <DialogDescription>{t('rejectRevisionDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor={reasonId}>{t('rejectRevisionDialog.reasonLabel')}</Label>
          <Textarea
            id={reasonId}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t('rejectRevisionDialog.reasonPlaceholder')}
            rows={5}
            disabled={submitting}
          />
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('rejectRevisionDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm} variant="destructive">
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <XCircle className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('rejectRevisionDialog.submitting') : t('rejectRevisionDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
