'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Search, FileText, Loader2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { TemplateCard } from './TemplateCard'
import { templatesApi, TemplateInfo } from '@/lib/api/templates'

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

  const filteredTemplates = templates.filter((template) => {
    const query = searchQuery.toLowerCase()
    return (
      template.name.toLowerCase().includes(query) ||
      template.code.toLowerCase().includes(query) ||
      template.description?.toLowerCase().includes(query)
    )
  })

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
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
        <Input
          placeholder={t('searchPlaceholder')}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10"
        />
      </div>

      {/* Templates Grid */}
      {filteredTemplates.length === 0 ? (
        <div className="text-center py-12">
          <FileText className="h-12 w-12 mx-auto text-gray-400 mb-4" />
          <p className="text-gray-600 dark:text-gray-400">
            {searchQuery ? t('noSearchResults') : t('noTemplates')}
          </p>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {filteredTemplates.map((template) => (
            <TemplateCard
              key={template.id}
              template={template}
              onPreview={onPreview}
              onCreate={onCreate}
              onEdit={onEdit}
              canEdit={canEdit}
            />
          ))}
        </div>
      )}
    </div>
  )
}
