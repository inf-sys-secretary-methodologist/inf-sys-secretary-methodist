// STUB (RED) — implemented in GREEN.
import type { ImportResult, StudentDebtsFilter } from '@/types/studentDebts'

export const studentDebtsApi = {
  async import(_file: File): Promise<ImportResult> {
    throw new Error('not implemented')
  },
  async export(_filter?: StudentDebtsFilter): Promise<Blob> {
    throw new Error('not implemented')
  },
}
