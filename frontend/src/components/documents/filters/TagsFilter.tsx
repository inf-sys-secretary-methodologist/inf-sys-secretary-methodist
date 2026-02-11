'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'

interface TagsFilterProps {
  tagInput: string
  onTagsChange: (value: string) => void
}

export const TagsFilter = memo(function TagsFilter({ tagInput, onTagsChange }: TagsFilterProps) {
  const t = useTranslations('documents.filters')

  /* c8 ignore start - JSX event handlers, tested in e2e */
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        {t('tags')}
      </label>
      <input
        type="text"
        value={tagInput}
        onChange={(e) => onTagsChange(e.target.value)}
        placeholder={t('tagsPlaceholder')}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                 bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                 focus:ring-2 focus:ring-blue-500 focus:border-transparent
                 placeholder:text-gray-400 dark:placeholder:text-gray-500"
      />
    </div>
  )
  /* c8 ignore stop */
})
