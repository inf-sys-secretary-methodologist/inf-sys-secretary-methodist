'use client'

import { useAuthCheck } from '@/hooks/useAuth'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { NavBar } from '@/components/ui/tubelight-navbar'
import { Megaphone } from 'lucide-react'
import { getAvailableNavItems } from '@/config/navigation'

export default function AnnouncementsPage() {
  const { user, isLoading } = useAuthCheck()

  // Get navigation items filtered by user role
  const navItems = getAvailableNavItems(user?.role)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background p-8">
      {/* Navigation Bar */}
      <NavBar items={navItems} />

      {/* Top Navigation */}
      <div
        className="fixed top-8 right-8 z-50 pointer-events-auto flex items-center gap-3"
        style={{ isolation: 'isolate' }}
      >
        <UserMenu />
        <ThemeToggleButton />
      </div>

      <div className="max-w-7xl mx-auto space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-4 pt-24">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">Объявления</h1>
          <p className="text-lg text-gray-600 dark:text-gray-300">
            Новости и важные объявления организации
          </p>
        </div>

        {/* Content Placeholder */}
        <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <div className="relative z-10 space-y-6 text-center">
            <Megaphone className="h-16 w-16 mx-auto text-gray-400" />
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">
              Раздел в разработке
            </h2>
            <p className="text-gray-600 dark:text-gray-400">
              Здесь будут отображаться объявления, новости и важные уведомления
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
