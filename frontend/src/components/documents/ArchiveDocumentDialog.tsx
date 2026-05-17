'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
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
import { archiveDocument } from '@/hooks/useDocumentWorkflow'

interface ArchiveDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onArchived?: () => void
}

// ArchiveDocumentDialog — admin-only confirmation modal для executed →
// archived terminal transition. Confirmation-only per ADR-5 (no input
// field — terminal step closing lifecycle).
//
// Issue: #233
export function ArchiveDocumentDialog({
  documentId,
  open,
  onClose,
  onArchived,
}: ArchiveDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await archiveDocument(documentId)
      toast.success(t('archiveToast.success'))
      onArchived?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'archiveToast.errors.notExecuted'
          break
        case 403:
          key = 'archiveToast.errors.forbidden'
          break
        case 404:
          key = 'archiveToast.errors.notFound'
          break
        default:
          key = 'archiveToast.errors.generic'
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
          <DialogTitle>{t('archive.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('archive.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('archive.cancelLabel')}
          </Button>
          <Button variant="destructive" onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Archive className="mr-2 h-4 w-4" />
            )}
            {t('archive.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
