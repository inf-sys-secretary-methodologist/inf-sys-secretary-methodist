'use client'

import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { useMood } from '@/hooks/useMood'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { MessageCircle, AlertTriangle, FileWarning } from 'lucide-react'
import type { MoodState } from '@/types/mood'

const moodConfig: Record<MoodState, { emoji: string; color: string; animation: string }> = {
  happy: { emoji: '\u{1F604}', color: 'text-green-500', animation: 'animate-bounce' },
  content: { emoji: '\u{1F60A}', color: 'text-blue-500', animation: '' },
  worried: { emoji: '\u{1F61F}', color: 'text-yellow-500', animation: 'animate-pulse' },
  stressed: { emoji: '\u{1F630}', color: 'text-orange-500', animation: 'animate-pulse' },
  panicking: { emoji: '\u{1F92F}', color: 'text-red-500', animation: 'animate-ping-slow' },
  relaxed: { emoji: '\u{1F60C}', color: 'text-indigo-400', animation: '' },
  inspired: { emoji: '\u2728', color: 'text-purple-500', animation: 'animate-spin-slow' },
}

export function MetodychWidget() {
  const t = useTranslations('metodych')
  const { mood, isLoading } = useMood()

  if (isLoading) {
    return (
      <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 animate-pulse">
        <div className="flex items-start gap-4">
          <div className="w-16 h-16 bg-gray-200 dark:bg-gray-700 rounded-full" />
          <div className="flex-1 space-y-3">
            <div className="h-4 w-48 bg-gray-200 dark:bg-gray-700 rounded" />
            <div className="h-3 w-full bg-gray-200 dark:bg-gray-700 rounded" />
            <div className="h-3 w-3/4 bg-gray-200 dark:bg-gray-700 rounded" />
          </div>
        </div>
      </div>
    )
  }

  if (!mood) return null

  const config = moodConfig[mood.state] || moodConfig.content

  return (
    <div className="relative overflow-hidden rounded-xl sm:rounded-2xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={2}
      />
      <div className="relative z-10">
        <div className="flex items-start gap-3 sm:gap-4">
          {/* Avatar */}
          <div className={`text-4xl sm:text-5xl ${config.animation} flex-shrink-0`}>
            {config.emoji}
          </div>

          {/* Speech Bubble */}
          <div className="flex-1 min-w-0">
            <div className="relative bg-gray-50 dark:bg-gray-800/50 rounded-xl p-3 sm:p-4">
              {/* Bubble arrow */}
              <div className="absolute left-[-8px] top-4 w-0 h-0 border-t-[8px] border-t-transparent border-r-[8px] border-r-gray-50 dark:border-r-gray-800/50 border-b-[8px] border-b-transparent" />

              <p className="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mb-1">
                {mood.greeting}
              </p>
              <p className="text-sm sm:text-base text-gray-900 dark:text-white font-medium">
                {mood.message}
              </p>
            </div>

            {/* Metrics */}
            <div className="flex flex-wrap gap-3 mt-3">
              {mood.overdue_documents > 0 && (
                <div className="flex items-center gap-1.5 text-xs sm:text-sm text-orange-600 dark:text-orange-400">
                  <FileWarning className="w-3.5 h-3.5" />
                  <span>{t('overdueDocuments', { count: mood.overdue_documents })}</span>
                </div>
              )}
              {mood.at_risk_students > 0 && (
                <div className="flex items-center gap-1.5 text-xs sm:text-sm text-red-600 dark:text-red-400">
                  <AlertTriangle className="w-3.5 h-3.5" />
                  <span>{t('atRiskStudents', { count: mood.at_risk_students })}</span>
                </div>
              )}
            </div>

            {/* Chat button */}
            <Link
              href="/ai"
              className="inline-flex items-center gap-2 mt-3 px-3 sm:px-4 py-2 rounded-lg text-xs sm:text-sm font-medium transition-all duration-300 bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-700 dark:hover:bg-gray-200 hover:scale-[1.02] active:scale-[0.98]"
            >
              <MessageCircle className="w-3.5 h-3.5" />
              {t('chatButton')}
            </Link>
          </div>
        </div>

        {/* Fun fact */}
        {mood.fun_fact && (
          <div className="mt-4 p-3 bg-indigo-50 dark:bg-indigo-950/30 rounded-lg border border-indigo-100 dark:border-indigo-900/50">
            <p className="text-xs sm:text-sm text-indigo-700 dark:text-indigo-300">
              {mood.fun_fact}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
