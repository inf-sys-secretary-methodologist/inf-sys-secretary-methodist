'use client'

import { useState } from 'react'
import dynamic from 'next/dynamic'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { canEdit } from '@/lib/auth/permissions'
import { useDashboardStats, useDashboardTrends, useDashboardActivity } from '@/hooks/useDashboard'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { StatsCard, ActivityFeed, ExportButton } from '@/components/dashboard'
import { FileText, Users, Calendar, TrendingUp, ClipboardList } from 'lucide-react'

// Lazy load MetodychWidget
const MetodychWidget = dynamic(
  () => import('@/components/dashboard/MetodychWidget').then((mod) => mod.MetodychWidget),
  {
    loading: () => (
      <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
        <div className="flex items-start gap-4">
          <div className="w-16 h-16 bg-gray-200 dark:bg-gray-700 rounded-full" />
          <div className="flex-1 space-y-3">
            <div className="h-4 w-48 bg-gray-200 dark:bg-gray-700 rounded" />
            <div className="h-3 w-full bg-gray-200 dark:bg-gray-700 rounded" />
          </div>
        </div>
      </div>
    ),
    ssr: false,
  }
)

// Lazy load TrendChart to reduce initial bundle (recharts ~200KB)
const TrendChart = dynamic(
  () => import('@/components/dashboard/TrendChart').then((mod) => mod.TrendChart),
  {
    loading: () => (
      <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
        <div className="h-4 w-32 bg-gray-200 dark:bg-gray-700 rounded mb-4" />
        <div className="h-48 sm:h-64 bg-gray-200 dark:bg-gray-700 rounded" />
      </div>
    ),
    ssr: false,
  }
)

type Period = 'week' | 'month' | 'quarter' | 'year'

