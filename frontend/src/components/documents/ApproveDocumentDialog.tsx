'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { CheckCircle2, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { approveDocument } from '@/hooks/useDocumentWorkflow'

interface ApproveDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onApproved?: () => void
}

// ApproveDocumentDialog — admin-only confirmation modal для
// approval → approved transition. Defended за route-level
// RequireRole(AcademicSecretary, SystemAdmin) middleware;
// dialog still branches on 403 для defense-in-depth.
//
// Issue: #227
export function ApproveDocumentDialog({
  documentId,
  open,
  onClose,
  onApproved,
}: ApproveDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await approveDocument(documentId)
      toast.success(t('approveToast.success'))
      onApproved?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'approveToast.errors.notApproval'
          break
        case 403:
          key = 'approveToast.errors.forbidden'
          break
        case 404:
          key = 'approveToast.errors.notFound'
          break
        default:
          key = 'approveToast.errors.generic'
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
          <DialogTitle>{t('approve.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('approve.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('approve.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <CheckCircle2 className="mr-2 h-4 w-4" />
            )}
            {t('approve.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
