'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { ChevronLeft, ChevronRight, FileText, Loader2, Plus } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { MinobrnaukiOrderCard } from '@/components/minobrnauki/MinobrnaukiOrderCard'
import { RecordMinobrnaukiOrderDialog } from '@/components/minobrnauki/RecordMinobrnaukiOrderDialog'
import { useMinobrnaukiOrders } from '@/hooks/useMinobrnaukiOrders'
import { useAuthCheck } from '@/hooks/useAuth'
import { canRecordMinobrnaukiOrder, canViewMinobrnaukiOrders } from '@/lib/auth/permissions'
import {
  MINOBRNAUKI_ORDER_CHANGE_SCOPES,
  type MinobrnaukiOrderChangeScope,
  type MinobrnaukiOrderListFilter,
} from '@/types/minobrnaukiOrder'

// MinobrnaukiOrdersPage — read-only browse of Минобрнауки orders (приказы)
// for non-student staff (system_admin / methodist / academic_secretary /
// teacher). Students are redirected to /forbidden because the backend list
// endpoint denies them (ADR-11 read gate); the page-shell guard skips a
// useless round-trip and FetchOpts.enabled=false prevents the SWR key from
// firing while the redirect is in flight. Mirrors CurriculumPage.
export default function MinobrnaukiOrdersPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('minobrnaukiOrder')

  const [scopeFilter, setScopeFilter] = useState<MinobrnaukiOrderChangeScope | ''>('')
  const [offset, setOffset] = useState(0)
  const [recordOpen, setRecordOpen] = useState(false)
  const limit = 20

  // Reset to the first page whenever the filter changes so the user does
  // not land on an out-of-range page from a previous filter.
  useEffect(() => {
    setOffset(0)
  }, [scopeFilter])

  const filter = useMemo<MinobrnaukiOrderListFilter>(
    () => ({
      change_scope: scopeFilter || undefined,
      limit,
      offset,
    }),
    [scopeFilter, offset]
  )

  // Fetch only for a confirmed viewer (non-student staff, mirror of the
  // backend ADR-11 read gate). A denied caller sees the redirect; pre-auth
  // sees the spinner — neither needs a fetch in flight.
  const canView = canViewMinobrnaukiOrders(user?.role)
  const canRecord = canRecordMinobrnaukiOrder(user?.role)
  const enabled = !isLoading && isAuthenticated && canView
  const {
    items,
    total,
    isLoading: listLoading,
    error,
    mutate,
  } = useMinobrnaukiOrders(filter, { enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && !canView) {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, canView, router])

  if (isLoading || !isAuthenticated || !canView) {
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
          {canRecord && (
            <Button onClick={() => setRecordOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              {t('recordButton')}
            </Button>
          )}
        </header>

        <section className="grid gap-3 sm:grid-cols-3">
          <div className="space-y-1.5">
            <Label htmlFor="filter-scope">{t('filters.changeScope')}</Label>
            <select
              id="filter-scope"
              aria-label={t('filters.changeScope')}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={scopeFilter}
              onChange={(e) => setScopeFilter(e.target.value as MinobrnaukiOrderChangeScope | '')}
            >
              <option value="">{t('filters.changeScopeOptions.all')}</option>
              {MINOBRNAUKI_ORDER_CHANGE_SCOPES.map((s) => (
                <option key={s} value={s}>
                  {t(`filters.changeScopeOptions.${s}`)}
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
            <FileText className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {items.map((order) => (
              <MinobrnaukiOrderCard key={order.id} order={order} />
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

        {canRecord && (
          <RecordMinobrnaukiOrderDialog
            open={recordOpen}
            onClose={() => setRecordOpen(false)}
            onCreated={() => mutate()}
          />
        )}
      </div>
    </AppLayout>
  )
}
