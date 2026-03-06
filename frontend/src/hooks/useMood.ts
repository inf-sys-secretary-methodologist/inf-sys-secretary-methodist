'use client'

import useSWR from 'swr'
import { swrFetcher } from '@/lib/api/fetchers'
import { SWR_DEDUPING, SWR_REFRESH } from '@/config/swr'
import type { MoodResponse } from '@/types/mood'

const MOOD_URL = '/api/ai/mood'

export function useMood() {
  const { data, error, isLoading, mutate } = useSWR<MoodResponse>(
    MOOD_URL,
    swrFetcher<MoodResponse>,
    {
      revalidateOnFocus: false,
      dedupingInterval: SWR_DEDUPING.EXTRA_LONG,
      refreshInterval: SWR_REFRESH.SLOW, // 5 minutes
    }
  )

  return {
    mood: data,
    isLoading,
    error,
    mutate,
  }
}
