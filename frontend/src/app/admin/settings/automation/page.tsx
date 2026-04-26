'use client'

import { useTranslations } from 'next-intl'
import { Workflow } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { AdminSettingsTabs } from '@/components/admin/AdminSettingsTabs'
import { useAuthCheck } from '@/hooks/useAuth'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export default function AdminAutomationPage() {
  const t = useTranslations('adminSettings')
  useAuthCheck()

  return (
    <AppLayout>
      <div className="mx-auto max-w-2xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <Workflow className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('automation.title')}</h1>
        </div>
        <p className="text-sm text-muted-foreground">{t('automation.subtitle')}</p>

        <AdminSettingsTabs />

        <Card>
          <CardHeader>
            <CardTitle>{t('automation.workflowsTitle')}</CardTitle>
            <CardDescription>{t('automation.workflowsDescription')}</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3 text-sm">
              <li className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Document notifications
              </li>
              <li className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Absence alerts
              </li>
              <li className="flex items-center gap-2">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Deadline reminders
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
