'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Plus } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import { recordMinobrnaukiOrder, pickMinobrnaukiOrderErrorKey } from '@/hooks/useMinobrnaukiOrders'
import {
  MINOBRNAUKI_ORDER_CHANGE_SCOPES,
  type MinobrnaukiOrderChangeScope,
} from '@/types/minobrnaukiOrder'

interface RecordMinobrnaukiOrderDialogProps {
  open: boolean
  onClose: () => void
  onCreated?: () => void
}

// Parses the optional comma-separated affected-РПД field into a unique,
// positive-integer id list, dropping anything non-numeric.
function parseAffectedIds(raw: string): number[] {
  const ids = raw
    .split(',')
    .map((s) => Number(s.trim()))
    .filter((n) => Number.isInteger(n) && n > 0)
  return Array.from(new Set(ids))
}

// RecordMinobrnaukiOrderDialog — modal for recording a new приказ Минобрнауки.
// Mirrors CreateCurriculumDialog (Radix dialog, reset-on-reopen, toast error
// mapping keeps the dialog open). The uploader is stamped from the JWT
// subject server-side; the page-level role gate (canRecordMinobrnaukiOrder)
// guards the button that opens this dialog, mirroring the backend write gate.
export function RecordMinobrnaukiOrderDialog({
  open,
  onClose,
  onCreated,
}: RecordMinobrnaukiOrderDialogProps) {
  const t = useTranslations('minobrnaukiOrder')
  const [orderNumber, setOrderNumber] = useState('')
  const [title, setTitle] = useState('')
  const [publishedAt, setPublishedAt] = useState('')
  const [changeScope, setChangeScope] = useState<MinobrnaukiOrderChangeScope>('minor')
  const [summary, setSummary] = useState('')
  const [affected, setAffected] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setOrderNumber('')
      setTitle('')
      setPublishedAt('')
      setChangeScope('minor')
      setSummary('')
      setAffected('')
    }
  }, [open])

  const valid =
    orderNumber.trim().length > 0 && title.trim().length > 0 && publishedAt.trim().length > 0

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleRecord = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      await recordMinobrnaukiOrder({
        order_number: orderNumber.trim(),
        title: title.trim(),
        published_at: publishedAt,
        change_scope: changeScope,
        summary: summary.trim(),
        affected_work_program_ids: parseAffectedIds(affected),
      })
      toast.success(t('recordDialog.successToast'))
      onCreated?.()
      onClose()
    } catch (err) {
      toast.error(t(`recordDialog.errors.${pickMinobrnaukiOrderErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('recordDialog.title')}</DialogTitle>
          <DialogDescription>{t('recordDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="record-number">{t('recordDialog.labels.orderNumber')}</Label>
            <Input
              id="record-number"
              value={orderNumber}
              onChange={(e) => setOrderNumber(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="record-title">{t('recordDialog.labels.title')}</Label>
            <Input
              id="record-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="record-published">{t('recordDialog.labels.publishedAt')}</Label>
            <Input
              id="record-published"
              type="date"
              value={publishedAt}
              onChange={(e) => setPublishedAt(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="record-scope">{t('recordDialog.labels.changeScope')}</Label>
            <select
              id="record-scope"
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={changeScope}
              onChange={(e) => setChangeScope(e.target.value as MinobrnaukiOrderChangeScope)}
              disabled={submitting}
            >
              {MINOBRNAUKI_ORDER_CHANGE_SCOPES.map((s) => (
                <option key={s} value={s}>
                  {t(`card.changeScope.${s}`)}
                </option>
              ))}
            </select>
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="record-summary">{t('recordDialog.labels.summary')}</Label>
            <Textarea
              id="record-summary"
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              rows={3}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="record-affected">{t('recordDialog.labels.affected')}</Label>
            <Input
              id="record-affected"
              value={affected}
              onChange={(e) => setAffected(e.target.value)}
              placeholder={t('recordDialog.affectedPlaceholder')}
              disabled={submitting}
            />
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('recordDialog.cancel')}
          </Button>
          <Button onClick={handleRecord} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Plus className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('recordDialog.recording') : t('recordDialog.record')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
