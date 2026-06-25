'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { CalendarClock, Loader2 } from 'lucide-react'

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
import { scheduleResit, pickStudentDebtErrorKey } from '@/hooks/useStudentDebts'

interface ScheduleResitDialogProps {
  debtId: number
  open: boolean
  onClose: () => void
  onScheduled?: () => void
}

// ScheduleResitDialog — appends a resit attempt and moves the debt to
// resit_scheduled (FSM: allowed from open / commission; from commission the
// attempt is itself a commission resit — the backend decides, not the UI).
// The date input is a calendar day; it is parsed as LOCAL midnight (per the
// project date rule) and serialized to an RFC3339 timestamp the backend
// expects. Errors (debt closed / invalid transition) route through
// pickStudentDebtErrorKey and keep the dialog open for a retry.
export function ScheduleResitDialog({
  debtId,
  open,
  onClose,
  onScheduled,
}: ScheduleResitDialogProps) {
  const t = useTranslations('studentDebts')
  const [date, setDate] = useState('')
  const [examiner, setExaminer] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setDate('')
      setExaminer('')
      setSubmitting(false)
    }
  }, [open])

  const trimmedExaminer = examiner.trim()
  const valid = date !== '' && trimmedExaminer.length > 0

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      // A resit date is a calendar day, not a local instant — serialize it as
      // UTC midnight so the day the user picked survives the round-trip
      // regardless of the viewer's timezone (a local-midnight ISO would shift
      // the calendar day under a positive UTC offset).
      const scheduled = `${date}T00:00:00Z`
      await scheduleResit(debtId, { scheduled_date: scheduled, examiner: trimmedExaminer })
      toast.success(t('scheduleDialog.successToast'))
      onScheduled?.()
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
          <DialogTitle>{t('scheduleDialog.title')}</DialogTitle>
          <DialogDescription>{t('scheduleDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="sd-resit-date">{t('scheduleDialog.labels.scheduledDate')}</Label>
            <Input
              id="sd-resit-date"
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="sd-resit-examiner">{t('scheduleDialog.labels.examiner')}</Label>
            <Input
              id="sd-resit-examiner"
              value={examiner}
              onChange={(e) => setExaminer(e.target.value)}
              placeholder={t('scheduleDialog.placeholders.examiner')}
              disabled={submitting}
            />
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('scheduleDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <CalendarClock className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('scheduleDialog.submitting') : t('scheduleDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
