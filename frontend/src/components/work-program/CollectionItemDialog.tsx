'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'
import { Loader2, Save } from 'lucide-react'

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
import { pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'
import type { WorkProgram } from '@/types/workProgram'

// CollectionItemDialog — one schema-driven modal for adding or editing
// any of the five РПД inner collections (goals / competences / topics /
// assessments / references). Rather than 10 near-duplicate dialogs, the
// caller passes a field schema + an onSubmit closure that maps the raw
// string form values into the typed *Input and calls the matching hook
// (addGoal / updateCompetence / …). The dialog stays dumb: it renders
// fields, runs presence validation on required fields, and surfaces
// errors via pickWorkProgramErrorKey (deeper invariants — text length,
// MaxScore range — are enforced by the domain and arrive as
// INVALID_WORK_PROGRAM → the invalidWorkProgram message). Mirrors the
// reset-on-open + stay-open-on-error contract of CreateRevisionDialog.

export type CollectionFieldType = 'text' | 'textarea' | 'number' | 'select'

export interface CollectionFieldOption {
  value: string
  labelKey: string
}

export interface CollectionField {
  name: string
  labelKey: string
  type: CollectionFieldType
  required?: boolean
  placeholderKey?: string
  options?: CollectionFieldOption[]
}

interface CollectionItemDialogProps {
  open: boolean
  onClose: () => void
  mode: 'add' | 'edit'
  // i18n key (under workProgram) for the dialog title, e.g.
  // 'collectionDialog.goals.addTitle'.
  titleKey: string
  fields: CollectionField[]
  // String-form initial values keyed by field name. Empty object for add.
  initialValues: Record<string, string>
  onSubmit: (values: Record<string, string>) => Promise<WorkProgram>
  onDone: (updated: WorkProgram) => void
}

export function CollectionItemDialog({
  open,
  onClose,
  mode,
  titleKey,
  fields,
  initialValues,
  onSubmit,
  onDone,
}: CollectionItemDialogProps) {
  const t = useTranslations('workProgram')
  const [values, setValues] = useState<Record<string, string>>(initialValues)
  const [submitting, setSubmitting] = useState(false)

  // Reset the form to the supplied initial values every reopen so a
  // canceled edit does not leak into the next item the author touches.
  useEffect(() => {
    if (open) setValues(initialValues)
    // initialValues identity is stable per open (caller memoizes or
    // passes a fresh object keyed by the row), so depend on open only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const setField = (name: string, value: string) =>
    setValues((prev) => ({ ...prev, [name]: value }))

  const valid = fields.every((f) => !f.required || (values[f.name] ?? '').trim().length > 0)

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleSave = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      const updated = await onSubmit(values)
      toast.success(t(`collectionDialog.${mode}SuccessToast`))
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
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t(titleKey)}</DialogTitle>
          <DialogDescription>{t('collectionDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          {fields.map((f) => {
            const fieldId = `wp-collection-field-${f.name}`
            const label = (
              <Label htmlFor={fieldId}>
                {t(f.labelKey)}
                {f.required ? <span className="text-destructive"> *</span> : null}
              </Label>
            )
            const common = {
              id: fieldId,
              value: values[f.name] ?? '',
              disabled: submitting,
            }
            return (
              <div key={f.name} className="grid gap-1.5">
                {label}
                {f.type === 'textarea' ? (
                  <Textarea
                    {...common}
                    rows={3}
                    placeholder={f.placeholderKey ? t(f.placeholderKey) : undefined}
                    onChange={(e) => setField(f.name, e.target.value)}
                  />
                ) : f.type === 'select' ? (
                  <select
                    {...common}
                    onChange={(e) => setField(f.name, e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="" disabled>
                      {t('collectionDialog.selectPlaceholder')}
                    </option>
                    {(f.options ?? []).map((o) => (
                      <option key={o.value} value={o.value}>
                        {t(o.labelKey)}
                      </option>
                    ))}
                  </select>
                ) : (
                  <Input
                    {...common}
                    type={f.type === 'number' ? 'number' : 'text'}
                    placeholder={f.placeholderKey ? t(f.placeholderKey) : undefined}
                    onChange={(e) => setField(f.name, e.target.value)}
                  />
                )}
              </div>
            )
          })}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('collectionDialog.cancel')}
          </Button>
          <Button onClick={handleSave} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('collectionDialog.saving') : t('collectionDialog.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
