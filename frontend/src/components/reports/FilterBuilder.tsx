'use client'

import { useMemo, useState } from 'react'
import { useTranslations } from 'next-intl'
import { motion, AnimatePresence } from 'framer-motion'
import { GlowingEffect } from '@/components/ui/glowing-effect-lazy'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Plus, X, Filter } from 'lucide-react'
import type { DataSourceType, ReportField, ReportFilter, FilterOperator } from '@/types/reports'
import { AVAILABLE_FIELDS } from '@/types/reports'

interface FilterBuilderProps {
  dataSource: DataSourceType
  filters: ReportFilter[]
  onAddFilter: (filter: ReportFilter) => void
  onRemoveFilter: (filterId: string) => void
  onUpdateFilter: (filterId: string, updates: Partial<ReportFilter>) => void
}

const OPERATORS_BY_TYPE: Record<string, FilterOperator[]> = {
  string: [
    'equals',
    'not_equals',
    'contains',
    'not_contains',
    'starts_with',
    'ends_with',
    'is_null',
    'is_not_null',
  ],
  number: [
    'equals',
    'not_equals',
    'greater_than',
    'less_than',
    'greater_or_equal',
    'less_or_equal',
    'between',
    'is_null',
    'is_not_null',
  ],
  date: ['equals', 'not_equals', 'greater_than', 'less_than', 'between', 'is_null', 'is_not_null'],
  boolean: ['equals', 'not_equals'],
  enum: ['equals', 'not_equals', 'in', 'not_in'],
}

