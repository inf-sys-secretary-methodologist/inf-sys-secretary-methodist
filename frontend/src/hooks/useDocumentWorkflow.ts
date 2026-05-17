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

export interface RegisterDocumentRequest {
  number: string
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

// registerDocument (v0.149.0 #230) transitions approved → registered с
// registration number. Backend validates: status invariant + non-empty
// trim length ≥3 + UNIQUE registration_number (DB constraint). Errors:
// 422 invalid number / 409 not_approved / 404 not_found / 403 forbidden.
export async function registerDocument(
  id: number,
  body: RegisterDocumentRequest
): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/register`,
    body
  )
  return response.data
}

// startRoutingDocument (v0.150.0 #231) transitions registered → routing.
// Single-step visa per ADR-1 — sends the document to one admin approver.
// Body empty (path id + JWT subject identify row + actor).
// Errors: 409 not_registered / 404 not_found / 403 forbidden.
export async function startRoutingDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/start-routing`,
    {}
  )
  return response.data
}

// signVisaDocument (v0.150.0 #231) transitions routing → execution
// when the visa is signed. Single-step — one approver completes the
// visa. Errors: 409 not_routing / 404 not_found / 403 forbidden.
export async function signVisaDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/sign-visa`,
    {}
  )
  return response.data
}

// AssignExecutorRequest body для AssignExecutor endpoint (v0.151.0 #232).
// dueDate optional — YYYY-MM-DD string или RFC3339; backend handles both.
export interface AssignExecutorRequest {
  executor_id: number
  due_date?: string
}

// assignExecutorDocument (v0.151.0 #232) shapes executor assignment on an
// execution-status document. Status stays execution — shape-only per ADR-1.
// Reassign overwrites prior. Errors: 422 invalid_executor / 422 invalid
// due_date / 409 not_execution / 404 not_found / 403 forbidden.
export async function assignExecutorDocument(
  id: number,
  body: AssignExecutorRequest
): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/assign-executor`,
    body
  )
  return response.data
}

// markExecutedDocument (v0.151.0 #232) transitions execution → executed.
// Body-less. Errors: 409 not_execution / 404 not_found / 403 forbidden.
export async function markExecutedDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/mark-executed`,
    {}
  )
  return response.data
}

// archiveDocument (v0.152.0 #233) transitions executed → archived
// (terminal step closing lifecycle). Admin-only via route gate.
// Body-less. Errors: 409 not_executed / 404 not_found / 403 forbidden.
export async function archiveDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${ADMIN_DOCUMENTS_URL}/${id}/archive`,
    {}
  )
  return response.data
}

// resubmitDocument (v0.152.0 #233) transitions rejected → draft (rework
// cycle). Clears RejectedBy/At/Reason audit fields. Author OR edit-role
// gated at use-case boundary (mirror к submit pattern, NOT admin-only)
// — mounted on non-admin /api/documents path.
// Body-less. Errors: 409 not_rejected / 404 not_found / 403 forbidden.
export async function resubmitDocument(id: number): Promise<DocumentWorkflowFields> {
  const response = await apiClient.post<ApiResponse<DocumentWorkflowFields>>(
    `${DOCUMENTS_URL}/${id}/resubmit`,
    {}
  )
  return response.data
}
