'use client'

// STUB (RED state). The real SWR implementation lands in the GREEN commit.
// Returns empty results without hitting the API so the hook tests fail
// behaviorally (expected items / apiClient.get call) while still resolving
// the import for the linter — the project's RED-under-linter pattern.

import type { MinobrnaukiOrder, MinobrnaukiOrderListFilter } from '@/types/minobrnaukiOrder'

interface FetchOpts {
  enabled?: boolean
}

const noopMutate = async (): Promise<undefined> => undefined

export function useMinobrnaukiOrders(_filter?: MinobrnaukiOrderListFilter, _opts?: FetchOpts) {
  return {
    items: [] as MinobrnaukiOrder[],
    total: 0,
    isLoading: false,
    error: undefined as unknown,
    mutate: noopMutate,
  }
}

export function useMinobrnaukiOrder(_id: number | null, _opts?: FetchOpts) {
  return {
    order: undefined as MinobrnaukiOrder | undefined,
    isLoading: false,
    error: undefined as unknown,
    mutate: noopMutate,
  }
}
