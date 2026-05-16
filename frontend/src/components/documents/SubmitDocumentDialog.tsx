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
import { submitDocument } from '@/hooks/useDocumentWorkflow'

interface SubmitDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onSubmitted?: () => void
}

// SubmitDocumentDialog — confirmation modal for the draft → approval
// transition. Mirror к SubmitCurriculumDialog: empty body submit
// + onSubmitted callback for caller-side SWR refresh + error
// branching by HTTP status. Codebase precedent: state transitions
// use dialogs (per feedback_state_transitions_use_dialogs).
//
// Issue: #227
export function SubmitDocumentDialog({
  documentId,
  open,
  onClose,
  onSubmitted,
}: SubmitDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await submitDocument(documentId)
      toast.success(t('submitToast.success'))
      onSubmitted?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'submitToast.errors.notDraft'
          break
        case 403:
          key = 'submitToast.errors.forbidden'
          break
        case 404:
          key = 'submitToast.errors.notFound'
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('submit.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('submit.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('submit.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Send className="mr-2 h-4 w-4" />
            )}
            {t('submit.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
