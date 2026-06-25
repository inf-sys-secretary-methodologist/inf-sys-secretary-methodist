'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { AlertTriangle, ChevronLeft, ChevronRight, Download, Loader2, Upload } from 'lucide-react'
import { toast } from 'sonner'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { StudentDebtCard } from '@/components/student-debts/StudentDebtCard'
import { ImportDebtsDialog } from '@/components/student-debts/ImportDebtsDialog'
import { statusKey } from '@/components/student-debts/status'
import { useStudentDebts } from '@/hooks/useStudentDebts'
import { studentDebtsApi } from '@/lib/api/studentDebts'
import { useAuthCheck } from '@/hooks/useAuth'
import { canManageStudentDebts } from '@/lib/auth/permissions'
import {
  STUDENT_DEBT_STATUSES,
  type StudentDebtStatus,
  type StudentDebtsFilter,
} from '@/types/studentDebts'

// StudentDebtsPage — the academic-debt registry. Staff (admin/methodist/
// secretary) read everything and may import/export + drive the resit
// lifecycle; a teacher reads the registry server-scoped to their own
// disciplines. A student is denied the registry endpoint, so the page
// redirects them to /student-debts/my (their own debts) instead of firing
// a guaranteed-403 fetch.
export default function StudentDebtsPage() {
  const { isAuthenticated, isLoading, user } = useAuthCheck()
  const t = useTranslations('studentDebts')
  const router = useRouter()

  const [groupFilter, setGroupFilter] = useState('')
  const [semesterFilter, setSemesterFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<StudentDebtStatus | ''>('')
  const [offset, setOffset] = useState(0)
  const [importOpen, setImportOpen] = useState(false)
  const [exporting, setExporting] = useState(false)
  const limit = 20

  const isStudent = user?.role === 'student'
  const canManage = canManageStudentDebts(user?.role)

  // Send students to their own-debts view — the registry endpoint denies
  // them (read_scope.go), so there is nothing to render here.
  useEffect(() => {
    if (!isLoading && isAuthenticated && isStudent) {
      router.replace('/student-debts/my')
    }
  }, [isLoading, isAuthenticated, isStudent, router])

  // Reset to the first page whenever a filter changes so the user does not
  // land on an out-of-range page from a previous filter.
  useEffect(() => {
    setOffset(0)
  }, [groupFilter, semesterFilter, statusFilter])

  const filter = useMemo<StudentDebtsFilter>(() => {
    const parsedSem = semesterFilter.trim() ? Number(semesterFilter.trim()) : undefined
    return {
      group_name: groupFilter.trim() || undefined,
      semester: typeof parsedSem === 'number' && Number.isFinite(parsedSem) ? parsedSem : undefined,
      status: statusFilter || undefined,
      limit,
      offset,
    }
  }, [groupFilter, semesterFilter, statusFilter, offset])

  // Fetch once auth resolves, but never for a student (they are redirected).
  const enabled = !isLoading && isAuthenticated && !isStudent
  const {
    items,
    total,
    isLoading: listLoading,
    error,
    mutate,
  } = useStudentDebts(filter, { enabled })

  const handleExport = async () => {
    if (exporting) return
    setExporting(true)
    try {
      const blob = await studentDebtsApi.export(filter)
      downloadBlob(blob, t('export.fileName'))
      toast.success(t('export.successToast'))
    } catch {
      toast.error(t('errors.forbidden'))
    } finally {
      setExporting(false)
    }
  }

  if (isLoading || !isAuthenticated || isStudent) {
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
          {canManage && (
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => setImportOpen(true)}>
                <Upload className="h-4 w-4 mr-2" />
                {t('importButton')}
              </Button>
              <Button variant="outline" onClick={handleExport} disabled={exporting}>
                {exporting ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Download className="h-4 w-4 mr-2" />
                )}
                {t('exportButton')}
              </Button>
            </div>
          )}
        </header>

        <section className="grid gap-3 sm:grid-cols-3">
          <div className="space-y-1.5">
            <Label htmlFor="filter-group">{t('filters.group')}</Label>
            <Input
              id="filter-group"
              value={groupFilter}
              onChange={(e) => setGroupFilter(e.target.value)}
              placeholder={t('filters.groupPlaceholder')}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-semester">{t('filters.semester')}</Label>
            <Input
              id="filter-semester"
              inputMode="numeric"
              value={semesterFilter}
              onChange={(e) => setSemesterFilter(e.target.value.replace(/[^0-9]/g, ''))}
              placeholder={t('filters.semesterPlaceholder')}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-status">{t('filters.status')}</Label>
            <select
              id="filter-status"
              aria-label={t('filters.status')}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as StudentDebtStatus | '')}
            >
              <option value="">{t('filters.statusOptions.all')}</option>
              {STUDENT_DEBT_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {t(`filters.statusOptions.${statusKey(s)}`)}
                </option>
              ))}
            </select>
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
            <AlertTriangle className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((debt) => (
              <StudentDebtCard key={debt.id} debt={debt} />
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

        <ImportDebtsDialog
          open={importOpen}
          onClose={() => setImportOpen(false)}
          onImported={() => mutate()}
        />
      </div>
    </AppLayout>
  )
}

// downloadBlob triggers a browser download of an in-memory blob. Guarded for
// the jsdom test environment where createObjectURL is absent.
function downloadBlob(blob: Blob, filename: string): void {
  if (typeof URL === 'undefined' || typeof URL.createObjectURL !== 'function') return
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
