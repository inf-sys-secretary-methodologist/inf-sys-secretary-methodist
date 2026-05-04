'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { GraduationCap, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AssignmentCard } from '@/components/assignments/AssignmentCard'
import { useAssignments } from '@/hooks/useAssignments'
import { useAuthCheck } from '@/hooks/useAuth'
import type { AssignmentListFilter } from '@/types/assignments'

// AssignmentsPage — read-only list. Students are redirected to
// /forbidden because the grading flow is teacher/methodist/secretary/
// admin-only by design (the backend enforces the same rule in
// RequireNonStudent middleware; this guard skips a useless round-trip).
export default function AssignmentsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('assignments')

  const [subject, setSubject] = useState('')
  const [group, setGroup] = useState('')

  const filter = useMemo<AssignmentListFilter>(
    () => ({
      subject: subject.trim() || undefined,
      group_name: group.trim() || undefined,
      page_size: 100,
    }),
    [subject, group]
  )

  const { items, total, isLoading: listLoading, error } = useAssignments(filter)

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
        <header>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground">{t('description')}</p>
        </header>

        <section className="grid gap-3 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="filter-subject">{t('filters.subject')}</Label>
            <Input
              id="filter-subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder={t('filters.subjectPlaceholder')}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="filter-group">{t('filters.group')}</Label>
            <Input
              id="filter-group"
              value={group}
              onChange={(e) => setGroup(e.target.value)}
              placeholder={t('filters.groupPlaceholder')}
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
            <GraduationCap className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((a) => (
              <AssignmentCard key={a.id} assignment={a} />
            ))}
          </div>
        )}

        {items.length > 0 && (
          <p className="text-right text-sm text-muted-foreground">
            {t('countLabel', { shown: items.length, total })}
          </p>
        )}
      </div>
    </AppLayout>
  )
}
