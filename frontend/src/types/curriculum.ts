// Curriculum module types matching backend DTOs at
// internal/modules/curriculum/interfaces/http/handlers/curriculum_handler.go
// (CurriculumDTO, CurriculaListResponse, CreateCurriculumRequest,
// UpdateCurriculumRequest, RejectCurriculumRequest).
//
// Bounded context: academic curriculum (учебный план). Distinct from
// assignments (homework grading) and tasks (project management) — same
// "academic" group in navigation but different aggregates. Do not
// cross-import.

// CurriculumStatus mirrors backend
// internal/modules/curriculum/domain/entities/curriculum_status.go.
// Wire format is verbatim — no translation in the type layer; UI
// labels go through next-intl keys (curriculum.filters.status.*).
export type CurriculumStatus = 'draft' | 'pending_approval' | 'approved' | 'archived'

export const CURRICULUM_STATUSES: CurriculumStatus[] = [
  'draft',
  'pending_approval',
  'approved',
  'archived',
]

export interface Curriculum {
  id: number
  title: string
  code: string
  specialty: string
  year: number
  description: string
  status: CurriculumStatus
  created_by: number
  approved_by?: number
  approved_at?: string
  created_at: string
  updated_at: string
}

export interface CurriculumListResponse {
  items: Curriculum[]
  total: number
}

// CurriculumListFilter matches the query string accepted by GET
// /api/curriculum (handler parseListInput). All fields optional — the
// backend uses sensible defaults (limit=50 if unset, offset=0).
export interface CurriculumListFilter {
  status?: CurriculumStatus
  year?: number
  specialty?: string
  created_by?: number
  limit?: number
  offset?: number
}

// UpdateCurriculumRequest matches handler UpdateCurriculumRequest
// at internal/modules/curriculum/interfaces/http/handlers/curriculum_handler.go.
// Consumed by EditCurriculumDialog (v0.119.0). Backend invariants:
// title / code / specialty trim non-empty, year ∈ [2000, 2100],
// description ≤ 4096 chars after trim.
export interface UpdateCurriculumRequest {
  title: string
  code: string
  specialty: string
  year: number
  description: string
}

// CreateCurriculumRequest / RejectCurriculumRequest land alongside
// their consumers in later releases (v0.121.0 polish create dialog /
// v0.120.0 admin reject dialog) — not pre-defined here.
