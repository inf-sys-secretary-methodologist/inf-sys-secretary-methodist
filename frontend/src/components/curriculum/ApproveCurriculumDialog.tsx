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
import { approveCurriculum } from '@/hooks/useCurricula'

interface ApproveCurriculumDialogProps {
  curriculumId: number
  open: boolean
  onClose: () => void
  onApproved?: () => void
}

// ApproveCurriculumDialog — admin-only confirmation modal для
// pending_approval → approved transition. Mirrors SubmitCurriculumDialog
// shape (no input — backend approve endpoint accepts an empty body,
// identifies admin via JWT subject). Codebase precedent (chronicled
// 2026-05-06): state transitions consistently use dialogs (Resubmit /
// Return / Submit / Approve / Reject all wrapped). On success fires
// onApproved (caller refreshes SWR list) and onClose. On failure stays
// open so admin can retry without re-opening.
export function ApproveCurriculumDialog({
  curriculumId,
  open,
  onClose,
  onApproved,
}: ApproveCurriculumDialogProps) {
  const t = useTranslations('curriculum')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await approveCurriculum(curriculumId)
      toast.success(t('approveToast.success'))
      onApproved?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'approveToast.errors.notPending'
          break
        case 403:
          key = 'approveToast.errors.forbidden'
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
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('approveDialog.title')}</DialogTitle>
          <DialogDescription>{t('approveDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('approveDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <CheckCircle2 className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('approveDialog.submitting') : t('approveDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
