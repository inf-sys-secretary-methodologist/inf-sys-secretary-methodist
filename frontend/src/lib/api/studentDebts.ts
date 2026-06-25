import { apiClient } from '../api'
import type { ImportResult, StudentDebtsFilter } from '@/types/studentDebts'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

// studentDebtsApi wraps the bulk file-transfer endpoints of the student
// debts registry. Import sends a multipart xlsx; export streams an xlsx
// back as a binary Blob (responseType: 'blob' bypasses JSON parsing).
// Authentication / EDIT_ROLES gating happen server-side. The resit-lifecycle
// JSON mutations and read hooks live in hooks/useStudentDebts.ts.
export const studentDebtsApi = {
  // import POSTs the registry document and returns the import log
  // (created/updated/skipped + per-row errors). A forbidden actor → 403,
  // a malformed document → 400 (callers branch via pickStudentDebtErrorKey).
  async import(file: File): Promise<ImportResult> {
    const formData = new FormData()
    formData.append('file', file)
    const response = await apiClient.post<ApiResponse<ImportResult>>(
      '/api/student-debts/import',
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    )
    return response.data
  },

  // export GETs the role-scoped registry as an xlsx Blob. The list filter
  // (group_name / status / semester) narrows the export; pagination is
  // irrelevant server-side. Undefined/empty params are omitted.
  async export(filter?: StudentDebtsFilter): Promise<Blob> {
    const params: Record<string, string | number> = {}
    if (filter) {
      Object.entries(filter).forEach(([key, value]) => {
        if (value === undefined || value === null || value === '') return
        params[key] = value as string | number
      })
    }
    const blob = await apiClient.get<Blob>('/api/student-debts/export', {
      params,
      responseType: 'blob',
    })
    return blob
  },
}
