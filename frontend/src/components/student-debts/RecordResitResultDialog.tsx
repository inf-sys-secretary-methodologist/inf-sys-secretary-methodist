'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { ClipboardCheck, Loader2 } from 'lucide-react'

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
import { recordResitResult, pickStudentDebtErrorKey } from '@/hooks/useStudentDebts'
import { RESIT_RESULTS, type ResitResult } from '@/types/studentDebts'
import { resitResultKey } from './status'

interface RecordResitResultDialogProps {
  debtId: number
  // attemptNo identifies the currently scheduled attempt (the last one);
  // the backend records the outcome of that resit and advances the FSM.
  attemptNo: number
  open: boolean
  onClose: () => void
  onRecorded?: () => void
}

// The recordable results — pending is the not-yet-recorded sentinel, so it is
// excluded from the picker (you record an outcome, never "pending").
const RECORDABLE: ResitResult[] = RESIT_RESULTS.filter((r) => r !== 'pending')

// RecordResitResultDialog — records the outcome of the scheduled resit.
// passed → closed_passed; a failed/no_show advances the FSM (back to open,
// or to commission once regular attempts are exhausted, or to closed_failed
// on a commission attempt — the backend decides). The grade is optional (a
// pass/fail zachet carries no numeric grade); an empty field is sent as null.
export function RecordResitResultDialog({
  debtId,
  attemptNo,
  open,
  onClose,
  onRecorded,
}: RecordResitResultDialogProps) {
  const t = useTranslations('studentDebts')
  const [result, setResult] = useState<ResitResult>('passed')
  const [grade, setGrade] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setResult('passed')
      setGrade('')
      setSubmitting(false)
    }
  }, [open])

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      const trimmed = grade.trim()
      await recordResitResult(debtId, attemptNo, {
        result,
        grade: trimmed === '' ? null : Number(trimmed),
      })
      toast.success(t('recordDialog.successToast'))
      onRecorded?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickStudentDebtErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('recordDialog.title')}</DialogTitle>
          <DialogDescription>{t('recordDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="sd-record-result">{t('recordDialog.labels.result')}</Label>
            <select
              id="sd-record-result"
              aria-label={t('recordDialog.labels.result')}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={result}
              onChange={(e) => setResult(e.target.value as ResitResult)}
              disabled={submitting}
            >
              {RECORDABLE.map((r) => (
                <option key={r} value={r}>
                  {t(`recordDialog.resultOptions.${resitResultKey(r)}`)}
                </option>
              ))}
            </select>
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="sd-record-grade">{t('recordDialog.labels.grade')}</Label>
            <Input
              id="sd-record-grade"
              inputMode="numeric"
              value={grade}
              onChange={(e) => setGrade(e.target.value.replace(/[^0-9]/g, ''))}
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">{t('recordDialog.gradeHint')}</p>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('recordDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <ClipboardCheck className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('recordDialog.submitting') : t('recordDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
