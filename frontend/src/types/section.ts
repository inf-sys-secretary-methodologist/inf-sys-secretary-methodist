// Section module types matching backend DTOs at
// internal/modules/curriculum/interfaces/http/handlers/section_handler.go
// (SectionDTO, SectionsListResponse).
//
// Bounded context: child of Curriculum aggregate (B1a). Sections do
// not have own status — editability inherits curriculum.status (plan
// ADR-2, doc 2026-05-09-v0128-section-aggregate.md). v0.128.4 frontend
// reads sections для bulk-edit-items context; section CRUD UI is out
// of scope (sections seeded externally).

export interface Section {
  id: number
  curriculum_id: number
  title: string
  description: string
  order_index: number
  version: number
  created_at: string
  updated_at: string
}

export interface SectionListResponse {
  items: Section[]
}
