'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type {
  AuditLog,
  AuditLogFilter,
  AuditLogListResult,
  AuditLogPagination,
} from '@/types/audit'

const AUDIT_URL = '/api/admin/audit-logs'

interface FetchOpts {
  // enabled defaults to true. Setting it to false short-circuits the
  // SWR key to null so the fetch never fires — used by the
  // /admin/audit-logs page to skip the round-trip while the role
  // guard is still resolving (mirrors useCurricula opts pattern).
  enabled?: boolean
}

// Envelope is the full response.List wire shape from the backend
// (response.Response with success / data / meta / error). The audit-
// log read endpoint puts pagination meta on the envelope rather than
// flat inside data — declared explicitly here so the fetcher can
// project both items and pagination into a single hook result.
interface Envelope {
  success: boolean
  data: AuditLog[]
  meta: { pagination: AuditLogPagination }
  error?: { code: string; message: string }
}

function buildAuditUrl(filter?: AuditLogFilter): string {
  if (!filter) return AUDIT_URL
  const params = new URLSearchParams()
  if (filter.action) params.append('action', filter.action)
  if (filter.resource) params.append('resource', filter.resource)
  if (typeof filter.user_id === 'number') {
    params.append('user_id', String(filter.user_id))
  }
  if (filter.from) params.append('from', filter.from)
  if (filter.to) params.append('to', filter.to)
  if (typeof filter.limit === 'number') {
    params.append('limit', String(filter.limit))
  }
  if (typeof filter.offset === 'number') {
    params.append('offset', String(filter.offset))
  }
  const qs = params.toString()
  return qs ? `${AUDIT_URL}?${qs}` : AUDIT_URL
}

// auditFetcher resolves to the lifted AuditLogListResult so the hook
// can return items + pagination as siblings — consumers do not have
// to traverse `data.meta.pagination` past the hook boundary.
const auditFetcher = async (url: string): Promise<AuditLogListResult> => {
  const envelope = await apiClient.get<Envelope>(url)
  return { items: envelope.data, pagination: envelope.meta.pagination }
}

// useAuditLogs reads GET /api/admin/audit-logs. Backend is gated by
// RequireRole(system_admin); a 403 propagates as an SWR error so the
// page can fall back to the forbidden state.
export function useAuditLogs(filter?: AuditLogFilter, opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? buildAuditUrl(filter) : null
  const { data, error, isLoading, mutate } = useSWR<AuditLogListResult>(key, auditFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    items: data?.items ?? [],
    pagination: data?.pagination,
    total: data?.pagination?.total ?? 0,
    isLoading,
    error,
    mutate,
  }
}
