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
import { approveWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface ApproveWorkProgramDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onApproved?: () => void
}

// ApproveWorkProgramDialog — approver-side confirmation modal for the
// pending_approval → approved transition (РПД approver = methodist /
// admin per ADR-5). Mirrors SubmitWorkProgramDialog: no input — the
// backend approve endpoint takes an empty body and derives the approver
// from the JWT subject (non-spoofable). Errors route through
// pickWorkProgramErrorKey; the dialog stays open on failure for retry.
export function ApproveWorkProgramDialog({
  workProgramId,
  open,
  onClose,
  onApproved,
}: ApproveWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await approveWorkProgram(workProgramId)
      toast.success(t('approveDialog.successToast'))
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
          <DialogTitle>{t('approveDialog.title')}</DialogTitle>
          <DialogDescription>{t('approveDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('approveDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <CheckCircle2 className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('approveDialog.submitting') : t('approveDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
