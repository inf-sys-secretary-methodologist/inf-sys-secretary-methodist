'use client'

import { apiClient } from '@/lib/api'

export interface MFABeginResponse {
  otpauth_uri: string
  secret: string
}

interface MFAEnvelope<T> {
  data: T
}

// RED stub — real implementation lands in the GREEN follow-up commit.
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
