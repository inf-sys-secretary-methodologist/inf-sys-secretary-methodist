'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { BookMarked, Loader2, Plus } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { CurriculumCard } from '@/components/curriculum/CurriculumCard'
import { CreateCurriculumDialog } from '@/components/curriculum/CreateCurriculumDialog'
import { useCurricula } from '@/hooks/useCurricula'
import { useAuthCheck } from '@/hooks/useAuth'
import { canWriteCurriculum } from '@/lib/auth/permissions'
import {
  CURRICULUM_STATUSES,
  type CurriculumListFilter,
  type CurriculumStatus,
} from '@/types/curriculum'

// CurriculumPage — read-only list of curricula visible to non-student
// roles (methodist / system_admin / academic_secretary / teacher).
// Students are redirected to /forbidden because the read endpoint is
// gated by RequireNonStudent on the backend (v0.116.0); the
// page-shell guard skips a useless round-trip and the
// FetchOpts.enabled=false flag prevents the SWR key from firing while
// the redirect is in flight.
export default function CurriculumPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('curriculum')

  const [statusFilter, setStatusFilter] = useState<CurriculumStatus | ''>('')
  const [yearFilter, setYearFilter] = useState('')
  const [specialty, setSpecialty] = useState('')
  const [createOpen, setCreateOpen] = useState(false)

  const filter = useMemo<CurriculumListFilter>(() => {
    const parsedYear = yearFilter.trim() ? Number(yearFilter.trim()) : undefined
    return {
      status: statusFilter || undefined,
      year: typeof parsedYear === 'number' && Number.isFinite(parsedYear) ? parsedYear : undefined,
      specialty: specialty.trim() || undefined,
      limit: 100,
    }
  }, [statusFilter, yearFilter, specialty])

  // Fetch only when the caller is a confirmed non-student. Student
  // sees the redirect; pre-auth sees the spinner — neither needs a
  // fetch in flight (would 401 for student / fire pre-redirect).
  const enabled = !isLoading && isAuthenticated && user?.role !== 'student'
  const { items, total, isLoading: listLoading, error, mutate } = useCurricula(filter, { enabled })
  const canCreate = canWriteCurriculum(user?.role)

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role === 'student') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

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
        <header className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-muted-foreground">{t('description')}</p>
          </div>
          {canCreate && (
            <Button onClick={() => setCreateOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              {t('createButton')}
            </Button>
          )}
        </header>

        <section className="grid gap-3 sm:grid-cols-3">
          <div className="space-y-1.5">
            <Label htmlFor="filter-status">{t('filters.status')}</Label>
            <select
              id="filter-status"
              aria-label={t('filters.status')}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as CurriculumStatus | '')}
            >
              <option value="">{t('filters.statusOptions.all')}</option>
              {CURRICULUM_STATUSES.map((s) => {
                const key = s === 'pending_approval' ? 'pending' : s
                return (
                  <option key={s} value={s}>
                    {t(`filters.statusOptions.${key}`)}
                  </option>
                )
              })}
            </select>
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
          <div className="space-y-1.5">
            <Label htmlFor="filter-specialty">{t('filters.specialty')}</Label>
            <Input
              id="filter-specialty"
              value={specialty}
              onChange={(e) => setSpecialty(e.target.value)}
              placeholder={t('filters.specialtyPlaceholder')}
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
            <BookMarked className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((c) => (
              <CurriculumCard key={c.id} curriculum={c} />
            ))}
          </div>
        )}

        {items.length > 0 && (
          <p className="text-right text-sm text-muted-foreground">
            {t('countLabel', { shown: items.length, total })}
          </p>
        )}
      </div>

      {canCreate && (
        <CreateCurriculumDialog
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          onCreated={() => mutate()}
        />
      )}
    </AppLayout>
  )
}
