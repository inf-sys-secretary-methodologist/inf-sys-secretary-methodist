'use client'

import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { IntegrationsConfig } from '@/types/integrations'

const URL = '/api/admin/integrations/config'

interface FetchOpts {
  enabled?: boolean
}

interface Envelope {
  success: boolean
  data: IntegrationsConfig
  meta?: unknown
  error?: { code: string; message: string }
}

const fetcher = async (url: string): Promise<IntegrationsConfig> => {
  const envelope = await apiClient.get<Envelope>(url)
  return envelope.data
}

// useIntegrationsConfig reads GET /api/admin/integrations/config.
// Backend gated by RequireRole(system_admin); a 403 propagates as
// an SWR error so the page can fall back to the forbidden state.
export function useIntegrationsConfig(opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const key = enabled ? URL : null
  const { data, error, isLoading, mutate } = useSWR<IntegrationsConfig>(key, fetcher, {
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
