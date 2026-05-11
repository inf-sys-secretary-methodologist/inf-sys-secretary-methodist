'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { ChevronLeft, ChevronRight, FileText, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAuthCheck } from '@/hooks/useAuth'
import { useAuditLogs } from '@/hooks/useAuditLogs'
import type { AuditLog, AuditLogFilter } from '@/types/audit'

// PAGE_SIZE matches the backend DefaultLimit constant. Hard-coded
// rather than fetched so the prev/next math (offset = (page-1) *
// pageSize) stays deterministic across the first render. The backend
// clamps anyway, so a drift only changes per-screen density.
const PAGE_SIZE = 50

// AdminAuditLogsPage — admin-only forensic timeline view of the
// audit_logs table. Mirror к /admin/curriculum/approve role-gate
// shape (single-role allowlist via useAuthCheck + replace('/forbidden')
// fallback), но table-based rather than card-based because audit
// rows are dense / tabular by nature.
export default function AdminAuditLogsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminAuditLogs')

  // Each filter dimension is its own controlled state so a partial
  // reset (e.g. clearing only `action`) does not stomp on the
  // remaining dimensions. Offset is split out so pagination clicks
  // don't reset filter typing.
  const [action, setAction] = useState('')
  const [resource, setResource] = useState('')
  const [userIdInput, setUserIdInput] = useState('')
  const [fromInput, setFromInput] = useState('')
  const [toInput, setToInput] = useState('')
  const [offset, setOffset] = useState(0)

  // userId / from / to are passed through to the SWR key only after
  // light validation so a half-typed value (e.g. user_id="4") does
  // not fire a request that the backend will 400. Empty / invalid
  // values are coerced to undefined and the SWR key omits them.
  const parsedUserId = useMemo(() => {
    if (!userIdInput.trim()) return undefined
    const n = Number(userIdInput.trim())
    return Number.isFinite(n) && n > 0 ? n : undefined
  }, [userIdInput])

  const filter = useMemo<AuditLogFilter>(
    () => ({
      action: action.trim() || undefined,
      resource: resource.trim() || undefined,
      user_id: parsedUserId,
      from: fromInput.trim() || undefined,
      to: toInput.trim() || undefined,
      limit: PAGE_SIZE,
      offset,
    }),
    [action, resource, parsedUserId, fromInput, toInput, offset]
  )

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const {
    items,
    pagination,
    total,
    isLoading: listLoading,
    error,
  } = useAuditLogs(filter, {
    enabled,
  })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  const handleReset = () => {
    setAction('')
    setResource('')
    setUserIdInput('')
    setFromInput('')
    setToInput('')
    setOffset(0)
  }

  const page = pagination?.page ?? 1
  const totalPages = pagination?.total_pages ?? 0
  const canPrev = offset > 0
  const canNext = totalPages > page

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <FileText className="h-7 w-7" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-bold">{t('title')}</h1>
              {total > 0 && (
                <span
                  data-testid="audit-logs-total"
                  className="inline-flex items-center justify-center rounded-full bg-muted px-2.5 py-0.5 text-sm font-semibold text-muted-foreground"
                >
                  {total}
                </span>
              )}
            </div>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {/* Filter bar */}
        <section
          aria-label="audit-log filters"
          className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-3 rounded-xl border border-border bg-card p-4"
        >
          <div className="space-y-1">
            <Label htmlFor="filter-action">{t('filters.action')}</Label>
            <Input
              id="filter-action"
              value={action}
              onChange={(e) => {
                setAction(e.target.value)
                setOffset(0)
              }}
              placeholder="curriculum.approved"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="filter-resource">{t('filters.resource')}</Label>
            <Input
              id="filter-resource"
              value={resource}
              onChange={(e) => {
                setResource(e.target.value)
                setOffset(0)
              }}
              placeholder="curriculum"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="filter-user-id">{t('filters.userId')}</Label>
            <Input
              id="filter-user-id"
              type="number"
              min={1}
              value={userIdInput}
              onChange={(e) => {
                setUserIdInput(e.target.value)
                setOffset(0)
              }}
              placeholder="42"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="filter-from">{t('filters.from')}</Label>
            <Input
              id="filter-from"
              type="datetime-local"
              value={fromInput}
              onChange={(e) => {
                // datetime-local emits "YYYY-MM-DDTHH:MM" — coerce
                // to RFC3339 with seconds + Z for the backend.
                setFromInput(e.target.value ? `${e.target.value}:00Z` : '')
                setOffset(0)
              }}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="filter-to">{t('filters.to')}</Label>
            <Input
              id="filter-to"
              type="datetime-local"
              value={toInput}
              onChange={(e) => {
                setToInput(e.target.value ? `${e.target.value}:00Z` : '')
                setOffset(0)
              }}
            />
          </div>
          <div className="flex items-end">
            <Button type="button" variant="outline" onClick={handleReset} className="w-full">
              {t('filters.reset')}
            </Button>
          </div>
        </section>

        {/* Body — loading / error / empty / table */}
        {listLoading ? (
          <div data-testid="audit-logs-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="audit-logs-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : items.length === 0 ? (
          <div
            data-testid="audit-logs-empty"
            className="flex flex-col items-center justify-center py-16 text-center"
          >
            <FileText className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="rounded-xl border border-border bg-card overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('columns.created_at')}</TableHead>
                  <TableHead>{t('columns.action')}</TableHead>
                  <TableHead>{t('columns.resource')}</TableHead>
                  <TableHead>{t('columns.actor')}</TableHead>
                  <TableHead>{t('columns.ip')}</TableHead>
                  <TableHead>{t('columns.correlation_id')}</TableHead>
                  <TableHead>{t('columns.fields')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((log) => (
                  <AuditLogRow key={log.id} log={log} />
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        {/* Pagination */}
        {items.length > 0 && totalPages > 1 && (
          <div
            data-testid="audit-logs-pagination"
            className="flex items-center justify-between gap-2"
          >
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!canPrev}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            >
              <ChevronLeft className="h-4 w-4 mr-2" />
              {t('pagination.prev')}
            </Button>
            <span className="text-sm text-muted-foreground">
              {t('pagination.pageOf', { page, totalPages })}
            </span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={!canNext}
              onClick={() => setOffset(offset + PAGE_SIZE)}
            >
              {t('pagination.next')}
              <ChevronRight className="h-4 w-4 ml-2" />
            </Button>
          </div>
        )}
      </div>
    </AppLayout>
  )
}

// AuditLogRow renders one row + a collapsible JSON Fields cell. The
// fields object is shown as monospace pretty-printed JSON behind a
// "Show JSON" toggle so the default table view stays scannable —
// dense JSON inline blows out row height on long correlation rows.
function AuditLogRow({ log }: { log: AuditLog }) {
  const t = useTranslations('adminAuditLogs')
  const [open, setOpen] = useState(false)
  const hasFields = log.fields && Object.keys(log.fields).length > 0

  return (
    <>
      <TableRow>
        <TableCell className="font-mono text-xs">{log.created_at}</TableCell>
        <TableCell className="font-medium">{log.action}</TableCell>
        <TableCell>{log.resource}</TableCell>
        <TableCell>{log.actor_user_id ?? '—'}</TableCell>
        <TableCell className="font-mono text-xs">{log.actor_ip ?? '—'}</TableCell>
        <TableCell className="font-mono text-xs">{log.correlation_id ?? '—'}</TableCell>
        <TableCell>
          {hasFields ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setOpen((prev) => !prev)}
              aria-expanded={open}
            >
              {open ? t('fields.hide') : t('fields.showJson')}
            </Button>
          ) : (
            <span className="text-muted-foreground">—</span>
          )}
        </TableCell>
      </TableRow>
      {open && hasFields && (
        <TableRow>
          <TableCell colSpan={7} className="bg-muted/30">
            <pre className="whitespace-pre-wrap break-all text-xs font-mono">
              {JSON.stringify(log.fields, null, 2)}
            </pre>
          </TableCell>
        </TableRow>
      )}
    </>
  )
}
