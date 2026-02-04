'use client'

import { useTranslations } from 'next-intl'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { RiskLevel } from '@/lib/api/analytics'

interface RiskLevelBadgeProps {
  level: RiskLevel
  className?: string
  showLabel?: boolean
}

const riskColors: Record<RiskLevel, { bg: string; text: string; border: string }> = {
  low: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-700 dark:text-green-400',
    border: 'border-green-300 dark:border-green-700',
  },
  medium: {
    bg: 'bg-yellow-100 dark:bg-yellow-900/30',
    text: 'text-yellow-700 dark:text-yellow-400',
    border: 'border-yellow-300 dark:border-yellow-700',
  },
  high: {
    bg: 'bg-orange-100 dark:bg-orange-900/30',
    text: 'text-orange-700 dark:text-orange-400',
    border: 'border-orange-300 dark:border-orange-700',
  },
  critical: {
    bg: 'bg-red-100 dark:bg-red-900/30',
    text: 'text-red-700 dark:text-red-400',
    border: 'border-red-300 dark:border-red-700',
  },
}

export function RiskLevelBadge({ level, className, showLabel = true }: RiskLevelBadgeProps) {
  const t = useTranslations('analytics')
  const colors = riskColors[level]

  return (
    <Badge
      variant="outline"
      className={cn(colors.bg, colors.text, colors.border, 'font-medium', className)}
    >
      {showLabel ? t(`riskLevel.${level}`) : level}
    </Badge>
  )
}
