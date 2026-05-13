'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { AlertCircle, Bell, CheckCircle2, Loader2, Plug, Workflow } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'
import { useIntegrationsConfig } from '@/hooks/useIntegrationsConfig'
import type { N8NConfig, VAPIDConfig } from '@/types/integrations'

// AdminIntegrationsPage — admin-only read-only view of the WebPush
// (VAPID) + n8n runtime configuration. Two status cards side-by-side
// (stack on mobile). DSN-style "configured boolean" pattern mirrors
// /admin/sentry; VAPID public key + subject + n8n webhook URL are
// non-secret operational fields and surface verbatim.
export default function AdminIntegrationsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminIntegrations')

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const { config, isLoading: configLoading, error } = useIntegrationsConfig({ enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-integrations-page" className="max-w-5xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <Plug className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {configLoading ? (
          <div
            data-testid="integrations-loading"
            className="flex items-center justify-center py-16"
          >
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="integrations-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : config ? (
          <section className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <VAPIDCard config={config.vapid} />
            <N8NCard config={config.n8n} />
          </section>
        ) : null}
      </div>
    </AppLayout>
  )
}

function VAPIDCard({ config }: { config: VAPIDConfig }) {
  const t = useTranslations('adminIntegrations.vapid')
  const ok = config.configured
  return (
    <div
      data-testid="vapid-status-card"
      className="rounded-xl border border-border bg-card p-5 space-y-3"
    >
      <div className="flex items-center gap-2">
        <Bell className="h-5 w-5" />
        <h2 className="font-medium">{t('sectionLabel')}</h2>
        {ok ? (
          <span
            data-testid="vapid-status-configured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 px-2 py-0.5 text-xs font-medium"
          >
            <CheckCircle2 className="h-3 w-3" /> {t('configured')}
          </span>
        ) : (
          <span
            data-testid="vapid-status-unconfigured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-muted text-muted-foreground px-2 py-0.5 text-xs font-medium"
          >
            <AlertCircle className="h-3 w-3" /> {t('unconfigured')}
          </span>
        )}
      </div>
      <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-2 text-sm">
        <dt className="text-muted-foreground">{t('publicKey')}</dt>
        <dd data-testid="vapid-public-key" className="font-mono text-xs break-all text-right">
          {config.public_key || '—'}
        </dd>
        <dt className="text-muted-foreground">{t('subject')}</dt>
        <dd data-testid="vapid-subject" className="font-mono text-xs text-right">
          {config.subject || '—'}
        </dd>
      </dl>
    </div>
  )
}

function N8NCard({ config }: { config: N8NConfig }) {
  const t = useTranslations('adminIntegrations.n8n')
  const ok = config.enabled
  return (
    <div
      data-testid="n8n-status-card"
      className="rounded-xl border border-border bg-card p-5 space-y-3"
    >
      <div className="flex items-center gap-2">
        <Workflow className="h-5 w-5" />
        <h2 className="font-medium">{t('sectionLabel')}</h2>
        {ok ? (
          <span
            data-testid="n8n-status-enabled"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 px-2 py-0.5 text-xs font-medium"
          >
            <CheckCircle2 className="h-3 w-3" /> {t('enabled')}
          </span>
        ) : (
          <span
            data-testid="n8n-status-disabled"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-muted text-muted-foreground px-2 py-0.5 text-xs font-medium"
          >
            <AlertCircle className="h-3 w-3" /> {t('disabled')}
          </span>
        )}
      </div>
      <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-2 text-sm">
        <dt className="text-muted-foreground">{t('webhookUrl')}</dt>
        <dd data-testid="n8n-webhook-url" className="font-mono text-xs break-all text-right">
          {config.webhook_url || '—'}
        </dd>
      </dl>
    </div>
  )
}
