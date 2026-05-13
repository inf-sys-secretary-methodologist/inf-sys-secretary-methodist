'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { BackupListResponse } from '@/types/backup'

const BACKUPS_URL = '/api/admin/backups'

interface FetchOpts {
  // enabled defaults to true. Setting it to false short-circuits the
  // SWR key to null so the fetch never fires — used by the
  // /admin/backups page to skip the round-trip while the role
  // guard is still resolving (mirrors useAuditLogs opts pattern).
  enabled?: boolean
}

// Envelope is the response.Response wire shape (success / data / meta
// / error). The admin/backups endpoint puts the BackupListResponse
// flat under `data` — no pagination meta (file count is bounded by
// sidecar retention so a single page covers it).
interface Envelope {
  success: boolean
  data: BackupListResponse
  meta?: unknown
  error?: { code: string; message: string }
}

const backupsFetcher = async (url: string): Promise<BackupListResponse> => {
  const envelope = await apiClient.get<Envelope>(url)
  return envelope.data
}

// useBackups reads GET /api/admin/backups. Backend is gated by
// RequireRole(system_admin); a 403 propagates as an SWR error so the
// page can fall back to the forbidden state.
export function useBackups(opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? BACKUPS_URL : null
  const { data, error, isLoading, mutate } = useSWR<BackupListResponse>(key, backupsFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    files: data?.files ?? [],
    metrics: data?.metrics ?? null,
    isLoading,
    error,
    mutate,
  }
}
