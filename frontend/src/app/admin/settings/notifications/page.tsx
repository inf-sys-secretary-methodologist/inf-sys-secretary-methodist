'use client'

import { useTranslations } from 'next-intl'
import { Bell, Mail, Smartphone, MessageCircle } from 'lucide-react'

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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function AdminNotificationsPage() {
  const t = useTranslations('adminSettings')
  useAuthCheck()

  return (
    <AppLayout>
      <div className="mx-auto max-w-2xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <Bell className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('notifications.title')}</h1>
        </div>
        <p className="text-sm text-muted-foreground">{t('notifications.subtitle')}</p>

        <AdminSettingsTabs />

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              {t('notifications.smtpTitle')}
            </CardTitle>
            <CardDescription>{t('notifications.smtpDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t('notifications.smtpHost')}</Label>
              <Input placeholder="smtp.example.com" />
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.smtpPort')}</Label>
              <Input type="number" placeholder="587" />
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.smtpUser')}</Label>
              <Input placeholder="noreply@example.com" />
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.smtpPassword')}</Label>
              <Input type="password" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Smartphone className="h-5 w-5" />
              {t('notifications.pushTitle')}
            </CardTitle>
            <CardDescription>{t('notifications.pushDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>{t('notifications.vapidPublicKey')}</Label>
              <Input placeholder="BEl62i..." />
            </div>
            <div className="space-y-2">
              <Label>{t('notifications.vapidPrivateKey')}</Label>
              <Input type="password" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MessageCircle className="h-5 w-5" />
              {t('notifications.telegramTitle')}
            </CardTitle>
            <CardDescription>{t('notifications.telegramDescription')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <Label>{t('notifications.botToken')}</Label>
              <Input type="password" placeholder="123456:ABC-DEF..." />
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
