'use client'

import { AppLayout } from '@/components/layout'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Users } from 'lucide-react'

export default function StudentsPage() {
  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto space-y-6 sm:space-y-8">
        {/* Page Header */}
        <div className="text-center space-y-2 sm:space-y-4">
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-gray-900 dark:text-white">
            Управление студентами
          </h1>
          <p className="text-base sm:text-lg text-gray-600 dark:text-gray-300">
            Список студентов и их данные
          </p>
        </div>

        {/* Content Placeholder */}
        <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-6 sm:p-8 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
          <GlowingEffect
            spread={40}
            glow={true}
            disabled={false}
            proximity={64}
            inactiveZone={0.01}
            borderWidth={3}
          />
          <div className="relative z-10 space-y-4 sm:space-y-6 text-center py-8 sm:py-12">
            <Users className="h-12 w-12 sm:h-16 sm:w-16 mx-auto text-gray-400" />
            <h2 className="text-xl sm:text-2xl font-semibold text-gray-900 dark:text-white">
              Раздел в разработке
            </h2>
            <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 max-w-md mx-auto">
              Здесь будет отображаться список студентов, их профили и статистика
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  )
}
