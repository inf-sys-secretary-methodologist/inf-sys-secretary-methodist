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
import { createWorkProgram, pickWorkProgramErrorKey } from '@/hooks/useWorkPrograms'
import type { WorkProgram } from '@/types/workProgram'

// Mirror the domain invariants in entities/work_program.go NewWorkProgram
// so client validation matches what the backend would reject (DDD: the
// aggregate owns the rule, this is a fast-fail echo, not a second source
// of truth).
const YEAR_MIN = 2000
const YEAR_MAX = 2100
const ANNOTATION_MAX = 8192

interface CreateWorkProgramDialogProps {
  open: boolean
  onClose: () => void
  onCreated?: (created: WorkProgram) => void
}

// CreateWorkProgramDialog — modal for creating a fresh draft РПД. Mirrors
// CreateCurriculumDialog (Radix dialog, empty form, client validation
// echoing domain invariants, error mapping keeps the dialog open on
// failure) but routes errors through pickWorkProgramErrorKey — the 8a
// sentinel→i18n mapper — so the toast matches the backend's canonical
// code (identityExists for a duplicate identity tuple, etc.). The author
// is stamped from the JWT subject server-side, never a form field.
// discipline_id is entered as a raw numeric id (no picker) — a discipline
// selector would cross into the curriculum bounded context and belongs to
// a later slice; the create DTO accepts the id directly.
export function CreateWorkProgramDialog({
  open,
  onClose,
  onCreated,
}: CreateWorkProgramDialogProps) {
  const t = useTranslations('workProgram')
  const [title, setTitle] = useState('')
  const [disciplineId, setDisciplineId] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [year, setYear] = useState('')
  const [annotation, setAnnotation] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset the form every time the dialog reopens so a canceled draft does
  // not leak into the next session.
  useEffect(() => {
    if (open) {
      setTitle('')
      setDisciplineId('')
      setSpecialty('')
      setYear('')
      setAnnotation('')
    }
  }, [open])

  const trimmedTitle = title.trim()
  const trimmedSpecialty = specialty.trim()
  const trimmedAnnotation = annotation.trim()
  const disciplineIdNum = Number(disciplineId)
  const yearNum = Number(year)
  const disciplineIdValid = Number.isInteger(disciplineIdNum) && disciplineIdNum > 0
  const yearValid = Number.isInteger(yearNum) && yearNum >= YEAR_MIN && yearNum <= YEAR_MAX
  const annotationValid = trimmedAnnotation.length <= ANNOTATION_MAX
  const valid =
    trimmedTitle.length > 0 &&
    trimmedSpecialty.length > 0 &&
    disciplineIdValid &&
    yearValid &&
    annotationValid

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleCreate = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      const created = await createWorkProgram({
        title: trimmedTitle,
        discipline_id: disciplineIdNum,
        specialty_code: trimmedSpecialty,
        applicable_from_year: yearNum,
        annotation: trimmedAnnotation,
      })
      toast.success(t('createDialog.successToast'))
      onCreated?.(created)
      onClose()
    } catch (err) {
      toast.error(t(`errors.${pickWorkProgramErrorKey(err)}`))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('createDialog.title')}</DialogTitle>
          <DialogDescription>{t('createDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="wp-create-title">{t('createDialog.labels.title')}</Label>
            <Input
              id="wp-create-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="wp-create-discipline">{t('createDialog.labels.disciplineId')}</Label>
            <Input
              id="wp-create-discipline"
              inputMode="numeric"
              value={disciplineId}
              onChange={(e) => setDisciplineId(e.target.value.replace(/[^0-9]/g, ''))}
              placeholder={t('createDialog.placeholders.disciplineId')}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="wp-create-specialty">{t('createDialog.labels.specialty')}</Label>
            <Input
              id="wp-create-specialty"
              value={specialty}
              onChange={(e) => setSpecialty(e.target.value)}
              placeholder={t('createDialog.placeholders.specialty')}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="wp-create-year">{t('createDialog.labels.year')}</Label>
            <Input
              id="wp-create-year"
              inputMode="numeric"
              value={year}
              onChange={(e) => setYear(e.target.value.replace(/[^0-9]/g, ''))}
              placeholder={t('createDialog.placeholders.year')}
              disabled={submitting}
            />
            {!yearValid && year !== '' && (
              <p className="text-xs text-destructive">
                {t('createDialog.validation.yearRange', { min: YEAR_MIN, max: YEAR_MAX })}
              </p>
            )}
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="wp-create-annotation">{t('createDialog.labels.annotation')}</Label>
            <Textarea
              id="wp-create-annotation"
              value={annotation}
              onChange={(e) => setAnnotation(e.target.value)}
              rows={4}
              disabled={submitting}
            />
            <p
              className={`text-xs ${annotationValid ? 'text-muted-foreground' : 'text-destructive'}`}
            >
              {trimmedAnnotation.length} / {ANNOTATION_MAX}
            </p>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('createDialog.cancel')}
          </Button>
          <Button onClick={handleCreate} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Plus className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('createDialog.creating') : t('createDialog.create')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
