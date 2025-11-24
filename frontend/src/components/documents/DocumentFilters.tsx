'use client'

import { useState } from 'react'
import { Search, Filter, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DocumentCategory,
  DocumentStatus,
  DocumentCategoryLabels,
  DocumentStatusLabels,
  type DocumentFilter,
  type DocumentSortOptions
} from '@/types/document'

interface DocumentFiltersProps {
  onFilterChange: (filters: DocumentFilter) => void
  onSortChange: (sort: DocumentSortOptions) => void
  currentFilters: DocumentFilter
  currentSort: DocumentSortOptions
  className?: string
}

export function DocumentFilters({
  onFilterChange,
  onSortChange,
  currentFilters,
  currentSort,
  className = ''
}: DocumentFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [searchQuery, setSearchQuery] = useState(currentFilters.search || '')
  const [selectedCategory, setSelectedCategory] = useState<DocumentCategory | 'all'>(
    currentFilters.category || 'all'
  )
  const [selectedStatus, setSelectedStatus] = useState<DocumentStatus | 'all'>(
    currentFilters.status || 'all'
  )
  const [tagInput, setTagInput] = useState(currentFilters.tags?.join(', ') || '')

  const handleSearchChange = (value: string) => {
    setSearchQuery(value)
    onFilterChange({
      ...currentFilters,
      search: value || undefined
    })
  }

  const handleCategoryChange = (category: DocumentCategory | 'all') => {
    setSelectedCategory(category)
    onFilterChange({
      ...currentFilters,
      category: category === 'all' ? undefined : category
    })
  }

  const handleStatusChange = (status: DocumentStatus | 'all') => {
    setSelectedStatus(status)
    onFilterChange({
      ...currentFilters,
      status: status === 'all' ? undefined : status
    })
  }

  const handleTagsChange = (value: string) => {
    setTagInput(value)
    const tags = value
      .split(',')
      .map(t => t.trim())
      .filter(Boolean)
    onFilterChange({
      ...currentFilters,
      tags: tags.length > 0 ? tags : undefined
    })
  }

  const handleSortChange = (field: DocumentSortOptions['field']) => {
    if (currentSort.field === field) {
      // Toggle order if same field
      onSortChange({
        field,
        order: currentSort.order === 'asc' ? 'desc' : 'asc'
      })
    } else {
      // Default to desc for new field
      onSortChange({
        field,
        order: 'desc'
      })
    }
  }

  const clearFilters = () => {
    setSearchQuery('')
    setSelectedCategory('all')
    setSelectedStatus('all')
    setTagInput('')
    onFilterChange({})
  }

  const hasActiveFilters =
    currentFilters.search ||
    currentFilters.category ||
    currentFilters.status ||
    currentFilters.tags?.length

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
            placeholder="Поиск документов..."
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
          Фильтры
          {hasActiveFilters && !isExpanded && (
            <span className="ml-2 px-2 py-0.5 bg-blue-500 text-white text-xs rounded-full">
              {Object.keys(currentFilters).length}
            </span>
          )}
        </Button>

        {hasActiveFilters && (
          <Button
            variant="outline"
            onClick={clearFilters}
            className="flex-shrink-0"
          >
            <X className="h-4 w-4 mr-2" />
            Сбросить
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
                Категория
              </label>
              <select
                value={selectedCategory}
                onChange={(e) => handleCategoryChange(e.target.value as DocumentCategory | 'all')}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                         focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="all">Все категории</option>
                {Object.entries(DocumentCategoryLabels).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Статус
              </label>
              <select
                value={selectedStatus}
                onChange={(e) => handleStatusChange(e.target.value as DocumentStatus | 'all')}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg
                         bg-white dark:bg-gray-900 text-gray-900 dark:text-white
                         focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="all">Все статусы</option>
                {Object.entries(DocumentStatusLabels).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            {/* Tags Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Теги
              </label>
              <input
                type="text"
                value={tagInput}
                onChange={(e) => handleTagsChange(e.target.value)}
                placeholder="Введите теги через запятую..."
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
              Сортировка
            </label>
            <div className="flex flex-wrap gap-2">
              {[
                { field: 'name' as const, label: 'Название' },
                { field: 'uploadedAt' as const, label: 'Дата загрузки' },
                { field: 'modifiedAt' as const, label: 'Дата изменения' },
                { field: 'size' as const, label: 'Размер' }
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
                    <span className="ml-2 text-xs">
                      {currentSort.order === 'asc' ? '↑' : '↓'}
                    </span>
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
              Поиск: {currentFilters.search}
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
              {DocumentCategoryLabels[currentFilters.category]}
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
              {DocumentStatusLabels[currentFilters.status]}
              <button
                onClick={() => handleStatusChange('all')}
                className="hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-full p-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}
          {currentFilters.tags && currentFilters.tags.length > 0 && (
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full text-sm flex items-center gap-2">
              Теги: {currentFilters.tags.join(', ')}
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
