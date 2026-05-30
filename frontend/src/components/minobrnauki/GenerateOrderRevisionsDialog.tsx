'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { CheckCircle2, Loader2, Sparkles } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { generateOrderRevisions, pickMinobrnaukiOrderErrorKey } from '@/hooks/useMinobrnaukiOrders'
import type { GenerateOrderRevisionsResult } from '@/types/minobrnaukiOrder'

interface GenerateOrderRevisionsDialogProps {
  orderId: number
  open: boolean
  onClose: () => void
  onGenerated?: () => void
}

// GenerateOrderRevisionsDialog — methodist-triggered AI bulk-revision over
// every РПД an order affects (ADR-12). Two phases: a confirmation prompt
// and, after the run, a result panel summarising the counts (generated /
// skipped / failures) so the methodist sees what the LLM produced before
// the teachers review each draft via the revision flow. Mirrors
// GenerateWorkProgramDialog (no input — the backend identifies the order by
// path id + the actor by JWT subject); failures route through
// pickMinobrnaukiOrderErrorKey so the toast matches the backend code
// (RATE_LIMITED / 403 / 404 / …). On failure the dialog stays open.
export function GenerateOrderRevisionsDialog({
  orderId,
  open,
  onClose,
  onGenerated,
}: GenerateOrderRevisionsDialogProps) {
  const t = useTranslations('minobrnaukiOrder')
  const [generating, setGenerating] = useState(false)
  const [result, setResult] = useState<GenerateOrderRevisionsResult | null>(null)

  // Reset the result whenever the dialog (re)opens so a prior run's summary
  // never bleeds into a fresh confirmation.
  useEffect(() => {
    if (open) setResult(null)
  }, [open])

  const handleOpenChange = (next: boolean) => {
    if (!next && !generating) onClose()
  }

  const handleConfirm = async () => {
    if (generating) return
    setGenerating(true)
    try {
      const summary = await generateOrderRevisions(orderId)
      toast.success(t('generateDialog.successToast'))
      setResult(summary)
      onGenerated?.()
    } catch (err) {
      toast.error(t(`generateDialog.errors.${pickMinobrnaukiOrderErrorKey(err)}`))
    } finally {
      setGenerating(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('generateDialog.title')}</DialogTitle>
          <DialogDescription>
            {result ? t('generateDialog.result.description') : t('generateDialog.description')}
          </DialogDescription>
        </DialogHeader>

        {result ? (
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-sm font-medium text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="h-5 w-5" />
              {t('generateDialog.result.title')}
            </div>
            <ul className="grid gap-2 text-sm">
              <li className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2">
                <span className="text-muted-foreground">
                  {t('generateDialog.result.generated')}
                </span>
                <span data-testid="generate-result-generated" className="font-semibold">
                  {result.generated}
                </span>
              </li>
              <li className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2">
                <span className="text-muted-foreground">{t('generateDialog.result.skipped')}</span>
                <span data-testid="generate-result-skipped" className="font-semibold">
                  {result.skipped}
                </span>
              </li>
              <li className="flex items-center justify-between rounded-lg border border-border bg-card px-3 py-2">
                <span className="text-muted-foreground">{t('generateDialog.result.failures')}</span>
                <span data-testid="generate-result-failures" className="font-semibold">
                  {result.failures}
                </span>
              </li>
            </ul>
            <p className="text-xs text-muted-foreground">{t('generateDialog.result.hint')}</p>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{t('generateDialog.prompt')}</p>
        )}

        <DialogFooter className="gap-2">
          {result ? (
            <Button onClick={onClose}>{t('generateDialog.close')}</Button>
          ) : (
            <>
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
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
