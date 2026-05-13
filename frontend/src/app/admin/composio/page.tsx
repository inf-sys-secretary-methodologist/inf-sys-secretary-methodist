'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { AlertCircle, Bot, CheckCircle2, Loader2 } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'
import { useComposioConfig } from '@/hooks/useComposioConfig'
import type { ComposioConfig } from '@/types/composio'

// AdminComposioPage — admin-only read-only view of the runtime
// Composio integration state. Single-card layout (mirror к
// /admin/sentry — one service, one card). Only booleans surface:
// the API key is a signing secret; entity ID and MCP config ID
// are opaque platform identifiers (per VAPID privacy precedent).
export default function AdminComposioPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()
  const t = useTranslations('adminComposio')

  const enabled = !isLoading && isAuthenticated && user?.role === 'system_admin'
  const { config, isLoading: configLoading, error } = useComposioConfig({ enabled })

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-composio-page" className="max-w-3xl mx-auto space-y-6">
        <header className="flex items-center gap-3">
          <Bot className="h-7 w-7" />
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{t('title')}</h1>
            <p className="text-sm text-muted-foreground">{t('description')}</p>
          </div>
        </header>

        {configLoading ? (
          <div data-testid="composio-loading" className="flex items-center justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : error ? (
          <div
            data-testid="composio-error"
            className="rounded-xl border border-destructive/30 bg-destructive/5 p-6 text-center"
          >
            <p className="font-medium text-destructive">{t('loadFailed')}</p>
          </div>
        ) : config ? (
          <ComposioStatusCard config={config} />
        ) : null}
      </div>
    </AppLayout>
  )
}

function ComposioStatusCard({ config }: { config: ComposioConfig }) {
  const t = useTranslations('adminComposio')
  return (
    <section
      data-testid="composio-status-card"
      aria-label={t('status.sectionLabel')}
      className="rounded-xl border border-border bg-card p-5 space-y-4"
    >
      <div className="flex items-center gap-2">
        <h2 className="font-medium">{t('status.sectionLabel')}</h2>
        {config.configured ? (
          <span
            data-testid="composio-status-configured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 px-2 py-0.5 text-xs font-medium"
          >
            <CheckCircle2 className="h-3 w-3" /> {t('status.configured')}
          </span>
        ) : (
          <span
            data-testid="composio-status-unconfigured"
            className="ml-auto inline-flex items-center gap-1 rounded-full bg-muted text-muted-foreground px-2 py-0.5 text-xs font-medium"
          >
            <AlertCircle className="h-3 w-3" /> {t('status.unconfigured')}
          </span>
        )}
      </div>
      <dl className="grid grid-cols-2 gap-x-3 gap-y-2 text-sm">
        <dt className="text-muted-foreground">{t('fields.apiKey')}</dt>
        <dd data-testid="composio-field-api-key" className="text-right">
          {config.api_key_configured ? t('fields.set') : t('fields.unset')}
        </dd>
        <dt className="text-muted-foreground">{t('fields.entityID')}</dt>
        <dd data-testid="composio-field-entity-id" className="text-right">
          {config.entity_id_set ? t('fields.set') : t('fields.unset')}
        </dd>
        <dt className="text-muted-foreground">{t('fields.mcpConfigID')}</dt>
        <dd data-testid="composio-field-mcp-config-id" className="text-right">
          {config.mcp_config_id_set ? t('fields.set') : t('fields.unset')}
        </dd>
      </dl>
    </section>
  )
}
