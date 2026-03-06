'use client'

import { useMemo, useState } from 'react'
import { useTranslations } from 'next-intl'
import { motion, AnimatePresence, Reorder } from 'motion/react'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { GripVertical, Plus, X, Hash, Type, Calendar, ToggleLeft, List, Search } from 'lucide-react'
import type { DataSourceType, ReportField, SelectedField } from '@/types/reports'
import { AVAILABLE_FIELDS } from '@/types/reports'

interface FieldSelectorProps {
  dataSource: DataSourceType
  selectedFields: SelectedField[]
  onAddField: (field: ReportField) => void
  onRemoveField: (fieldId: string) => void
  onReorderFields: (fields: SelectedField[]) => void
}

const TYPE_ICONS: Record<string, React.ElementType> = {
  string: Type,
  number: Hash,
  date: Calendar,
  boolean: ToggleLeft,
  enum: List,
}

export function FieldSelector({
  dataSource,
  selectedFields,
  onAddField,
  onRemoveField,
  onReorderFields,
}: FieldSelectorProps) {
  const t = useTranslations('reports.builder')
  const [searchQuery, setSearchQuery] = useState('')

  const availableFields = useMemo(() => {
    /* c8 ignore next */
    const fields = AVAILABLE_FIELDS[dataSource] || []
    const selectedIds = new Set(selectedFields.map((f) => f.field.id))
    const unselected = fields.filter((f) => !selectedIds.has(f.id))

    if (!searchQuery) return unselected

    const query = searchQuery.toLowerCase()
    return unselected.filter(
      (f) => f.label.toLowerCase().includes(query) || f.name.toLowerCase().includes(query)
    )
  }, [dataSource, selectedFields, searchQuery])

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Available Fields */}
      <div className="relative overflow-hidden rounded-xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
        <GlowingEffect
          spread={40}
          glow={true}
          disabled={false}
          proximity={64}
          inactiveZone={0.01}
          borderWidth={2}
        />
        <div className="relative z-10">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('availableFields')}
          </h3>

          {/* Search */}
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={t('searchFields')}
              className="w-full pl-10 pr-4 py-2 rounded-lg bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm focus:outline-none focus:ring-2 focus:ring-gray-300 dark:focus:ring-gray-600"
            />
          </div>

          {/* Field List */}
          <div className="space-y-2 max-h-[400px] overflow-y-auto">
            <AnimatePresence>
              {availableFields.length === 0 ? (
                <motion.p
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="text-center text-gray-500 dark:text-gray-400 py-8"
                >
                  {searchQuery ? t('noFieldsFound') : t('allFieldsSelected')}
                </motion.p>
              ) : (
                availableFields.map((field) => {
                  /* c8 ignore next */
                  const Icon = TYPE_ICONS[field.type] || Type
                  return (
                    <motion.button
                      key={field.id}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, x: -20 }}
                      onClick={() => onAddField(field)}
                      className="group w-full flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 border border-transparent hover:border-gray-200 dark:hover:border-gray-700 transition-all text-left"
                    >
                      <div className="w-8 h-8 rounded-md flex items-center justify-center bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300">
                        <Icon className="h-4 w-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-gray-900 dark:text-white truncate">
                          {t(`fields.${field.source}.${field.name}`, { defaultValue: field.label })}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {field.type}
                          {/* c8 ignore next */}
                          {field.enumValues && ` (${field.enumValues.length} ${t('options')})`}
                        </p>
                      </div>
                      <div className="opacity-0 group-hover:opacity-100 transition-opacity">
                        <Plus className="h-5 w-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300" />
                      </div>
                    </motion.button>
                  )
                })
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      {/* Selected Fields */}
      <div className="relative overflow-hidden rounded-xl p-4 sm:p-6 bg-white dark:bg-black/95 border border-gray-200 dark:border-gray-700">
        <GlowingEffect
          spread={40}
          glow={true}
          disabled={false}
          proximity={64}
          inactiveZone={0.01}
          borderWidth={2}
        />
        <div className="relative z-10">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('selectedFields')}
            {selectedFields.length > 0 && (
              <span className="ml-2 text-sm font-normal text-gray-500 dark:text-gray-400">
                ({selectedFields.length})
              </span>
            )}
          </h3>

          {selectedFields.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
                <Plus className="h-8 w-8 text-gray-400" />
              </div>
              <p className="text-gray-500 dark:text-gray-400 mb-2">{t('noFieldsSelected')}</p>
              <p className="text-sm text-gray-400 dark:text-gray-500">{t('clickToAddFields')}</p>
            </div>
          ) : (
            <Reorder.Group
              axis="y"
              values={selectedFields}
              onReorder={onReorderFields}
              className="space-y-2"
            >
              <AnimatePresence>
                {selectedFields.map((selected) => {
                  /* c8 ignore next */
                  const Icon = TYPE_ICONS[selected.field.type] || Type
                  return (
                    <Reorder.Item
                      key={selected.field.id}
                      value={selected}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      exit={{ opacity: 0, x: 20 }}
                      className="group flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 cursor-grab active:cursor-grabbing"
                    >
                      <GripVertical className="h-4 w-4 text-gray-400 flex-shrink-0" />
                      <div className="w-8 h-8 rounded-md flex items-center justify-center bg-gray-900 dark:bg-white text-white dark:text-gray-900 flex-shrink-0">
                        <Icon className="h-4 w-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-gray-900 dark:text-white truncate">
                          {t(`fields.${selected.field.source}.${selected.field.name}`, {
                            defaultValue: selected.field.label,
                          })}
                        </p>
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          {selected.field.type}
                        </p>
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          onRemoveField(selected.field.id)
                        }}
                        className="opacity-0 group-hover:opacity-100 p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-all"
                      >
                        <X className="h-4 w-4 text-gray-500 hover:text-red-500" />
                      </button>
                    </Reorder.Item>
                  )
                })}
              </AnimatePresence>
            </Reorder.Group>
          )}
        </div>
      </div>
    </div>
  )
}
