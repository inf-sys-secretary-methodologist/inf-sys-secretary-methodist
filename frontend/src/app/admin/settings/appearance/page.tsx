'use client'

import { useTranslations } from 'next-intl'
import { Palette, Image, Pipette } from 'lucide-react'

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

export default function AdminAppearancePage() {
  const t = useTranslations('adminSettings')
  useAuthCheck()

  return (
    <AppLayout>
      <div className="mx-auto max-w-2xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <Palette className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('appearance.title')}</h1>
        </div>
        <p className="text-sm text-muted-foreground">{t('appearance.subtitle')}</p>

        <AdminSettingsTabs />

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Image className="h-5 w-5" />
              {t('appearance.brandTitle')}
            </CardTitle>
            <CardDescription>{t('appearance.brandDescription')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>{t('appearance.logoLabel')}</Label>
              <Input type="file" accept="image/*" />
            </div>
            <div className="space-y-2">
              <Label>{t('appearance.primaryColorLabel')}</Label>
              <div className="flex items-center gap-2">
                <Pipette className="h-4 w-4 text-muted-foreground" />
                <Input type="color" defaultValue="#6366f1" className="h-10 w-20 p-1" />
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t('appearance.faviconLabel')}</Label>
              <Input type="file" accept="image/x-icon,image/png" />
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  )
}
