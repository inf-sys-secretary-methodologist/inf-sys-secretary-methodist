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

// StudentAssignmentView mirrors backend
// internal/modules/assignments/interfaces/http/handlers/my_assignments_handler.go
// StudentAssignmentDTO. Denormalised assignment + submission projection
// returned by GET /api/assignments/my and GET /api/assignments/:id/my.
export interface StudentAssignmentView {
  // Assignment columns.
  assignment_id: number
  title: string
  description?: string
  subject: string
  group_name: string
  max_score: number
  due_date?: string
  assignment_created_at: string
  assignment_updated_at: string

  // Submission columns.
  submission_id: number
  student_id: number
  grade_value?: number
  feedback?: string
  graded_by?: number
  graded_at?: string
  return_reason?: string
  returned_by?: number
  returned_at?: string
  status: SubmissionStatus
  submission_created_at: string
  submission_updated_at: string
}

export interface MyAssignmentListResponse {
  items: StudentAssignmentView[]
  total: number
}

// Mirrors backend resubmit_handler.go success payload — assignment id
// from path + student id from JWT subject. The student supplies no
// other input on resubmit (body is empty), so the response is a pure
// echo confirming which row was transitioned.
export interface ResubmitSubmissionResponse {
  assignment_id: number
  student_id: number
}
