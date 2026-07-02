'use client'

import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useStudentGroups, useDisciplines, useSemesters, useLessonTypes } from '@/hooks/useSchedule'
import { useUsers } from '@/hooks/useUsers'
import { WEEK_TYPES } from '@/types/schedule'
import type { TeachingLoad, TeachingLoadInput, WeekType } from '@/types/teachingLoad'

interface TeachingLoadFormProps {
  entity?: TeachingLoad
  onSubmit: (input: TeachingLoadInput) => Promise<void>
  onCancel: () => void
}

// numOrZero maps a shadcn Select string value back to a numeric id (0 = unset).
function numOrZero(v: number | undefined): string {
  return v && v > 0 ? String(v) : ''
}

export function TeachingLoadForm({ entity, onSubmit, onCancel }: TeachingLoadFormProps) {
  const t = useTranslations('teachingLoad')
  const { semesters } = useSemesters()
  const { groups } = useStudentGroups()
  const { disciplines } = useDisciplines()
  const { lessonTypes } = useLessonTypes()
  const { users: teachers } = useUsers({ role: 'teacher', limit: 200 })

  const [semesterId, setSemesterId] = useState(entity?.semester_id ?? 0)
  const [groupId, setGroupId] = useState(entity?.group_id ?? 0)
  const [disciplineId, setDisciplineId] = useState(entity?.discipline_id ?? 0)
  const [teacherId, setTeacherId] = useState(entity?.teacher_id ?? 0)
  const [lessonTypeId, setLessonTypeId] = useState(entity?.lesson_type_id ?? 0)
  const [pairsPerWeek, setPairsPerWeek] = useState(entity?.pairs_per_week ?? 1)
  const [weekType, setWeekType] = useState<WeekType>(entity?.week_type ?? 'all')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!semesterId || !groupId || !disciplineId || !teacherId || !lessonTypeId) {
      setError(t('form.requiredFields'))
      return
    }
    if (pairsPerWeek < 1 || pairsPerWeek > 20) {
      setError(t('form.invalidPairs'))
      return
    }
    setError('')
    setSubmitting(true)
    try {
      await onSubmit({
        semester_id: semesterId,
        group_id: groupId,
        discipline_id: disciplineId,
        teacher_id: teacherId,
        lesson_type_id: lessonTypeId,
        pairs_per_week: pairsPerWeek,
        week_type: weekType,
      })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>{t('form.semester')}</Label>
        <Select value={numOrZero(semesterId)} onValueChange={(v) => setSemesterId(Number(v))}>
          <SelectTrigger>
            <SelectValue placeholder={t('form.selectPlaceholder')} />
          </SelectTrigger>
          <SelectContent>
            {semesters.map((s) => (
              <SelectItem key={s.id} value={String(s.id)}>
                {s.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>{t('form.group')}</Label>
        <Select value={numOrZero(groupId)} onValueChange={(v) => setGroupId(Number(v))}>
          <SelectTrigger>
            <SelectValue placeholder={t('form.selectPlaceholder')} />
          </SelectTrigger>
          <SelectContent>
            {groups.map((g) => (
              <SelectItem key={g.id} value={String(g.id)}>
                {g.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>{t('form.discipline')}</Label>
        <Select value={numOrZero(disciplineId)} onValueChange={(v) => setDisciplineId(Number(v))}>
          <SelectTrigger>
            <SelectValue placeholder={t('form.selectPlaceholder')} />
          </SelectTrigger>
          <SelectContent>
            {disciplines.map((d) => (
              <SelectItem key={d.id} value={String(d.id)}>
                {d.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>{t('form.teacher')}</Label>
        <Select value={numOrZero(teacherId)} onValueChange={(v) => setTeacherId(Number(v))}>
          <SelectTrigger>
            <SelectValue placeholder={t('form.selectPlaceholder')} />
          </SelectTrigger>
          <SelectContent>
            {teachers.map((u) => (
              <SelectItem key={u.id} value={String(u.id)}>
                {u.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label>{t('form.lessonType')}</Label>
        <Select value={numOrZero(lessonTypeId)} onValueChange={(v) => setLessonTypeId(Number(v))}>
          <SelectTrigger>
            <SelectValue placeholder={t('form.selectPlaceholder')} />
          </SelectTrigger>
          <SelectContent>
            {lessonTypes.map((lt) => (
              <SelectItem key={lt.id} value={String(lt.id)}>
                {lt.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="pairs-per-week">{t('form.pairsPerWeek')}</Label>
          <Input
            id="pairs-per-week"
            type="number"
            min={1}
            max={20}
            value={pairsPerWeek}
            onChange={(e) => {
              const n = parseInt(e.target.value, 10)
              setPairsPerWeek(Number.isNaN(n) ? 0 : n)
            }}
          />
        </div>
        <div className="space-y-2">
          <Label>{t('form.weekType')}</Label>
          <Select value={weekType} onValueChange={(v) => setWeekType(v as WeekType)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {WEEK_TYPES.map((wt) => (
                <SelectItem key={wt} value={wt}>
                  {t(`weekType.${wt}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel} disabled={submitting}>
          {t('form.cancel')}
        </Button>
        <Button type="submit" disabled={submitting}>
          {t('form.save')}
        </Button>
      </div>
    </form>
  )
}
