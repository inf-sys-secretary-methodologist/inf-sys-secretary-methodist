'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { ComposioConfig } from '@/types/composio'

const URL = '/api/admin/composio/config'

interface FetchOpts {
  enabled?: boolean
}

interface Envelope {
  success: boolean
  data: ComposioConfig
  meta?: unknown
  error?: { code: string; message: string }
}

const fetcher = async (url: string): Promise<ComposioConfig> => {
  const envelope = await apiClient.get<Envelope>(url)
  return envelope.data
}

// useComposioConfig reads GET /api/admin/composio/config. Backend
// gated by RequireRole(system_admin); a 403 propagates as an SWR
// error so the page can fall back to the forbidden state. Mirror к
// useIntegrationsConfig + useSentryConfig SWR setup.
export function useComposioConfig(opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? URL : null
  const { data, error, isLoading, mutate } = useSWR<ComposioConfig>(key, fetcher, {
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
