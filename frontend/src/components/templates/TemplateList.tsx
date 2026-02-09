'use client'

import { useState, useEffect, useMemo } from 'react'
import { useTranslations } from 'next-intl'
import { Search, FileText, Loader2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { BlurFade } from '@/components/ui/blur-fade'
import { TemplateCard } from './TemplateCard'
import {
  TemplateCategoryTabs,
  TemplateCategory,
  filterTemplatesByCategory,
  countTemplatesByCategory,
} from './TemplateCategoryTabs'
import { templatesApi, TemplateInfo } from '@/lib/api/templates'
import { useFavoriteTemplates } from '@/hooks/useFavoriteTemplates'

interface TemplateListProps {
  onPreview: (template: TemplateInfo) => void
  onCreate: (template: TemplateInfo) => void
  onEdit?: (template: TemplateInfo) => void
  canEdit?: boolean
}

export function TemplateList({ onPreview, onCreate, onEdit, canEdit = false }: TemplateListProps) {
  const t = useTranslations('templates')
  const [templates, setTemplates] = useState<TemplateInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [activeCategory, setActiveCategory] = useState<TemplateCategory>('all')

  const { favorites, recentlyUsed, isFavorite, toggleFavorite, addToRecent } =
    useFavoriteTemplates()

  useEffect(() => {
    const fetchTemplates = async () => {
      try {
        setIsLoading(true)
        setError(null)
        const data = await templatesApi.getAll()
        setTemplates(data)
      } catch (err) {
        console.error('Failed to fetch templates:', err)
        setError(t('loadError'))
      } finally {
        setIsLoading(false)
      }
    }

    fetchTemplates()
  }, [t])

  // Filter templates by search query
  const searchFilteredTemplates = useMemo(() => {
    if (!searchQuery) return templates
    const query = searchQuery.toLowerCase()
    return templates.filter(
      (template) =>
        template.name.toLowerCase().includes(query) ||
        template.code.toLowerCase().includes(query) ||
        template.description?.toLowerCase().includes(query)
    )
  }, [templates, searchQuery])

  // Filter by category
  const filteredTemplates = useMemo(() => {
    return filterTemplatesByCategory(
      searchFilteredTemplates,
      activeCategory,
      favorites,
      recentlyUsed
    )
  }, [searchFilteredTemplates, activeCategory, favorites, recentlyUsed])

  // Count templates per category
  const categoryCounts = useMemo(() => {
    return countTemplatesByCategory(searchFilteredTemplates, favorites, recentlyUsed)
  }, [searchFilteredTemplates, favorites, recentlyUsed])

  const handleCreate = (template: TemplateInfo) => {
    addToRecent(template.id)
    onCreate(template)
  }

  const handleToggleFavorite = (template: TemplateInfo) => {
    toggleFavorite(template.id)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">{error}</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Category Tabs */}
      <BlurFade delay={0.1}>
        <TemplateCategoryTabs
          activeCategory={activeCategory}
          onCategoryChange={setActiveCategory}
          counts={categoryCounts}
        />
      </BlurFade>

      {/* Search */}
      <BlurFade delay={0.15}>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
          <Input
            placeholder={t('searchPlaceholder')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      </BlurFade>

      {/* Templates Grid */}
      {filteredTemplates.length === 0 ? (
        <BlurFade delay={0.2}>
          <div className="text-center py-12">
            <FileText className="h-12 w-12 mx-auto text-gray-400 mb-4" />
            <p className="text-gray-600 dark:text-gray-400">
              {searchQuery
                ? t('noSearchResults')
                : activeCategory === 'favorites'
                  ? t('noFavorites')
                  : activeCategory === 'recent'
                    ? t('noRecent')
                    : t('noTemplates')}
            </p>
          </div>
        </BlurFade>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {filteredTemplates.map((template, index) => (
            <BlurFade key={template.id} delay={0.2 + index * 0.05} inView>
              <TemplateCard
                template={template}
                onPreview={onPreview}
                onCreate={handleCreate}
                onEdit={onEdit}
                canEdit={canEdit}
                isFavorite={isFavorite(template.id)}
                onToggleFavorite={handleToggleFavorite}
              />
            </BlurFade>
          ))}
        </div>
      )}
    </div>
  )
}
