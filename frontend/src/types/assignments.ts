// Assignments module types matching backend DTOs at
// internal/modules/assignments/interfaces/http/handlers/assignments_handler.go
// (AssignmentDTO, SubmissionViewDTO) and grade_handler.go (SaveGradeRequest).
//
// Bounded context: academic homework. Distinct from src/types/tasks.ts which
// describes the project-management tasks module — same word, different
// aggregate. Do not cross-import.

export type SubmissionStatus = 'pending' | 'graded' | 'returned'

export const SUBMISSION_STATUSES: SubmissionStatus[] = ['pending', 'graded', 'returned']

export interface Assignment {
  id: number
  title: string
  description?: string
  teacher_id: number
  group_name: string
  subject: string
  max_score: number
  due_date?: string
  created_at: string
  updated_at: string
}

export interface AssignmentListResponse {
  items: Assignment[]
  total: number
}

export interface SubmissionView {
  id: number
  assignment_id: number
  student_id: number
  student_name: string
  grade_value?: number
  feedback?: string
  graded_by?: number
  graded_at?: string
  return_reason?: string
  returned_by?: number
  returned_at?: string
  status: SubmissionStatus
  created_at: string
  updated_at: string
}

export interface SubmissionListResponse {
  items: SubmissionView[]
}

export interface SaveGradeRequest {
  student_id: number
  value: number
  feedback?: string
}

export interface SaveGradeResponse {
  assignment_id: number
  student_id: number
  value: number
}

export interface ReturnSubmissionRequest {
  student_id: number
  reason: string
}

export interface ReturnSubmissionResponse {
  assignment_id: number
  student_id: number
  reason: string
}

export interface AssignmentListFilter {
  subject?: string
  group_name?: string
  page_size?: number
  offset?: number
}
