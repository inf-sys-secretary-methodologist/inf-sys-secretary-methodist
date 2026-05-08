'use client'

import { useTranslations } from 'next-intl'
import { ShieldCheck } from 'lucide-react'

import { AppLayout } from '@/components/layout'
import { AdminSettingsTabs } from '@/components/admin/AdminSettingsTabs'
import { MFASettingsCard } from '@/components/admin/MFASettingsCard'
import { useAuthCheck } from '@/hooks/useAuth'

export default function AdminSecurityPage() {
  const t = useTranslations('adminSettings')
  useAuthCheck()

  return (
    <AppLayout>
      <div className="mx-auto max-w-2xl space-y-6 p-4 md:p-6">
        <div className="flex items-center gap-2">
          <ShieldCheck className="h-6 w-6" />
          <h1 className="text-2xl font-bold">{t('security.title')}</h1>
        </div>
        <p className="text-sm text-muted-foreground">{t('security.subtitle')}</p>

        <AdminSettingsTabs />

        <MFASettingsCard mfaEnabled={false} />
      </div>
    </AppLayout>
  )
}
