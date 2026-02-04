'use client'

import { FileText, Eye, Plus, Settings } from 'lucide-react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { TemplateInfo } from '@/lib/api/templates'
import { useTranslations } from 'next-intl'

interface TemplateCardProps {
  template: TemplateInfo
  onPreview: (template: TemplateInfo) => void
  onCreate: (template: TemplateInfo) => void
  onEdit?: (template: TemplateInfo) => void
  canEdit?: boolean
  className?: string
}

export function TemplateCard({
  template,
  onPreview,
  onCreate,
  onEdit,
  canEdit = false,
  className,
}: TemplateCardProps) {
  const t = useTranslations('templates')
  const variablesCount = template.template_variables?.length || 0

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700 transition-all duration-300 hover:shadow-xl group',
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
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
            <FileText className="h-6 w-6" />
          </div>
          {variablesCount > 0 && (
            <Badge variant="secondary" className="text-xs">
              {variablesCount} {t('variables')}
            </Badge>
          )}
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
            className="flex-1 bg-blue-600 hover:bg-blue-700 text-white"
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
