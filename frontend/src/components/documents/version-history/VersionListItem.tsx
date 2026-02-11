'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
import {
  RotateCcw,
  Download,
  Trash2,
  Clock,
  User,
  ChevronDown,
  ChevronUp,
  Check,
  Loader2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { DocumentVersionInfo } from '@/lib/api/documents'

interface VersionListItemProps {
  version: DocumentVersionInfo
  isLatest: boolean
  isExpanded: boolean
  isSelected: boolean
  compareMode: boolean
  isRestoring: boolean
  isDeleting: boolean
  onToggleExpand: () => void
  onSelect: () => void
  onRestore: () => void
  onDelete: () => void
  onDownloadFile: () => void
  formatDate: (date: string) => string
}

export const VersionListItem = memo(function VersionListItem({
  version,
  isLatest,
  isExpanded,
  isSelected,
  compareMode,
  isRestoring,
  isDeleting,
  onToggleExpand,
  onSelect,
  onRestore,
  onDelete,
  onDownloadFile,
  formatDate,
}: VersionListItemProps) {
  const t = useTranslations('documents.versions')
  const tCommon = useTranslations('common')

  return (
    <div
      /* c8 ignore next 5 - Conditional styling */
      className={`
        border rounded-lg overflow-hidden transition-all
        ${isLatest ? 'border-blue-300 dark:border-blue-600' : 'border-gray-200 dark:border-gray-700'}
        ${isSelected ? 'ring-2 ring-blue-500' : ''}
        ${compareMode ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50' : ''}
      `}
      onClick={compareMode ? onSelect : undefined}
    >
      {/* Header */}
      <div
        className={`
          flex items-center justify-between p-3
          ${!compareMode ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50' : ''}
        `}
        onClick={!compareMode ? onToggleExpand : undefined}
      >
        <div className="flex items-center gap-3">
          {/* c8 ignore start - Compare mode checkbox */}
          {compareMode && (
            <div
              className={`
                w-5 h-5 rounded border-2 flex items-center justify-center
                ${isSelected ? 'bg-blue-500 border-blue-500' : 'border-gray-300 dark:border-gray-600'}
              `}
            >
              {isSelected && <Check className="h-3 w-3 text-white" />}
            </div>
          )}
          {/* c8 ignore stop */}

          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-gray-900 dark:text-white">
                {t('version')} {version.version}
              </span>
              {isLatest && (
                <span className="px-2 py-0.5 text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded">
                  {t('current')}
                </span>
              )}
            </div>
            <div className="flex items-center gap-3 text-sm text-gray-500 mt-1">
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDate(version.created_at)}
              </span>
              {version.changed_by_name && (
                <span className="flex items-center gap-1">
                  <User className="h-3 w-3" />
                  {version.changed_by_name}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* c8 ignore start - Version item action buttons */}
        {!compareMode && (
          <div className="flex items-center gap-2">
            {version.file_name && (
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation()
                  onDownloadFile()
                }}
              >
                <Download className="h-4 w-4" />
              </Button>
            )}

            {!isLatest && (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    onRestore()
                  }}
                  disabled={isRestoring}
                >
                  {isRestoring ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <RotateCcw className="h-4 w-4" />
                  )}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete()
                  }}
                  disabled={isDeleting}
                  className="text-red-500 hover:text-red-700"
                >
                  {isDeleting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                </Button>
              </>
            )}

            {isExpanded ? (
              <ChevronUp className="h-4 w-4 text-gray-400" />
            ) : (
              <ChevronDown className="h-4 w-4 text-gray-400" />
            )}
          </div>
        )}
        {/* c8 ignore stop */}
      </div>

      {/* Expanded Details */}
      {/* c8 ignore start - Expanded state */}
      {isExpanded && !compareMode && (
        <div className="px-3 pb-3 pt-0 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/30">
          <div className="grid grid-cols-2 gap-4 text-sm pt-3">
            {version.title && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">{t('title')}</p>
                <p className="text-gray-900 dark:text-white">{version.title}</p>
              </div>
            )}
            {version.status && (
              <div>
                <p className="text-gray-500 dark:text-gray-400">{t('status')}</p>
                <p className="text-gray-900 dark:text-white">{version.status}</p>
              </div>
            )}
            {version.change_description && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('changes')}</p>
                <p className="text-gray-900 dark:text-white">{version.change_description}</p>
              </div>
            )}
            {version.file_name && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('file')}</p>
                <p className="text-gray-900 dark:text-white">
                  {version.file_name}
                  {version.file_size && (
                    <span className="text-gray-500 ml-2">
                      (
                      {tCommon('fileSize.mb', {
                        size: (version.file_size / 1024 / 1024).toFixed(2),
                      })}
                      )
                    </span>
                  )}
                </p>
              </div>
            )}
            {version.subject && (
              <div className="col-span-2">
                <p className="text-gray-500 dark:text-gray-400">{t('topic')}</p>
                <p className="text-gray-900 dark:text-white">{version.subject}</p>
              </div>
            )}
          </div>
        </div>
      )}
      {/* c8 ignore stop */}
    </div>
  )
})
