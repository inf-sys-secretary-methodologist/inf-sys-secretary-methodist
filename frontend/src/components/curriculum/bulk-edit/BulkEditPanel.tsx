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
  hasPendingChanges,
  initialBulkEditState,
} from './bulkEditReducer'
import { pickBulkEditErrorKey } from './pickBulkEditErrorKey'

interface BulkEditPanelProps {
  sectionID: number
  curriculumStatus: CurriculumStatus
}

// canEdit gates writes. Per backend ADR-2 lifecycle inheritance, items
// are editable iff curriculum.status === 'draft'.
function deriveCanEdit(status: CurriculumStatus): boolean {
  return status === 'draft'
}

export function BulkEditPanel({ sectionID, curriculumStatus }: BulkEditPanelProps) {
  const t = useTranslations('curriculum')
  const [state, dispatch] = useReducer(bulkEditReducer, initialBulkEditState)
  const [confirmCancelOpen, setConfirmCancelOpen] = useState(false)

  const canEdit = deriveCanEdit(curriculumStatus)
  const { items, isLoading, mutate } = useDisciplineItems(sectionID)

  const dirty = hasPendingChanges(state)

  const handleSubmit = async () => {
    if (state.submitting || !dirty) return
    dispatch({ type: 'SUBMIT_START' })
    try {
      const result = await bulkEditDisciplineItems(sectionID, buildBulkEditRequest(state))
      if (result.kind === 'success') {
        dispatch({ type: 'SUBMIT_SUCCESS' })
        toast.success(t('disciplineItems.bulkEdit.successToast'))
        mutate()
        return
      }
      // 409 conflict — store conflicts (Submit stays disabled because
      // SUBMIT_CONFLICT keeps submitting=true), then refetch each id
      // outside the failed tx per plan ADR-12 (CurrentVersion=0 hint,
      // real value lives on a fresh GET /api/items/:id snapshot).
      dispatch({ type: 'SUBMIT_CONFLICT', payload: { conflicts: result.conflicts } })
      await Promise.all(
        result.conflicts.map(async (c) => {
          try {
            const refreshed = await fetchDisciplineItem(c.id)
            dispatch({ type: 'SET_REFRESHED_CONFLICT_ITEM', payload: refreshed })
          } catch (refetchErr) {
            // Refetch failure is non-fatal — banner still shows the
            // expected_version hint и user can retry. Log so оператор
            // увидит при triage; no toast чтобы не дублировать main
            // 409 visual signal.
            console.error('bulk-edit conflict refetch failed for id', c.id, refetchErr)
          }
        })
      )
      // Re-enable Submit only after refetch loop has resolved — без
      // этого user мог re-click submit с stale expected_version.
      dispatch({ type: 'SUBMIT_CONFLICT_REFRESHED' })
    } catch (err) {
      const status = axios.isAxiosError(err) ? err.response?.status : undefined
      const code =
        axios.isAxiosError(err) &&
        err.response?.data &&
        typeof err.response.data === 'object' &&
        'error' in err.response.data
          ? (err.response.data as { error?: { code?: string } }).error?.code
          : undefined
      const errorKey = pickBulkEditErrorKey(status, code)
      dispatch({ type: 'SUBMIT_ERROR', payload: { errorKey } })
      toast.error(t(`disciplineItems.bulkEdit.${errorKey}`))
    }
  }

  const handleCancelClick = () => {
    // Guard against the in-flight submit / refetch race — cancelling
    // mid-refetch would dispatch DISCARD_ALL while async
    // SET_REFRESHED_CONFLICT_ITEM dispatches are queued. The reducer
    // also guards via the ghost-write check, but blocking here keeps
    // the UI affordance honest.
    if (dirty && !state.submitting) setConfirmCancelOpen(true)
  }

  const handleConfirmCancel = () => {
    dispatch({ type: 'DISCARD_ALL' })
    setConfirmCancelOpen(false)
  }

  if (isLoading) {
    return (
      <div data-testid="bulk-edit-panel-loading" className="p-4 text-sm text-muted-foreground">
        {t('disciplineItems.bulkEdit.loading')}
      </div>
    )
  }

  return (
    <div data-testid="bulk-edit-panel" className="space-y-4">
      {state.conflicts.length > 0 && (
        <div className="space-y-2 rounded border border-amber-300 bg-amber-50/60 p-3 dark:border-amber-700 dark:bg-amber-950/30">
          <p className="text-sm font-medium">
            {t('disciplineItems.bulkEdit.conflictBanner.heading')}
          </p>
          {state.conflicts.map((conflict) => {
            const refreshed = state.refreshedConflictItems[conflict.id]
            return (
              <div
                key={conflict.id}
                data-testid={`bulk-edit-conflict-banner-${conflict.id}`}
                className="flex items-center justify-between gap-3 rounded border border-amber-200 bg-background p-2 text-sm"
              >
                <div>
                  <p className="font-medium">{refreshed ? refreshed.title : `#${conflict.id}`}</p>
                  <p className="text-xs text-muted-foreground">
                    {t('disciplineItems.bulkEdit.conflictBanner.message', {
                      expected: conflict.expected_version,
                    })}
                  </p>
                </div>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  data-testid={`bulk-edit-conflict-banner-${conflict.id}-apply-server`}
                  onClick={() => dispatch({ type: 'REVERT_ITEM', payload: { id: conflict.id } })}
                >
                  {t('disciplineItems.bulkEdit.conflictBanner.applyServer')}
                </Button>
              </div>
            )
          })}
        </div>
      )}

      <BulkEditTable
        sectionID={sectionID}
        items={items}
        state={state}
        dispatch={dispatch}
        canEdit={canEdit}
      />

      {canEdit && (
        <div className="flex items-center justify-end gap-2">
          {dirty && (
            <Button
              type="button"
              variant="outline"
              onClick={handleCancelClick}
              disabled={state.submitting}
            >
              {t('disciplineItems.bulkEdit.cancel')}
            </Button>
          )}
          <Button type="button" onClick={handleSubmit} disabled={!dirty || state.submitting}>
            {state.submitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            {t('disciplineItems.bulkEdit.submit')}
          </Button>
        </div>
      )}

      <Dialog open={confirmCancelOpen} onOpenChange={setConfirmCancelOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('disciplineItems.bulkEdit.cancelDialog.title')}</DialogTitle>
            <DialogDescription>
              {t('disciplineItems.bulkEdit.cancelDialog.description')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={() => setConfirmCancelOpen(false)}>
              {t('disciplineItems.bulkEdit.cancelDialog.keepEditing')}
            </Button>
            <Button type="button" variant="destructive" onClick={handleConfirmCancel}>
              {t('disciplineItems.bulkEdit.cancelDialog.confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
