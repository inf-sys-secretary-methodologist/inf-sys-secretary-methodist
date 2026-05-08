'use client'

import { apiClient } from '@/lib/api'

export interface MFABeginResponse {
  otpauth_uri: string
  secret: string
}

interface MFAEnvelope<T> {
  data: T
}

// useMFA wraps the three MFA enrollment endpoints. The hook is stateless —
// callers (typically MFASettingsCard) own UI state.
export function useMFA() {
  return {
    beginEnrollment: async (): Promise<MFABeginResponse> => {
      const resp = await apiClient.post<MFAEnvelope<MFABeginResponse>>('/api/auth/mfa/begin')
      return resp.data
    },
    confirmEnrollment: async (code: string): Promise<void> => {
      await apiClient.post('/api/auth/mfa/confirm', { code })
    },
    disable: async (code: string): Promise<void> => {
      await apiClient.post('/api/auth/mfa/disable', { code })
    },
  }
}
