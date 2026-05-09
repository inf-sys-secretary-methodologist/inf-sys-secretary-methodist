'use client'

import { useReducer, useState } from 'react'
import { useTranslations } from 'next-intl'
import axios from 'axios'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  bulkEditDisciplineItems,
  fetchDisciplineItem,
  useDisciplineItems,
} from '@/hooks/useDisciplineItems'
import type { CurriculumStatus } from '@/types/curriculum'

import { BulkEditTable } from './BulkEditTable'
import {
  bulkEditReducer,
  buildBulkEditRequest,
  getConflictForItem,
  hasPendingChanges,
  initialBulkEditState,
} from './bulkEditReducer'
import { pickBulkEditErrorKey } from './pickBulkEditErrorKey'

interface BulkEditPanelProps {
  sectionID: number
  curriculumStatus: CurriculumStatus
}

// canEdit gates writes. Per backend ADR-2 lifecycle inheritance, items
// are editable iff curriculum.status === 'draft'. Approved / archived /
// pending_approval = frozen for everyone.
function deriveCanEdit(status: CurriculumStatus): boolean {
  return status === 'draft'
}

// Pair 7 RED stub — minimal panel rendering. Tests asserting Submit
// button, Cancel confirm dialog, conflict banner UI, refetch-on-409
// behavior fail. GREEN replaces с full handlers.
export function BulkEditPanel({ sectionID, curriculumStatus }: BulkEditPanelProps) {
  const t = useTranslations('curriculum')
  const [state, dispatch] = useReducer(bulkEditReducer, initialBulkEditState)
  const [confirmCancelOpen, setConfirmCancelOpen] = useState(false)

  const canEdit = deriveCanEdit(curriculumStatus)
  const { items, isLoading, mutate } = useDisciplineItems(sectionID)

  // Used by tests but not wired до GREEN.
  void state
  void dispatch
  void mutate
  void confirmCancelOpen
  void setConfirmCancelOpen
  void buildBulkEditRequest
  void hasPendingChanges
  void getConflictForItem
  void bulkEditDisciplineItems
  void fetchDisciplineItem
  void axios
  void toast
  void Button
  void Loader2
  void Dialog
  void DialogContent
  void DialogDescription
  void DialogFooter
  void DialogHeader
  void DialogTitle
  void pickBulkEditErrorKey

  if (isLoading) {
    return (
      <div data-testid="bulk-edit-panel-loading" className="p-4 text-sm text-muted-foreground">
        {t('disciplineItems.bulkEdit.loading')}
      </div>
    )
  }

  return (
    <div data-testid="bulk-edit-panel" className="space-y-4">
      <BulkEditTable
        sectionID={sectionID}
        items={items}
        state={state}
        dispatch={dispatch}
        canEdit={canEdit}
      />
    </div>
  )
}
