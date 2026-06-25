// Student debts (Долги студентов) module types, mirroring the backend DTOs
// at internal/modules/student_debts/interfaces/http/handlers/dto.go
// (DebtDTO + AttemptDTO / DebtListItemDTO / DebtListResponse / DebtStatsDTO /
// ImportResultDTO). Enum const arrays mirror the domain wire values at
// internal/modules/student_debts/domain/entities/{debt_status,control_form,
// resit_result}.go. Error codes mirror handler mapDebtError.
//
// Bounded context: академические задолженности студентов (a discipline a
// student failed and must re-sit), with a full lifecycle (attempts →
// commission → closed). Wire format is verbatim — no translation in the
// type layer; UI labels go through next-intl keys (studentDebts.*).

export type StudentDebtStatus =
  | 'open'
  | 'resit_scheduled'
  | 'commission'
  | 'closed_passed'
  | 'closed_failed'

export const STUDENT_DEBT_STATUSES: StudentDebtStatus[] = [
  'open',
  'resit_scheduled',
  'commission',
  'closed_passed',
  'closed_failed',
]

export type ControlForm = 'zachet' | 'exam' | 'course_project' | 'differential_zachet'

export const CONTROL_FORMS: ControlForm[] = [
  'zachet',
  'exam',
  'course_project',
  'differential_zachet',
]

export type ResitResult = 'pending' | 'passed' | 'failed' | 'no_show'

export const RESIT_RESULTS: ResitResult[] = ['pending', 'passed', 'failed', 'no_show']

// === Aggregate shapes (mirror AttemptDTO / DebtDTO) ===

export interface ResitAttempt {
  id: number
  attempt_no: number
  is_commission: boolean
  scheduled_date: string
  examiner: string
  result: ResitResult
  grade?: number
  recorded_by?: number
  recorded_at?: string
}

// StudentDebt is the full aggregate (root + attempt timeline) returned by
// GET /api/student-debts/:id.
export interface StudentDebt {
  id: number
  student_full_name: string
  group_name: string
  discipline_name: string
  semester: number
  control_form: ControlForm
  student_user_id?: number
  discipline_id?: number
  source_ref?: string
  status: StudentDebtStatus
  version: number
  created_at: string
  updated_at: string
  attempts: ResitAttempt[]
}

// StudentDebtListItem is the lightweight list-row projection (root-only,
// no attempts) returned by the list endpoints.
export interface StudentDebtListItem {
  id: number
  student_full_name: string
  group_name: string
  discipline_name: string
  semester: number
  control_form: ControlForm
  student_user_id?: number
  status: StudentDebtStatus
  version: number
}

export interface StudentDebtListResponse {
  items: StudentDebtListItem[]
  total: number
}

export interface StudentDebtStats {
  total: number
  open: number
  resit_scheduled: number
  commission: number
  closed_passed: number
  closed_failed: number
}

export interface StudentDebtsFilter {
  group_name?: string
  status?: StudentDebtStatus
  semester?: number
  limit?: number
  offset?: number
}

// === Mutation inputs (mirror ScheduleResitRequest / RecordResitResultRequest) ===

export interface ScheduleResitInput {
  scheduled_date: string // RFC3339 timestamp
  examiner: string
}

export interface RecordResitResultInput {
  result: ResitResult
  grade?: number | null
}

// === Import result (mirror ImportResultDTO / ImportRowErrorDTO) ===

export interface ImportRowError {
  row: number
  identity: string
  message: string
}

export interface ImportResult {
  created: number
  updated: number
  skipped: number
  errors: ImportRowError[]
}
