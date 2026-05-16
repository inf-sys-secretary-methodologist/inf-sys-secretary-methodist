'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2, FileSignature } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { registerDocument } from '@/hooks/useDocumentWorkflow'

// Backend invariant: rune count ≥3 после trim, UNIQUE constraint
// on registration_number AT registered_by IS NOT NULL.
const MIN_NUMBER_LEN = 3

interface RegisterDocumentDialogProps {
  documentId: number
  open: boolean
  onClose: () => void
  onRegistered?: () => void
}

// RegisterDocumentDialog — admin-only modal для approved → registered
// transition с required registration_number. Mirror к Reject dialog
// pattern (input + length validation + status-aware error mapping).
//
// Issue: #230
export function RegisterDocumentDialog({
  documentId,
  open,
  onClose,
  onRegistered,
}: RegisterDocumentDialogProps) {
  const t = useTranslations('documentsWorkflow')
  const [number, setNumber] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) setNumber('')
  }, [open])

  const trimmed = number.trim()
  const length = Array.from(trimmed).length
  const tooShort = length < MIN_NUMBER_LEN
  const canConfirm = !tooShort && !submitting

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!canConfirm) return
    setSubmitting(true)
    try {
      await registerDocument(documentId, { number: trimmed })
      toast.success(t('registerToast.success'))
      onRegistered?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 422:
          key = 'registerToast.errors.invalidNumber'
          break
        case 409:
          key = 'registerToast.errors.notApproved'
          break
        case 403:
          key = 'registerToast.errors.forbidden'
          break
        case 404:
          key = 'registerToast.errors.notFound'
          break
        default:
          key = 'registerToast.errors.generic'
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
          <DialogTitle>{t('register.dialogTitle')}</DialogTitle>
          <DialogDescription>{t('register.dialogBody')}</DialogDescription>
        </DialogHeader>
        <div className="space-y-2 py-2">
          <Label htmlFor="register-number">{t('register.numberLabel')}</Label>
          <Input
            id="register-number"
            value={number}
            onChange={(e) => setNumber(e.target.value)}
            placeholder={t('register.numberPlaceholder')}
            disabled={submitting}
          />
          {tooShort && trimmed.length > 0 && (
            <p className="text-destructive text-sm">
              {t('register.numberTooShort', { min: MIN_NUMBER_LEN })}
            </p>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('register.cancelLabel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!canConfirm}>
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <FileSignature className="mr-2 h-4 w-4" />
            )}
            {t('register.confirmLabel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
