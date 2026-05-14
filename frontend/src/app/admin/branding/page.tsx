'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

import { AppLayout } from '@/components/layout'
import { useAuthCheck } from '@/hooks/useAuth'

// AdminBrandingPage — RED skeleton. GREEN ships the form,
// hook wiring, success / error banners and color pickers.
export default function AdminBrandingPage() {
  const router = useRouter()
  const { user, isAuthenticated, isLoading } = useAuthCheck()

  useEffect(() => {
    if (!isLoading && isAuthenticated && user?.role !== 'system_admin') {
      router.replace('/forbidden')
    }
  }, [isLoading, isAuthenticated, user, router])

  return (
    <AppLayout>
      <div data-testid="admin-branding-page" />
    </AppLayout>
  )
}
