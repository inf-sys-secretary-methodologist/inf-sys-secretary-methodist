'use client'

import type { Dispatch } from 'react'
import { useTranslations } from 'next-intl'

import type { DisciplineItem } from '@/types/disciplineItem'
import { CONTROL_FORMS } from '@/types/disciplineItem'
import type { BulkEditAction, BulkEditState, PendingCreate } from './bulkEditReducer'

interface BulkEditTableProps {
  sectionID: number
  items: DisciplineItem[]
  state: BulkEditState
  dispatch: Dispatch<BulkEditAction>
  canEdit: boolean
}

// Pair 6 RED stub — minimal placeholder. Tests asserting column headers,
// per-row inputs, Add button, ControlForm options fail. GREEN replaces
// с full implementation.
export function BulkEditTable(props: BulkEditTableProps) {
  void props
  void CONTROL_FORMS
  const _ = useTranslations('curriculum')
  void _
  return <div data-testid="bulk-edit-table-placeholder" />
}

// Helper retained для GREEN pair — produces a stable localKey for new
// pending creates. Using crypto.randomUUID when available, fallback к
// Date.now()+counter; collisions tolerated since localKey is purely
// client-side row identity.
let _localKeyCounter = 0
export function nextLocalKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  _localKeyCounter += 1
  return `local-${Date.now()}-${_localKeyCounter}`
}

// Helper exported для GREEN: composes the displayed value of an item
// cell — returns the pending update's value if one exists, otherwise
// falls back к the server snapshot.
export function displayedField<K extends keyof DisciplineItem>(
  item: DisciplineItem,
  pending: PendingCreate | undefined,
  state: BulkEditState,
  field: K
): DisciplineItem[K] {
  void pending
  const pendingUpdate = state.pendingUpdates.find((u) => u.id === item.id)
  if (pendingUpdate && field in pendingUpdate) {
    return (pendingUpdate as unknown as DisciplineItem)[field]
  }
  return item[field]
}
