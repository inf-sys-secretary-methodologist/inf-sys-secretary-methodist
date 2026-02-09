'use client'

import {
  FileText,
  Eye,
  Plus,
  Settings,
  Star,
  ScrollText,
  Mail,
  ClipboardList,
  FileSignature,
  Briefcase,
} from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { TemplateInfo } from '@/lib/api/templates'
import { useTranslations } from 'next-intl'

// Icon mapping for template types
const TEMPLATE_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  order_main: ScrollText,
  order_hr: ScrollText,
  order_admin: ScrollText,
  memo: FileText,
  directive: FileText,
  business_letter: Mail,
  protocol: ClipboardList,
  contract: FileSignature,
  job_instruction: Briefcase,
}

// Color mapping for template types
const TEMPLATE_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  order_main: {
    bg: 'bg-blue-100 dark:bg-blue-900/30',
    text: 'text-blue-600 dark:text-blue-400',
    border: 'border-blue-200 dark:border-blue-800',
  },
  order_hr: {
    bg: 'bg-purple-100 dark:bg-purple-900/30',
    text: 'text-purple-600 dark:text-purple-400',
    border: 'border-purple-200 dark:border-purple-800',
  },
  order_admin: {
    bg: 'bg-indigo-100 dark:bg-indigo-900/30',
    text: 'text-indigo-600 dark:text-indigo-400',
    border: 'border-indigo-200 dark:border-indigo-800',
  },
  memo: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-600 dark:text-green-400',
    border: 'border-green-200 dark:border-green-800',
  },
  directive: {
    bg: 'bg-teal-100 dark:bg-teal-900/30',
    text: 'text-teal-600 dark:text-teal-400',
    border: 'border-teal-200 dark:border-teal-800',
  },
  business_letter: {
    bg: 'bg-amber-100 dark:bg-amber-900/30',
    text: 'text-amber-600 dark:text-amber-400',
    border: 'border-amber-200 dark:border-amber-800',
  },
  protocol: {
    bg: 'bg-cyan-100 dark:bg-cyan-900/30',
    text: 'text-cyan-600 dark:text-cyan-400',
    border: 'border-cyan-200 dark:border-cyan-800',
  },
  contract: {
    bg: 'bg-rose-100 dark:bg-rose-900/30',
    text: 'text-rose-600 dark:text-rose-400',
    border: 'border-rose-200 dark:border-rose-800',
  },
  job_instruction: {
    bg: 'bg-orange-100 dark:bg-orange-900/30',
    text: 'text-orange-600 dark:text-orange-400',
    border: 'border-orange-200 dark:border-orange-800',
  },
}

const DEFAULT_COLORS = {
  bg: 'bg-gray-100 dark:bg-gray-900/30',
  text: 'text-gray-600 dark:text-gray-400',
  border: 'border-gray-200 dark:border-gray-700',
}

interface TemplateCardProps {
  template: TemplateInfo
  onPreview: (template: TemplateInfo) => void
  onCreate: (template: TemplateInfo) => void
  onEdit?: (template: TemplateInfo) => void
  canEdit?: boolean
  isFavorite?: boolean
  onToggleFavorite?: (template: TemplateInfo) => void
  className?: string
}

export function TemplateCard({
  template,
  onPreview,
  onCreate,
  onEdit,
  canEdit = false,
  isFavorite = false,
  onToggleFavorite,
  className,
}: TemplateCardProps) {
  const t = useTranslations('templates')
  const variablesCount = template.template_variables?.length || 0

  // Get icon and colors for this template type
  const Icon = TEMPLATE_ICONS[template.code] || FileText
  const colors = TEMPLATE_COLORS[template.code] || DEFAULT_COLORS

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border transition-all duration-300 hover:shadow-xl group',
        colors.border,
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
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div
            className={cn(
              'flex h-12 w-12 items-center justify-center rounded-lg',
              colors.bg,
              colors.text
            )}
          >
            <Icon className="h-6 w-6" />
          </div>
          <div className="flex items-center gap-2">
            {onToggleFavorite && (
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  onToggleFavorite(template)
                }}
                className={cn(
                  'p-1.5 rounded-lg transition-all duration-200',
                  isFavorite
                    ? 'text-yellow-500 hover:text-yellow-600'
                    : 'text-gray-400 hover:text-yellow-500 opacity-0 group-hover:opacity-100'
                )}
                aria-label={isFavorite ? t('removeFromFavorites') : t('addToFavorites')}
              >
                <Star className={cn('h-5 w-5', isFavorite && 'fill-current')} />
              </button>
            )}
            {variablesCount > 0 && (
              <Badge variant="secondary" className="text-xs">
                {variablesCount} {t('variables')}
              </Badge>
            )}
          </div>
        </div>

        {/* Content */}
        <div className="mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">
            {template.name}
          </h3>
          {template.description && (
            <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
              {template.description}
            </p>
          )}
          <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
            {t('code')}: {template.code}
          </p>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPreview(template)}
            className="flex-1"
          >
            <Eye className="h-4 w-4 mr-1" />
            {t('preview')}
          </Button>
          <Button
            size="sm"
            onClick={() => onCreate(template)}
            className="flex-1 bg-primary hover:bg-primary/90 text-primary-foreground"
          >
            <Plus className="h-4 w-4 mr-1" />
            {t('create')}
          </Button>
          {canEdit && onEdit && (
            <Button variant="ghost" size="sm" onClick={() => onEdit(template)} className="px-2">
              <Settings className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
