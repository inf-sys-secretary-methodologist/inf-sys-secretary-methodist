'use client'

import { useTranslations } from 'next-intl'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { FileText, Users, Calendar, ClipboardList, GraduationCap } from 'lucide-react'
import type { DataSourceType } from '@/types/reports'

interface DataSourceSelectorProps {
  selected: DataSourceType
  onChange: (source: DataSourceType) => void
}

const DATA_SOURCES: { type: DataSourceType; icon: React.ElementType; color: string }[] = [
  { type: 'documents', icon: FileText, color: 'text-blue-500' },
  { type: 'users', icon: Users, color: 'text-green-500' },
  { type: 'events', icon: Calendar, color: 'text-purple-500' },
  { type: 'tasks', icon: ClipboardList, color: 'text-orange-500' },
  { type: 'students', icon: GraduationCap, color: 'text-cyan-500' },
]

export function DataSourceSelector({ selected, onChange }: DataSourceSelectorProps) {
  const t = useTranslations('reports.builder')

  return (
    <div className="relative overflow-hidden rounded-xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={2}
      />
      <div className="relative z-10">
        <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">
          {t('dataSource')}
        </h3>
        <div className="flex flex-wrap gap-2">
          {DATA_SOURCES.map(({ type, icon: Icon, color }) => (
            <button
              key={type}
              onClick={() => onChange(type)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                selected === type
                  ? 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              <Icon className={`h-4 w-4 ${selected === type ? '' : color}`} />
              {t(`sources.${type}`)}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
