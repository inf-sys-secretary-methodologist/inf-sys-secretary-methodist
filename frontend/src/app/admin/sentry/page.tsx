'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

// AdminSentryPage — RED skeleton. The GREEN commit replaces the body
// with the actual status card + metadata grid driven by
// useSentryConfig + i18n × 4. The skeleton keeps the role-gate live
// so the redirect test can be authored ahead of the rendered UI.
export default function AdminSentryPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-sentry-page" />
    </AppLayout>
  )
}
