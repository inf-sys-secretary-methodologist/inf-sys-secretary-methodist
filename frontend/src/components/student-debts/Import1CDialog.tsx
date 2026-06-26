'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Database, Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { studentDebtsApi } from '@/lib/api/studentDebts'
import { pickStudentDebtErrorKey } from '@/hooks/useStudentDebts'
import type { ImportResult } from '@/types/studentDebts'

interface Import1CDialogProps {
  open: boolean
  onClose: () => void
  // onImported fires after a successful sync (even one with per-row errors)
  // so the parent list can revalidate.
  onImported?: () => void
}

// Import1CDialog triggers a server-side pull of the 1С academic-debt catalog
// into the registry. Unlike the xlsx import there is no file picker — the
// source is the 1С OData API, so the dialog is a single confirm action. The
// backend sync is idempotent on the identity tuple and per-row problems do
// NOT fail the whole pull (they come back in ImportResult.Errors with the
// catalog still applied), so the dialog stays open to show the summary + any
// row errors and calls onImported to refresh the registry. A forbidden actor
// (teacher/student) is a 403; a 1С upstream failure is a 502 — both route
// through pickStudentDebtErrorKey to a toast.
export function Import1CDialog({ open, onClose, onImported }: Import1CDialogProps) {
  const t = useTranslations('studentDebts')
  const [importing, setImporting] = useState(false)
  const [result, setResult] = useState<ImportResult | null>(null)

  // Reset on every reopen so a previous run's result does not leak.
  useEffect(() => {
    if (open) {
      setImporting(false)
      setResult(null)
    }
  }, [open])

  const handleOpenChange = (next: boolean) => {
    if (!next && !importing) onClose()
  }

  const handleImport = async () => {
    if (importing) return
    setImporting(true)
    try {
      const res = await studentDebtsApi.import1C()
      setResult(res)
      toast.success(
        t('import1CDialog.successToast', {
          created: res.created,
          updated: res.updated,
          skipped: res.skipped,
        })
      )
      onImported?.()
    } catch (err) {
      toast.error(t(`errors.${pickStudentDebtErrorKey(err)}`))
    } finally {
      setImporting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('import1CDialog.title')}</DialogTitle>
          <DialogDescription>{t('import1CDialog.description')}</DialogDescription>
        </DialogHeader>

        {result && result.errors.length > 0 && (
          <div className="rounded-md border border-amber-300/50 bg-amber-50 dark:bg-amber-950/20 p-3">
            <p className="mb-2 text-sm font-medium text-amber-800 dark:text-amber-300">
              {t('import1CDialog.rowErrorsTitle')}
            </p>
            <ul className="space-y-1 text-xs text-amber-700 dark:text-amber-400">
              {result.errors.map((e) => (
                <li key={`${e.row}-${e.identity}`}>
                  {t('import1CDialog.rowError', {
                    row: e.row,
                    identity: e.identity,
                    message: e.message,
                  })}
                </li>
              ))}
            </ul>
          </div>
        )}

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={importing}>
            {t('import1CDialog.cancel')}
          </Button>
          <Button onClick={handleImport} disabled={importing}>
            {importing ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Database className="h-4 w-4 mr-2" />
            )}
            {importing ? t('import1CDialog.importing') : t('import1CDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
