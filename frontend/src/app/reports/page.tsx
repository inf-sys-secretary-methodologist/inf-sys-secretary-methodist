'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useAuthCheck } from '@/hooks/useAuth'
import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import {
  Plus,
  FileText,
  Clock,
  Users,
  Calendar,
  ClipboardList,
  GraduationCap,
  ChevronRight,
} from 'lucide-react'
import type { DataSourceType } from '@/types/reports'

interface QuickReportTemplate {
  id: string
  titleKey: string
  descriptionKey: string
  icon: React.ElementType
  dataSource: DataSourceType
  color: string
}

const QUICK_TEMPLATES: QuickReportTemplate[] = [
  {
    id: 'documents_by_category',
    titleKey: 'documentsByCategory',
    descriptionKey: 'documentsByCategoryDesc',
    icon: FileText,
    dataSource: 'documents',
    color: 'text-blue-500',
  },
  {
    id: 'users_by_role',
    titleKey: 'usersByRole',
    descriptionKey: 'usersByRoleDesc',
    icon: Users,
    dataSource: 'users',
    color: 'text-green-500',
  },
  {
    id: 'events_this_month',
    titleKey: 'eventsThisMonth',
    descriptionKey: 'eventsThisMonthDesc',
    icon: Calendar,
    dataSource: 'events',
    color: 'text-purple-500',
  },
  {
    id: 'tasks_by_status',
    titleKey: 'tasksByStatus',
    descriptionKey: 'tasksByStatusDesc',
    icon: ClipboardList,
    dataSource: 'tasks',
    color: 'text-orange-500',
  },
  {
    id: 'students_by_group',
    titleKey: 'studentsByGroup',
    descriptionKey: 'studentsByGroupDesc',
    icon: GraduationCap,
    dataSource: 'students',
    color: 'text-cyan-500',
  },
  {
    id: 'recent_documents',
    titleKey: 'recentDocuments',
    descriptionKey: 'recentDocumentsDesc',
    icon: Clock,
    dataSource: 'documents',
    color: 'text-pink-500',
  },
]

export default function ReportsPage() {
  const router = useRouter()
  const { user: _user } = useAuthCheck()
  const t = useTranslations('reports')
  const [savedReports] = useState<{ id: string; name: string; updatedAt: string }[]>([])

  const handleCreateNew = () => {
    router.push('/reports/builder')
  }

  const handleQuickTemplate = (template: QuickReportTemplate) => {
    router.push(`/reports/builder?source=${template.dataSource}&template=${template.id}`)
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">
              {t('title')}
            </h1>
            <p className="text-sm sm:text-base text-gray-600 dark:text-gray-300 mt-1">
              {t('subtitle')}
            </p>
          </div>
          <Button
            onClick={handleCreateNew}
            className="flex items-center gap-2 bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-100"
          >
            <Plus className="h-4 w-4" />
            {t('createNew')}
          </Button>
        </div>

        {/* Quick Templates */}
        <section>
          <h2 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white mb-4">
            {t('quickTemplates')}
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {QUICK_TEMPLATES.map((template) => {
              const Icon = template.icon
              return (
                <button
                  key={template.id}
                  onClick={() => handleQuickTemplate(template)}
                  className="group relative overflow-hidden rounded-xl p-5 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 text-left transition-all hover:shadow-lg hover:-translate-y-0.5"
                >
                  <GlowingEffect
                    spread={40}
                    glow={true}
                    disabled={false}
                    proximity={64}
                    inactiveZone={0.01}
                    borderWidth={2}
                  />
                  <div className="relative z-10">
                    <div
                      className={`w-10 h-10 rounded-lg flex items-center justify-center bg-gray-100 dark:bg-gray-800 mb-3 ${template.color}`}
                    >
                      <Icon className="h-5 w-5" />
                    </div>
                    <h3 className="font-medium text-gray-900 dark:text-white mb-1">
                      {t(`templates.${template.titleKey}`)}
                    </h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {t(`templates.${template.descriptionKey}`)}
                    </p>
                    <div className="flex items-center gap-1 mt-3 text-sm text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors">
                      <span>{t('useTemplate')}</span>
                      <ChevronRight className="h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
                    </div>
                  </div>
                </button>
              )
            })}
          </div>
        </section>

        {/* Saved Reports */}
        <section>
          <h2 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white mb-4">
            {t('savedReports')}
          </h2>
          {savedReports.length === 0 ? (
            <div className="relative overflow-hidden rounded-xl p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 text-center">
              <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={64}
                inactiveZone={0.01}
                borderWidth={2}
              />
              <div className="relative z-10">
                <FileText className="h-12 w-12 mx-auto text-gray-300 dark:text-gray-600 mb-4" />
                <p className="text-gray-500 dark:text-gray-400 mb-4">{t('noSavedReports')}</p>
                <Button
                  onClick={handleCreateNew}
                  variant="outline"
                  className="border-gray-300 dark:border-gray-600"
                >
                  {t('createFirstReport')}
                </Button>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {savedReports.map((report) => (
                <button
                  key={report.id}
                  onClick={() => router.push(`/reports/builder?id=${report.id}`)}
                  className="relative overflow-hidden rounded-xl p-5 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 text-left transition-all hover:shadow-lg"
                >
                  <GlowingEffect
                    spread={40}
                    glow={true}
                    disabled={false}
                    proximity={64}
                    inactiveZone={0.01}
                    borderWidth={2}
                  />
                  <div className="relative z-10">
                    <h3 className="font-medium text-gray-900 dark:text-white mb-1">
                      {report.name}
                    </h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {t('lastUpdated', { date: report.updatedAt })}
                    </p>
                  </div>
                </button>
              ))}
            </div>
          )}
        </section>
      </div>
    </AppLayout>
  )
}
