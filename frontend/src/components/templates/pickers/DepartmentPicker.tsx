'use client'

import { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { Check, ChevronsUpDown, Loader2, Building2, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { departmentsApi, Department } from '@/lib/api/users'

interface DepartmentPickerProps {
  value: string
  onChange: (value: string, department?: Department) => void
  dataField?: string // Which field to use: 'name', 'code'
  placeholder?: string
  disabled?: boolean
}

export function DepartmentPicker({
  value,
  onChange,
  dataField = 'name',
  placeholder,
  disabled = false,
}: DepartmentPickerProps) {
  const t = useTranslations('templates.pickers')
  const [open, setOpen] = useState(false)
  const [departments, setDepartments] = useState<Department[]>([])
  const [isLoading, setIsLoading] = useState(false)

  // Fetch departments when popover opens
  useEffect(() => {
    if (!open) return

    const fetchDepartments = async () => {
      try {
        setIsLoading(true)
        const response = await departmentsApi.list(1, 100, true)
        setDepartments(response.data.departments || [])
      } catch (error) {
        console.error('Failed to fetch departments:', error)
        setDepartments([])
      } finally {
        setIsLoading(false)
      }
    }

    fetchDepartments()
  }, [open])

  // Get display value for the selected department
  const getDepartmentFieldValue = (department: Department): string => {
    switch (dataField) {
      case 'name':
        return department.name
      case 'code':
        return department.code
      default:
        return department.name
    }
  }

  const handleSelect = (department: Department) => {
    const fieldValue = getDepartmentFieldValue(department)
    onChange(fieldValue, department)
    setOpen(false)
  }

  const handleClear = () => {
    onChange('')
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className="w-full justify-between font-normal h-10"
        >
          {value ? (
            <span className="flex items-center gap-2 truncate flex-1">
              <Building2 className="h-4 w-4 shrink-0 text-green-500" />
              <span className="truncate">{value}</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{placeholder || t('selectDepartment')}</span>
          )}
          <div className="flex items-center gap-1 ml-2">
            {value && (
              <X
                className="h-4 w-4 shrink-0 opacity-50 hover:opacity-100"
                onClick={(e) => {
                  e.stopPropagation()
                  handleClear()
                }}
              />
            )}
            <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" />
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[350px] p-0" align="start">
        <ScrollArea className="h-[280px]">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : departments.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {t('noDepartmentsFound')}
            </div>
          ) : (
            <div className="p-1">
              {departments.map((department) => {
                const fieldValue = getDepartmentFieldValue(department)
                const isSelected = value === fieldValue
                return (
                  <button
                    key={department.id}
                    onClick={() => handleSelect(department)}
                    className={cn(
                      'flex items-center gap-3 w-full px-3 py-2 rounded-md text-left transition-colors',
                      'hover:bg-accent hover:text-accent-foreground',
                      isSelected && 'bg-accent'
                    )}
                  >
                    <Check
                      className={cn('h-4 w-4 shrink-0', isSelected ? 'opacity-100' : 'opacity-0')}
                    />
                    <Building2 className="h-4 w-4 shrink-0 text-green-500" />
                    <div className="flex flex-col min-w-0 flex-1">
                      <span className="font-medium truncate text-sm">{department.name}</span>
                      {department.description && (
                        <span className="text-xs text-muted-foreground truncate">
                          {department.description}
                        </span>
                      )}
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </ScrollArea>
      </PopoverContent>
    </Popover>
  )
}
