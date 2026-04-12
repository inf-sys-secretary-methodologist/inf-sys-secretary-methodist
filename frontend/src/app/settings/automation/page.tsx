'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Workflow, ExternalLink, CheckCircle2, XCircle, RefreshCw } from 'lucide-react'
import { AppLayout } from '@/components/layout'
import { SettingsTabs } from '@/components/settings/SettingsTabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

export default function AutomationSettingsPage() {
  const t = useTranslations('settings.automation')
  const [healthy, setHealthy] = useState<boolean | null>(null)
  const [loading, setLoading] = useState(true)

  const n8nURL = process.env.NEXT_PUBLIC_N8N_URL || 'http://localhost:5678'

  const checkHealth = async () => {
    setLoading(true)
    try {
      await fetch(`${n8nURL}/healthz`, { mode: 'no-cors', signal: AbortSignal.timeout(5000) })
      setHealthy(true)
    } catch {
      setHealthy(false)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    checkHealth()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const workflows = [
    { nameKey: 'workflow1Name', descKey: 'workflow1Desc' },
    { nameKey: 'workflow2Name', descKey: 'workflow2Desc' },
    { nameKey: 'workflow3Name', descKey: 'workflow3Desc' },
  ]

  return (
    <AppLayout>
      <SettingsTabs />
      <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Workflow className="h-6 w-6 text-primary" />
        <div>
          <h1 className="text-2xl font-bold">{t('title')}</h1>
          <p className="text-muted-foreground">{t('description')}</p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{t('platform')}</CardTitle>
              <CardDescription>{t('platformDesc')}</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              {loading ? (
                <Badge variant="outline">
                  <RefreshCw className="mr-1 h-3 w-3 animate-spin" />
                  {t('checking')}
                </Badge>
              ) : healthy ? (
                <Badge variant="default" className="bg-green-600">
                  <CheckCircle2 className="mr-1 h-3 w-3" />
                  {t('connected')}
                </Badge>
              ) : (
                <Badge variant="destructive">
                  <XCircle className="mr-1 h-3 w-3" />
                  {t('offline')}
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div>
              <p className="text-sm font-medium">{t('dashboard')}</p>
              <p className="text-sm text-muted-foreground">{n8nURL}</p>
            </div>
            <Button variant="outline" size="sm" asChild>
              <a href={n8nURL} target="_blank" rel="noopener noreferrer">
                {t('openN8n')}
                <ExternalLink className="ml-2 h-3 w-3" />
              </a>
            </Button>
          </div>

          <Button variant="ghost" size="sm" onClick={checkHealth} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            {t('refreshStatus')}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('workflowsTitle')}</CardTitle>
          <CardDescription>{t('workflowsDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {workflows.map((wf, i) => (
              <div key={i} className="flex items-center justify-between rounded-lg border p-4">
                <div>
                  <p className="text-sm font-medium">{t(wf.nameKey)}</p>
                  <p className="text-sm text-muted-foreground">{t(wf.descKey)}</p>
                </div>
                <Badge variant="outline">{t('ready')}</Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
      </div>
    </AppLayout>
  )
}
