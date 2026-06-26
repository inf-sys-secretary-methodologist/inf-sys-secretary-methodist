'use client'

// Document e-signature (#140) SWR hook + JSON mutation actions. Talks to
// /api/documents/:id/{sign,signatures,signatures/:sigId/verify} (see
// internal/modules/documents/interfaces/http/handlers/signature_handler.go).
// Signing is server-side (per-user key) — the sign mutation sends no body.

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { DocumentSignature, SignatureVerification } from '@/types/document'

const BASE_URL = '/api/documents'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T>>(url)
  return response.data
}

interface FetchOpts {
  // enabled defaults to true; false short-circuits the SWR key to null so the
  // fetch is skipped while auth resolves. Mirrors useStudentDebts.
  enabled?: boolean
}

// useDocumentSignatures returns a document's signatures (oldest first).
export function useDocumentSignatures(documentId: number | null, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = !enabled || documentId == null ? null : `${BASE_URL}/${documentId}/signatures`
  const { data, error, isLoading, mutate } = useSWR<DocumentSignature[]>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })
  return { signatures: data ?? [], isLoading, error, mutate }
}

// signDocument applies the current actor's signature to a document. The server
// signs with the actor's key, so the request carries no body.
export async function signDocument(documentId: number): Promise<DocumentSignature> {
  const response = await apiClient.post<ApiResponse<DocumentSignature>>(
    `${BASE_URL}/${documentId}/sign`
  )
  return response.data
}

// verifySignature re-checks a stored signature against the document's current
// state and returns the verdict.
export async function verifySignature(
  documentId: number,
  signatureId: number
): Promise<SignatureVerification> {
  const response = await apiClient.get<ApiResponse<SignatureVerification>>(
    `${BASE_URL}/${documentId}/signatures/${signatureId}/verify`
  )
  return response.data
}

// pickSignatureErrorKey maps a failed request to an i18n message key under the
// documentSignatures.errors namespace. The backend returns plain HTTP statuses
// for these endpoints, so we branch on status.
export function pickSignatureErrorKey(err: unknown): string {
  const status = (err as { response?: { status?: number } })?.response?.status
  if (status === 403) return 'forbidden'
  if (status === 404) return 'notFound'
  if (status === 422) return 'invalid'
  return 'generic'
}
