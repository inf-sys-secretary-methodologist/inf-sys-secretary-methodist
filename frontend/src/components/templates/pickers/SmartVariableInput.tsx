'use client'

import { useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Calendar, User } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { TemplateVariable } from '@/lib/api/templates'
import { StudentPicker } from './StudentPicker'
import { EmployeePicker } from './EmployeePicker'
import { DepartmentPicker } from './DepartmentPicker'
import { useAuthStore } from '@/stores/authStore'

interface SmartVariableInputProps {
  variable: TemplateVariable
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

export function SmartVariableInput({
  variable,
  value,
  onChange,
  disabled = false,
}: SmartVariableInputProps) {
  const t = useTranslations('templates')
  const user = useAuthStore((state) => state.user)

  // Auto-fill current_date and current_user on mount
  useEffect(() => {
    if (variable.variable_type === 'current_date' && !value) {
      const today = new Date().toISOString().split('T')[0]
      onChange(today)
    }
    if (variable.variable_type === 'current_user' && !value && user) {
      onChange(user.name || user.email)
    }
  }, [variable.variable_type, value, onChange, user])

  // Render based on variable type
  switch (variable.variable_type) {
    case 'student':
      return (
        <StudentPicker
          value={value}
          onChange={(val) => onChange(val)}
          dataField={variable.data_field || 'full_name'}
          disabled={disabled}
        />
      )

    case 'employee':
      return (
        <EmployeePicker
          value={value}
          onChange={(val) => onChange(val)}
          dataField={variable.data_field || 'full_name'}
          disabled={disabled}
        />
      )

    case 'department':
      return (
        <DepartmentPicker
          value={value}
          onChange={(val) => onChange(val)}
          dataField={variable.data_field || 'name'}
          disabled={disabled}
        />
      )

    case 'current_date':
      return (
        <div className="relative">
          <Input
            type="date"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            disabled={disabled}
            className="pr-24"
          />
          <Badge
            variant="secondary"
            className="absolute right-2 top-1/2 -translate-y-1/2 text-xs gap-1"
          >
            <Calendar className="h-3 w-3" />
            {t('autoFilled')}
          </Badge>
        </div>
      )

    case 'current_user':
      return (
        <div className="relative">
          <Input
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            disabled={disabled}
            className="pr-24"
          />
          <Badge
            variant="secondary"
            className="absolute right-2 top-1/2 -translate-y-1/2 text-xs gap-1"
          >
            <User className="h-3 w-3" />
            {t('autoFilled')}
          </Badge>
        </div>
      )

    case 'select':
      if (variable.options && variable.options.length > 0) {
        return (
          <select
            value={value}
            onChange={(e) => onChange(e.target.value)}
            disabled={disabled}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          >
            <option value="">{t('selectOption')}</option>
            {variable.options.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
        )
      }
      return (
        <Input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          placeholder={variable.description || ''}
        />
      )

    case 'date':
      return (
        <Input
          type="date"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
        />
      )

    case 'number':
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          placeholder={variable.description || ''}
        />
      )

    case 'user':
      // For now, use employee picker for users
      // Could be extended to use usersApi if needed
      return (
        <EmployeePicker
          value={value}
          onChange={(val) => onChange(val)}
          dataField={variable.data_field || 'full_name'}
          disabled={disabled}
        />
      )

    case 'text':
    default:
      return (
        <Input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          placeholder={variable.description || ''}
        />
      )
  }
}

// Helper function to get icon for variable type
export function getVariableTypeIcon(type: TemplateVariable['variable_type']) {
  switch (type) {
    case 'student':
      return '🎓'
    case 'employee':
    case 'user':
      return '👤'
    case 'department':
      return '🏢'
    case 'current_date':
    case 'date':
      return '📅'
    case 'current_user':
      return '👤'
    case 'number':
      return '🔢'
    case 'select':
      return '📋'
    default:
      return '📝'
  }
}

// Helper function to get label for variable type
export function getVariableTypeLabel(
  type: TemplateVariable['variable_type'],
  t: (key: string) => string
) {
  const labels: Record<string, string> = {
    text: t('variableTypes.text'),
    date: t('variableTypes.date'),
    number: t('variableTypes.number'),
    select: t('variableTypes.select'),
    student: t('variableTypes.student'),
    employee: t('variableTypes.employee'),
    user: t('variableTypes.user'),
    department: t('variableTypes.department'),
    current_date: t('variableTypes.currentDate'),
    current_user: t('variableTypes.currentUser'),
  }
  return labels[type] || type
}
