'use client'

import { useEffect, useMemo, useState } from 'react'
import { useTranslations } from 'next-intl'
import { Loader2, Wand2, Check } from 'lucide-react'
import { toast } from 'sonner'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'
import { useSemesters } from '@/hooks/useSchedule'
import { canGenerateSchedule } from '@/lib/auth/permissions'
import { scheduleGenerateApi } from '@/lib/api/schedule'
import { DAY_NAMES } from '@/types/schedule'
import type {
  SchedulePreview,
  GeneratedLesson,
  GenerateScheduleRequest,
  WeekType,
} from '@/types/schedule'

// Full working week (Monday..Saturday); day_of_week uses 1=Monday.
const ALL_DAYS = [1, 2, 3, 4, 5, 6]

// isConflict detects the 409 the apply endpoint returns when the semester
// already has a schedule (axios rejects with the response attached).
function isConflict(err: unknown): boolean {
  return (
    typeof err === 'object' &&
    err !== null &&
    (err as { response?: { status?: number } }).response?.status === 409
  )
}

// slotRows collapses the placed lessons into the distinct time slots present,
// sorted by slot number, so the grid renders exactly the bells that were used.
function slotRows(lessons: GeneratedLesson[]): { slot: number; start: string; end: string }[] {
  const seen = new Map<number, { slot: number; start: string; end: string }>()
  for (const l of lessons) {
    if (!seen.has(l.slot_number)) {
      seen.set(l.slot_number, { slot: l.slot_number, start: l.time_start, end: l.time_end })
    }
  }
  return Array.from(seen.values()).sort((a, b) => a.slot - b.slot)
}

