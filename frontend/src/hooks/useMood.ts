'use client'

import useSWR from 'swr'
import { swrFetcher } from '@/lib/api/fetchers'
import type { MoodResponse } from '@/types/mood'

const MOOD_URL = '/api/ai/mood'

export function useMood() {
  const { data, error, isLoading, mutate } = useSWR<MoodResponse>(
    MOOD_URL,
    swrFetcher<MoodResponse>,
    {
      revalidateOnFocus: false,
      dedupingInterval: 60000,
      refreshInterval: 300000, // 5 minutes
    }
  )

  return {
    mood: data,
    isLoading,
    error,
    mutate,
  }
}