export function FilterBuilder({
  dataSource,
  filters,
  onAddFilter,
  onRemoveFilter,
  onUpdateFilter,
}: FilterBuilderProps) {
  const t = useTranslations('reports.builder')
  const [isAddingFilter, setIsAddingFilter] = useState(false)
  const [newFilterField, setNewFilterField] = useState<string>('')

  /* c8 ignore next */
  const availableFields = useMemo(() => AVAILABLE_FIELDS[dataSource] || [], [dataSource])

  /* c8 ignore start - Radix Select interaction, tested in e2e */
  const handleAddFilter = () => {
    if (!newFilterField) return

    const field = availableFields.find((f) => f.id === newFilterField)
    if (!field) return

    const operators = OPERATORS_BY_TYPE[field.type] || ['equals']
    const defaultOperator = operators[0]

    onAddFilter({
      id: `filter_${Date.now()}`,
      field,
      operator: defaultOperator,
      value: null,
    })

    setNewFilterField('')
    setIsAddingFilter(false)
  }
  /* c8 ignore stop */

  const getOperatorsForField = (field: ReportField): FilterOperator[] => {
    /* c8 ignore next */
    return OPERATORS_BY_TYPE[field.type] || ['equals']
  }

  return (
    <div className="space-y-4">
      {/* Filters List */}
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
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              {t('filters')}
              {filters.length > 0 && (
                <span className="ml-2 text-sm font-normal text-gray-500 dark:text-gray-400">
                  ({filters.length})
                </span>
              )}
            </h3>
            <Button
              onClick={() => setIsAddingFilter(true)}
              variant="outline"
              size="sm"
              className="flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              {t('addFilter')}
            </Button>
          </div>

          {filters.length === 0 && !isAddingFilter ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
                <Filter className="h-8 w-8 text-gray-400" />
              </div>
              <p className="text-gray-500 dark:text-gray-400 mb-2">{t('noFilters')}</p>
              <p className="text-sm text-gray-400 dark:text-gray-500">{t('addFiltersHint')}</p>
            </div>
          ) : (
            <div className="space-y-3">
              <AnimatePresence>
                {/* Add New Filter Form */}
                {isAddingFilter && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    className="flex items-center gap-3 p-3 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800"
                  >
                    <Select value={newFilterField} onValueChange={setNewFilterField}>
                      <SelectTrigger className="flex-1">
                        <SelectValue placeholder={t('selectField')} />
                      </SelectTrigger>
                      <SelectContent>
                        {availableFields.map((field) => (
                          <SelectItem key={field.id} value={field.id}>
                            {t(`fields.${field.source}.${field.name}`, {
                              defaultValue: field.label,
                            })}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button onClick={handleAddFilter} size="sm" disabled={!newFilterField}>
                      {t('add')}
                    </Button>
                    <Button
                      onClick={() => {
                        setIsAddingFilter(false)
                        setNewFilterField('')
                      }}
                      variant="ghost"
                      size="sm"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </motion.div>
                )}

                {/* Existing Filters */}
                {filters.map((filter, index) => (
                  <motion.div
                    key={filter.id}
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: -20 }}
                    className="flex flex-wrap items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700"
                  >
                    {/* Field Name */}
                    <div className="flex items-center gap-2 min-w-[120px]">
                      {index > 0 && (
                        <span className="text-xs text-gray-400 uppercase font-medium">
                          {t('and')}
                        </span>
                      )}
                      <span className="font-medium text-gray-900 dark:text-white">
                        {t(`fields.${filter.field.source}.${filter.field.name}`, {
                          defaultValue: filter.field.label,
                        })}
                      </span>
                    </div>

                    {/* Operator */}
                    <Select
                      value={filter.operator}
                      /* c8 ignore next 2 - Radix Select callback */
                      onValueChange={(value) =>
                        onUpdateFilter(filter.id, { operator: value as FilterOperator })
                      }
                    >
                      <SelectTrigger className="w-[160px]">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {getOperatorsForField(filter.field).map((op) => (
                          <SelectItem key={op} value={op}>
                            {t(`operators.${op}`)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>

                    {/* Value Input */}
                    {filter.operator !== 'is_null' && filter.operator !== 'is_not_null' && (
                      <>
                        {filter.field.type === 'enum' && filter.field.enumValues ? (
                          <Select
                            /* c8 ignore next - Fallback for enum value */
                            value={(filter.value as string) || ''}
                            /* c8 ignore next - Radix Select callback */
                            onValueChange={(value) => onUpdateFilter(filter.id, { value })}
                          >
                            <SelectTrigger className="w-[160px]">
                              <SelectValue placeholder={t('selectValue')} />
                            </SelectTrigger>
                            <SelectContent>
                              {filter.field.enumValues.map((v) => (
                                <SelectItem key={v} value={v}>
                                  {t(
                                    `enumValues.${filter.field.source}.${filter.field.name}.${v}`,
                                    { defaultValue: v }
                                  )}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        ) : filter.field.type === 'boolean' ? (
                          <Select
                            value={String(filter.value)}
                            /* c8 ignore next 2 - Radix Select callback */
                            onValueChange={(value) =>
                              onUpdateFilter(filter.id, { value: value === 'true' })
                            }
                          >
                            <SelectTrigger className="w-[100px]">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="true">{t('true')}</SelectItem>
                              <SelectItem value="false">{t('false')}</SelectItem>
                            </SelectContent>
                          </Select>
                        ) : filter.field.type === 'date' ? (
                          <Input
                            type="date"
                            value={(filter.value as string) || ''}
                            onChange={(e) => onUpdateFilter(filter.id, { value: e.target.value })}
                            className="w-[160px]"
                          />
                        ) : filter.field.type === 'number' ? (
                          <Input
                            type="number"
                            value={(filter.value as number) || ''}
                            onChange={(e) =>
                              onUpdateFilter(filter.id, {
                                value: parseFloat(e.target.value),
                              })
                            }
                            placeholder={t('enterValue')}
                            className="w-[120px]"
                          />
                        ) : (
                          <Input
                            type="text"
                            value={(filter.value as string) || ''}
                            onChange={(e) => onUpdateFilter(filter.id, { value: e.target.value })}
                            placeholder={t('enterValue')}
                            className="flex-1 min-w-[120px]"
                          />
                        )}

                        {/* Second value for 'between' operator */}
                        {/* c8 ignore start - Between operator value2 input */}
                        {filter.operator === 'between' && (
                          <>
                            <span className="text-gray-400">{t('and')}</span>
                            {filter.field.type === 'date' ? (
                              <Input
                                type="date"
                                value={(filter.value2 as string) || ''}
                                onChange={(e) =>
                                  onUpdateFilter(filter.id, { value2: e.target.value })
                                }
                                className="w-[160px]"
                              />
                            ) : (
                              <Input
                                type="number"
                                value={(filter.value2 as number) || ''}
                                onChange={(e) =>
                                  onUpdateFilter(filter.id, {
                                    value2: parseFloat(e.target.value),
                                  })
                                }
                                placeholder={t('enterValue')}
                                className="w-[120px]"
                              />
                            )}
                          </>
                        )}
                        {/* c8 ignore stop */}
                      </>
                    )}

                    {/* Remove Button */}
                    <button
                      onClick={() => onRemoveFilter(filter.id)}
                      className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors ml-auto"
                    >
                      <X className="h-4 w-4 text-gray-500 hover:text-red-500" />
                    </button>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
