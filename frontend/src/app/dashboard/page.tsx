'use client'

import { useAuthCheck } from '@/hooks/useAuth'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { GlowingEffect } from '@/components/ui/glowing-effect'
import { NavBar } from '@/components/ui/tubelight-navbar'
import { FileText, Users, Calendar, TrendingUp, LayoutDashboard } from 'lucide-react'

export default function DashboardPage() {
  const { user, isLoading } = useAuthCheck()

  const navItems = [
    { name: 'Dashboard', url: '/dashboard', icon: LayoutDashboard },
    { name: 'Студенты', url: '/students', icon: Users },
    { name: 'Документы', url: '/documents', icon: FileText }
  ]

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
      <div className="fixed top-8 right-8 z-50 pointer-events-auto flex items-center gap-3" style={{ isolation: 'isolate' }}>
        <UserMenu />
        <ThemeToggleButton />
      </div>

      <div className="max-w-7xl mx-auto space-y-8">
        {/* Welcome Header */}
        <div className="text-center space-y-4 pt-24">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
            Добро пожаловать, {user?.name}!
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-300">
            <span className="font-semibold">{getRoleDisplayName(user?.role || '')}</span>
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[
            { icon: FileText, title: 'Документы', value: 24, trend: '+12%' },
            { icon: Users, title: 'Студенты', value: 156, trend: '+8%' },
            { icon: Calendar, title: 'Мероприятия', value: 8, trend: '+3%' },
            { icon: TrendingUp, title: 'Отчеты', value: 12, trend: '+15%' },
          ].map((stat, index) => (
            <div
              key={index}
              className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:scale-105 hover:shadow-xl"
            >
              <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
              <div className="relative z-10">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 dark:bg-white/10 text-gray-900 dark:text-white">
                    <stat.icon className="h-6 w-6" />
                  </div>
                  {stat.trend && (
                    <span className="text-xs px-2 py-1 rounded-full font-medium bg-green-500/20 text-green-600 dark:text-green-400 border border-green-500/50">
                      {stat.trend}
                    </span>
                  )}
                </div>
                <h3 className="text-sm font-medium mb-2 text-gray-600 dark:text-gray-400">{stat.title}</h3>
                <div className="text-3xl font-bold text-gray-900 dark:text-white">{stat.value}</div>
              </div>
            </div>
          ))}
        </div>

        {/* Quick Actions */}
        <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
          <div className="relative z-10 space-y-6">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">Быстрые действия</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Загрузить документ
              </button>
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Добавить студента
              </button>
              <button className="px-6 py-4 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg">
                Создать мероприятие
              </button>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect spread={40} glow={true} disabled={false} proximity={64} inactiveZone={0.01} borderWidth={3} />
          <div className="relative z-10 space-y-6">
            <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">Последние действия</h2>
            <div className="space-y-4">
              {[
                { action: 'Загружен документ', details: 'Отчет за октябрь 2024', time: '2 часа назад' },
                { action: 'Добавлен студент', details: 'Иванов Иван Иванович', time: '5 часов назад' },
                { action: 'Создано мероприятие', details: 'Конференция по IT', time: '1 день назад' },
              ].map((activity, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-4 rounded-lg bg-gray-50 dark:bg-white/5 hover:bg-gray-100 dark:hover:bg-white/10 transition-colors"
                >
                  <div>
                    <p className="font-medium text-gray-900 dark:text-white">{activity.action}</p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">{activity.details}</p>
                  </div>
                  <span className="text-xs text-gray-500 dark:text-gray-500">{activity.time}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function getRoleDisplayName(role: string): string {
  const roleMap: Record<string, string> = {
    system_admin: 'Администратор',
    methodist: 'Методист',
    academic_secretary: 'Секретарь',
    teacher: 'Преподаватель',
    student: 'Студент',
  }
  return roleMap[role] || role
}
