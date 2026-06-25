'use client'

// STUB (RED) — implemented in GREEN.
import type {
  StudentDebt,
  StudentDebtListResponse,
  StudentDebtStats,
  StudentDebtsFilter,
  ScheduleResitInput,
  RecordResitResultInput,
} from '@/types/studentDebts'

interface FetchOpts {
  enabled?: boolean
}

export function useStudentDebts(_filter?: StudentDebtsFilter, _opts?: FetchOpts) {
  return { items: [], total: 0, isLoading: false, error: undefined, mutate: async () => {} } as {
    items: StudentDebtListResponse['items']
    total: number
    isLoading: boolean
    error: unknown
    mutate: () => Promise<unknown>
  }
}

export function useStudentDebt(_id: number | null, _opts?: FetchOpts) {
  return { debt: undefined, isLoading: false, error: undefined, mutate: async () => {} } as {
    debt: StudentDebt | undefined
    isLoading: boolean
    error: unknown
    mutate: () => Promise<unknown>
  }
}

export function useMyStudentDebts(_filter?: StudentDebtsFilter, _opts?: FetchOpts) {
  return { items: [], total: 0, isLoading: false, error: undefined, mutate: async () => {} } as {
    items: StudentDebtListResponse['items']
    total: number
    isLoading: boolean
    error: unknown
    mutate: () => Promise<unknown>
  }
}

export function useDebtStats(_filter?: StudentDebtsFilter, _opts?: FetchOpts) {
  return { stats: undefined, isLoading: false, error: undefined, mutate: async () => {} } as {
    stats: StudentDebtStats | undefined
    isLoading: boolean
    error: unknown
    mutate: () => Promise<unknown>
  }
}

export async function scheduleResit(_id: number, _input: ScheduleResitInput): Promise<StudentDebt> {
  throw new Error('not implemented')
}

export async function recordResitResult(
  _id: number,
  _attemptNo: number,
  _input: RecordResitResultInput
): Promise<StudentDebt> {
  throw new Error('not implemented')
}

export function pickStudentDebtErrorKey(_err: unknown): string {
  return 'generic'
}
