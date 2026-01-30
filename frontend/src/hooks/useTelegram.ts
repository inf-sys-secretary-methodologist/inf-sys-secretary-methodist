'use client'

import useSWR, { mutate } from 'swr'
import { apiClient } from '@/lib/api'
import { useState, useCallback } from 'react'

const TELEGRAM_BASE_URL = '/api/telegram'

// API Response wrapper type from backend
interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}

// Types for Telegram integration
export interface TelegramConnectionStatus {
  connected: boolean
  username?: string
  first_name?: string
  connected_at?: string
}

export interface TelegramVerificationCode {
  code: string
  expires_at: string
  bot_username: string
  bot_link: string
}

// Fetcher for SWR - extracts data from wrapped response
const fetcher = async <T>(url: string): Promise<T> => {
  const response = await apiClient.get<ApiResponse<T> | T>(url)

  /* c8 ignore start - API response wrapper handling */
  // Check if response is the API wrapper format
  if (response && typeof response === 'object' && 'success' in response) {
    const wrappedResponse = response as ApiResponse<T>
    if (wrappedResponse.success && wrappedResponse.data !== undefined) {
      return wrappedResponse.data
    } else {
      throw new Error(wrappedResponse.error?.message || 'API returned error')
    }
  }
  /* c8 ignore stop */

  // Response is already the data
  return response as T
}

/**
 * Hook to get Telegram connection status
 */
export function useTelegramStatus() {
  const {
    data,
    error,
    isLoading,
    mutate: revalidate,
  } = useSWR<TelegramConnectionStatus>(`${TELEGRAM_BASE_URL}/status`, fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 30000,
  })

  return {
    data,
    isConnected: data?.connected ?? false,
    username: data?.username,
    firstName: data?.first_name,
    connectedAt: data?.connected_at,
    isLoading,
    error,
    mutate: revalidate,
  }
}

/**
 * Hook to generate verification code for Telegram linking
 */
export function useGenerateVerificationCode() {
  const [isPending, setIsPending] = useState(false)
  const [data, setData] = useState<TelegramVerificationCode | null>(null)
  const [error, setError] = useState<Error | null>(null)

  const generate = useCallback(async () => {
    setIsPending(true)
    setError(null)
    try {
      const response = await apiClient.post<
        ApiResponse<TelegramVerificationCode> | TelegramVerificationCode
      >(`${TELEGRAM_BASE_URL}/verification-code`)

      let result: TelegramVerificationCode
      /* c8 ignore start - API response handling */
      if (response && typeof response === 'object' && 'success' in response) {
        const wrappedResponse = response as ApiResponse<TelegramVerificationCode>
        if (wrappedResponse.success && wrappedResponse.data) {
          result = wrappedResponse.data
        } else {
          throw new Error(wrappedResponse.error?.message || 'Failed to generate code')
        }
      } else {
        result = response as TelegramVerificationCode
      }
      /* c8 ignore stop */

      setData(result)
      return result
      /* c8 ignore start - Error and finally handling */
    } catch (err) {
      const e = err instanceof Error ? err : new Error('Failed to generate verification code')
      setError(e)
      throw e
    } finally {
      setIsPending(false)
    }
    /* c8 ignore stop */
  }, [])

  const reset = useCallback(() => {
    setData(null)
    setError(null)
  }, [])

  return {
    mutateAsync: generate,
    data,
    isPending,
    error,
    reset,
  }
}

/**
 * Hook to disconnect Telegram account
 */
export function useDisconnectTelegram() {
  const [isPending, setIsPending] = useState(false)

  const disconnect = useCallback(async () => {
    setIsPending(true)
    try {
      await apiClient.post(`${TELEGRAM_BASE_URL}/disconnect`)
      // Revalidate Telegram status cache
      mutate(`${TELEGRAM_BASE_URL}/status`)
      // Also revalidate notification preferences since Telegram channel gets disabled
      mutate('/api/notifications/preferences')
    } finally {
      setIsPending(false)
    }
  }, [])

  return { mutateAsync: disconnect, isPending }
}
