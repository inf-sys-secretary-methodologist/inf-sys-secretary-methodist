'use client'

import { useTranslations } from 'next-intl'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type {
  LessonFilterParams,
  Semester,
  StudentGroup,
  Classroom,
} from '@/types/schedule'

interface ScheduleFiltersProps {
  filters: LessonFilterParams
  onFiltersChange: (filters: LessonFilterParams) => void
  semesters: Semester[]
  groups: StudentGroup[]
  classrooms: Classroom[]
}

const ALL_VALUE = '__all__'

export function ScheduleFilters({
  filters,
  onFiltersChange,
  semesters,
  groups,
  classrooms,
}: ScheduleFiltersProps) {
  const t = useTranslations('schedule')

  const handleSemesterChange = (value: string) => {
    onFiltersChange({
      ...filters,
      semester_id: value === ALL_VALUE ? undefined : Number(value),
    })
  }

  const handleGroupChange = (value: string) => {
    onFiltersChange({
      ...filters,
      group_id: value === ALL_VALUE ? undefined : Number(value),
    })
  }

  const handleClassroomChange = (value: string) => {
    onFiltersChange({
      ...filters,
      classroom_id: value === ALL_VALUE ? undefined : Number(value),
    })
  }

  return (
    <div className="flex flex-wrap gap-3" data-testid="schedule-filters">
      {/* Semester */}
      <Select
        value={filters.semester_id ? String(filters.semester_id) : ALL_VALUE}
        onValueChange={handleSemesterChange}
      >
        <SelectTrigger className="w-[200px]">
          <SelectValue placeholder={t('filters.semester')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_VALUE}>{t('filters.semester')}</SelectItem>
          {semesters.map((s) => (
            <SelectItem key={s.id} value={String(s.id)}>
              {s.name}{s.is_active ? ' *' : ''}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Group */}
      <Select
        value={filters.group_id ? String(filters.group_id) : ALL_VALUE}
        onValueChange={handleGroupChange}
      >
        <SelectTrigger className="w-[200px]">
          <SelectValue placeholder={t('filters.group')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_VALUE}>{t('filters.group')}</SelectItem>
          {groups.map((g) => (
            <SelectItem key={g.id} value={String(g.id)}>
              {g.name} ({g.course} курс)
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Classroom */}
      <Select
        value={filters.classroom_id ? String(filters.classroom_id) : ALL_VALUE}
        onValueChange={handleClassroomChange}
      >
        <SelectTrigger className="w-[200px]">
          <SelectValue placeholder={t('filters.classroom')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL_VALUE}>{t('filters.classroom')}</SelectItem>
          {classrooms.map((c) => (
            <SelectItem key={c.id} value={String(c.id)}>
              {c.building}-{c.number}{c.name ? ` (${c.name})` : ''}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
