'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Sparkles } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { generateWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'

interface GenerateWorkProgramDialogProps {
  workProgramId: number
  open: boolean
  onClose: () => void
  onGenerated?: () => void
}

// GenerateWorkProgramDialog — confirmation modal for filling an empty
// draft РПД from the LLM (GenerateDraftUseCase, see PR 5a/5b). Mirrors
// SubmitWorkProgramDialog: no input — the backend generate endpoint takes
// an empty body and identifies the row by path id + the actor by JWT
// subject. The dialog never replicates the backend invariants (draft must
// be empty, hourly quota) — those live in the domain; instead failures
// route through pickWorkProgramErrorKey so the toast matches the backend's
// canonical code (DRAFT_NOT_EMPTY / RATE_LIMITED / …). On failure the
// dialog stays open for retry.
export function GenerateWorkProgramDialog({
  workProgramId,
  open,
  onClose,
  onGenerated,
}: GenerateWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [generating, setGenerating] = useState(false)

  const handleOpenChange = (next: boolean) => {
    if (!next && !generating) onClose()
  }

  const handleConfirm = async () => {
    if (generating) return
    setGenerating(true)
    try {
      await generateWorkProgram(workProgramId)
      toast.success(t('generateDialog.successToast'))
      onGenerated?.()
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setGenerating(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('generateDialog.title')}</DialogTitle>
          <DialogDescription>{t('generateDialog.description')}</DialogDescription>
        </DialogHeader>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={generating}>
            {t('generateDialog.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={generating}>
            {generating ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Sparkles className="h-4 w-4 mr-2" />
            )}
            {generating ? t('generateDialog.generating') : t('generateDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