export default function GenerateSchedulePage() {
  const t = useTranslations('scheduleGenerate')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const canGenerate = canGenerateSchedule(user?.role)

  const { semesters } = useSemesters()
  const [semesterId, setSemesterId] = useState<number | undefined>(undefined)
  const [days, setDays] = useState<number[]>(ALL_DAYS)
  const [preview, setPreview] = useState<SchedulePreview | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)
  const [isApplying, setIsApplying] = useState(false)

  // Preselect the active semester (fallback: the first) once semesters load, so
  // the generate action is usable without a manual pick in the common case.
  useEffect(() => {
    if (semesterId !== undefined || semesters.length === 0) return
    const active = semesters.find((s) => s.is_active) ?? semesters[0]
    setSemesterId(active.id)
  }, [semesters, semesterId])

  const rows = useMemo(() => (preview ? slotRows(preview.lessons) : []), [preview])

  if (!canGenerate) {
    return (
      <AppLayout>
        <p className="text-sm text-muted-foreground">{t('accessDenied')}</p>
      </AppLayout>
    )
  }

  const buildRequest = (): GenerateScheduleRequest => {
    const req: GenerateScheduleRequest = { semester_id: semesterId! }
    // Omit days when the full week is selected — the backend defaults to Mon-Sat.
    if (days.length > 0 && days.length < ALL_DAYS.length) {
      req.days = [...days].sort((a, b) => a - b)
    }
    return req
  }

  const toggleDay = (day: number) => {
    setDays((prev) => (prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]))
  }

  const handleGenerate = async () => {
    if (semesterId === undefined) return
    setIsGenerating(true)
    try {
      const draft = await scheduleGenerateApi.preview(buildRequest())
      setPreview(draft)
    } catch {
      toast.error(t('errors.generateFailed'))
    } finally {
      setIsGenerating(false)
    }
  }

  const handleApply = async () => {
    if (semesterId === undefined) return
    setIsApplying(true)
    try {
      const result = await scheduleGenerateApi.apply(buildRequest())
      toast.success(t('applied', { count: result.created }))
      setPreview(null)
    } catch (err) {
      toast.error(isConflict(err) ? t('errors.alreadyExists') : t('errors.applyFailed'))
    } finally {
      setIsApplying(false)
    }
  }

  const weekBadge = (wt: WeekType) =>
    wt === 'all' ? null : (
      <Badge variant="secondary" className="text-[10px]">
        {t(`weekType.${wt}`)}
      </Badge>
    )

  return (
    <AppLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
          <p className="text-sm text-muted-foreground">{t('description')}</p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t('params.title')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="max-w-xs space-y-2">
              <Label>{t('params.semester')}</Label>
              <Select
                value={semesterId ? String(semesterId) : undefined}
                onValueChange={(v) => setSemesterId(Number(v))}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t('params.selectSemester')} />
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
              <Label>{t('params.days')}</Label>
              <div className="flex flex-wrap gap-2">
                {DAY_NAMES.map((name, i) => {
                  const day = i + 1
                  const on = days.includes(day)
                  return (
                    <Button
                      key={name}
                      type="button"
                      size="sm"
                      variant={on ? 'default' : 'outline'}
                      onClick={() => toggleDay(day)}
                    >
                      {t(`days.${name}`)}
                    </Button>
                  )
                })}
              </div>
            </div>

            <Button
              onClick={handleGenerate}
              disabled={isGenerating || semesterId === undefined || days.length === 0}
            >
              {isGenerating ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <Wand2 className="h-4 w-4 mr-2" />
              )}
              {t('generate')}
            </Button>
          </CardContent>
        </Card>

        {preview && (
          <>
            <Card>
              <CardContent className="pt-6">
                <div className="flex flex-wrap items-center gap-6 text-sm">
                  <span>
                    {t('summary.total')}:{' '}
                    <span data-testid="summary-total" className="font-semibold">
                      {preview.total_requested}
                    </span>
                  </span>
                  <span className="text-green-600 dark:text-green-500">
                    {t('summary.placed')}:{' '}
                    <span data-testid="summary-placed" className="font-semibold">
                      {preview.placed_count}
                    </span>
                  </span>
                  <span className={preview.unplaced_count > 0 ? 'text-destructive' : ''}>
                    {t('summary.unplaced')}:{' '}
                    <span data-testid="summary-unplaced" className="font-semibold">
                      {preview.unplaced_count}
                    </span>
                  </span>
                  <div className="ml-auto">
                    <Button
                      onClick={handleApply}
                      disabled={isApplying || preview.placed_count === 0}
                    >
                      {isApplying ? (
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      ) : (
                        <Check className="h-4 w-4 mr-2" />
                      )}
                      {t('apply')}
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">{t('preview.title')}</CardTitle>
              </CardHeader>
              <CardContent>
                {rows.length === 0 ? (
                  <p className="text-sm text-muted-foreground py-6 text-center">
                    {t('preview.empty')}
                  </p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full border-collapse text-sm min-w-[52rem]">
                      <thead>
                        <tr>
                          <th className="p-2 text-xs font-medium text-muted-foreground border-b w-24" />
                          {DAY_NAMES.map((name) => (
                            <th
                              key={name}
                              className="p-2 text-xs font-semibold text-center border-b min-w-[9rem]"
                            >
                              {t(`days.${name}`)}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {rows.map((row) => (
                          <tr key={row.slot}>
                            <td className="p-2 text-xs text-muted-foreground text-center border-r whitespace-nowrap align-top">
                              <div className="font-medium">{row.start}</div>
                              <div className="text-[10px]">{row.end}</div>
                            </td>
                            {DAY_NAMES.map((name, i) => {
                              const day = i + 1
                              const cell = preview.lessons.filter(
                                (l) => l.day_of_week === day && l.slot_number === row.slot
                              )
                              return (
                                <td
                                  key={name}
                                  className="p-1 border border-gray-100 dark:border-gray-800 align-top"
                                >
                                  <div className="space-y-1">
                                    {cell.map((l, idx) => (
                                      <div
                                        key={`${l.load_id}-${idx}`}
                                        className="rounded-md bg-muted/50 p-1.5 text-xs leading-tight"
                                      >
                                        <div className="flex items-center justify-between gap-1">
                                          <span className="font-semibold">{l.group_name}</span>
                                          {weekBadge(l.week_type as WeekType)}
                                        </div>
                                        <div>{l.discipline_name}</div>
                                        <div className="text-muted-foreground">
                                          {l.teacher_name}
                                        </div>
                                        <div className="text-muted-foreground">
                                          {l.classroom_name} · {l.lesson_type_name}
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                </td>
                              )
                            })}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </CardContent>
            </Card>

            {preview.unplaced.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base text-destructive">
                    {t('unplaced.title')}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground mb-3">{t('unplaced.hint')}</p>
                  <ul className="space-y-2">
                    {preview.unplaced.map((u, idx) => (
                      <li
                        key={`${u.load_id}-${idx}`}
                        className="flex flex-wrap items-center gap-2 text-sm border-b pb-2 last:border-0"
                      >
                        <span className="font-semibold">{u.group_name}</span>
                        <span>{u.discipline_name}</span>
                        <span className="text-muted-foreground">{u.lesson_type_name}</span>
                        {weekBadge(u.week_type as WeekType)}
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            )}
          </>
        )}
      </div>
    </AppLayout>
  )
}
