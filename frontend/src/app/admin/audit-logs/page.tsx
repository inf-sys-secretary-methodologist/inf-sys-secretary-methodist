'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'

// RED stub — full implementation lands in the matching GREEN commit.
// Only the role guard is wired so the redirect contract test passes;
// every other test must observe the stub state.
export default function AdminAuditLogsPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="audit-logs-stub">RED stub: audit logs page not yet implemented</div>
    </AppLayout>
  )
}
