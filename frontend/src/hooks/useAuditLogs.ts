'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { AuditLogFilter, AuditLogListResult } from '@/types/audit'

const AUDIT_URL = '/api/admin/audit-logs'

interface FetchOpts {
  // enabled defaults to true. Setting it to false short-circuits the
  // SWR key to null so the fetch never fires — used by the
  // /admin/audit-logs page to skip the round-trip while the role
  // guard is still resolving (mirrors useCurricula opts pattern).
  enabled?: boolean
}

// Stub: behavior deferred to the matching GREEN commit. Returns an
// empty state so the test file compiles against the public API.
export function useAuditLogs(
  _filter?: AuditLogFilter,
  _opts?: FetchOpts
): {
  items: AuditLogListResult['items']
  pagination: AuditLogListResult['pagination'] | undefined
  total: number
  isLoading: boolean
  error: unknown
  mutate: () => Promise<unknown>
} {
  // No-op references to keep TS/lint quiet until GREEN replaces this body.
  void _filter
  void _opts
  void useSWR
  void apiClient
  void AUDIT_URL
  void SWR_DEDUPING

  return {
    items: [],
    pagination: undefined,
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: async () => undefined,
  }
}
