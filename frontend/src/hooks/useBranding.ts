'use client'

import { useState } from 'react'
import useSWR from 'swr'
import { apiClient } from '@/lib/api'
import { SWR_DEDUPING } from '@/config/swr'
import type { BrandSettings, UpdateBrandingRequest } from '@/types/branding'

const ADMIN_URL = '/api/admin/branding'
const PUBLIC_URL = '/api/public/branding'

interface FetchOpts {
  enabled?: boolean
  // public=true reads /api/public/branding (no auth required, longer
  // deduping). Default reads /api/admin/branding (admin only).
  public?: boolean
}

interface Envelope<T> {
  success: boolean
  data: T
  meta?: unknown
  error?: { code: string; message: string }
}

const fetcher = async (url: string): Promise<BrandSettings> => {
  const envelope = await apiClient.get<Envelope<BrandSettings>>(url)
  return envelope.data
}

// useBranding fetches the brand settings. With { public: true } it
// reads the unauth endpoint used by the login page; without it
// reads the admin endpoint (gated by RequireRole(system_admin) —
// non-admin callers receive an SWR error).
export function useBranding(opts?: FetchOpts) {
  const enabled = opts?.enabled ?? true
  const isPublic = opts?.public ?? false
  const url = isPublic ? PUBLIC_URL : ADMIN_URL
  const key = enabled ? url : null
  // Public branding rarely changes; use a longer dedup interval to
  // reduce DB roundtrips on the login page.
  const dedupingInterval = isPublic ? SWR_DEDUPING.LONG : SWR_DEDUPING.SHORT
  const { data, error, isLoading, mutate } = useSWR<BrandSettings>(key, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval,
  })
  return { config: data, isLoading, error, mutate }
}

interface MutationState {
  isLoading: boolean
  error: Error | null
  errorCode: string | null
}

// useUpdateBranding wraps PUT /api/admin/branding. Tracks
// isLoading + error + errorCode (the typed code surfaced by the
// backend on domain validation failures — INVALID_APP_NAME /
// INVALID_TAGLINE / INVALID_COLOR / INVALID_URL) so the form can
// render field-specific inline messages without parsing error
// message strings.
export function useUpdateBranding() {
  const [state, setState] = useState<MutationState>({
    isLoading: false,
    error: null,
    errorCode: null,
  })

  const updateBranding = async (req: UpdateBrandingRequest): Promise<BrandSettings> => {
    setState({ isLoading: true, error: null, errorCode: null })
    try {
      const envelope = await apiClient.put<Envelope<BrandSettings>>(ADMIN_URL, req)
      setState({ isLoading: false, error: null, errorCode: null })
      return envelope.data
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e))
      const code = extractErrorCode(e)
      setState({ isLoading: false, error: err, errorCode: code })
      throw err
    }
  }

  return { updateBranding, ...state }
}

// extractErrorCode reads the backend-typed error code from the
// thrown error. apiClient throws errors that may carry the
// envelope's error.code on a `.code` property; we read it
// defensively, returning null on any miss.
function extractErrorCode(e: unknown): string | null {
  if (e && typeof e === 'object' && 'code' in e) {
    const c = (e as { code?: unknown }).code
    if (typeof c === 'string') return c
  }
  if (e && typeof e === 'object' && 'response' in e) {
    const r = (e as { response?: { data?: { error?: { code?: unknown } } } }).response
    const c = r?.data?.error?.code
    if (typeof c === 'string') return c
  }
  return null
}
