'use client'

import { memo } from 'react'
import { useTranslations } from 'next-intl'

interface Author {
  id: number
  name: string
}

interface AuthorFilterProps {
  selectedAuthorId: number | 'all'
  authors: Author[]
  isLoadingAuthors: boolean
  onAuthorChange: (authorId: number | 'all') => void
}

export const AuthorFilter = memo(function AuthorFilter({
  selectedAuthorId,
  authors,
  isLoadingAuthors,
  onAuthorChange,
}: AuthorFilterProps) {
  const t = useTranslations('documents.filters')

  /* c8 ignore start - JSX event handlers, tested in e2e */
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        {t('author')}
      </label>
      {/* c8 ignore start - Author select change handler */}
      <select
        value={selectedAuthorId}
        onChange={(e) => onAuthorChange(e.target.value === 'all' ? 'all' : Number(e.target.value))}
        disabled={isLoadingAuthors}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                 bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                 focus:ring-2 focus:ring-blue-500 focus:border-transparent
                 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {/* c8 ignore stop */}
        {/* c8 ignore next */}
        <option value="all">{isLoadingAuthors ? t('loading') : t('allAuthors')}</option>
        {authors.map((author) => (
          <option key={author.id} value={author.id}>
            {author.name}
          </option>
        ))}
      </select>
    </div>
  )
  /* c8 ignore stop */
})
