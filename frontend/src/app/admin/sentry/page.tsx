'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Activity, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'
import { useSentryConfig } from '@/hooks/useSentryConfig'

// AdminSentryPage — admin-only read-only observability for the runtime
// Sentry integration. Mirrors /admin/backups read-only status card
// pattern. DSN is exposed as a boolean only (raw value is a secret
// even on an admin-gated endpoint).
export default function AdminSentryPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminSentry')

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const { config, isLoading: configLoading, error } = useSentryConfig({ enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-sentry-page" className="max-w-3xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <Activity className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {configLoading ? (
          <div data-testid="sentry-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="sentry-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : config ? (
          <SentryStatusCard config={config} />
        ) : null}
      </div>
    </AppLayout>
  )
}

function SentryStatusCard({
  config,
}: {
  config: {
    dsn_configured: boolean
    environment: string
    release: string
    traces_sample_rate: number
    tracing_enabled: boolean
  }
}) {
  const t = useTranslations('adminSentry')
  const isConfigured = config.dsn_configured
  return (
    <section
      data-testid="sentry-status-card"
      aria-label={t('status.sectionLabel')}
      className="rounded-xl border border-border bg-card p-5 space-y-4"
    >
      <div className="flex items-center gap-2">
        <h2 className="font-medium">{t('status.sectionLabel')}</h2>
        {isConfigured ? (
          <span
            data-testid="sentry-status-configured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 px-2 py-0.5 text-xs font-medium"
          >
            <CheckCircle2 className="h-3 w-3" /> {t('status.configured')}
          </span>
        ) : (
          <span
            data-testid="sentry-status-unconfigured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-muted text-muted-foreground px-2 py-0.5 text-xs font-medium"
          >
            <AlertCircle className="h-3 w-3" /> {t('status.unconfigured')}
          </span>
        )}
      </div>
      <dl className="grid grid-cols-2 gap-x-3 gap-y-2 text-sm">
        <dt className="text-muted-foreground">{t('fields.environment')}</dt>
        <dd data-testid="sentry-meta-environment" className="font-mono text-xs text-right">
          {config.environment}
        </dd>
        <dt className="text-muted-foreground">{t('fields.release')}</dt>
        <dd data-testid="sentry-meta-release" className="font-mono text-xs text-right">
          {config.release}
        </dd>
        <dt className="text-muted-foreground">{t('fields.tracesSampleRate')}</dt>
        <dd data-testid="sentry-meta-traces" className="font-mono text-xs text-right">
          {config.traces_sample_rate}
        </dd>
        <dt className="text-muted-foreground">{t('fields.tracingEnabled')}</dt>
        <dd data-testid="sentry-meta-tracing" className="text-right">
          {config.tracing_enabled ? t('fields.enabled') : t('fields.disabled')}
        </dd>
      </dl>
    </section>
  )
}
