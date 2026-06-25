'use client'

import { useEffect, useRef, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Upload } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
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

interface ImportDebtsDialogProps {
  open: boolean
  onClose: () => void
  // onImported fires after any successful upload (even one with per-row
  // errors) so the parent list can revalidate — created/updated rows are
  // already persisted regardless of the skipped ones.
  onImported?: () => void
}

// ImportDebtsDialog — uploads an xlsx debt registry and surfaces the import
// log. The backend import is idempotent on the identity tuple (existing rows
// update, new ones are created), and per-row problems do NOT fail the whole
// upload — they come back in ImportResult.Errors with the document still
// applied. So the dialog stays open after a successful upload to show the
// summary + any row errors, and calls onImported to refresh the registry.
// A forbidden actor (teacher/student) is a 403; a malformed document is a
// 400 — both route through pickStudentDebtErrorKey to a toast.
export function ImportDebtsDialog({ open, onClose, onImported }: ImportDebtsDialogProps) {
  const t = useTranslations('studentDebts')
  const [file, setFile] = useState<File | null>(null)
  const [importing, setImporting] = useState(false)
  const [result, setResult] = useState<ImportResult | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Reset on every reopen so a previous run's file/result does not leak.
  useEffect(() => {
    if (open) {
      setFile(null)
      setImporting(false)
      setResult(null)
      if (inputRef.current) inputRef.current.value = ''
    }
  }, [open])

  const handleOpenChange = (next: boolean) => {
    if (!next && !importing) onClose()
  }

  const handleImport = async () => {
    if (!file || importing) return
    setImporting(true)
    try {
      const res = await studentDebtsApi.import(file)
      setResult(res)
      toast.success(
        t('importDialog.successToast', {
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
          <DialogTitle>{t('importDialog.title')}</DialogTitle>
          <DialogDescription>{t('importDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor="sd-import-file">{t('importDialog.selectFile')}</Label>
          <input
            id="sd-import-file"
            ref={inputRef}
            type="file"
            accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
            aria-label={t('importDialog.selectFile')}
            disabled={importing}
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="block w-full text-sm text-muted-foreground file:mr-4 file:rounded-md file:border-0 file:bg-primary file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary-foreground hover:file:bg-primary/90"
          />
          <p className="text-xs text-muted-foreground">{t('importDialog.fileHint')}</p>
        </div>

        {result && result.errors.length > 0 && (
          <div className="rounded-md border border-amber-300/50 bg-amber-50 dark:bg-amber-950/20 p-3">
            <p className="mb-2 text-sm font-medium text-amber-800 dark:text-amber-300">
              {t('importDialog.rowErrorsTitle')}
            </p>
            <ul className="space-y-1 text-xs text-amber-700 dark:text-amber-400">
              {result.errors.map((e) => (
                <li key={`${e.row}-${e.identity}`}>
                  {t('importDialog.rowError', {
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
            {t('importDialog.cancel')}
          </Button>
          <Button onClick={handleImport} disabled={!file || importing}>
            {importing ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Upload className="h-4 w-4 mr-2" />
            )}
            {importing ? t('importDialog.importing') : t('importDialog.confirm')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
