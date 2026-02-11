'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
import { DocumentStatus } from '@/types/document'

interface StatusFilterProps {
  selectedStatus: DocumentStatus | 'all'
  onStatusChange: (status: DocumentStatus | 'all') => void
}

export const StatusFilter = memo(function StatusFilter({
  selectedStatus,
  onStatusChange,
}: StatusFilterProps) {
  const t = useTranslations('documents.filters')
  const tDocs = useTranslations('documents')

  /* c8 ignore start - JSX event handlers, tested in e2e */
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        {t('status')}
      </label>
      <select
        value={selectedStatus}
        onChange={(e) => onStatusChange(e.target.value as DocumentStatus | 'all')}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                 bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      >
        <option value="all">{t('allStatuses')}</option>
        {Object.values(DocumentStatus).map((value) => (
          <option key={value} value={value}>
            {tDocs(`statuses.${value}`)}
          </option>
        ))}
      </select>
    </div>
  )
  /* c8 ignore stop */
})
