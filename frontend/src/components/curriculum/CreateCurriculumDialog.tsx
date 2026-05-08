'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
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
import { createCurriculum } from '@/hooks/useCurricula'

const YEAR_MIN = 2000
const YEAR_MAX = 2100
const DESCRIPTION_MAX = 4096

interface CreateCurriculumDialogProps {
  open: boolean
  onClose: () => void
  onCreated?: () => void
}

// CreateCurriculumDialog — modal for creating a new draft curriculum.
// Mirrors EditCurriculumDialog (Radix dialog with 5-field form, client
// validation matches domain invariants, error mapping by HTTP status
// keeps the dialog open). Diverges from Edit: starts empty (not
// pre-filled), labels namespaced под createDialog.*, 422 maps to
// invalidInput (not notEditable). Backend write-whitelist methodist+
// admin v0.116.0 is enforced server-side; the page-level role check
// gates the button that opens this dialog.
export function CreateCurriculumDialog({ open, onClose, onCreated }: CreateCurriculumDialogProps) {
  const t = useTranslations('curriculum')
  const [title, setTitle] = useState('')
  const [code, setCode] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [year, setYear] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Reset form state every time the dialog reopens so a previously
  // canceled draft does not leak into the next session.
  useEffect(() => {
    if (open) {
      setTitle('')
      setCode('')
      setSpecialty('')
      setYear('')
      setDescription('')
    }
  }, [open])

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

  const handleCreate = async () => {
    if (!valid || submitting) return
    setSubmitting(true)
    try {
      await createCurriculum({
        title: trimmedTitle,
        code: trimmedCode,
        specialty: trimmedSpecialty,
        year: yearNum,
        description: description.trim(),
      })
      toast.success(t('createDialog.successToast'))
      onCreated?.()
      onClose()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let key: string
      switch (status) {
        case 409:
          key = 'createDialog.errors.codeExists'
          break
        case 422:
          key = 'createDialog.errors.invalidInput'
          break
        case 403:
          key = 'createDialog.errors.forbidden'
          break
        default:
          key = 'createDialog.errors.generic'
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
          <DialogTitle>{t('createDialog.title')}</DialogTitle>
          <DialogDescription>{t('createDialog.description')}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <div className="grid gap-1.5">
            <Label htmlFor="create-title">{t('createDialog.labels.title')}</Label>
            <Input
              id="create-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="create-code">{t('createDialog.labels.code')}</Label>
            <Input
              id="create-code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="create-specialty">{t('createDialog.labels.specialty')}</Label>
            <Input
              id="create-specialty"
              value={specialty}
              onChange={(e) => setSpecialty(e.target.value)}
              disabled={submitting}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="create-year">{t('createDialog.labels.year')}</Label>
            <Input
              id="create-year"
              inputMode="numeric"
              value={year}
              onChange={(e) => setYear(e.target.value.replace(/[^0-9]/g, ''))}
              disabled={submitting}
            />
            {!yearValid && year !== '' && (
              <p className="text-xs text-destructive">
                {t('createDialog.validation.yearRange', { min: YEAR_MIN, max: YEAR_MAX })}
              </p>
            )}
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="create-description">{t('createDialog.labels.description')}</Label>
            <Textarea
              id="create-description"
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
