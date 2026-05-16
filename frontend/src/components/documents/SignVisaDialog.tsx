'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, Stamp } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { signVisaDocument } from '@/hooks/useDocumentWorkflow'

interface SignVisaDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onSigned?: () => void
}

// SignVisaDialog — admin-only confirmation modal для
// routing → execution transition (single-step visa per ADR-1).
// Defended за route-level RequireRole middleware; dialog still
// branches on 403 для defense-in-depth.
//
// Issue: #231
export function SignVisaDialog({ documentId, open, onClose, onSigned }: SignVisaDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await signVisaDocument(documentId)
      toast.success(t('signVisaToast.success'))
      onSigned?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'signVisaToast.errors.notRouting'
          break
        case 403:
          key = 'signVisaToast.errors.forbidden'
          break
        case 404:
          key = 'signVisaToast.errors.notFound'
          break
        default:
          key = 'signVisaToast.errors.generic'
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
          <DialogTitle>{t('signVisa.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('signVisa.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('signVisa.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Stamp className="mr-2 h-4 w-4" />
            )}
            {t('signVisa.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
