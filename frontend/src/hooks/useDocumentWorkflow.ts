// v0.148.0 — documents workflow API client functions (#227).
//
// Mirror к useCurricula.ts submit/approve/reject pattern:
// thin axios POST wrappers that unwrap ApiResponse + propagate
// axios errors so the dialog components can branch on HTTP status.

import { apiClient } from '@/lib/api'

const DOCUMENTS_URL = '/api/documents'
const ADMIN_DOCUMENTS_URL = '/api/admin/documents'

interface ApiResponse<T> {
  data: T
}

// Document workflow fields the backend includes in responses post-
// v0.148.0. The full Document type lives in components/documents но
// these audit fields are workflow-specific.
export interface DocumentWorkflowFields {
  status: string
  submitted_by?: number | null
  submitted_at?: string | null
  approved_by?: number | null
  approved_at?: string | null
  rejected_by?: number | null
  rejected_at?: string | null
  rejected_reason?: string | null
}

export interface RejectDocumentRequest {
  reason: string
}

// submitDocument transitions a draft document into the approval
// queue. Body empty per backend contract (path id + JWT subject
// identify row + actor). Axios errors propagate — caller branches
// on 409 (NOT_DRAFT) / 403 (FORBIDDEN) / 404 by HTTP status.
export async function submitDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${DOCUMENTS_URL}/${id}/submit`,
    {}
  )
  return response.data
}

// approveDocument transitions approval → approved. Admin-only via
// route gate (academic_secretary / system_admin). Axios errors
// propagate — caller branches on 409 / 403 / 404.
export async function approveDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/approve`,
    {}
  )
  return response.data
}

// rejectDocument transitions approval → rejected. Body carries the
// rejection reason — backend validates 10..500 char rune count after
// trim (RejectionReason VO) и stores на the entity for future rework
// context (diverges от curriculum's audit-only approach per ADR-3 v0.148.0).
// Axios errors propagate — caller branches on 422 (INVALID_REASON or
// NOT_APPROVAL) / 403 / 404.
export async function rejectDocument(
  id: number,
  body: RejectDocumentRequest
): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/reject`,
    body
  )
  return response.data
}
