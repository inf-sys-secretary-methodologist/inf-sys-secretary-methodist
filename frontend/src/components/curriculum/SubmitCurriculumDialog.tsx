'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
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
import { submitCurriculum } from '@/hooks/useCurricula'

interface SubmitCurriculumDialogProps {
  curriculumId: number
  open: boolean
  onClose: () => void
  onSubmitted?: () => void
}

// SubmitCurriculumDialog — confirmation modal for the methodist-side
// status='draft' → 'pending_approval' transition. Mirrors ResubmitDialog
// (no input — backend submit endpoint accepts an empty body, identifies
// the row by path id + JWT subject). Keeping the dialog wrapper instead
// of an inline button matches the codebase precedent (state transitions
// use dialogs: ResubmitDialog, ReturnDialog, …) and prevents
// methodist mis-clicks on "Submit" while a curriculum is still being
// drafted. On success fires onSubmitted (caller refreshes SWR) and
// onClose. On failure stays open so methodist can retry without
// re-opening.
export function SubmitCurriculumDialog({
  curriculumId,
  open,
  onClose,
  onSubmitted,
}: SubmitCurriculumDialogProps) {
  const t = useTranslations('curriculum')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await submitCurriculum(curriculumId)
      toast.success(t('submitToast.success'))
      onSubmitted?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'submitToast.errors.notDraft'
          break
        case 403:
          key = 'submitToast.errors.forbidden'
          break
        default:
          key = 'submitToast.errors.generic'
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
