'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Send } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { submitRevision, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface SubmitRevisionDialogProps {
  workProgramId: number
  revisionId: number
  open: boolean
  onClose: () => void
  onSubmitted?: () => void
}

// SubmitRevisionDialog — confirmation modal for the author-side
// draft → pending_approval transition of a лист актуализации (revision).
// Mirrors SubmitWorkProgramDialog: empty body — path ids (РПД + revision)
// + JWT subject identify the row + actor. Errors route through
// pickWorkProgramErrorKey; on failure the dialog stays open for retry.
export function SubmitRevisionDialog({
  workProgramId,
  revisionId,
  open,
  onClose,
  onSubmitted,
}: SubmitRevisionDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await submitRevision(workProgramId, revisionId)
      toast.success(t('submitRevisionDialog.successToast'))
      onSubmitted?.()
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
          <DialogTitle>{t('submitRevisionDialog.title')}</DialogTitle>
          <DialogDescription>{t('submitRevisionDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('submitRevisionDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Send className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('submitRevisionDialog.submitting') : t('submitRevisionDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
