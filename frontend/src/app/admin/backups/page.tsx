'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import {
  AlertCircle,
  CheckCircle2,
  Database,
  Download,
  HardDrive,
  Loader2,
  Server,
} from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAuthCheck } from '@/hooks/useAuth'
import { useBackups } from '@/hooks/useBackups'
import { formatFileSize } from '@/lib/utils'
import type { BackupFile, TypeMetrics, RemoteSyncMetrics } from '@/types/backup'

// AdminBackupsPage — admin-only read-only observability surface for
// the /backup sidecar. Lists artifacts produced by pg_dump + MinIO
// tarball jobs and surfaces the sidecar's Prometheus textfile
// metrics. Mirrors /admin/audit-logs role-gate + layout shape.
export default function AdminBackupsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminBackups')

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const { files, metrics, isLoading: listLoading, error } = useBackups({ enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-backups-page" className="max-w-7xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <HardDrive className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {/* Metrics tile — postgres + minio cards stack on mobile, side
            by side on md+. Each card surfaces an "ok / failed / no data"
            banner driven by last_run_success + last_run_at. */}
        <section
          aria-label={t('metrics.sectionLabel')}
          className="grid grid-cols-1 md:grid-cols-2 gap-4"
        >
          <MetricsCard
            testId="backup-metrics-postgres"
            icon={<Database className="h-5 w-5" />}
            title={t('metrics.postgres')}
            metrics={metrics?.postgres ?? null}
          />
          <MetricsCard
            testId="backup-metrics-minio"
            icon={<Server className="h-5 w-5" />}
            title={t('metrics.minio')}
            metrics={metrics?.minio ?? null}
          />
        </section>

        {metrics?.remote_sync && <RemoteSyncBanner metrics={metrics.remote_sync} />}

        {/* Body */}
        {listLoading ? (
          <div data-testid="backups-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="backups-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : files.length === 0 ? (
          <div
            data-testid="backups-empty"
            className="flex flex-col items-center justify-center py-16 text-center"
          >
            <HardDrive className="h-16 w-16 text-muted-foreground/30 mb-4" />
            <h3 className="text-lg font-medium">{t('empty.title')}</h3>
            <p className="text-muted-foreground">{t('empty.description')}</p>
          </div>
        ) : (
          <div className="rounded-xl border border-border bg-card overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('columns.name')}</TableHead>
                  <TableHead>{t('columns.type')}</TableHead>
                  <TableHead>{t('columns.size')}</TableHead>
                  <TableHead>{t('columns.modifiedAt')}</TableHead>
                  <TableHead>{t('columns.encryption')}</TableHead>
                  <TableHead className="text-right">{t('columns.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {files.map((file) => (
                  <BackupFileRow key={file.name} file={file} />
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </AppLayout>
  )
}

function MetricsCard({
  testId,
  icon,
  title,
  metrics,
}: {
  testId: string
  icon: React.ReactNode
  title: string
  metrics: TypeMetrics | null
}) {
  const t = useTranslations('adminBackups.metrics')
  const hasData = metrics !== null && metrics.last_run_at > 0
  const ok = hasData && metrics.last_run_success
  return (
    <div data-testid={testId} className="rounded-xl border border-border bg-card p-4 space-y-2">
      <div className="flex items-center gap-2">
        {icon}
        <h2 className="font-medium">{title}</h2>
        <span
          className={`ml-auto inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
            !hasData
              ? 'bg-muted text-muted-foreground'
              : ok
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
                : 'bg-destructive/15 text-destructive'
          }`}
        >
          {!hasData ? (
            t('never')
          ) : ok ? (
            <>
              <CheckCircle2 className="h-3 w-3" /> {t('ok')}
            </>
          ) : (
            <>
              <AlertCircle className="h-3 w-3" /> {t('failed')}
            </>
          )}
        </span>
      </div>
      {hasData && (
        <dl className="grid grid-cols-2 gap-x-3 gap-y-1 text-sm">
          <dt className="text-muted-foreground">{t('lastRun')}</dt>
          <dd className="font-mono text-xs text-right">
            {new Date(metrics.last_run_at * 1000).toISOString()}
          </dd>
          <dt className="text-muted-foreground">{t('age')}</dt>
          <dd className="text-right">{formatDuration(metrics.age_seconds)}</dd>
          <dt className="text-muted-foreground">{t('duration')}</dt>
          <dd className="text-right">{metrics.duration_seconds}s</dd>
          <dt className="text-muted-foreground">{t('sizeBytes')}</dt>
          <dd className="text-right">{formatFileSize(metrics.size_bytes)}</dd>
          <dt className="text-muted-foreground">{t('totalCount')}</dt>
          <dd className="text-right">
            {metrics.success_count}/{metrics.total_count} (
            <span className="text-destructive">{metrics.failure_count}</span>)
          </dd>
        </dl>
      )}
    </div>
  )
}

function RemoteSyncBanner({ metrics }: { metrics: RemoteSyncMetrics }) {
  const t = useTranslations('adminBackups.remoteSync')
  const ok = metrics.last_run_success
  return (
    <div className="flex items-center gap-3 rounded-xl border border-border bg-card p-3 text-sm">
      {ok ? (
        <CheckCircle2 className="h-5 w-5 text-green-600" />
      ) : (
        <AlertCircle className="h-5 w-5 text-destructive" />
      )}
      <span className="font-medium">{t('title')}</span>
      <span className="text-muted-foreground">
        {ok ? t('ok') : t('failed')} · {new Date(metrics.last_run_at * 1000).toISOString()}
      </span>
      <span className="ml-auto text-muted-foreground">
        {metrics.success_count}/{metrics.total_count}
      </span>
    </div>
  )
}

function BackupFileRow({ file }: { file: BackupFile }) {
  const t = useTranslations('adminBackups')
  const downloadUrl = `/api/admin/backups/${file.type}/${encodeURIComponent(file.name)}/download`
  return (
    <TableRow>
      <TableCell className="font-mono text-xs">{file.name}</TableCell>
      <TableCell>
        <span className="inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
          {t(`types.${file.type}`)}
        </span>
      </TableCell>
      <TableCell>{formatFileSize(file.size)}</TableCell>
      <TableCell className="font-mono text-xs">
        {new Date(file.modified_at * 1000).toISOString()}
      </TableCell>
      <TableCell>
        <span
          data-testid={`backup-encryption-${file.name}`}
          className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
            file.encryption === ''
              ? 'bg-muted text-muted-foreground'
              : 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300'
          }`}
          title={file.encryption !== '' ? t('encryption.tooltip') : undefined}
        >
          {file.encryption === '' ? t('encryption.none') : t(`encryption.${file.encryption}`)}
        </span>
      </TableCell>
      <TableCell className="text-right">
        <Button asChild variant="ghost" size="sm">
          <a href={downloadUrl} download={file.name}>
            <Download className="h-4 w-4 mr-2" />
            {t('actions.download')}
          </a>
        </Button>
      </TableCell>
    </TableRow>
  )
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  return `${Math.floor(seconds / 86400)}d`
}
