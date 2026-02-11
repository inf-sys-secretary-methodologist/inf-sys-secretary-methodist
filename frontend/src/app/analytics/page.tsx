'use client'

import { useState, useEffect } from 'react'
import dynamic from 'next/dynamic'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AlertTriangle, Users, TrendingUp, BarChart3, Loader2 } from 'lucide-react'
import { analyticsApi, GroupSummaryInfo, RiskLevel } from '@/lib/api/analytics'

// Динамический импорт тяжелых компонентов с графиками
const AtRiskStudentsList = dynamic(
  () => import('@/components/analytics').then((mod) => ({ default: mod.AtRiskStudentsList })),
  {
    loading: () => (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    ),
  }
)
const GroupSummaryCard = dynamic(
  () => import('@/components/analytics').then((mod) => ({ default: mod.GroupSummaryCard })),
  { loading: () => <div className="animate-pulse h-32 bg-gray-200 dark:bg-gray-800 rounded-lg" /> }
)
const AttendanceTrendChart = dynamic(
  () => import('@/components/analytics').then((mod) => ({ default: mod.AttendanceTrendChart })),
  {
    loading: () => (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    ),
    ssr: false,
  }
)
const RiskDistributionChart = dynamic(
  () => import('@/components/analytics').then((mod) => ({ default: mod.RiskDistributionChart })),
  {
    loading: () => (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    ),
    ssr: false,
  }
)

export default function AnalyticsPage() {
  useAuthCheck()
  const t = useTranslations('analytics')
  const [activeTab, setActiveTab] = useState('overview')
  const [selectedRiskLevel, setSelectedRiskLevel] = useState<RiskLevel | 'all'>('all')
  const [groups, setGroups] = useState<GroupSummaryInfo[]>([])
  const [isLoadingGroups, setIsLoadingGroups] = useState(true)

  useEffect(() => {
    const fetchGroups = async () => {
      try {
        setIsLoadingGroups(true)
        const data = await analyticsApi.getAllGroupsSummary()
        setGroups(data)
      } catch (err) {
        console.error('Failed to fetch groups:', err)
      } finally {
        setIsLoadingGroups(false)
      }
    }

    fetchGroups()
  }, [])

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            {t('title')}
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">{t('subtitle')}</p>
        </div>

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-4 mb-6">
            <TabsTrigger value="overview" className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.overview')}</span>
            </TabsTrigger>
            <TabsTrigger value="at-risk" className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.atRisk')}</span>
            </TabsTrigger>
            <TabsTrigger value="groups" className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.groups')}</span>
            </TabsTrigger>
            <TabsTrigger value="trends" className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              <span className="hidden sm:inline">{t('tabs.trends')}</span>
            </TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-6">
            {/* Stats Row */}
            <div className="grid gap-6 sm:grid-cols-2">
              <RiskDistributionChart />
              <AttendanceTrendChart months={6} />
            </div>

            {/* Quick At-Risk View */}
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400">
                      <AlertTriangle className="h-5 w-5" />
                    </div>
                    <div>
                      <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                        {t('criticalRiskStudents')}
                      </h2>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {t('requiresAttention')}
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setSelectedRiskLevel('critical')
                      setActiveTab('at-risk')
                    }}
                  >
                    {t('viewAll')}
                  </Button>
                </div>
                <AtRiskStudentsList riskLevel="critical" pageSize={3} />
              </div>
            </div>
          </TabsContent>

          {/* At-Risk Students Tab */}
          <TabsContent value="at-risk" className="space-y-6">
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400">
                      <AlertTriangle className="h-5 w-5" />
                    </div>
                    <div>
                      <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                        {t('atRiskStudents')}
                      </h2>
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        {t('filterByRisk')}
                      </p>
                    </div>
                  </div>

                  <Select
                    value={selectedRiskLevel}
                    onValueChange={(v) => setSelectedRiskLevel(v as RiskLevel | 'all')}
                  >
                    <SelectTrigger className="w-40">
                      <SelectValue placeholder={t('selectRiskLevel')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">{t('allLevels')}</SelectItem>
                      <SelectItem value="critical">{t('riskLevel.critical')}</SelectItem>
                      <SelectItem value="high">{t('riskLevel.high')}</SelectItem>
                      <SelectItem value="medium">{t('riskLevel.medium')}</SelectItem>
                      <SelectItem value="low">{t('riskLevel.low')}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <AtRiskStudentsList
                  riskLevel={selectedRiskLevel === 'all' ? undefined : selectedRiskLevel}
                  pageSize={9}
                />
              </div>
            </div>
          </TabsContent>

          {/* Groups Tab */}
          <TabsContent value="groups" className="space-y-6">
            <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={3}
              />
              <div className="relative z-10">
                <div className="flex items-center gap-3 mb-6">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400">
                    <Users className="h-5 w-5" />
                  </div>
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                      {t('groupsSummary')}
                    </h2>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {t('analyticsPerGroup')}
                    </p>
                  </div>
                </div>

                {isLoadingGroups ? (
                  <div className="flex items-center justify-center py-12">
                    <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
                  </div>
                ) : groups.length === 0 ? (
                  <div className="text-center py-12">
                    <Users className="h-12 w-12 mx-auto text-gray-400 mb-4" />
                    <p className="text-gray-600 dark:text-gray-400">{t('noGroups')}</p>
                  </div>
                ) : (
                  <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {groups.map((group) => (
                      <GroupSummaryCard key={group.group_name} group={group} />
                    ))}
                  </div>
                )}
              </div>
            </div>
          </TabsContent>

          {/* Trends Tab */}
          <TabsContent value="trends" className="space-y-6">
            <div className="grid gap-6 lg:grid-cols-2">
              <AttendanceTrendChart months={12} />
              <RiskDistributionChart />
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </AppLayout>
  )
}
