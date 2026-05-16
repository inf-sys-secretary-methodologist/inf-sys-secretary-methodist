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
import { rejectDocument } from '@/hooks/useDocumentWorkflow'

// Backend RejectionReason VO invariant: rune count в [10, 500] после trim.
const MIN_REASON = 10
const MAX_REASON = 500

interface RejectDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onRejected?: () => void
}

// RejectDocumentDialog — admin-only modal для approval → rejected
// transition с required reason. Reason VO validated server-side:
// rune count [10, 500] после trim. Frontend pre-validates to give
// immediate feedback и cuts down round-trips. Diverges from
// RejectCurriculumDialog (max 4096, audit-only) — backend documents
// stores the reason on the entity для rework context.
//
// Issue: #227
export function RejectDocumentDialog({
  documentId,
  open,
  onClose,
  onRejected,
}: RejectDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Clear reason on every open так что admin starts blank on each
  // document в queue (mirror к RejectCurriculumDialog approach).
  useEffect(() => {
    if (open) setReason('')
  }, [open])

  const trimmed = reason.trim()
  const length = Array.from(trimmed).length // rune-aware length (matches backend Go []rune count)
  const tooShort = length < MIN_REASON
  const tooLong = length > MAX_REASON
  const canConfirm = !tooShort && !tooLong && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await rejectDocument(documentId, { reason: trimmed })
      toast.success(t('rejectToast.success'))
      onRejected?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          // 422 covers both invalid reason AND not-in-approval-status.
          // Backend wraps either ErrRejectionReasonInvalid или
          // ErrCannotReject — frontend cannot distinguish without
          // parsing the message, so falls к generic invalid-reason
          // copy. Acceptable trade-off для defense window.
          key = 'rejectToast.errors.invalidOrConflict'
          break
        case 409:
          key = 'rejectToast.errors.notApproval'
          break
        case 403:
          key = 'rejectToast.errors.forbidden'
          break
        case 404:
          key = 'rejectToast.errors.notFound'
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('reject.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('reject.dialogBody')}</DialogDescription>
        </DialogHeader>
        <div className="space-y-2 py-2">
          <Label htmlFor="reject-reason">{t('reject.reasonLabel')}</Label>
          <Textarea
            id="reject-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t('reject.reasonPlaceholder')}
            disabled={submitting}
            rows={5}
          />
          <p className="text-muted-foreground text-sm">
            {length} / {MAX_REASON}
            {tooShort && (
              <span className="text-destructive ml-2">
                {t('reject.reasonTooShort', { min: MIN_REASON })}
              </span>
            )}
            {tooLong && (
              <span className="text-destructive ml-2">
                {t('reject.reasonTooLong', { max: MAX_REASON })}
              </span>
            )}
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('reject.cancelLabel')}
          </Button>
          <Button variant="destructive" onClick={handleConfirm} disabled={!canConfirm}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <XCircle className="mr-2 h-4 w-4" />
            )}
            {t('reject.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
