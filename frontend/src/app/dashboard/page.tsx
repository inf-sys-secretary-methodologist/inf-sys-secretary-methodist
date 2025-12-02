'use client'

import { useState } from 'react'
import { useAuthCheck } from '@/hooks/useAuth'
import { useDashboardStats, useDashboardTrends, useDashboardActivity } from '@/hooks/useDashboard'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggleButton } from '@/components/theme-toggle-button'
import { GlowingEffect } from '@/components/ui/glowing-effect'
import { NavBar } from '@/components/ui/tubelight-navbar'
import { StatsCard, TrendChart, ActivityFeed, ExportButton } from '@/components/dashboard'
import { FileText, Users, Calendar, TrendingUp, ClipboardList } from 'lucide-react'
import { getAvailableNavItems } from '@/config/navigation'

type Period = 'week' | 'month' | 'quarter' | 'year'

const periodLabels: Record<Period, string> = {
  week: 'неделю',
  month: 'месяц',
  quarter: 'квартал',
  year: 'год',
}

export default function DashboardPage() {
  const { user, isLoading: authLoading } = useAuthCheck()
  const [period, setPeriod] = useState<Period>('month')

  // Fetch dashboard data
  const { stats, isLoading: statsLoading } = useDashboardStats(period)
  const { trends, isLoading: trendsLoading } = useDashboardTrends(period)
  const { activities, isLoading: activityLoading } = useDashboardActivity(10)

  // Get navigation items filtered by user role
  const navItems = getAvailableNavItems(user?.role)

  if (authLoading) {
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
        {/* Welcome Header */}
        <div className="text-center space-y-4 pt-24">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
            Добро пожаловать, {user?.name}!
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-300">
            <span className="font-semibold">{getRoleDisplayName(user?.role || '')}</span>
          </p>
        </div>

        {/* Period Selector & Export */}
        <div className="flex items-center justify-between">
          <div className="relative overflow-hidden rounded-xl p-1 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect
              spread={40}
              glow={true}
              disabled={false}
              proximity={64}
              inactiveZone={0.01}
              borderWidth={2}
            />
            <div className="relative z-10 flex gap-1">
              {(['week', 'month', 'quarter', 'year'] as Period[]).map((p) => (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                    period === p
                      ? 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/10'
                  }`}
                >
                  {p === 'week' && 'Неделя'}
                  {p === 'month' && 'Месяц'}
                  {p === 'quarter' && 'Квартал'}
                  {p === 'year' && 'Год'}
                </button>
              ))}
            </div>
          </div>

          <ExportButton />
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
          {statsLoading ? (
            // Loading skeletons
            Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse"
              >
                <div className="h-12 w-12 bg-gray-200 dark:bg-gray-700 rounded-lg mb-4" />
                <div className="h-4 w-20 bg-gray-200 dark:bg-gray-700 rounded mb-2" />
                <div className="h-8 w-16 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
            ))
          ) : (
            <>
              <StatsCard
                icon={FileText}
                title="Документы"
                value={stats?.documents.total || 0}
                change={stats?.documents.change || 0}
                period={periodLabels[period]}
              />
              <StatsCard
                icon={Users}
                title="Студенты"
                value={stats?.students.total || 0}
                change={stats?.students.change || 0}
                period={periodLabels[period]}
              />
              <StatsCard
                icon={Calendar}
                title="Мероприятия"
                value={stats?.events.total || 0}
                change={stats?.events.change || 0}
                period={periodLabels[period]}
              />
              <StatsCard
                icon={TrendingUp}
                title="Отчеты"
                value={stats?.reports.total || 0}
                change={stats?.reports.change || 0}
                period={periodLabels[period]}
              />
              <StatsCard
                icon={ClipboardList}
                title="Задачи"
                value={stats?.tasks.total || 0}
                change={stats?.tasks.change || 0}
                period={periodLabels[period]}
              />
            </>
          )}
        </div>

        {/* Charts Row */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {trendsLoading ? (
            <>
              <div className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                <div className="h-64 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
              <div className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                <div className="h-64 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
            </>
          ) : (
            <>
              <TrendChart
                title="Документы и отчеты"
                datasets={[
                  {
                    name: 'Документы',
                    data: trends?.documents_trend || [],
                    color: '#3b82f6',
                  },
                  {
                    name: 'Отчеты',
                    data: trends?.reports_trend || [],
                    color: '#10b981',
                  },
                ]}
              />
              <TrendChart
                title="Задачи и мероприятия"
                datasets={[
                  {
                    name: 'Задачи',
                    data: trends?.tasks_trend || [],
                    color: '#f59e0b',
                  },
                  {
                    name: 'Мероприятия',
                    data: trends?.events_trend || [],
                    color: '#8b5cf6',
                  },
                ]}
              />
            </>
          )}
        </div>

        {/* Quick Actions & Activity */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Quick Actions */}
          <div className="relative overflow-hidden rounded-2xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect
              spread={40}
              glow={true}
              disabled={false}
              proximity={64}
              inactiveZone={0.01}
              borderWidth={3}
            />
            <div className="relative z-10 space-y-6">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                Быстрые действия
              </h2>
              <div className="space-y-3">
                <button className="w-full px-4 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg text-left">
                  Загрузить документ
                </button>
                <button className="w-full px-4 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg text-left">
                  Добавить студента
                </button>
                <button className="w-full px-4 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg text-left">
                  Создать мероприятие
                </button>
                <button className="w-full px-4 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-105 active:scale-95 hover:shadow-lg text-left">
                  Создать задачу
                </button>
              </div>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="lg:col-span-2">
            {activityLoading ? (
              <div className="relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                <div className="h-96 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
            ) : (
              <ActivityFeed activities={activities} />
            )}
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
