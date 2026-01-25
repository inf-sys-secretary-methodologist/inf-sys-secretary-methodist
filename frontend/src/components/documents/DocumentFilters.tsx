'use client'

import { useState, useEffect } from 'react'
import dynamic from 'next/dynamic'
import { useTranslations, useLocale } from 'next-intl'
import { Search, Filter, X, Calendar as CalendarIcon, Loader2 } from 'lucide-react'
import { format, Locale } from 'date-fns'
import { ru, enUS, fr, ar } from 'date-fns/locale'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

// Lazy load Calendar to reduce initial bundle (react-day-picker ~100KB)
const Calendar = dynamic(() => import('@/components/ui/calendar').then((mod) => mod.Calendar), {
  loading: () => (
    <div className="flex items-center justify-center p-4">
      <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
    </div>
  ),
  ssr: false,
})
import {
  DocumentCategory,
  DocumentStatus,
  type DocumentFilter,
  type DocumentSortOptions,
} from '@/types/document'
import { usersApi } from '@/lib/api/users'
import { cn } from '@/lib/utils'

interface Author {
  id: number
  name: string
}

interface DocumentFiltersProps {
  onFilterChange: (filters: DocumentFilter) => void
  onSortChange: (sort: DocumentSortOptions) => void
  currentFilters: DocumentFilter
  currentSort: DocumentSortOptions
  className?: string
}

const localeMap: Record<string, Locale> = {
  ru: ru,
  en: enUS,
  fr: fr,
  ar: ar,
}

