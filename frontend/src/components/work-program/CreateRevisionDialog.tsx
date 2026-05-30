'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Plus } from 'lucide-react'

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
import { createRevision, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'
import {
  REVISION_CHANGE_TYPES,
  type RevisionChangeType,
  type WorkProgram,
} from '@/types/workProgram'

interface CreateRevisionDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onCreated?: (created: WorkProgram) => void
}

// CreateRevisionDialog — author-side modal for proposing a лист
// актуализации (revision) on an approved / needs_revision РПД. Mirrors
// CreateWorkProgramDialog (reset-on-open, client validation echoing the
// domain, errors via pickWorkProgramErrorKey keep the dialog open). The
// change_type uses a native <select> (the codebase uses native selects
// for simple enums — EventModal/RegisterForm) so the five domain values
// stay 1:1 with RevisionChangeType; an empty placeholder option forces an
// explicit choice. change-type labels are reused from the detail page
// namespace (detail.revisionChangeType.*) to avoid duplicating the enum.
// The author is stamped from the JWT subject server-side. diff_payload is
// omitted — structured diffs arrive later via AI bulk-revision; here the
// author records the categorized change + a human summary.
export function CreateRevisionDialog({
  workProgramId,
  open,
  onClose,
  onCreated,
}: CreateRevisionDialogProps) {
  const t = useTranslations('workProgram')
  const [changeType, setChangeType] = useState<RevisionChangeType | ''>('')
  const [summary, setSummary] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset the form every reopen so a canceled draft does not leak into
  // the next revision the author proposes.
  useEffect(() => {
    if (open) {
      setChangeType('')
      setSummary('')
    }
  }, [open])

  const trimmedSummary = summary.trim()
  const valid = changeType !== '' && trimmedSummary.length > 0

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleCreate = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      const created = await createRevision(workProgramId, {
        change_type: changeType as RevisionChangeType,
        change_summary: trimmedSummary,
      })
      toast.success(t('createRevisionDialog.successToast'))
      onCreated?.(created)
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  const selectId = `wp-create-revision-type-${workProgramId}`
  const summaryId = `wp-create-revision-summary-${workProgramId}`

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('createRevisionDialog.title')}</DialogTitle>
          <DialogDescription>{t('createRevisionDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor={selectId}>{t('createRevisionDialog.changeTypeLabel')}</Label>
            <select
              id={selectId}
              value={changeType}
              onChange={(e) => setChangeType(e.target.value as RevisionChangeType | '')}
              disabled={submitting}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="" disabled>
                {t('createRevisionDialog.changeTypePlaceholder')}
              </option>
              {REVISION_CHANGE_TYPES.map((ct) => (
                <option key={ct} value={ct}>
                  {t(`detail.revisionChangeType.${ct}`)}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor={summaryId}>{t('createRevisionDialog.summaryLabel')}</Label>
            <Textarea
              id={summaryId}
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder={t('createRevisionDialog.summaryPlaceholder')}
              rows={4}
              disabled={submitting}
            />
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('createRevisionDialog.cancel')}
          </Button>
          <Button onClick={handleCreate} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Plus className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('createRevisionDialog.creating') : t('createRevisionDialog.create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
