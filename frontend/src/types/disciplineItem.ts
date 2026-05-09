// DisciplineItem module types matching backend DTOs at
// internal/modules/curriculum/interfaces/http/handlers/discipline_item_handler.go
// (DisciplineItemDTO, DisciplineItemsListResponse).
//
// Bounded context: child of Section aggregate (B1a Layer 2). Editability
// inherits curriculum.status (plan ADR-2). v0.128.4 frontend reads items
// для bulk-edit table view; single-row CRUD UI deferred (bulk-only path).

// ControlForm mirrors backend domain VO entities/control_form.go.
// Per CLAUDE.md ubiquitous-language gate: typed string union, не
// magic strings. Wire format verbatim — no UI translation in the
// type layer; labels go through next-intl key
// curriculum.disciplineItems.controlForm.{zachet|exam|course_project|differential_zachet}.
export type ControlForm = 'zachet' | 'exam' | 'course_project' | 'differential_zachet'

export const CONTROL_FORMS: ControlForm[] = [
  'zachet',
  'exam',
  'course_project',
  'differential_zachet',
]

export interface DisciplineItem {
  id: number
  section_id: number
  title: string
  hours_lectures: number
  hours_practice: number
  hours_lab: number
  hours_self: number
  control_form: ControlForm
  credits: number
  semester: number
  order_index: number
  version: number
  created_at: string
  updated_at: string
}

export interface DisciplineItemListResponse {
  items: DisciplineItem[]
}
