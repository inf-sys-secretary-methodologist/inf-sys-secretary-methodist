'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { RotateCcw, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { resubmitDocument } from '@/hooks/useDocumentWorkflow'

interface ResubmitDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onResubmitted?: () => void
}

// ResubmitDocumentDialog — author/admin confirmation modal для rejected →
// draft rework cycle. Clears RejectedBy/At/Reason audit fields server-side.
// Confirmation-only per ADR-5 (no input field — author then revises via
// existing draft edit flow).
//
// Issue: #233
export function ResubmitDocumentDialog({
  documentId,
  open,
  onClose,
  onResubmitted,
}: ResubmitDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await resubmitDocument(documentId)
      toast.success(t('resubmitToast.success'))
      onResubmitted?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'resubmitToast.errors.notRejected'
          break
        case 403:
          key = 'resubmitToast.errors.forbidden'
          break
        case 404:
          key = 'resubmitToast.errors.notFound'
          break
        default:
          key = 'resubmitToast.errors.generic'
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
          <DialogTitle>{t('resubmit.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('resubmit.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('resubmit.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <RotateCcw className="mr-2 h-4 w-4" />
            )}
            {t('resubmit.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