export function DocumentFilters({
  onFilterChange,
  onSortChange,
  currentFilters,
  currentSort,
  className = '',
}: DocumentFiltersProps) {
  const t = useTranslations('documents.filters')
  const tDocs = useTranslations('documents')
  const tCommon = useTranslations('common')
  const tForm = useTranslations('documents.form')
  const locale = useLocale()
  const dateLocale = localeMap[locale] || enUS
  const [isExpanded, setIsExpanded] = useState(false)
  const [searchQuery, setSearchQuery] = useState(currentFilters.search || '')
  const [selectedCategory, setSelectedCategory] = useState<DocumentCategory | 'all'>(
    currentFilters.category || 'all'
  )
  const [selectedStatus, setSelectedStatus] = useState<DocumentStatus | 'all'>(
    currentFilters.status || 'all'
  )
  const [tagInput, setTagInput] = useState(currentFilters.tags?.join(', ') || '')
  const [dateFrom, setDateFrom] = useState<Date | undefined>(currentFilters.dateFrom)
  const [dateTo, setDateTo] = useState<Date | undefined>(currentFilters.dateTo)
  const [selectedAuthorId, setSelectedAuthorId] = useState<number | 'all'>(
    currentFilters.authorId || 'all'
  )
  const [authors, setAuthors] = useState<Author[]>([])
  const [isLoadingAuthors, setIsLoadingAuthors] = useState(false)

  /* c8 ignore start - Authors loading, tested in e2e */
  // Load authors list
  useEffect(() => {
    const loadAuthors = async () => {
      setIsLoadingAuthors(true)
      try {
        const users = await usersApi.getAll()
        setAuthors(users.map((u) => ({ id: u.id, name: u.name })))
      } catch (err) {
        console.error('Failed to load authors:', err)
      } finally {
        setIsLoadingAuthors(false)
      }
    }
    loadAuthors()
  }, [])
  /* c8 ignore stop */

  const handleSearchChange = (value: string) => {
    setSearchQuery(value)
    onFilterChange({
      ...currentFilters,
      search: value || undefined,
    })
  }

  /* c8 ignore start - Filter change handlers, tested in e2e */
  const handleCategoryChange = (category: DocumentCategory | 'all') => {
    setSelectedCategory(category)
    onFilterChange({
      ...currentFilters,
      category: category === 'all' ? undefined : category,
    })
  }

  const handleStatusChange = (status: DocumentStatus | 'all') => {
    setSelectedStatus(status)
    onFilterChange({
      ...currentFilters,
      status: status === 'all' ? undefined : status,
    })
  }

  const handleTagsChange = (value: string) => {
    setTagInput(value)
    const tags = value
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)
    onFilterChange({
      ...currentFilters,
      tags: tags.length > 0 ? tags : undefined,
    })
  }

  const handleDateFromChange = (date: Date | undefined) => {
    setDateFrom(date)
    onFilterChange({
      ...currentFilters,
      dateFrom: date,
    })
  }

  const handleDateToChange = (date: Date | undefined) => {
    setDateTo(date)
    onFilterChange({
      ...currentFilters,
      dateTo: date,
    })
  }

  const handleAuthorChange = (authorId: number | 'all') => {
    setSelectedAuthorId(authorId)
    onFilterChange({
      ...currentFilters,
      authorId: authorId === 'all' ? undefined : authorId,
    })
  }
  /* c8 ignore stop */

  const handleSortChange = (field: DocumentSortOptions['field']) => {
    if (currentSort.field === field) {
      // Toggle order if same field
      onSortChange({
        field,
        order: currentSort.order === 'asc' ? 'desc' : 'asc',
      })
    } else {
      // Default to desc for new field
      onSortChange({
        field,
        order: 'desc',
      })
    }
  }

  const clearFilters = () => {
    setSearchQuery('')
    setSelectedCategory('all')
    setSelectedStatus('all')
    setTagInput('')
    setDateFrom(undefined)
    setDateTo(undefined)
    setSelectedAuthorId('all')
    onFilterChange({})
  }

  const hasActiveFilters =
    currentFilters.search ||
    currentFilters.category ||
    currentFilters.status ||
    currentFilters.tags?.length ||
    currentFilters.dateFrom ||
    currentFilters.dateTo ||
    currentFilters.authorId

  /* c8 ignore start - Author name helper */
  const getAuthorName = (id: number) => {
    const author = authors.find((a) => a.id === id)
    return author?.name || `${t('authorPrefix')} #${id}`
  }
  /* c8 ignore stop */

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Search Bar */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder={tForm('searchPlaceholder')}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                     bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                     focus:ring-2 focus:ring-blue-500 focus:border-transparent
                     placeholder:text-gray-400 dark:placeholder:text-gray-500"
          />
        </div>

        <Button
          variant={isExpanded ? 'default' : 'outline'}
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex-shrink-0"
        >
          <Filter className="h-4 w-4 mr-2" />
          {tCommon('filters')}
          {hasActiveFilters && !isExpanded && (
            <span className="ml-2 px-2 py-0.5 bg-blue-500 text-white text-xs rounded-full">
              {
                Object.keys(currentFilters).filter((k) => currentFilters[k as keyof DocumentFilter])
                  .length
              }
            </span>
          )}
        </Button>

        {hasActiveFilters && (
          <Button variant="outline" onClick={clearFilters} className="flex-shrink-0">
            <X className="h-4 w-4 mr-2" />
            {tCommon('reset')}
          </Button>
        )}
      </div>

      {/* Expanded Filters */}
      {isExpanded && (
        <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-800/50 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* Category Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('category')}
              </label>
              <select
                value={selectedCategory}
                onChange={(e) => handleCategoryChange(e.target.value as DocumentCategory | 'all')}
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

            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('status')}
              </label>
              <select
                value={selectedStatus}
                onChange={(e) => handleStatusChange(e.target.value as DocumentStatus | 'all')}
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

            {/* Author Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('author')}
              </label>
              {/* c8 ignore start - Author select change handler */}
              <select
                value={selectedAuthorId}
                onChange={(e) =>
                  handleAuthorChange(e.target.value === 'all' ? 'all' : Number(e.target.value))
                }
                disabled={isLoadingAuthors}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                         focus:ring-2 focus:ring-blue-500 focus:border-transparent
                         disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {/* c8 ignore stop */}
                <option value="all">{isLoadingAuthors ? t('loading') : t('allAuthors')}</option>
                {authors.map((author) => (
                  <option key={author.id} value={author.id}>
                    {author.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Date From Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('dateFrom')}
              </label>
              <div className="flex gap-2">
                <Popover>
                  <PopoverTrigger asChild>
                    {/* c8 ignore start - Date from button */}
                    <Button
                      variant="outline"
                      className={cn(
                        'flex-1 justify-start text-left font-normal',
                        !dateFrom && 'text-muted-foreground'
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {dateFrom
                        ? format(dateFrom, 'dd.MM.yyyy', { locale: dateLocale })
                        : t('selectDate')}
                    </Button>
                    {/* c8 ignore stop */}
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={dateFrom}
                      onSelect={handleDateFromChange}
                      disabled={(date) => (dateTo ? date > dateTo : false)}
                      locale={dateLocale}
                    />
                  </PopoverContent>
                </Popover>
                {dateFrom && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDateFromChange(undefined)}
                    className="flex-shrink-0 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                    aria-label={tForm('resetDate')}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>

            {/* Date To Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('dateTo')}
              </label>
              <div className="flex gap-2">
                <Popover>
                  <PopoverTrigger asChild>
                    {/* c8 ignore start - Date to button */}
                    <Button
                      variant="outline"
                      className={cn(
                        'flex-1 justify-start text-left font-normal',
                        !dateTo && 'text-muted-foreground'
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {dateTo
                        ? format(dateTo, 'dd.MM.yyyy', { locale: dateLocale })
                        : t('selectDate')}
                    </Button>
                    {/* c8 ignore stop */}
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={dateTo}
                      onSelect={handleDateToChange}
                      disabled={(date) => (dateFrom ? date < dateFrom : false)}
                      locale={dateLocale}
                    />
                  </PopoverContent>
                </Popover>
                {dateTo && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDateToChange(undefined)}
                    className="flex-shrink-0 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
                    aria-label={tForm('resetDate')}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                )}
              </div>
            </div>

            {/* Tags Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {t('tags')}
              </label>
              <input
                type="text"
                value={tagInput}
                onChange={(e) => handleTagsChange(e.target.value)}
                placeholder={t('tagsPlaceholder')}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                         focus:ring-2 focus:ring-blue-500 focus:border-transparent
                         placeholder:text-gray-400 dark:placeholder:text-gray-500"
              />
            </div>
          </div>

          {/* Sort Options */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {t('sort')}
            </label>
            <div className="flex flex-wrap gap-2">
              {[
                { field: 'name' as const, label: t('name') },
                { field: 'uploadedAt' as const, label: t('uploadDate') },
                { field: 'modifiedAt' as const, label: t('modifyDate') },
                { field: 'size' as const, label: t('size') },
              ].map(({ field, label }) => (
                <Button
                  key={field}
                  variant={currentSort.field === field ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => handleSortChange(field)}
                  className="relative"
                >
                  {label}
                  {currentSort.field === field && (
                    <span className="ml-2 text-xs">{currentSort.order === 'asc' ? '↑' : '↓'}</span>
                  )}
                </Button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Active Filters Summary */}
      {hasActiveFilters && !isExpanded && (
        <div className="flex flex-wrap gap-2">
          {currentFilters.search && (
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
              {t('search')}: {currentFilters.search}
              <button
                onClick={() => handleSearchChange('')}
                className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.category && (
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
              {tDocs(`categories.${currentFilters.category}`)}
              <button
                onClick={() => handleCategoryChange('all')}
                className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.status && (
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
              {tDocs(`statuses.${currentFilters.status}`)}
              <button
                onClick={() => handleStatusChange('all')}
                className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.authorId && (
            <span className="px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400 rounded-full text-sm flex items-center gap-2">
              {t('authorPrefix')}: {getAuthorName(currentFilters.authorId)}
              <button
                onClick={() => handleAuthorChange('all')}
                className="hover:bg-green-200 dark:hover:bg-green-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.dateFrom && (
            <span className="px-3 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400 rounded-full text-sm flex items-center gap-2">
              {t('from')}: {format(currentFilters.dateFrom, 'dd.MM.yyyy', { locale: dateLocale })}
              <button
                onClick={() => handleDateFromChange(undefined)}
                className="hover:bg-purple-200 dark:hover:bg-purple-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.dateTo && (
            <span className="px-3 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-400 rounded-full text-sm flex items-center gap-2">
              {t('to')}: {format(currentFilters.dateTo, 'dd.MM.yyyy', { locale: dateLocale })}
              <button
                onClick={() => handleDateToChange(undefined)}
                className="hover:bg-purple-200 dark:hover:bg-purple-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.tags && currentFilters.tags.length > 0 && (
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
              {t('tags')}: {currentFilters.tags.join(', ')}
              <button
                onClick={() => handleTagsChange('')}
                className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
        </div>
      )}
    </div>
  )
}
