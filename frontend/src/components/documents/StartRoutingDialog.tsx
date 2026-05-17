'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, Route } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { startRoutingDocument } from '@/hooks/useDocumentWorkflow'

interface StartRoutingDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onRouted?: () => void
}

// StartRoutingDialog — admin-only confirmation modal для
// registered → routing transition. Defended за route-level
// RequireRole(AcademicSecretary, SystemAdmin) middleware;
// dialog still branches on 403 для defense-in-depth.
//
// Issue: #231
export function StartRoutingDialog({
  documentId,
  open,
  onClose,
  onRouted,
}: StartRoutingDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await startRoutingDocument(documentId)
      toast.success(t('routingToast.success'))
      onRouted?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'routingToast.errors.notRegistered'
          break
        case 403:
          key = 'routingToast.errors.forbidden'
          break
        case 404:
          key = 'routingToast.errors.notFound'
          break
        default:
          key = 'routingToast.errors.generic'
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
          <DialogTitle>{t('routing.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('routing.dialogBody')}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('routing.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Route className="mr-2 h-4 w-4" />
            )}
            {t('routing.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
