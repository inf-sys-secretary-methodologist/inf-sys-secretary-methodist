'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import {
  BookMarked,
  Calendar,
  CheckCircle2,
  ClipboardCheck,
  GraduationCap,
  Loader2,
  XCircle,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { useCurricula } from '@/hooks/useCurricula'
import { useAuthCheck } from '@/hooks/useAuth'
import { ApproveCurriculumDialog } from '@/components/curriculum/ApproveCurriculumDialog'
import { RejectCurriculumDialog } from '@/components/curriculum/RejectCurriculumDialog'
import { STATUS_STYLES, statusKey } from '@/components/curriculum/status'
import type { CurriculumListFilter } from '@/types/curriculum'
import { cn } from '@/lib/utils'

// AdminCurriculumApprovePage — admin-only queue of pending curricula
// awaiting approval. Mirror к /curriculum list page-shell guard но
// с inverse role gate (single-role allowlist: system_admin only).
// Filter pinned to status='pending_approval' — admin focus is the
// actionable queue. Approved/archived curricula visible через main
// /curriculum list (admin тоже sees them там).
export default function AdminCurriculumApprovePage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('curriculum')

  const [approveTargetId, setApproveTargetId] = useState<number | null>(null)
  const [rejectTargetId, setRejectTargetId] = useState<number | null>(null)

  const filter = useMemo<CurriculumListFilter>(
    () => ({ status: 'pending_approval', limit: 100 }),
    []
  )

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const { items, total, isLoading: listLoading, error, mutate } = useCurricula(filter, { enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
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
      <div className="max-w-5xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <ClipboardCheck className="h-7 w-7" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold">{t('adminApprove.title')}</h1>
              {total > 0 && (
                <span
                  data-testid="pending-count-badge"
                  className="inline-flex items-center justify-center rounded-full bg-amber-500/15 px-2.5 py-0.5 text-sm font-semibold text-amber-700 dark:text-amber-300"
                  aria-label={t('adminApprove.pendingCountLabel', { count: total })}
                >
                  {total}
                </span>
              )}
            </div>
            <p className="text-sm text-muted-foreground">{t('adminApprove.description')}</p>
          </div>
        </header>

        {listLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('adminApprove.loadFailed')}</p>
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <CheckCircle2 className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('adminApprove.empty.title')}</h3>
            <p className="text-muted-foreground">{t('adminApprove.empty.description')}</p>
          </div>
        ) : (
          <div className="space-y-3">
            {items.map((c) => {
              const styles = STATUS_STYLES[c.status]
              const Icon = styles.Icon
              return (
                <article key={c.id} className="rounded-xl border border-border bg-card p-4 sm:p-5">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <h3 className="font-semibold leading-tight">{c.title}</h3>
                      {c.description && (
                        <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                          {c.description}
                        </p>
                      )}
                      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-muted-foreground">
                        <span className="inline-flex items-center gap-1.5">
                          <BookMarked className="h-3.5 w-3.5" />
                          {c.code}
                        </span>
                        <span className="inline-flex items-center gap-1.5">
                          <GraduationCap className="h-3.5 w-3.5" />
                          {c.specialty}
                        </span>
                        <span className="inline-flex items-center gap-1.5">
                          <Calendar className="h-3.5 w-3.5" />
                          {c.year}
                        </span>
                      </div>
                    </div>
                    <div
                      className={cn(
                        'inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
                        styles.bg,
                        styles.text
                      )}
                    >
                      <Icon className="h-3.5 w-3.5" />
                      {t(`card.status.${statusKey(c.status)}`)}
                    </div>
                  </div>

                  <div className="mt-4 flex flex-wrap justify-end gap-2">
                    <Button variant="outline" size="sm" onClick={() => setRejectTargetId(c.id)}>
                      <XCircle className="h-4 w-4 mr-2" />
                      {t('adminApprove.actions.reject')}
                    </Button>
                    <Button size="sm" onClick={() => setApproveTargetId(c.id)}>
                      <CheckCircle2 className="h-4 w-4 mr-2" />
                      {t('adminApprove.actions.approve')}
                    </Button>
                  </div>
                </article>
              )
            })}
          </div>
        )}

        {/*
          Render dialogs only when a target id is set. Avoids passing
          a sentinel id (0) into the dialog when closed — defensive
          against transient confirm fires during open=true→false
          transition, even though the double-click guard makes that
          near-impossible in practice.
        */}
        {approveTargetId !== null && (
          <ApproveCurriculumDialog
            curriculumId={approveTargetId}
            open={true}
            onClose={() => setApproveTargetId(null)}
            onApproved={() => mutate()}
          />
        )}
        {rejectTargetId !== null && (
          <RejectCurriculumDialog
            curriculumId={rejectTargetId}
            open={true}
            onClose={() => setRejectTargetId(null)}
            onRejected={() => mutate()}
          />
        )}
      </div>
    </AppLayout>
  )
}
