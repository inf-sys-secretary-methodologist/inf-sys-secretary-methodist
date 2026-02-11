'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
import { History, GitCompare, Plus, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface VersionListHeaderProps {
  totalVersions: number
  compareMode: boolean
  selectedVersionsCount: number
  isComparing: boolean
  canCompare: boolean
  onCompare: () => void
  onCancelCompare: () => void
  onToggleCompareMode: () => void
  onShowCreateForm: () => void
}

export const VersionListHeader = memo(function VersionListHeader({
  totalVersions,
  compareMode,
  selectedVersionsCount,
  isComparing,
  canCompare,
  onCompare,
  onCancelCompare,
  onToggleCompareMode,
  onShowCreateForm,
}: VersionListHeaderProps) {
  const t = useTranslations('documents.versions')
  const tCommon = useTranslations('common')

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <History className="h-5 w-5 text-gray-500" />
        <h3 className="font-semibold text-gray-900 dark:text-white">{t('history')}</h3>
        <span className="text-sm text-gray-500">
          ({totalVersions} {t('versionsCount')})
        </span>
      </div>

      <div className="flex items-center gap-2">
        {compareMode ? (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={onCompare}
              /* c8 ignore next - Compare button disabled condition */
              disabled={selectedVersionsCount !== 2 || isComparing}
            >
              {/* c8 ignore start - Loading state conditional */}
              {isComparing ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <GitCompare className="h-4 w-4 mr-2" />
              )}
              {/* c8 ignore stop */}
              {t('compareCount', { count: selectedVersionsCount })}
            </Button>
            <Button variant="ghost" size="sm" onClick={onCancelCompare}>
              {tCommon('cancel')}
            </Button>
          </>
        ) : (
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={onToggleCompareMode}
              disabled={!canCompare}
            >
              <GitCompare className="h-4 w-4 mr-2" />
              {t('compare')}
            </Button>
            <Button variant="outline" size="sm" onClick={onShowCreateForm}>
              <Plus className="h-4 w-4 mr-2" />
              {t('createVersion')}
            </Button>
          </>
        )}
      </div>
    </div>
  )
})
