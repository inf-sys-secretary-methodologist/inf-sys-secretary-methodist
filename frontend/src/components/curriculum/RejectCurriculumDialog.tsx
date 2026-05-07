'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, XCircle } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { rejectCurriculum } from '@/hooks/useCurricula'

const MAX_REASON = 4096

interface RejectCurriculumDialogProps {
  curriculumId: number
  open: boolean
  onClose: () => void
  onRejected?: () => void
}

// RejectCurriculumDialog — admin-only modal для pending_approval →
// draft transition с required reason. Mirrors ReturnDialog shape
// (textarea + character counter + error mapping). Reason trim
// non-empty (handler enforces; 400 if empty) and ≤ 4096 chars (mirror
// to backend validation). Per backend ADR-3 (v0.117.0) the reason is
// audited only — not stored on the entity, so a future rework cycle
// (Reject → edit → SubmitForApproval) starts с clean slate.
export function RejectCurriculumDialog({
  curriculumId,
  open,
  onClose,
  onRejected,
}: RejectCurriculumDialogProps) {
  const t = useTranslations('curriculum')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset reason on each fresh open так что admin starts blank for
  // each curriculum в queue (mirror к EditCurriculumDialog pattern
  // for stale-state avoidance).
  useEffect(() => {
    if (open) setReason('')
  }, [open])

  const trimmed = reason.trim()
  const tooLong = reason.length > MAX_REASON
  const canConfirm = trimmed.length > 0 && !tooLong && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await rejectCurriculum(curriculumId, { reason: trimmed })
      toast.success(t('rejectToast.success'))
      onRejected?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'rejectToast.errors.notPending'
          break
        case 400:
          key = 'rejectToast.errors.invalidReason'
          break
        case 403:
          key = 'rejectToast.errors.forbidden'
          break
        default:
          key = 'rejectToast.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('rejectDialog.title')}</DialogTitle>
          <DialogDescription>{t('rejectDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor={`reject-reason-${curriculumId}`}>{t('rejectDialog.reasonLabel')}</Label>
          <Textarea
            id={`reject-reason-${curriculumId}`}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t('rejectDialog.reasonPlaceholder')}
            rows={5}
            disabled={submitting}
          />
          <p className={`text-xs ${tooLong ? 'text-destructive' : 'text-muted-foreground'}`}>
            {reason.length} / {MAX_REASON}
          </p>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('rejectDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm} variant="destructive">
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <XCircle className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('rejectDialog.submitting') : t('rejectDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
