'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Save, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import axios from 'axios'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { saveGrade } from '@/hooks/useAssignments'
import { validateGrade } from '@/lib/assignments/validation'
import type { SubmissionView } from '@/types/assignments'

interface GradeFormProps {
  assignmentId: number
  maxScore: number
  submission: SubmissionView
  onSaved?: () => void
}

// GradeForm — inline form for grading a single submission. Validation
// runs on the client (validateGrade) before the round-trip; the
// backend is the source of truth and re-validates, but the early
// feedback keeps the teacher's flow uninterrupted on obvious typos.
export function GradeForm({ assignmentId, maxScore, submission, onSaved }: GradeFormProps) {
  const t = useTranslations('assignments.gradeForm')

  const initialValue = submission.grade_value != null ? String(submission.grade_value) : ''
  const [value, setValue] = useState(initialValue)
  const [feedback, setFeedback] = useState(submission.feedback ?? '')
  const [submitting, setSubmitting] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const validation = validateGrade(value, maxScore)
    if (!validation.ok) {
      const reasonKey = `validation.${validation.reason}`
      const message = t(reasonKey, { max: maxScore })
      setValidationError(message)
      return
    }
    setValidationError(null)

    setSubmitting(true)
    try {
      await saveGrade(assignmentId, {
        student_id: submission.student_id,
        value: validation.value,
        feedback: feedback.trim() || undefined,
      })
      toast.success(t('saveSuccess', { name: submission.student_name }))
      onSaved?.()
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      let messageKey: string
      switch (status) {
        case 409:
          messageKey = 'errors.alreadyGraded'
          break
        case 422:
          messageKey = 'errors.invalidValue'
          break
        case 403:
          messageKey = 'errors.forbidden'
          break
        default:
          messageKey = 'errors.generic'
      }
      toast.error(t(messageKey))
    } finally {
      setSubmitting(false)
    }
  }

  // Disable the grade form for any status that is NOT pending: graded
  // stays disabled (re-grading must go through Return), and returned
  // stays disabled until the student resubmits (resubmission flow is
  // v0.112.0). When the status flips back to pending, the SubmissionRow
  // remounts GradeForm via key={...:status:...} so fresh state is
  // guaranteed on each transition.
  const isAlreadyGraded = submission.status !== 'pending'
  const buttonLabel =
    submission.status === 'graded'
      ? t('alreadySaved')
      : submission.status === 'returned'
        ? t('returnedReadOnly')
        : t('save')

  return (
    <form onSubmit={handleSubmit} className="grid gap-3 md:grid-cols-[120px_1fr_auto] md:items-end">
      <div className="space-y-1.5">
        <Label htmlFor={`grade-${submission.id}`}>{t('valueLabel', { max: maxScore })}</Label>
        <Input
          id={`grade-${submission.id}`}
          inputMode="numeric"
          pattern="[0-9]*"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          disabled={submitting || isAlreadyGraded}
          aria-invalid={validationError != null}
          aria-describedby={validationError ? `grade-error-${submission.id}` : undefined}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor={`feedback-${submission.id}`}>{t('feedbackLabel')}</Label>
        <Input
          id={`feedback-${submission.id}`}
          value={feedback}
          onChange={(e) => setFeedback(e.target.value)}
          maxLength={4096}
          disabled={submitting || isAlreadyGraded}
          placeholder={t('feedbackPlaceholder')}
        />
      </div>

      <Button type="submit" disabled={submitting || isAlreadyGraded}>
        {submitting ? (
          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
        ) : (
          <Save className="h-4 w-4 mr-2" />
        )}
        {buttonLabel}
      </Button>

      {validationError && (
        <p
          id={`grade-error-${submission.id}`}
          className="md:col-span-3 text-sm text-destructive"
          role="alert"
        >
          {validationError}
        </p>
      )}
    </form>
  )
}
