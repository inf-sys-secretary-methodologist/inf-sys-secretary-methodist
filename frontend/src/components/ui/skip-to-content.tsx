'use client'

import { useTranslations } from 'next-intl'
import { cn } from '@/lib/utils'

interface SkipToContentProps {
  contentId?: string
  className?: string
}

/**
 * Skip to content link for keyboard navigation.
 * Allows users to skip navigation and go directly to main content.
 * WCAG 2.4.1: Bypass Blocks
 */
export function SkipToContent({ contentId = 'main-content', className }: SkipToContentProps) {
  const t = useTranslations()
  return (
    <a
      href={`#${contentId}`}
      className={cn(
        'sr-only focus:not-sr-only',
        'fixed top-4 left-4 z-[100]',
        'px-4 py-2 rounded-md',
        'bg-primary text-primary-foreground',
        'font-medium text-sm',
        'focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
        'transition-all duration-200',
        className
      )}
    >
      {t('skipToContent')}
    </a>
  )
}
