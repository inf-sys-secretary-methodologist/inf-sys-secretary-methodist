// Work program (РПД — рабочая программа дисциплины) module types,
// mirroring the backend DTOs at
// internal/modules/work_program/interfaces/http/handlers/work_program_handler.go
// (WorkProgramDTO + 6 child DTOs / WorkProgramSummaryDTO /
// WorkProgramsListResponse / CreateWorkProgramRequest /
// RejectWorkProgramRequest). Enum const arrays mirror the domain types
// at internal/modules/work_program/domain/types.go. Error codes mirror
// handler mapWorkProgramError.
//
// Bounded context: рабочая программа дисциплины (how one discipline is
// taught), orthogonal to curriculum (учебный план — what a specialty
// studies). Distinct lifecycle, distinct author (teacher, not
// academic_secretary). Do not cross-import with curriculum types.

// Wire format is verbatim — no translation in the type layer; UI labels
// go through next-intl keys (workProgram.card.status.* etc.) via the
// statusKey() mapper in components/work-program/status.ts.
export type WorkProgramStatus =
  | 'draft'
  | 'pending_approval'
  | 'approved'
  | 'needs_revision'
  | 'archived'

export const WORK_PROGRAM_STATUSES: WorkProgramStatus[] = [
  'draft',
  'pending_approval',
  'approved',
  'needs_revision',
  'archived',
]

export type CompetenceType = 'pk' | 'ok' | 'uk'
export const COMPETENCE_TYPES: CompetenceType[] = ['pk', 'ok', 'uk']

export type TopicKind = 'lecture' | 'practice' | 'lab' | 'self_study'
export const TOPIC_KINDS: TopicKind[] = ['lecture', 'practice', 'lab', 'self_study']

export type AssessmentType = 'current' | 'intermediate' | 'final'
export const ASSESSMENT_TYPES: AssessmentType[] = ['current', 'intermediate', 'final']

export type ReferenceKind = 'main' | 'additional' | 'electronic'
export const REFERENCE_KINDS: ReferenceKind[] = ['main', 'additional', 'electronic']

export type RevisionStatus = 'draft' | 'pending_approval' | 'approved' | 'rejected'

export type RevisionChangeType = 'hours' | 'semester' | 'literature' | 'assessment' | 'other'

export const REVISION_CHANGE_TYPES: RevisionChangeType[] = [
  'hours',
  'semester',
  'literature',
  'assessment',
  'other',
]

// === Child collection shapes (mirror GoalDTO / CompetenceDTO / ... ) ===

export interface WorkProgramGoal {
  id: number
  text: string
  order_index: number
}

export interface WorkProgramCompetence {
  id: number
  code: string
  type: CompetenceType
  description: string
}

export interface WorkProgramTopic {
  id: number
  kind: TopicKind
  title: string
  hours: number
  week_number?: number | null
  learning_outcomes: string
  order_index: number
}

export interface WorkProgramAssessment {
  id: number
  type: AssessmentType
  description: string
  max_score: number
  example_questions: string[]
}

export interface WorkProgramReference {
  id: number
  kind: ReferenceKind
  citation: string
  year?: number | null
  isbn?: string
  url?: string
  order_index: number
}

export interface WorkProgramRevision {
  id: number
  revision_number: number
  change_type: RevisionChangeType
  change_summary: string
  status: RevisionStatus
  author_id: number
  approver_id?: number | null
  approved_at?: string | null
  reject_reason?: string
  created_at: string
  updated_at: string
}

// === Root aggregate (mirror WorkProgramDTO, all 6 collections hydrated) ===

export interface WorkProgram {
  id: number
  discipline_id: number
  specialty_code: string
  applicable_from_year: number
  title: string
  annotation: string
  status: WorkProgramStatus
  author_id: number
  approver_id?: number | null
  approved_at?: string | null
  reject_reason?: string
  version: number
  created_at: string
  updated_at: string
  goals: WorkProgramGoal[]
  competences: WorkProgramCompetence[]
  topics: WorkProgramTopic[]
  assessments: WorkProgramAssessment[]
  references: WorkProgramReference[]
  revisions: WorkProgramRevision[]
}

