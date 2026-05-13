'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

// AdminUsersPage — RED skeleton. The GREEN commit replaces the body
// with the user list + filters + pagination, mirror к /admin/audit-logs
// shape. Skeleton keeps the role gate live so the redirect test can
// be authored ahead of the rendered table.
export default function AdminUsersPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-users-page" />
    </AppLayout>
  )
}
