'use client'

import { ReactNode } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { getAvailableNavItems } from '@/config/navigation'
import { AppHeader } from './AppHeader'
import { InstallPrompt } from '@/components/pwa/install-prompt'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const { user, isLoading } = useAuthCheck()

  // Get navigation items filtered by user role
  const navItems = getAvailableNavItems(user?.role)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen">
      <AppHeader items={navItems} />
      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 sm:py-8 lg:px-8 pb-16">{children}</main>
      <InstallPrompt />
    </div>
  )
}
