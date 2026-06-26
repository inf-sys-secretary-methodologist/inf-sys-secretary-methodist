'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, PenLine } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { signDocument, pickSignatureErrorKey } from '@/hooks/useDocumentSignatures'

interface SignDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onSigned?: () => void
}

// SignDocumentDialog — confirmation modal for applying a cryptographic
// signature to a document (#140). The server signs with the actor's per-user
// key, so there is nothing to enter — the dialog just confirms intent. On
// success it fires onSigned for caller-side SWR refresh. Mirrors
// SubmitDocumentDialog (empty-body action + status-branched error toast).
export function SignDocumentDialog({
  documentId,
  open,
  onClose,
  onSigned,
}: SignDocumentDialogProps) {
  const t = useTranslations('documentSignatures')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await signDocument(documentId)
      toast.success(t('signDialog.successToast'))
      onSigned?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickSignatureErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('signDialog.title')}</DialogTitle>
          <DialogDescription>{t('signDialog.description')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('signDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <PenLine className="mr-2 h-4 w-4" />
            )}
            {t('signDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
