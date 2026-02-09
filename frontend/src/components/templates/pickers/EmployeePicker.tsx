'use client'

import { useState, useEffect, useRef } from 'react'
import { useTranslations } from 'next-intl'
import { Check, ChevronsUpDown, Loader2, Briefcase, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { employeesApi, ExternalEmployee } from '@/lib/api/integration'

interface EmployeePickerProps {
  value: string
  onChange: (value: string, employee?: ExternalEmployee) => void
  dataField?: string // Which field to use: 'full_name', 'email', 'position', etc.
  placeholder?: string
  disabled?: boolean
}

export function EmployeePicker({
  value,
  onChange,
  dataField = 'full_name',
  placeholder,
  disabled = false,
}: EmployeePickerProps) {
  const t = useTranslations('templates.pickers')
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [employees, setEmployees] = useState<ExternalEmployee[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  // Fetch employees on search
  useEffect(() => {
    if (!open) return

    const fetchEmployees = async () => {
      try {
        setIsLoading(true)
        const response = await employeesApi.list({
          search: search || undefined,
          is_active: true,
          limit: 20,
        })
        setEmployees(response.data.employees || [])
      } catch (error) {
        console.error('Failed to fetch employees:', error)
        setEmployees([])
      } finally {
        setIsLoading(false)
      }
    }

    const debounce = setTimeout(fetchEmployees, 300)
    return () => clearTimeout(debounce)
  }, [search, open])

  // Format employee name
  const formatEmployeeName = (employee: ExternalEmployee) => {
    const parts = [employee.last_name, employee.first_name, employee.middle_name].filter(Boolean)
    return parts.join(' ')
  }

  // Get display value for the selected employee
  const getEmployeeFieldValue = (employee: ExternalEmployee): string => {
    switch (dataField) {
      case 'full_name':
        return formatEmployeeName(employee)
      case 'email':
        return employee.email || ''
      case 'position':
        return employee.position || ''
      case 'department':
        return employee.department || ''
      default:
        return formatEmployeeName(employee)
    }
  }

  const handleSelect = (employee: ExternalEmployee) => {
    const fieldValue = getEmployeeFieldValue(employee)
    onChange(fieldValue, employee)
    setOpen(false)
    setSearch('')
  }

  const handleClear = () => {
    onChange('')
    setSearch('')
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
              <Briefcase className="h-4 w-4 shrink-0 text-purple-500" />
              <span className="truncate">{value}</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{placeholder || t('selectEmployee')}</span>
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
      <PopoverContent className="w-[400px] p-0" align="start">
        <div className="p-2 border-b">
          <Input
            ref={inputRef}
            placeholder={t('searchEmployee')}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9"
            autoFocus
          />
        </div>
        <ScrollArea className="h-[280px]">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : employees.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {t('noEmployeesFound')}
            </div>
          ) : (
            <div className="p-1">
              {employees.map((employee) => {
                const fieldValue = getEmployeeFieldValue(employee)
                const isSelected = value === fieldValue
                return (
                  <button
                    key={employee.id}
                    onClick={() => handleSelect(employee)}
                    className={cn(
                      'flex items-center gap-3 w-full px-3 py-2 rounded-md text-left transition-colors',
                      'hover:bg-accent hover:text-accent-foreground',
                      isSelected && 'bg-accent'
                    )}
                  >
                    <Check
                      className={cn('h-4 w-4 shrink-0', isSelected ? 'opacity-100' : 'opacity-0')}
                    />
                    <Briefcase className="h-4 w-4 shrink-0 text-purple-500" />
                    <div className="flex flex-col min-w-0 flex-1">
                      <span className="font-medium truncate text-sm">
                        {formatEmployeeName(employee)}
                      </span>
                      <span className="text-xs text-muted-foreground truncate">
                        {employee.position && `${employee.position}`}
                        {employee.position && employee.department && ' • '}
                        {employee.department}
                      </span>
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
