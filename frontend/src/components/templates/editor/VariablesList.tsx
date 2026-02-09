'use client'

import { useTranslations } from 'next-intl'
import { Plus, Trash2, GripVertical, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { TemplateVariable, TemplateVariableType } from '@/lib/api/templates'
import { getVariableTypeIcon } from '../pickers'

interface VariablesListProps {
  variables: TemplateVariable[]
  onChange: (variables: TemplateVariable[]) => void
  onAdd: () => void
  onEdit: (variable: TemplateVariable, index: number) => void
  undeclaredVariables?: string[] // Variables found in content but not declared
  unusedVariables?: string[] // Variables declared but not used in content
  disabled?: boolean
}

export function VariablesList({
  variables,
  onChange,
  onAdd,
  onEdit,
  undeclaredVariables = [],
  unusedVariables = [],
  disabled = false,
}: VariablesListProps) {
  const t = useTranslations('templates.editor')

  const handleRemove = (index: number) => {
    const newVariables = [...variables]
    newVariables.splice(index, 1)
    onChange(newVariables)
  }

  const handleMoveUp = (index: number) => {
    if (index === 0) return
    const newVariables = [...variables]
    ;[newVariables[index - 1], newVariables[index]] = [newVariables[index], newVariables[index - 1]]
    onChange(newVariables)
  }

  const handleMoveDown = (index: number) => {
    if (index === variables.length - 1) return
    const newVariables = [...variables]
    ;[newVariables[index], newVariables[index + 1]] = [newVariables[index + 1], newVariables[index]]
    onChange(newVariables)
  }

  const getTypeColor = (type: TemplateVariableType) => {
    const colors: Record<TemplateVariableType, string> = {
      text: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
      date: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
      number: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      select: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
      student: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
      employee: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      user: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
      department: 'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400',
      current_date: 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400',
      current_user: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    }
    return colors[type] || colors.text
  }

  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-900 dark:text-white">{t('variables')}</h4>
        <Button size="sm" variant="outline" onClick={onAdd} disabled={disabled} className="gap-1">
          <Plus className="h-4 w-4" />
          {t('addVariable')}
        </Button>
      </div>

      {/* Undeclared variables warning */}
      {undeclaredVariables.length > 0 && (
        <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
          <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
          <div className="text-sm">
            <p className="font-medium text-amber-800 dark:text-amber-200">
              {t('undeclaredVariables')}
            </p>
            <div className="flex flex-wrap gap-1 mt-1">
              {undeclaredVariables.map((name) => (
                <Badge
                  key={name}
                  variant="outline"
                  className="text-xs border-amber-300 dark:border-amber-700"
                >
                  {`{{${name}}}`}
                </Badge>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Variables list */}
      {variables.length === 0 ? (
        <div className="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
          {t('noVariablesYet')}
        </div>
      ) : (
        <div className="space-y-2">
          {variables.map((variable, index) => {
            const isUnused = unusedVariables.includes(variable.name)
            return (
              <div
                key={`${variable.name}-${index}`}
                className={cn(
                  'flex items-center gap-2 p-3 rounded-lg border transition-colors',
                  'bg-white dark:bg-gray-900/50',
                  isUnused
                    ? 'border-gray-300 dark:border-gray-600 opacity-60'
                    : 'border-gray-200 dark:border-gray-700',
                  !disabled && 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
                )}
              >
                {/* Drag handle */}
                <div className="flex flex-col gap-0.5">
                  <button
                    onClick={() => handleMoveUp(index)}
                    disabled={disabled || index === 0}
                    className="p-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30"
                    title={t('moveUp')}
                  >
                    <GripVertical className="h-3 w-3 rotate-90" />
                  </button>
                  <button
                    onClick={() => handleMoveDown(index)}
                    disabled={disabled || index === variables.length - 1}
                    className="p-0.5 text-gray-400 hover:text-gray-600 disabled:opacity-30"
                    title={t('moveDown')}
                  >
                    <GripVertical className="h-3 w-3 rotate-90" />
                  </button>
                </div>

                {/* Variable info */}
                <button
                  onClick={() => onEdit(variable, index)}
                  disabled={disabled}
                  className="flex-1 text-left"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-base">{getVariableTypeIcon(variable.variable_type)}</span>
                    <span className="font-medium text-gray-900 dark:text-white">
                      {variable.name}
                    </span>
                    {variable.required && <span className="text-red-500 text-xs">*</span>}
                    {isUnused && (
                      <Badge variant="outline" className="text-xs text-gray-500">
                        {t('unused')}
                      </Badge>
                    )}
                  </div>
                  {variable.description && (
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">
                      {variable.description}
                    </p>
                  )}
                </button>

                {/* Type badge */}
                <Badge className={cn('text-xs shrink-0', getTypeColor(variable.variable_type))}>
                  {variable.variable_type}
                </Badge>

                {/* Remove button */}
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleRemove(index)}
                  disabled={disabled}
                  className="h-8 w-8 p-0 text-gray-400 hover:text-red-500"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