// WorkProgramSummary mirrors WorkProgramSummaryDTO — list-row projection,
// root fields only (no inner collections).
export interface WorkProgramSummary {
  id: number
  discipline_id: number
  specialty_code: string
  applicable_from_year: number
  title: string
  status: WorkProgramStatus
  author_id: number
  version: number
}

export interface WorkProgramListResponse {
  items: WorkProgramSummary[]
  total: number
}

// WorkProgramListFilter matches the query string accepted by GET
// /api/v1/work-programs. All fields optional — the use case applies
// role-scoping (teacher → own / student → approved) + pagination
// defaults server-side.
export interface WorkProgramListFilter {
  status?: WorkProgramStatus
  discipline_id?: number
  specialty_code?: string
  applicable_from_year?: number
  author_id?: number
  limit?: number
  offset?: number
}

// CreateWorkProgramInput matches CreateWorkProgramRequest. The author is
// stamped from the JWT subject server-side — never a client field.
export interface CreateWorkProgramInput {
  discipline_id: number
  specialty_code: string
  applicable_from_year: number
  title: string
  annotation?: string
}

// RejectWorkProgramInput matches RejectWorkProgramRequest. The reason is
// mandatory (domain enforces non-empty after trim; binding fails fast).
// Reused by both the РПД reject and the revision (лист актуализации)
// reject endpoints — identical shape.
export interface RejectWorkProgramInput {
  reason: string
}

// CreateRevisionInput matches CreateRevisionRequest (revision_handler.go).
// The author is stamped from the JWT subject server-side — never a client
// field. diff_payload (optional raw before/after blob) is omitted here:
// the create-revision dialog only collects the categorized change_type +
// a human summary; structured diffs arrive later via AI bulk-revision.
export interface CreateRevisionInput {
  change_type: RevisionChangeType
  change_summary: string
}

// Error codes consumed by pickWorkProgramErrorKey (→ i18n keys). The
// first five are sentinel codes the backend emits in the error body via
// mapWorkProgramError; FORBIDDEN / NOT_FOUND are derived frontend-side
// from the HTTP status (the handler returns 403/404 with no code), and
// GENERIC is the catch-all fallback for unknown 5xx — neither group is
// an actual mapWorkProgramError code.
export type WorkProgramErrorCode =
  | 'IDENTITY_EXISTS'
  | 'VERSION_CONFLICT'
  | 'INVALID_TRANSITION'
  | 'REJECT_REASON_REQUIRED'
  | 'INVALID_WORK_PROGRAM'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'GENERIC'

export const WORK_PROGRAM_ERROR_CODES: WorkProgramErrorCode[] = [
  'IDENTITY_EXISTS',
  'VERSION_CONFLICT',
  'INVALID_TRANSITION',
  'REJECT_REASON_REQUIRED',
  'INVALID_WORK_PROGRAM',
  'FORBIDDEN',
  'NOT_FOUND',
  'GENERIC',
]

// === Collection-edit inputs (slice 12c) ===
//
// Payloads for the 15 manual collection-edit endpoints (add/update ×
// goals/competences/topics/assessments/references). Mirror the request
// DTOs at work_program_content_handler.go. Identity fields are required
// by backend binding tags; deeper invariants (text length, MaxScore
// range) live in the domain and surface as INVALID_WORK_PROGRAM → 422.
// The actor is stamped from the JWT subject server-side — never a field.

export interface GoalInput {
  text: string
  order_index: number
}

export interface CompetenceInput {
  code: string
  type: CompetenceType
  description: string
}

// week_number is optional in the domain (a topic may not be pinned to a
// teaching week); the empty form field maps to null so the backend's
// *int pointer stays absent rather than coerced to week 0.
export interface TopicInput {
  kind: TopicKind
  title: string
  hours: number
  week_number?: number | null
  learning_outcomes: string
  order_index: number
}

// max_score per-item range is enforced by the domain ([1,100]); the form
// only collects the value. example_questions (ФОС) is a free list — the
// dialog edits it as one textarea, one question per line.
export interface AssessmentInput {
  type: AssessmentType
  description: string
  max_score: number
  example_questions: string[]
}

// year is optional in the domain (an electronic source may have none) —
// the empty form field maps to null. isbn/url are plain optional strings
// (backend non-pointer), so they default to '' rather than null.
export interface ReferenceInput {
  kind: ReferenceKind
  citation: string
  year?: number | null
  isbn: string
  url: string
  order_index: number
}
