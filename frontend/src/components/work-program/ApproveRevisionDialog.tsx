'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { CheckCircle2, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { approveRevision, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface ApproveRevisionDialogProps {
  workProgramId: number
  revisionId: number
  open: boolean
  onClose: () => void
  onApproved?: () => void
}

// ApproveRevisionDialog — approver-side confirmation modal for the
// pending_approval → approved transition of a лист актуализации (revision).
// Mirrors ApproveWorkProgramDialog: empty body — path ids (РПД + revision)
// + JWT subject identify the row + approver. Errors route through
// pickWorkProgramErrorKey; the dialog stays open on failure for retry.
export function ApproveRevisionDialog({
  workProgramId,
  revisionId,
  open,
  onClose,
  onApproved,
}: ApproveRevisionDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await approveRevision(workProgramId, revisionId)
      toast.success(t('approveRevisionDialog.successToast'))
      onApproved?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('approveRevisionDialog.title')}</DialogTitle>
          <DialogDescription>{t('approveRevisionDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('approveRevisionDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <CheckCircle2 className="h-4 w-4 mr-2" />
            )}
            {submitting
              ? t('approveRevisionDialog.submitting')
              : t('approveRevisionDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
