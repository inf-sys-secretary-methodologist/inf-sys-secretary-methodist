// Teaching load (Нагрузка) wire types. Mirror of the backend DTO in
// internal/modules/schedule/application/dto/teaching_load_dto.go and the
// TeachingLoad entity. The auto-scheduler (issue #139) expands each load line
// into schedulable lessons; this screen is the methodist's source of truth.

import type { WeekType } from './schedule'

export type { WeekType }

// TeachingLoad is one planned assignment with denormalized reference names.
export interface TeachingLoad {
  id: number
  semester_id: number
  group_id: number
  group_name: string
  discipline_id: number
  discipline_name: string
  teacher_id: number
  teacher_name: string
  lesson_type_id: number
  lesson_type_name: string
  pairs_per_week: number
  week_type: WeekType
}

// TeachingLoadInput is the create/update request body (7 mutable fields).
export interface TeachingLoadInput {
  semester_id: number
  group_id: number
  discipline_id: number
  teacher_id: number
  lesson_type_id: number
  pairs_per_week: number
  week_type: WeekType
}

// TeachingLoadFilter narrows the list; all fields optional.
export interface TeachingLoadFilter {
  semester_id?: number
  group_id?: number
  teacher_id?: number
}

// TeachingLoadListResponse is the wrapped list payload.
export interface TeachingLoadListResponse {
  teaching_loads: TeachingLoad[]
}
