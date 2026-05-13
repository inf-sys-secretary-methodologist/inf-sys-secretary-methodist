'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

// AdminIntegrationsPage — RED skeleton. The GREEN commit replaces
// the body with the VAPID + n8n status cards driven by
// useIntegrationsConfig + i18n × 4.
export default function AdminIntegrationsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-integrations-page" />
    </AppLayout>
  )
}
