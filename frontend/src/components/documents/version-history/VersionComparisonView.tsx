'use client'

import { memo } from 'react'
import dynamic from 'next/dynamic'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/ui/button'
import { VersionDiffOutput } from '@/lib/api/documents'

// Lazy load TextDiff to reduce initial bundle (diff library ~15KB)
const TextDiff = dynamic(() => import('../TextDiff').then((mod) => mod.TextDiff), {
  loading: () => <div className="animate-pulse bg-gray-100 dark:bg-gray-800 rounded h-24" />,
  ssr: false,
})

interface VersionComparisonViewProps {
  comparisonResult: VersionDiffOutput
  onClose: () => void
}

export const VersionComparisonView = memo(function VersionComparisonView({
  comparisonResult,
  onClose,
}: VersionComparisonViewProps) {
  const t = useTranslations('documents.versions')
  const tCommon = useTranslations('common')

  return (
    <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="font-medium text-blue-900 dark:text-blue-100">
          {t('comparisonTitle', {
            from: comparisonResult.from_version,
            to: comparisonResult.to_version,
          })}
        </h4>
        {/* c8 ignore next - Close comparison button */}
        <Button variant="ghost" size="sm" onClick={onClose}>
          {tCommon('close')}
        </Button>
      </div>

      {comparisonResult.changed_fields.length === 0 ? (
        <p className="text-sm text-blue-700 dark:text-blue-300">{t('versionsIdentical')}</p>
      ) : (
        <div className="space-y-2">
          <p className="text-sm text-blue-700 dark:text-blue-300">
            {t('changedFields')}: {comparisonResult.changed_fields.join(', ')}
          </p>
          {/* c8 ignore start - Diff rendering, complex UI tested in e2e */}
          {comparisonResult.diff_data && (
            <div className="space-y-4 mt-3">
              {Object.entries(comparisonResult.diff_data).map(([field, diff]) => {
                const fieldLabels: Record<string, string> = {
                  title: t('fieldLabels.title'),
                  subject: t('fieldLabels.subject'),
                  content: t('fieldLabels.content'),
                  status: t('fieldLabels.status'),
                  file_name: t('fieldLabels.file_name'),
                }
                const isTextContent =
                  field === 'content' || field === 'subject' || field === 'title'
                const oldValue = String(diff.from || '')
                const newValue = String(diff.to || '')

                return (
                  <div
                    key={field}
                    className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-blue-200 dark:border-blue-700"
                  >
                    <div className="font-medium text-sm text-gray-700 dark:text-gray-300 mb-3">
                      {fieldLabels[field] || field}
                    </div>

                    {isTextContent && (oldValue.length > 20 || newValue.length > 20) ? (
                      <TextDiff
                        oldText={oldValue}
                        newText={newValue}
                        oldLabel={`${t('version')} ${comparisonResult.from_version}`}
                        newLabel={`${t('version')} ${comparisonResult.to_version}`}
                      />
                    ) : (
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="text-red-600 dark:text-red-400 font-medium">
                            {t('before')}:
                          </span>
                          <p className="mt-1 text-gray-600 dark:text-gray-400 bg-red-50 dark:bg-red-900/20 p-2 rounded break-words">
                            {oldValue || t('emptyValue')}
                          </p>
                        </div>
                        <div>
                          <span className="text-green-600 dark:text-green-400 font-medium">
                            {t('after')}:
                          </span>
                          <p className="mt-1 text-gray-600 dark:text-gray-400 bg-green-50 dark:bg-green-900/20 p-2 rounded break-words">
                            {newValue || t('emptyValue')}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
          {/* c8 ignore stop */}
        </div>
      )}
    </div>
  )
})
