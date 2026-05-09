// Bulk-edit module types matching backend at
// internal/modules/curriculum/interfaces/http/handlers/bulk_discipline_items_handler.go
// (BulkCreateItemRequest, BulkUpdateItemRequest, BulkEditRequest,
// BulkEditSuccessResponse, BulkEditConflictItem, BulkEditConflictResponse).
//
// Endpoint: POST /api/sections/:sectionID/items/bulk — atomic
// creates+updates+deletes within a single tx (Repeatable Read isolation
// per plan ADR-12). Single-section per request — sectionID lives в URL
// path, не в body.

import type { ControlForm, DisciplineItem } from './disciplineItem'

// Per-item create payload. No `id` — backend assigns после INSERT.
// No `section_id` — inherited from URL path. No `version` — INSERT
// stamps version=0.
export interface BulkEditCreateInput {
  title: string
  hours_lectures: number
  hours_practice: number
  hours_lab: number
  hours_self: number
  control_form: ControlForm
  credits: number
  semester: number
  order_index: number
}

// Per-item update payload. Includes `id` (target row); no `section_id`
// (filtered server-side, must equal URL path sectionID или 422
// CROSS_SECTION_BULK_EDIT). No `version` — backend handler comment
// (bulk_discipline_items_handler.go:60): "version intentionally NOT in
// DTO (repo loads server-side fresh entity, optimistic-lock SQL guards
// race)". Conflict response carries expected_version for UI display
// only.
export interface BulkEditUpdateInput {
  id: number
  title: string
  hours_lectures: number
  hours_practice: number
  hours_lab: number
  hours_self: number
  control_form: ControlForm
  credits: number
  semester: number
  order_index: number
}

// Combined request body. Empty bulk (all 3 arrays empty) → 422
// EMPTY_BULK_INPUT per plan ADR-11.
export interface BulkEditRequest {
  creates: BulkEditCreateInput[]
  updates: BulkEditUpdateInput[]
  deletes: number[]
}

export interface BulkEditSuccessResponse {
  created: DisciplineItem[]
  updated: DisciplineItem[]
  deleted: number[]
}

// Per-item conflict entry в 409 response. CurrentVersion=0 always under
// Repeatable Read isolation (plan ADR-12 — re-fetch within failed tx
// returns same snapshot, defeating purpose). Frontend MUST refetch
// outside the failed tx via GET /api/items/:id (fetchDisciplineItem) для
// accurate current value.
export interface BulkEditConflict {
  id: number
  expected_version: number
  current_version: number
}

export interface BulkEditConflictResponse {
  error: 'VERSION_CONFLICT'
  conflicts: BulkEditConflict[]
}

// Discriminated union returned by bulkEditDisciplineItems. 409 conflict
// is an EXPECTED business outcome (concurrent edit by another methodist),
// modeled as a result variant — not thrown. Other errors (404 / 422 /
// 403 / 500) propagate as axios exceptions for caller to map via
// pickErrorKey utility (Pair 4).
export type BulkEditResult =
  | { kind: 'success'; data: BulkEditSuccessResponse }
  | { kind: 'conflict'; conflicts: BulkEditConflict[] }
