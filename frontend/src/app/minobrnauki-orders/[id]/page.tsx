'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { useTranslations } from 'next-intl'
import {
  ArrowLeft,
  Calendar,
  FileText,
  Hash,
  Loader2,
  ScrollText,
  Sparkles,
  User,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { GenerateOrderRevisionsDialog } from '@/components/minobrnauki/GenerateOrderRevisionsDialog'
import { useMinobrnaukiOrder } from '@/hooks/useMinobrnaukiOrders'
import { useAuthCheck } from '@/hooks/useAuth'
import { canViewMinobrnaukiOrders, canRecordMinobrnaukiOrder } from '@/lib/auth/permissions'
import { cn } from '@/lib/utils'

const SCOPE_STYLES = {
  minor: { bg: 'bg-sky-100 dark:bg-sky-950/40', text: 'text-sky-700 dark:text-sky-300' },
  major: { bg: 'bg-amber-100 dark:bg-amber-950/40', text: 'text-amber-700 dark:text-amber-300' },
} as const

// MinobrnaukiOrderDetailPage — single order view for non-student staff:
// metadata + the РПД the order affects (each a link to its detail page,
// so a methodist can jump straight to a touched work program). Mirrors
// the role-gate + param-parse pattern of the other detail pages. Recorders
// (canRecordMinobrnaukiOrder — methodist/secretary/admin) also get the
// "Сгенерировать правки" action that triggers AI bulk-revision over the
// affected set (ADR-12, slice 11c-3).
export default function MinobrnaukiOrderDetailPage() {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const t = useTranslations('minobrnaukiOrder')
  const [generateOpen, setGenerateOpen] = useState(false)

  const id = useMemo(() => {
    const raw = params?.id
    const parsed = typeof raw === 'string' ? Number(raw) : NaN
    return Number.isInteger(parsed) && parsed > 0 ? parsed : null
  }, [params])

  const { user, isAuthenticated, isLoading: authLoading } = useAuthCheck()
  const canView = canViewMinobrnaukiOrders(user?.role)
  const canGenerate = canRecordMinobrnaukiOrder(user?.role)

  const enabled = !authLoading && isAuthenticated && canView && id !== null
  const { order, isLoading: detailLoading, error, mutate } = useMinobrnaukiOrder(id, { enabled })

  useEffect(() => {
    if (!authLoading && isAuthenticated && !canView) {
      router.replace('/forbidden')
    }
  }, [authLoading, isAuthenticated, canView, router])

  if (authLoading || !isAuthenticated || !canView) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AppLayout>
    )
  }

  const scope = order ? SCOPE_STYLES[order.change_scope] : SCOPE_STYLES.minor

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-6">
        <Link
          href="/minobrnauki-orders"
          className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('detail.back')}
        </Link>

        {detailLoading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error || !order ? (
          <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center">
            <p className="font-medium text-destructive">{t('detail.notFound')}</p>
          </div>
        ) : (
          <>
            <header className="space-y-3">
              <div className="flex items-start justify-between gap-3">
                <h1 className="text-2xl font-bold leading-tight">{order.title}</h1>
                <div className="flex shrink-0 items-center gap-2">
                  {canGenerate && (
                    <Button size="sm" onClick={() => setGenerateOpen(true)}>
                      <Sparkles className="h-4 w-4 mr-2" />
                      {t('generateButton')}
                    </Button>
                  )}
                  <div
                    className={cn(
                      'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium',
                      scope.bg,
                      scope.text
                    )}
                  >
                    <FileText className="h-3.5 w-3.5" />
                    {t(`card.changeScope.${order.change_scope}`)}
                  </div>
                </div>
              </div>
              <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
                <span className="inline-flex items-center gap-1.5">
                  <Hash className="h-4 w-4" />
                  {order.order_number}
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <Calendar className="h-4 w-4" />
                  {t('detail.publishedAt', { date: order.published_at })}
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <User className="h-4 w-4" />
                  {t('detail.uploadedBy', { id: order.uploaded_by })}
                </span>
              </div>
            </header>

            {order.summary && (
              <section className="rounded-xl border border-border bg-card p-5">
                <h2 className="mb-2 text-sm font-semibold text-muted-foreground">
                  {t('detail.summary')}
                </h2>
                <p className="whitespace-pre-wrap text-sm">{order.summary}</p>
              </section>
            )}

            <section className="space-y-3">
              <h2 className="text-lg font-semibold">
                {t('detail.affectedTitle', { count: order.affected_work_program_ids.length })}
              </h2>
              {order.affected_work_program_ids.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t('detail.affectedEmpty')}</p>
              ) : (
                <ul className="grid gap-2 sm:grid-cols-2">
                  {order.affected_work_program_ids.map((wpId) => (
                    <li key={wpId}>
                      <Link
                        href={`/work-programs/${wpId}`}
                        className="group flex items-center gap-2 rounded-lg border border-border bg-card px-4 py-3 text-sm transition hover:border-primary/40 hover:shadow-sm"
                      >
                        <ScrollText className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
                        {t('detail.affectedItem', { id: wpId })}
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            {canGenerate && (
              <GenerateOrderRevisionsDialog
                orderId={order.id}
                open={generateOpen}
                onClose={() => setGenerateOpen(false)}
                onGenerated={() => mutate()}
              />
            )}
          </>
        )}
      </div>
    </AppLayout>
  )
}
