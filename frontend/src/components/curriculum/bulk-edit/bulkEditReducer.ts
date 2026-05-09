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

export function bulkEditReducer(state: BulkEditState, action: BulkEditAction): BulkEditState {
  switch (action.type) {
    case 'ADD_CREATE':
      return { ...state, pendingCreates: [...state.pendingCreates, action.payload] }
    case 'EDIT_CREATE':
      return {
        ...state,
        pendingCreates: state.pendingCreates.map((c) =>
          c.localKey === action.payload.localKey ? { ...c, ...action.payload.patch } : c
        ),
      }
    case 'REMOVE_CREATE':
      return {
        ...state,
        pendingCreates: state.pendingCreates.filter((c) => c.localKey !== action.payload.localKey),
      }
    case 'EDIT_ITEM': {
      const exists = state.pendingUpdates.some((u) => u.id === action.payload.id)
      const nextUpdates = exists
        ? state.pendingUpdates.map((u) => (u.id === action.payload.id ? action.payload : u))
        : [...state.pendingUpdates, action.payload]
      return { ...state, pendingUpdates: nextUpdates }
    }
    case 'REVERT_ITEM':
      return {
        ...state,
        pendingUpdates: state.pendingUpdates.filter((u) => u.id !== action.payload.id),
        pendingDeletes: state.pendingDeletes.filter((d) => d !== action.payload.id),
      }
    case 'TOGGLE_DELETE': {
      const id = action.payload.id
      const alreadyMarked = state.pendingDeletes.includes(id)
      if (alreadyMarked) {
        return { ...state, pendingDeletes: state.pendingDeletes.filter((d) => d !== id) }
      }
      return {
        ...state,
        pendingDeletes: [...state.pendingDeletes, id],
        pendingUpdates: state.pendingUpdates.filter((u) => u.id !== id),
      }
    }
    case 'SUBMIT_START':
      return { ...state, submitting: true, lastErrorKey: null }
    case 'SUBMIT_SUCCESS':
      return initialBulkEditState
    case 'SUBMIT_CONFLICT':
      return { ...state, submitting: false, conflicts: action.payload.conflicts }
    case 'SUBMIT_ERROR':
      return { ...state, submitting: false, lastErrorKey: action.payload.errorKey }
    case 'SET_REFRESHED_CONFLICT_ITEM':
      return {
        ...state,
        refreshedConflictItems: {
          ...state.refreshedConflictItems,
          [action.payload.id]: action.payload,
        },
      }
    case 'CLEAR_CONFLICTS':
      return { ...state, conflicts: [], refreshedConflictItems: {} }
    case 'DISCARD_ALL':
      return initialBulkEditState
  }
}

// ===== Selectors =====

export function hasPendingChanges(state: BulkEditState): boolean {
  return (
    state.pendingCreates.length > 0 ||
    state.pendingUpdates.length > 0 ||
    state.pendingDeletes.length > 0
  )
}

export function buildBulkEditRequest(state: BulkEditState): BulkEditRequest {
  return {
    creates: state.pendingCreates.map(({ localKey: _key, ...rest }) => rest),
    updates: state.pendingUpdates,
    deletes: state.pendingDeletes,
  }
}

export function getConflictForItem(
  state: BulkEditState,
  itemID: number
): BulkEditConflict | undefined {
  return state.conflicts.find((c) => c.id === itemID)
}
