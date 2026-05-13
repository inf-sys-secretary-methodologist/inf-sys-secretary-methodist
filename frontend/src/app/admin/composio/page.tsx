'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

// AdminComposioPage — RED skeleton. The GREEN commit replaces the
// body with the Composio status card driven by useComposioConfig +
// i18n × 4 (mirror к admin/sentry single-card layout).
export default function AdminComposioPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-composio-page" />
    </AppLayout>
  )
}
