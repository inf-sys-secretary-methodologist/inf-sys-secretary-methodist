'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { SentryConfig } from '@/types/sentry'

const SENTRY_CONFIG_URL = '/api/admin/sentry/config'

interface FetchOpts {
  // enabled defaults to true. Setting false short-circuits the SWR
  // key to null so the fetch never fires — used by /admin/sentry to
  // skip the round-trip while the role guard is still resolving.
  enabled?: boolean
}

interface Envelope {
  success: boolean
  data: SentryConfig
  meta?: unknown
  error?: { code: string; message: string }
}

const sentryFetcher = async (url: string): Promise<SentryConfig> => {
  const envelope = await apiClient.get<Envelope>(url)
  return envelope.data
}

// useSentryConfig reads GET /api/admin/sentry/config. Backend is
// gated by RequireRole(system_admin); a 403 propagates as an SWR
// error so the page can fall back to the forbidden state.
export function useSentryConfig(opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? SENTRY_CONFIG_URL : null
  const { data, error, isLoading, mutate } = useSWR<SentryConfig>(key, sentryFetcher, {
    revalidateOnFocus: false,
    dedupingInterval: SWR_DEDUPING.SHORT,
  })

  return {
    config: data,
    isLoading,
    error,
    mutate,
  }
}
