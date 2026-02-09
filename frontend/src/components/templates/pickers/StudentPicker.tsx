'use client'

import { useState, useEffect, useRef } from 'react'
import { useTranslations } from 'next-intl'
import { Check, ChevronsUpDown, Loader2, GraduationCap, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { studentsApi, ExternalStudent } from '@/lib/api/integration'

interface StudentPickerProps {
  value: string
  onChange: (value: string, student?: ExternalStudent) => void
  dataField?: string // Which field to use: 'full_name', 'email', 'student_id', etc.
  placeholder?: string
  disabled?: boolean
}

export function StudentPicker({
  value,
  onChange,
  dataField = 'full_name',
  placeholder,
  disabled = false,
}: StudentPickerProps) {
  const t = useTranslations('templates.pickers')
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [students, setStudents] = useState<ExternalStudent[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  // Fetch students on search
  useEffect(() => {
    if (!open) return

    const fetchStudents = async () => {
      try {
        setIsLoading(true)
        const response = await studentsApi.list({
          search: search || undefined,
          is_active: true,
          limit: 20,
        })
        setStudents(response.data.students || [])
      } catch (error) {
        console.error('Failed to fetch students:', error)
        setStudents([])
      } finally {
        setIsLoading(false)
      }
    }

    const debounce = setTimeout(fetchStudents, 300)
    return () => clearTimeout(debounce)
  }, [search, open])

  // Format student name
  const formatStudentName = (student: ExternalStudent) => {
    const parts = [student.last_name, student.first_name, student.middle_name].filter(Boolean)
    return parts.join(' ')
  }

  // Get display value for the selected student
  const getStudentFieldValue = (student: ExternalStudent): string => {
    switch (dataField) {
      case 'full_name':
        return formatStudentName(student)
      case 'email':
        return student.email || ''
      case 'student_id':
        return student.student_id || ''
      case 'group_name':
        return student.group_name || ''
      case 'faculty':
        return student.faculty || ''
      default:
        return formatStudentName(student)
    }
  }

  const handleSelect = (student: ExternalStudent) => {
    const fieldValue = getStudentFieldValue(student)
    onChange(fieldValue, student)
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
              <GraduationCap className="h-4 w-4 shrink-0 text-blue-500" />
              <span className="truncate">{value}</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{placeholder || t('selectStudent')}</span>
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
            placeholder={t('searchStudent')}
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
          ) : students.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {t('noStudentsFound')}
            </div>
          ) : (
            <div className="p-1">
              {students.map((student) => {
                const fieldValue = getStudentFieldValue(student)
                const isSelected = value === fieldValue
                return (
                  <button
                    key={student.id}
                    onClick={() => handleSelect(student)}
                    className={cn(
                      'flex items-center gap-3 w-full px-3 py-2 rounded-md text-left transition-colors',
                      'hover:bg-accent hover:text-accent-foreground',
                      isSelected && 'bg-accent'
                    )}
                  >
                    <Check
                      className={cn('h-4 w-4 shrink-0', isSelected ? 'opacity-100' : 'opacity-0')}
                    />
                    <GraduationCap className="h-4 w-4 shrink-0 text-blue-500" />
                    <div className="flex flex-col min-w-0 flex-1">
                      <span className="font-medium truncate text-sm">
                        {formatStudentName(student)}
                      </span>
                      <span className="text-xs text-muted-foreground truncate">
                        {student.group_name && `${student.group_name}`}
                        {student.group_name && student.faculty && ' • '}
                        {student.faculty}
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
