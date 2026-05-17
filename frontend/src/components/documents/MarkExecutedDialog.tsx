'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { CheckCheck, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { markExecutedDocument } from '@/hooks/useDocumentWorkflow'

interface MarkExecutedDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onMarked?: () => void
}

// MarkExecutedDialog — admin-only confirmation modal для execution →
// executed transition. Confirmation-only per Phase 3 ADR-5 (no input
// field — terminal step).
//
// Issue: #232
export function MarkExecutedDialog({
  documentId,
  open,
  onClose,
  onMarked,
}: MarkExecutedDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await markExecutedDocument(documentId)
      toast.success(t('markExecutedToast.success'))
      onMarked?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'markExecutedToast.errors.notExecution'
          break
        case 403:
          key = 'markExecutedToast.errors.forbidden'
          break
        case 404:
          key = 'markExecutedToast.errors.notFound'
          break
        default:
          key = 'markExecutedToast.errors.generic'
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
          <DialogTitle>{t('markExecuted.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('markExecuted.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('markExecuted.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <CheckCheck className="mr-2 h-4 w-4" />
            )}
            {t('markExecuted.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
