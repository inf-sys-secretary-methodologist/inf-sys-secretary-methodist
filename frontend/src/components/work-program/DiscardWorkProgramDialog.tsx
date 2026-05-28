'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Archive, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { discardWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface DiscardWorkProgramDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onDiscarded?: () => void
}

// DiscardWorkProgramDialog — confirmation modal for the author-side
// draft → archived transition. Archiving is terminal in the FSM (a
// discarded draft leaves the active workflow), so the confirm button is
// destructive-styled to signal the consequence. No input — the backend
// discard endpoint takes an empty body, identifying the row by path id
// + actor by JWT subject. Errors route through pickWorkProgramErrorKey;
// the dialog stays open on failure for retry.
export function DiscardWorkProgramDialog({
  workProgramId,
  open,
  onClose,
  onDiscarded,
}: DiscardWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await discardWorkProgram(workProgramId)
      toast.success(t('discardDialog.successToast'))
      onDiscarded?.()
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
          <DialogTitle>{t('discardDialog.title')}</DialogTitle>
          <DialogDescription>{t('discardDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('discardDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting} variant="destructive">
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Archive className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('discardDialog.submitting') : t('discardDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
