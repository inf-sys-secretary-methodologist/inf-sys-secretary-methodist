'use client'

import { LucideIcon } from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { NumberTicker } from '@/components/ui/number-ticker'
import { cn } from '@/lib/utils'

interface StatsCardProps {
  icon: LucideIcon
  title: string
  value: number
  change: number
  period: string
  className?: string
}

export function StatsCard({ icon: Icon, title, value, change, period, className }: StatsCardProps) {
  const isPositive = change >= 0
  const changeFormatted = `${isPositive ? '+' : ''}${change.toFixed(1)}%`

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:scale-105 hover:shadow-xl',
        className
      )}
    >
      <GlowingEffect
        spread={40}
        glow={true}
        disabled={false}
        proximity={64}
        inactiveZone={0.01}
        borderWidth={3}
      />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 dark:bg-white/10 text-gray-900 dark:text-white">
            <Icon className="h-6 w-6" />
          </div>
          <span
            className={cn(
              'text-xs px-2 py-1 rounded-full font-medium border',
              isPositive
                ? 'bg-green-500/20 text-green-600 dark:text-green-400 border-green-500/50'
                : 'bg-red-500/20 text-red-600 dark:text-red-400 border-red-500/50'
            )}
          >
            {changeFormatted}
          </span>
        </div>
        <p className="text-sm font-medium mb-2 text-gray-700 dark:text-gray-300">{title}</p>
        <div className="text-3xl font-bold text-gray-900 dark:text-white">
          <NumberTicker value={value} />
        </div>
        <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">за {period}</p>
      </div>
    </div>
  )
}
