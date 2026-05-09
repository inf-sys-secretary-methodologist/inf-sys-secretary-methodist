// Pure state machine для BulkEditTable. useReducer per Q1 brainstorm
// (resolved A — local page state, не Zustand). Encodes:
//  - pending creates / updates / deletes accumulated через cell editing
//  - submit lifecycle (idle / submitting / success / conflict / error)
//  - 409 conflict state с refreshed-item snapshots from
//    fetchDisciplineItem (per plan ADR-12 outside-tx refetch)
//
// Pure: no React, no axios, no side effects. All async (apiClient,
// fetchDisciplineItem) lives in BulkEditTable component which dispatches
// actions on resolution. Selectors exported separately (used by Submit
// affordance + per-row conflict banner).

import type {
  BulkEditConflict,
  BulkEditCreateInput,
  BulkEditRequest,
  BulkEditUpdateInput,
} from '@/types/bulkEdit'
import type { DisciplineItem } from '@/types/disciplineItem'
import type { BulkEditErrorKey } from './pickBulkEditErrorKey'

// Pending create rows carry a local key — backend has not assigned an
// id yet. Stable identity for React list rendering + selection state.
export interface PendingCreate extends BulkEditCreateInput {
  localKey: string
}

export interface BulkEditState {
  pendingCreates: PendingCreate[]
  pendingUpdates: BulkEditUpdateInput[]
  pendingDeletes: number[]
  conflicts: BulkEditConflict[]
  refreshedConflictItems: Record<number, DisciplineItem>
  submitting: boolean
  lastErrorKey: BulkEditErrorKey | null
}

export const initialBulkEditState: BulkEditState = {
  pendingCreates: [],
  pendingUpdates: [],
  pendingDeletes: [],
  conflicts: [],
  refreshedConflictItems: {},
  submitting: false,
  lastErrorKey: null,
}

export type BulkEditAction =
  | { type: 'ADD_CREATE'; payload: PendingCreate }
  | { type: 'EDIT_CREATE'; payload: { localKey: string; patch: Partial<BulkEditCreateInput> } }
  | { type: 'REMOVE_CREATE'; payload: { localKey: string } }
  | { type: 'EDIT_ITEM'; payload: BulkEditUpdateInput }
  | { type: 'REVERT_ITEM'; payload: { id: number } }
  | { type: 'TOGGLE_DELETE'; payload: { id: number } }
  | { type: 'SUBMIT_START' }
  | { type: 'SUBMIT_SUCCESS' }
  | { type: 'SUBMIT_CONFLICT'; payload: { conflicts: BulkEditConflict[] } }
  | { type: 'SUBMIT_ERROR'; payload: { errorKey: BulkEditErrorKey } }
  | { type: 'SET_REFRESHED_CONFLICT_ITEM'; payload: DisciplineItem }
  | { type: 'CLEAR_CONFLICTS' }
  | { type: 'DISCARD_ALL' }

// Pair 5 RED stub — identity reducer ignores actions. Tests asserting
// post-action state fail. GREEN replaces с full switch.
export function bulkEditReducer(state: BulkEditState, action: BulkEditAction): BulkEditState {
  void action
  return state
}

// ===== Selectors =====

export function hasPendingChanges(state: BulkEditState): boolean {
  // RED stub — always false. GREEN returns OR of three lengths.
  void state
  return false
}

export function buildBulkEditRequest(state: BulkEditState): BulkEditRequest {
  // RED stub — always empty body. GREEN composes from pending state.
  void state
  return { creates: [], updates: [], deletes: [] }
}

export function getConflictForItem(
  state: BulkEditState,
  itemID: number
): BulkEditConflict | undefined {
  // RED stub — always undefined. GREEN finds in conflicts array.
  void state
  void itemID
  return undefined
}
