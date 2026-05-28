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
import { submitWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface SubmitWorkProgramDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onSubmitted?: () => void
}

// SubmitWorkProgramDialog — confirmation modal for the author-side
// draft → pending_approval transition (РПД author = teacher / methodist
// / admin per ADR-5). Mirrors SubmitCurriculumDialog: no input — the
// backend submit endpoint accepts an empty body and identifies the row
// by path id + the actor by JWT subject. Wrapping the transition in a
// dialog matches the codebase precedent (state transitions use dialogs)
// and prevents accidental submits of a still-being-drafted programme.
// Errors route through pickWorkProgramErrorKey (the 8a sentinel→i18n
// mapper) so the toast matches the backend's canonical error code; on
// failure the dialog stays open for retry.
export function SubmitWorkProgramDialog({
  workProgramId,
  open,
  onClose,
  onSubmitted,
}: SubmitWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await submitWorkProgram(workProgramId)
      toast.success(t('submitDialog.successToast'))
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
          <DialogTitle>{t('submitDialog.title')}</DialogTitle>
          <DialogDescription>{t('submitDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('submitDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Send className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('submitDialog.submitting') : t('submitDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
