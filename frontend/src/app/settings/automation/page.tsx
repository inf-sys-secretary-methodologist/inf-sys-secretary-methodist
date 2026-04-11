'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Workflow, ExternalLink, CheckCircle2, XCircle, RefreshCw } from 'lucide-react'
import { AppLayout } from '@/components/layout'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface N8NStatus {
  healthy: boolean
  url: string
}

export default function AutomationSettingsPage() {
  const t = useTranslations('settings')
  const [status, setStatus] = useState<N8NStatus | null>(null)
  const [loading, setLoading] = useState(true)

  const n8nURL = process.env.NEXT_PUBLIC_N8N_URL || 'http://localhost:5678'

  const checkHealth = async () => {
    setLoading(true)
    try {
      await fetch(`${n8nURL}/healthz`, { mode: 'no-cors', signal: AbortSignal.timeout(5000) })
      setStatus({ healthy: true, url: n8nURL })
    } catch {
      setStatus({ healthy: false, url: n8nURL })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    checkHealth()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const workflows = [
    {
      name: 'Document Created → Telegram',
      description: 'Sends notification when a new document is created',
      status: 'ready',
    },
    {
      name: 'Student Absences → Curator Alert',
      description: 'Alerts curator when student has high risk score',
      status: 'ready',
    },
    {
      name: 'Deadline → Reminder',
      description: 'Daily reminders for approaching task deadlines',
      status: 'ready',
    },
  ]

  return (
    <AppLayout>
      <div className="container mx-auto max-w-4xl space-y-6 px-4 py-6 sm:px-6">
        <div className="flex items-center gap-3">
          <Workflow className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-2xl font-bold">{t('automation.title')}</h1>
            <p className="text-muted-foreground">{t('automation.description')}</p>
          </div>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>n8n Platform</CardTitle>
                <CardDescription>
                  Self-hosted workflow automation with 400+ integrations
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                {loading ? (
                  <Badge variant="outline">
                    <RefreshCw className="mr-1 h-3 w-3 animate-spin" />
                    Checking...
                  </Badge>
                ) : status?.healthy ? (
                  <Badge variant="default" className="bg-green-600">
                    <CheckCircle2 className="mr-1 h-3 w-3" />
                    Connected
                  </Badge>
                ) : (
                  <Badge variant="destructive">
                    <XCircle className="mr-1 h-3 w-3" />
                    Offline
                  </Badge>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-lg border p-4">
              <div>
                <p className="text-sm font-medium">n8n Dashboard</p>
                <p className="text-sm text-muted-foreground">{n8nURL}</p>
              </div>
              <Button variant="outline" size="sm" asChild>
                <a href={n8nURL} target="_blank" rel="noopener noreferrer">
                  Open n8n
                  <ExternalLink className="ml-2 h-3 w-3" />
                </a>
              </Button>
            </div>

            <Button variant="ghost" size="sm" onClick={checkHealth} disabled={loading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh Status
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Available Workflows</CardTitle>
            <CardDescription>
              Pre-configured automation workflows. Import them into n8n to activate.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {workflows.map((wf, i) => (
                <div key={i} className="flex items-center justify-between rounded-lg border p-4">
                  <div>
                    <p className="text-sm font-medium">{wf.name}</p>
                    <p className="text-sm text-muted-foreground">{wf.description}</p>
                  </div>
                  <Badge variant="outline">{wf.status}</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