export default function DashboardPage() {
  const router = useRouter()
  const { user } = useAuthCheck()
  const t = useTranslations('dashboard')
  const tRoles = useTranslations('roles')
  const userCanEdit = canEdit(user?.role)
  const [period, setPeriod] = useState<Period>('month')

  // Fetch dashboard data
  const { stats, isLoading: statsLoading } = useDashboardStats(period)
  const { trends, isLoading: trendsLoading } = useDashboardTrends(period)
  const { activities, isLoading: activityLoading } = useDashboardActivity(10)

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Welcome Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            {t('welcome', { name: user?.name || '' })}
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">
            <span className="font-semibold">{user?.role ? tRoles(user.role) : ''}</span>
          </p>
        </div>

        {/* Period Selector & Export */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-4">
          <div className="relative overflow-hidden rounded-xl p-1 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
            <GlowingEffect
              spread={40}
              glow={true}
              disabled={false}
              proximity={64}
              inactiveZone={0.01}
              borderWidth={2}
            />
            <div className="relative z-10 flex flex-wrap gap-1">
              {(['week', 'month', 'quarter', 'year'] as Period[]).map((p) => (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className={`flex-1 sm:flex-none px-3 sm:px-4 py-2 rounded-lg text-xs sm:text-sm font-medium transition-all ${
                    period === p
                      ? 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-white/10'
                  }`}
                >
                  {t(`periods.${p}`)}
                </button>
              ))}
            </div>
          </div>

          <ExportButton />
        </div>

        {/* Stats Grid */}
        <section aria-labelledby="stats-heading">
          <h2 id="stats-heading" className="sr-only">
            {t('stats.statistics')}
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 sm:gap-4 lg:gap-6">
            {statsLoading ? (
              // Loading skeletons
              Array.from({ length: 5 }).map((_, i) => (
                <div
                  key={i}
                  className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse"
                >
                  <div className="h-8 sm:h-12 w-8 sm:w-12 bg-gray-200 dark:bg-gray-700 rounded-lg mb-3 sm:mb-4" />
                  <div className="h-3 sm:h-4 w-16 sm:w-20 bg-gray-200 dark:bg-gray-700 rounded mb-2" />
                  <div className="h-6 sm:h-8 w-12 sm:w-16 bg-gray-200 dark:bg-gray-700 rounded" />
                </div>
              ))
            ) : (
              <>
                <StatsCard
                  icon={FileText}
                  title={t('stats.documents')}
                  value={stats?.documents.total || 0}
                  change={stats?.documents.change || 0}
                  period={t(`periodLabels.${period}`)}
                />
                <StatsCard
                  icon={Users}
                  title={t('stats.students')}
                  value={stats?.students.total || 0}
                  change={stats?.students.change || 0}
                  period={t(`periodLabels.${period}`)}
                />
                <StatsCard
                  icon={Calendar}
                  title={t('stats.events')}
                  value={stats?.events.total || 0}
                  change={stats?.events.change || 0}
                  period={t(`periodLabels.${period}`)}
                />
                <StatsCard
                  icon={TrendingUp}
                  title={t('stats.reports')}
                  value={stats?.reports.total || 0}
                  change={stats?.reports.change || 0}
                  period={t(`periodLabels.${period}`)}
                />
                <StatsCard
                  icon={ClipboardList}
                  title={t('stats.tasks')}
                  value={stats?.tasks.total || 0}
                  change={stats?.tasks.change || 0}
                  period={t(`periodLabels.${period}`)}
                  className="col-span-2 sm:col-span-1"
                />
              </>
            )}
          </div>
        </section>

        {/* Metodych AI Assistant */}
        <MetodychWidget />

        {/* Charts Row */}
        <section aria-labelledby="trends-heading">
          <h2 id="trends-heading" className="sr-only">
            {t('stats.trends')}
          </h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
            {trendsLoading ? (
              <>
                <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                  <div className="h-48 sm:h-64 bg-gray-200 dark:bg-gray-700 rounded" />
                </div>
                <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                  <div className="h-48 sm:h-64 bg-gray-200 dark:bg-gray-700 rounded" />
                </div>
              </>
            ) : (
              <>
                <TrendChart
                  title={t('charts.documentsAndReports')}
                  period={period}
                  datasets={[
                    {
                      name: t('stats.documents'),
                      data: trends?.documents_trend || [],
                      color: '#3b82f6',
                    },
                    {
                      name: t('stats.reports'),
                      data: trends?.reports_trend || [],
                      color: '#10b981',
                    },
                  ]}
                />
                <TrendChart
                  title={t('charts.tasksAndEvents')}
                  period={period}
                  datasets={[
                    {
                      name: t('stats.tasks'),
                      data: trends?.tasks_trend || [],
                      color: '#f59e0b',
                    },
                    {
                      name: t('stats.events'),
                      data: trends?.events_trend || [],
                      color: '#8b5cf6',
                    },
                  ]}
                />
              </>
            )}
          </div>
        </section>

        {/* Quick Actions & Activity */}
        <div className={`grid grid-cols-1 ${userCanEdit ? 'lg:grid-cols-3' : ''} gap-4 sm:gap-6`}>
          {/* Quick Actions - only for users with edit permissions */}
          {userCanEdit && (
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-6 sm:p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10 space-y-4 sm:space-y-6">
                <h2 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white">
                  {t('quickActions.title')}
                </h2>
                <div className="space-y-2 sm:space-y-3">
                  <button
                    onClick={() => router.push('/documents')}
                    className="w-full px-3 sm:px-4 py-2.5 sm:py-3 rounded-lg font-medium text-sm sm:text-base transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-[1.02] active:scale-[0.98] hover:shadow-lg text-left"
                  >
                    {t('quickActions.uploadDocument')}
                  </button>
                  <button
                    onClick={() => router.push('/students')}
                    className="w-full px-3 sm:px-4 py-2.5 sm:py-3 rounded-lg font-medium text-sm sm:text-base transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-[1.02] active:scale-[0.98] hover:shadow-lg text-left"
                  >
                    {t('quickActions.addStudent')}
                  </button>
                  <button
                    onClick={() => router.push('/calendar')}
                    className="w-full px-3 sm:px-4 py-2.5 sm:py-3 rounded-lg font-medium text-sm sm:text-base transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-[1.02] active:scale-[0.98] hover:shadow-lg text-left"
                  >
                    {t('quickActions.createEvent')}
                  </button>
                  <button
                    onClick={() => router.push('/calendar')}
                    className="w-full px-3 sm:px-4 py-2.5 sm:py-3 rounded-lg font-medium text-sm sm:text-base transition-all duration-300 bg-white dark:bg-white text-gray-900 hover:bg-gray-900 dark:hover:bg-gray-900 hover:text-white dark:hover:text-white border border-gray-200 hover:border-gray-900 dark:hover:border-gray-700 hover:scale-[1.02] active:scale-[0.98] hover:shadow-lg text-left"
                  >
                    {t('quickActions.createTask')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Recent Activity */}
          <div className={userCanEdit ? 'lg:col-span-2' : ''}>
            {activityLoading ? (
              <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
                <div className="h-72 sm:h-96 bg-gray-200 dark:bg-gray-700 rounded" />
              </div>
            ) : (
              <ActivityFeed activities={activities} />
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  )
}
