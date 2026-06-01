'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Trash2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'
import type { WorkProgram } from '@/types/workProgram'

// DeleteCollectionItemDialog — one generic confirm modal for removing an
// item from any of the five РПД inner collections. The caller passes the
// delete closure (deleteGoal / deleteCompetence / …) and a short preview
// of the item being removed so the author sees what they are about to
// delete. Destructive-styled confirm; stays open on error for retry.

interface DeleteCollectionItemDialogProps {
  open: boolean
  onClose: () => void
  // Short human preview of the item (goal text / competence code / topic
  // title …) rendered in the confirmation body.
  itemLabel: string
  onConfirm: () => Promise<WorkProgram>
  onDone: (updated: WorkProgram) => void
}

export function DeleteCollectionItemDialog({
  open,
  onClose,
  itemLabel,
  onConfirm,
  onDone,
}: DeleteCollectionItemDialogProps) {
  const t = useTranslations('workProgram')
  const [submitting, setSubmitting] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleConfirm = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      const updated = await onConfirm()
      toast.success(t('collectionDialog.deleteSuccessToast'))
      onDone(updated)
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('collectionDialog.deleteTitle')}</DialogTitle>
          <DialogDescription>{t('collectionDialog.deleteDescription')}</DialogDescription>
        </DialogHeader>

        {itemLabel ? (
          <p className="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm">
            {itemLabel}
          </p>
        ) : null}

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('collectionDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={submitting} variant="destructive">
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Trash2 className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('collectionDialog.deleting') : t('collectionDialog.deleteConfirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
