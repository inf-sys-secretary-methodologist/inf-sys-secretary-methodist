// Schedule module types matching backend DTOs at
// internal/modules/schedule/application/dto/lesson_dto.go

export interface Lesson {
  id: number
  semester_id: number
  discipline_id: number
  lesson_type_id: number
  teacher_id: number
  group_id: number
  classroom_id: number
  day_of_week: number // 1=Monday, 7=Sunday
  time_start: string // "09:00"
  time_end: string // "10:30"
  week_type: 'all' | 'odd' | 'even'
  date_start: string
  date_end: string
  notes?: string
  is_cancelled: boolean
  created_at: string
  updated_at: string
  discipline?: Discipline
  lesson_type?: LessonTypeInfo
  classroom?: Classroom
  group?: StudentGroup
  teacher?: TeacherInfo
}

export interface Classroom {
  id: number
  building: string
  number: string
  name?: string
  capacity: number
  type?: string
  is_available: boolean
}

export interface StudentGroup {
  id: number
  name: string
  course: number
}

export interface Discipline {
  id: number
  name: string
  code?: string
}

export interface LessonTypeInfo {
  id: number
  name: string
  short_name: string
  color?: string
}

export interface Semester {
  id: number
  name: string
  number: number
  start_date: string
  end_date: string
  is_active: boolean
}

export interface TeacherInfo {
  id: number
  name: string
  email: string
}

export interface ScheduleChange {
  id: number
  lesson_id: number
  change_type: 'cancelled' | 'moved' | 'replaced_teacher' | 'replaced_classroom'
  original_date: string
  new_date?: string
  new_classroom_id?: number
  new_teacher_id?: number
  reason?: string
  created_by: number
  created_at: string
}

export interface CreateLessonInput {
  semester_id: number
  discipline_id: number
  lesson_type_id: number
  teacher_id: number
  group_id: number
  classroom_id: number
  day_of_week: number
  time_start: string
  time_end: string
  week_type: 'all' | 'odd' | 'even'
  date_start: string
  date_end: string
  notes?: string
}

export interface UpdateLessonInput {
  classroom_id?: number
  teacher_id?: number
  day_of_week?: number
  time_start?: string
  time_end?: string
  week_type?: 'all' | 'odd' | 'even'
  notes?: string
}

export interface LessonFilterParams {
  semester_id?: number
  group_id?: number
  teacher_id?: number
  classroom_id?: number
  discipline_id?: number
  day_of_week?: number
}

export interface CreateChangeInput {
  lesson_id: number
  change_type: 'cancelled' | 'moved' | 'replaced_teacher' | 'replaced_classroom'
  original_date: string
  new_date?: string
  new_classroom_id?: number
  new_teacher_id?: number
  reason?: string
}

export type WeekType = 'all' | 'odd' | 'even'

export const WEEK_TYPES: WeekType[] = ['all', 'odd', 'even']

export const DAY_NAMES = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'] as const

export const TIME_SLOTS = [
  { start: '09:00', end: '10:30' },
  { start: '10:45', end: '12:15' },
  { start: '13:00', end: '14:30' },
  { start: '14:45', end: '16:15' },
  { start: '16:30', end: '18:00' },
] as const
