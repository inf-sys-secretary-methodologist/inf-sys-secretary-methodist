'use client'

import { useState, useMemo } from 'react'
import { useTranslations } from 'next-intl'
import { Calendar, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TimetableGrid } from '@/components/schedule/TimetableGrid'
import { ScheduleFilters } from '@/components/schedule/ScheduleFilters'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuthStore } from '@/stores/authStore'
import { can, Resource, Action } from '@/lib/auth/permissions'
import {
  useScheduleTimetable,
  useClassrooms,
  useStudentGroups,
  useSemesters,
} from '@/hooks/useSchedule'
import type { LessonFilterParams, WeekType, Lesson } from '@/types/schedule'
import { WEEK_TYPES } from '@/types/schedule'

export default function SchedulePage() {
  const t = useTranslations('schedule')
  useAuthCheck()
  const user = useAuthStore((s) => s.user)
  const userCanEdit = can(user?.role, Resource.SCHEDULE, Action.CREATE)

  const [weekType, setWeekType] = useState<WeekType>('all')
  const [filters, setFilters] = useState<LessonFilterParams>({})

  // Auto-select active semester on load
  const { semesters } = useSemesters()
  const effectiveFilters = useMemo(() => {
    const f = { ...filters }
    if (!f.semester_id && semesters.length > 0) {
      const active = semesters.find((s) => s.is_active)
      if (active) f.semester_id = active.id
    }
    return f
  }, [filters, semesters])

  const { lessons, isLoading, error } = useScheduleTimetable(effectiveFilters)
  const { classrooms } = useClassrooms()
  const { groups } = useStudentGroups()

  // Filter lessons by week_type on the client side
  const filteredLessons = useMemo(() => {
    if (weekType === 'all') return lessons
    return lessons.filter(
      (l: Lesson) => l.week_type === 'all' || l.week_type === weekType
    )
  }, [lessons, weekType])

  return (
    <AppLayout>
      <div className="mx-auto max-w-7xl space-y-6 p-4 md:p-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-2">
            <Calendar className="h-6 w-6" />
            <div>
              <h1 className="text-2xl font-bold">{t('title')}</h1>
              <p className="text-sm text-muted-foreground">{t('description')}</p>
            </div>
          </div>
        </div>

        {/* Filters */}
        <ScheduleFilters
          filters={effectiveFilters}
          onFiltersChange={setFilters}
          semesters={semesters}
          groups={groups}
          classrooms={classrooms}
        />

        {/* Week type tabs */}
        <Tabs value={weekType} onValueChange={(v) => setWeekType(v as WeekType)}>
          <TabsList>
            {WEEK_TYPES.map((wt) => (
              <TabsTrigger key={wt} value={wt}>
                {t(`filters.${wt === 'all' ? 'allWeeks' : wt === 'odd' ? 'oddWeeks' : 'evenWeeks'}`)}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        {/* Content */}
        {isLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl bg-card border border-border p-8 text-center">
            <p className="text-destructive font-medium">{t('errors.loadFailed')}</p>
          </div>
        ) : filteredLessons.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Calendar className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <p className="text-lg font-medium">{t('empty')}</p>
          </div>
        ) : (
          <TimetableGrid
            lessons={filteredLessons}
            canEdit={userCanEdit}
          />
        )}
      </div>
    </AppLayout>
  )
}
