'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
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
import { updateCurriculum } from '@/hooks/useCurricula'
import type { Curriculum } from '@/types/curriculum'

const YEAR_MIN = 2000
const YEAR_MAX = 2100
const DESCRIPTION_MAX = 4096

interface EditCurriculumDialogProps {
  curriculum: Curriculum
  open: boolean
  onClose: () => void
  onSaved?: () => void
}

// EditCurriculumDialog — modal for editing a draft curriculum.
// Mirrors ReturnDialog's shape (Radix dialog с form, validate before
// save, error mapping with dialog staying open on error). Form fields
// mirror domain invariants verbatim (curriculum.go): title / code /
// specialty trim non-empty, year ∈ [2000, 2100], description ≤ 4096.
// Backend remains authoritative — client validation is for UX.
export function EditCurriculumDialog({
  curriculum,
  open,
  onClose,
  onSaved,
}: EditCurriculumDialogProps) {
  const t = useTranslations('curriculum')
  const [title, setTitle] = useState(curriculum.title)
  const [code, setCode] = useState(curriculum.code)
  const [specialty, setSpecialty] = useState(curriculum.specialty)
  const [year, setYear] = useState(String(curriculum.year))
  const [description, setDescription] = useState(curriculum.description)
  const [submitting, setSubmitting] = useState(false)

  // When the dialog re-opens for a different (or refreshed) curriculum
  // entity, reset local form state to match. Without this, after an
  // edit-save-close-open cycle the inputs would briefly show stale
  // values until the next keystroke.
  useEffect(() => {
    if (open) {
      setTitle(curriculum.title)
      setCode(curriculum.code)
      setSpecialty(curriculum.specialty)
      setYear(String(curriculum.year))
      setDescription(curriculum.description)
    }
  }, [open, curriculum])

  const yearNum = Number(year)
  const trimmedTitle = title.trim()
  const trimmedCode = code.trim()
  const trimmedSpecialty = specialty.trim()
  const yearValid = Number.isInteger(yearNum) && yearNum >= YEAR_MIN && yearNum <= YEAR_MAX
  const descriptionValid = description.trim().length <= DESCRIPTION_MAX
  const valid =
    trimmedTitle.length > 0 &&
    trimmedCode.length > 0 &&
    trimmedSpecialty.length > 0 &&
    yearValid &&
    descriptionValid

  const handleOpenChange = (next: boolean) => {
    if (!next && !submitting) onClose()
  }

  const handleSave = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      await updateCurriculum(curriculum.id, {
        title: trimmedTitle,
        code: trimmedCode,
        specialty: trimmedSpecialty,
        year: yearNum,
        description: description.trim(),
      })
      toast.success(t('editDialog.successToast'))
      onSaved?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'editDialog.errors.codeExists'
          break
        case 422:
          key = 'editDialog.errors.notEditable'
          break
        case 403:
          key = 'editDialog.errors.forbidden'
          break
        default:
          key = 'editDialog.errors.generic'
      }
      toast.error(t(key))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('editDialog.title')}</DialogTitle>
          <DialogDescription>{t('editDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="edit-title">{t('editDialog.labels.title')}</Label>
            <Input
              id="edit-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="edit-code">{t('editDialog.labels.code')}</Label>
            <Input
              id="edit-code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="edit-specialty">{t('editDialog.labels.specialty')}</Label>
            <Input
              id="edit-specialty"
              value={specialty}
              onChange={(e) => setSpecialty(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="edit-year">{t('editDialog.labels.year')}</Label>
            <Input
              id="edit-year"
              inputMode="numeric"
              value={year}
              onChange={(e) => setYear(e.target.value.replace(/[^0-9-]/g, ''))}
              disabled={submitting}
            />
            {!yearValid && year !== '' && (
              <p className="text-xs text-destructive">
                {t('editDialog.validation.yearRange', { min: YEAR_MIN, max: YEAR_MAX })}
              </p>
            )}
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="edit-description">{t('editDialog.labels.description')}</Label>
            <Textarea
              id="edit-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              disabled={submitting}
            />
            <p
              className={`text-xs ${descriptionValid ? 'text-muted-foreground' : 'text-destructive'}`}
            >
              {description.trim().length} / {DESCRIPTION_MAX}
            </p>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={onClose} disabled={submitting}>
            {t('editDialog.cancel')}
          </Button>
          <Button onClick={handleSave} disabled={!valid || submitting}>
            {submitting ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Save className="h-4 w-4 mr-2" />
            )}
            {submitting ? t('editDialog.saving') : t('editDialog.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
