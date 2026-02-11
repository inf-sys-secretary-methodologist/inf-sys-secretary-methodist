'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'
import { DocumentCategory } from '@/types/document'

interface CategoryFilterProps {
  selectedCategory: DocumentCategory | 'all'
  onCategoryChange: (category: DocumentCategory | 'all') => void
}

export const CategoryFilter = memo(function CategoryFilter({
  selectedCategory,
  onCategoryChange,
}: CategoryFilterProps) {
  const t = useTranslations('documents.filters')
  const tDocs = useTranslations('documents')

  /* c8 ignore start - JSX event handlers, tested in e2e */
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        {t('category')}
      </label>
      <select
        value={selectedCategory}
        onChange={(e) => onCategoryChange(e.target.value as DocumentCategory | 'all')}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                 bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      >
        <option value="all">{t('allCategories')}</option>
        {Object.values(DocumentCategory).map((value) => (
          <option key={value} value={value}>
            {tDocs(`categories.${value}`)}
          </option>
        ))}
      </select>
    </div>
  )
  /* c8 ignore stop */
})
