'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { TemplateVariable, TemplateVariableType } from '@/lib/api/templates'
import { getVariableTypeIcon } from '../pickers'
import { cn } from '@/lib/utils'

interface AddVariableDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (variable: TemplateVariable) => void
  editVariable?: TemplateVariable | null
  existingNames?: string[]
}

const VARIABLE_TYPES: { value: TemplateVariableType; label: string; description: string }[] = [
  { value: 'text', label: 'Text', description: 'Free text input' },
  { value: 'number', label: 'Number', description: 'Numeric value' },
  { value: 'date', label: 'Date', description: 'Date picker' },
  { value: 'select', label: 'Select', description: 'Dropdown with options' },
  { value: 'student', label: 'Student', description: 'Select from students database' },
  { value: 'employee', label: 'Employee', description: 'Select from employees database' },
  { value: 'department', label: 'Department', description: 'Select from departments' },
  { value: 'current_date', label: 'Auto Date', description: 'Auto-fill with current date' },
  { value: 'current_user', label: 'Current User', description: 'Auto-fill with current user' },
]

export function AddVariableDialog({
  open,
  onOpenChange,
  onSave,
  editVariable,
  existingNames = [],
}: AddVariableDialogProps) {
  const t = useTranslations('templates.editor')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [variableType, setVariableType] = useState<TemplateVariableType>('text')
  const [required, setRequired] = useState(false)
  const [defaultValue, setDefaultValue] = useState('')
  const [options, setOptions] = useState('')
  const [dataField, setDataField] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Reset form when dialog opens/closes or editVariable changes
  useEffect(() => {
    if (open) {
      if (editVariable) {
        setName(editVariable.name)
        setDescription(editVariable.description || '')
        setVariableType(editVariable.variable_type)
        setRequired(editVariable.required)
        setDefaultValue(editVariable.default_value || '')
        setOptions(editVariable.options?.join(', ') || '')
        setDataField(editVariable.data_field || '')
      } else {
        setName('')
        setDescription('')
        setVariableType('text')
        setRequired(false)
        setDefaultValue('')
        setOptions('')
        setDataField('')
      }
      setErrors({})
    }
  }, [open, editVariable])

  const validateName = (value: string) => {
    if (!value.trim()) {
      return t('nameRequired')
    }
    if (!/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(value)) {
      return t('nameInvalid')
    }
    if (existingNames.includes(value) && (!editVariable || editVariable.name !== value)) {
      return t('nameDuplicate')
    }
    return null
  }

  const handleSave = () => {
    const nameError = validateName(name)
    if (nameError) {
      setErrors({ name: nameError })
      return
    }

    const variable: TemplateVariable = {
      name: name.trim(),
      description: description.trim() || undefined,
      variable_type: variableType,
      required,
      default_value: defaultValue.trim() || undefined,
      options:
        variableType === 'select' && options.trim()
          ? options
              .split(',')
              .map((o) => o.trim())
              .filter(Boolean)
          : undefined,
      data_field: ['student', 'employee', 'department', 'user'].includes(variableType)
        ? dataField.trim() || undefined
        : undefined,
    }

    onSave(variable)
    onOpenChange(false)
  }

  const isSmartType = ['student', 'employee', 'department', 'user'].includes(variableType)
  const isAutoType = ['current_date', 'current_user'].includes(variableType)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{editVariable ? t('editVariable') : t('addVariable')}</DialogTitle>
          <DialogDescription>{t('variableDialogDescription')}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Name */}
          <div className="space-y-2">
            <Label htmlFor="var-name">{t('variableName')} *</Label>
            <Input
              id="var-name"
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                setErrors((prev) => ({ ...prev, name: '' }))
              }}
              placeholder="employee_name"
              className={cn(errors.name && 'border-red-500')}
            />
            {errors.name && <p className="text-xs text-red-500">{errors.name}</p>}
            <p className="text-xs text-gray-500">{t('nameHint')}</p>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="var-description">{t('variableDescription')}</Label>
            <Input
              id="var-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('descriptionPlaceholder')}
            />
          </div>

          {/* Type */}
          <div className="space-y-2">
            <Label>{t('variableType')} *</Label>
            <div className="grid grid-cols-3 gap-2">
              {VARIABLE_TYPES.map((type) => (
                <button
                  key={type.value}
                  onClick={() => setVariableType(type.value)}
                  className={cn(
                    'flex flex-col items-center gap-1 p-3 rounded-lg border transition-all text-center',
                    variableType === type.value
                      ? 'border-primary bg-primary/5 ring-1 ring-primary'
                      : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                  )}
                >
                  <span className="text-lg">{getVariableTypeIcon(type.value)}</span>
                  <span className="text-xs font-medium">{type.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Options for select type */}
          {variableType === 'select' && (
            <div className="space-y-2">
              <Label htmlFor="var-options">{t('selectOptions')} *</Label>
              <Input
                id="var-options"
                value={options}
                onChange={(e) => setOptions(e.target.value)}
                placeholder={t('optionsPlaceholder')}
              />
              <p className="text-xs text-gray-500">{t('optionsHint')}</p>
            </div>
          )}

          {/* Data field for smart types */}
          {isSmartType && (
            <div className="space-y-2">
              <Label htmlFor="var-datafield">{t('dataField')}</Label>
              <Input
                id="var-datafield"
                value={dataField}
                onChange={(e) => setDataField(e.target.value)}
                placeholder="full_name"
              />
              <p className="text-xs text-gray-500">{t('dataFieldHint')}</p>
            </div>
          )}

          {/* Default value (not for auto types) */}
          {!isAutoType && (
            <div className="space-y-2">
              <Label htmlFor="var-default">{t('defaultValue')}</Label>
              <Input
                id="var-default"
                value={defaultValue}
                onChange={(e) => setDefaultValue(e.target.value)}
                placeholder={t('defaultPlaceholder')}
              />
            </div>
          )}

          {/* Required checkbox (not for auto types) */}
          {!isAutoType && (
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="var-required"
                checked={required}
                onChange={(e) => setRequired(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
              />
              <Label htmlFor="var-required" className="cursor-pointer">
                {t('requiredField')}
              </Label>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('cancel')}
          </Button>
          <Button onClick={handleSave}>{editVariable ? t('save') : t('add')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
