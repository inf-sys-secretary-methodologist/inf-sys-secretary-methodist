'use client'

import { useEffect, useMemo, useState } from 'react'
import { useTranslations } from 'next-intl'
import { BookOpen, ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { WorkProgramCard } from '@/components/work-program/WorkProgramCard'
import { statusKey } from '@/components/work-program/status'
import { useWorkPrograms } from '@/hooks/useWorkPrograms'
import { useAuthCheck } from '@/hooks/useAuth'
import {
  WORK_PROGRAM_STATUSES,
  type WorkProgramListFilter,
  type WorkProgramStatus,
} from '@/types/workProgram'

// WorkProgramsPage — РПД list visible to ALL authenticated roles.
// Unlike /curriculum, students are NOT redirected: 273-ФЗ ст. 29
// mandates open access to approved work programs (ЭИОС), and the
// backend List use case forces status=approved for students and
// scopes teachers to their own. So the page just gates the fetch on
// auth resolution and lets the server narrow the rows.
export default function WorkProgramsPage() {
  const { isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('workProgram')

  const [statusFilter, setStatusFilter] = useState<WorkProgramStatus | ''>('')
  const [yearFilter, setYearFilter] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [offset, setOffset] = useState(0)
  const limit = 20

  // Reset to the first page whenever a filter changes so the user does
  // not land on an out-of-range page from a previous filter.
  useEffect(() => {
    setOffset(0)
  }, [statusFilter, yearFilter, specialty])

  const filter = useMemo<WorkProgramListFilter>(() => {
    const parsedYear = yearFilter.trim() ? Number(yearFilter.trim()) : undefined
    return {
      status: statusFilter || undefined,
      applicable_from_year:
        typeof parsedYear === 'number' && Number.isFinite(parsedYear) ? parsedYear : undefined,
      specialty_code: specialty.trim() || undefined,
      limit,
      offset,
    }
  }, [statusFilter, yearFilter, specialty, offset])

  // Fetch once auth has resolved — every role (incl. student) is allowed
  // to read; the server role-scopes the result.
  const enabled = !isLoading && isAuthenticated
  const { items, total, isLoading: listLoading, error } = useWorkPrograms(filter, { enabled })

  if (isLoading || !isAuthenticated) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto space-y-6">
        <header>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground">{t('description')}</p>
        </header>

        <section className="grid gap-3 sm:grid-cols-3">
          <div className="space-y-1.5">
            <Label htmlFor="filter-status">{t('filters.status')}</Label>
            <select
              id="filter-status"
              aria-label={t('filters.status')}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as WorkProgramStatus | '')}
            >
              <option value="">{t('filters.statusOptions.all')}</option>
              {WORK_PROGRAM_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {t(`filters.statusOptions.${statusKey(s)}`)}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-specialty">{t('filters.specialty')}</Label>
            <Input
              id="filter-specialty"
              value={specialty}
              onChange={(e) => setSpecialty(e.target.value)}
              placeholder={t('filters.specialtyPlaceholder')}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-year">{t('filters.year')}</Label>
            <Input
              id="filter-year"
              inputMode="numeric"
              value={yearFilter}
              onChange={(e) => setYearFilter(e.target.value.replace(/[^0-9]/g, ''))}
              placeholder={t('filters.yearPlaceholder')}
            />
          </div>
        </section>

        {listLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <BookOpen className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((wp) => (
              <WorkProgramCard key={wp.id} workProgram={wp} />
            ))}
          </div>
        )}

        {items.length > 0 && (
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              {t('countLabel', { shown: items.length, total })}
            </p>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(Math.max(0, offset - limit))}
                disabled={offset === 0}
              >
                <ChevronLeft className="h-4 w-4 mr-1" />
                {t('pagination.prev')}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setOffset(offset + limit)}
                disabled={offset + limit >= total}
              >
                {t('pagination.next')}
                <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </AppLayout>
  )
}
