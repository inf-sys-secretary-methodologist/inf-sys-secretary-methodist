'use client'

import type { Dispatch } from 'react'
import { useTranslations } from 'next-intl'

import type { DisciplineItem } from '@/types/disciplineItem'
import { CONTROL_FORMS, type ControlForm } from '@/types/disciplineItem'
import type { BulkEditCreateInput, BulkEditUpdateInput } from '@/types/bulkEdit'
import type { BulkEditAction, BulkEditState, PendingCreate } from './bulkEditReducer'

interface BulkEditTableProps {
  sectionID: number
  items: DisciplineItem[]
  state: BulkEditState
  dispatch: Dispatch<BulkEditAction>
  canEdit: boolean
}

let _localKeyCounter = 0
export function nextLocalKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  _localKeyCounter += 1
  return `local-${Date.now()}-${_localKeyCounter}`
}

// Composes the live row view = server item ⊕ any pending update.
// Returned shape matches DisciplineItem so cell readers stay simple.
function effectiveRow(item: DisciplineItem, state: BulkEditState): DisciplineItem {
  const pending = state.pendingUpdates.find((u) => u.id === item.id)
  return pending ? { ...item, ...pending } : item
}

// Strips DisciplineItem to the BulkEditUpdateInput contract (drops id-
// adjacent server-only fields: section_id / version / created_at /
// updated_at). Keeps id + всех editable cells.
function toUpdateInput(row: DisciplineItem): BulkEditUpdateInput {
  return {
    id: row.id,
    title: row.title,
    hours_lectures: row.hours_lectures,
    hours_practice: row.hours_practice,
    hours_lab: row.hours_lab,
    hours_self: row.hours_self,
    control_form: row.control_form,
    credits: row.credits,
    semester: row.semester,
    order_index: row.order_index,
  }
}

// Number input parser — empty / non-numeric strings collapse к 0 for
// downstream backend numeric invariants. Backend re-validates ranges
// server-side; frontend coercion is purely UX.
function asInt(value: string): number {
  const n = Number.parseInt(value, 10)
  return Number.isFinite(n) ? n : 0
}

const DEFAULT_PENDING_CREATE: Omit<BulkEditCreateInput, 'order_index'> = {
  title: '',
  hours_lectures: 0,
  hours_practice: 0,
  hours_lab: 0,
  hours_self: 0,
  control_form: 'zachet',
  credits: 1,
  semester: 1,
}

export function BulkEditTable({ sectionID, items, state, dispatch, canEdit }: BulkEditTableProps) {
  const t = useTranslations('curriculum')

  const isEmpty = items.length === 0 && state.pendingCreates.length === 0

  const handleAddRow = () => {
    const order_index = items.length + state.pendingCreates.length // append at the end of the list
    dispatch({
      type: 'ADD_CREATE',
      payload: { localKey: nextLocalKey(), ...DEFAULT_PENDING_CREATE, order_index },
    })
  }

  const editExistingField = <K extends keyof BulkEditUpdateInput>(
    item: DisciplineItem,
    field: K,
    value: BulkEditUpdateInput[K]
  ) => {
    const row = effectiveRow(item, state)
    const next = toUpdateInput({ ...row, [field]: value })
    dispatch({ type: 'EDIT_ITEM', payload: next })
  }

  const editPendingCreateField = <K extends keyof BulkEditCreateInput>(
    pending: PendingCreate,
    field: K,
    value: BulkEditCreateInput[K]
  ) => {
    dispatch({
      type: 'EDIT_CREATE',
      payload: {
        localKey: pending.localKey,
        patch: { [field]: value } as Partial<BulkEditCreateInput>,
      },
    })
  }

  return (
    <div className="space-y-3" data-testid={`bulk-edit-table-${sectionID}`}>
      {!isEmpty && (
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="bg-muted/30 text-left">
              {canEdit && <th className="p-2" aria-label="select" />}
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.title')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.hoursLectures')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.hoursPractice')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.hoursLab')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.hoursSelf')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.controlForm')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.credits')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.semester')}</th>
              <th className="p-2">{t('disciplineItems.bulkEdit.columns.order')}</th>
            </tr>
          </thead>
          <tbody>
            {items.map((item) => {
              const row = effectiveRow(item, state)
              const isPendingDelete = state.pendingDeletes.includes(item.id)
              return (
                <tr
                  key={item.id}
                  data-testid={`bulk-edit-row-${item.id}`}
                  data-pending-delete={isPendingDelete ? 'true' : 'false'}
                  className={isPendingDelete ? 'opacity-50 line-through' : ''}
                >
                  {canEdit && (
                    <td className="p-2">
                      <input
                        type="checkbox"
                        data-testid={`bulk-edit-row-${item.id}-delete-toggle`}
                        checked={isPendingDelete}
                        onChange={() =>
                          dispatch({ type: 'TOGGLE_DELETE', payload: { id: item.id } })
                        }
                      />
                    </td>
                  )}
                  <td className="p-2">
                    <input
                      type="text"
                      value={row.title}
                      readOnly={!canEdit}
                      onChange={(e) => editExistingField(item, 'title', e.target.value)}
                      className="w-full rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.hours_lectures}
                      readOnly={!canEdit}
                      onChange={(e) =>
                        editExistingField(item, 'hours_lectures', asInt(e.target.value))
                      }
                      className="w-20 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.hours_practice}
                      readOnly={!canEdit}
                      onChange={(e) =>
                        editExistingField(item, 'hours_practice', asInt(e.target.value))
                      }
                      className="w-20 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.hours_lab}
                      readOnly={!canEdit}
                      onChange={(e) => editExistingField(item, 'hours_lab', asInt(e.target.value))}
                      className="w-20 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.hours_self}
                      readOnly={!canEdit}
                      onChange={(e) => editExistingField(item, 'hours_self', asInt(e.target.value))}
                      className="w-20 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <select
                      data-testid={`bulk-edit-row-${item.id}-control-form-select`}
                      value={row.control_form}
                      disabled={!canEdit}
                      onChange={(e) =>
                        editExistingField(item, 'control_form', e.target.value as ControlForm)
                      }
                      className="rounded border bg-background p-1"
                    >
                      {CONTROL_FORMS.map((cf) => (
                        <option key={cf} value={cf}>
                          {t(`disciplineItems.controlForm.${cf}`)}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.credits}
                      readOnly={!canEdit}
                      onChange={(e) => editExistingField(item, 'credits', asInt(e.target.value))}
                      className="w-16 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.semester}
                      readOnly={!canEdit}
                      onChange={(e) => editExistingField(item, 'semester', asInt(e.target.value))}
                      className="w-16 rounded border bg-background p-1"
                    />
                  </td>
                  <td className="p-2">
                    <input
                      type="number"
                      min={0}
                      value={row.order_index}
                      readOnly={!canEdit}
                      onChange={(e) =>
                        editExistingField(item, 'order_index', asInt(e.target.value))
                      }
                      className="w-16 rounded border bg-background p-1"
                    />
                  </td>
                </tr>
              )
            })}
            {state.pendingCreates.map((pending) => (
              <tr
                key={pending.localKey}
                data-testid={`bulk-edit-row-create-${pending.localKey}`}
                className="bg-emerald-50/40 dark:bg-emerald-950/20"
              >
                {canEdit && (
                  <td className="p-2">
                    <button
                      type="button"
                      data-testid={`bulk-edit-row-create-${pending.localKey}-remove`}
                      onClick={() =>
                        dispatch({
                          type: 'REMOVE_CREATE',
                          payload: { localKey: pending.localKey },
                        })
                      }
                      aria-label={t('disciplineItems.bulkEdit.removeRow')}
                      className="text-destructive"
                    >
                      ×
                    </button>
                  </td>
                )}
                <td className="p-2">
                  <input
                    type="text"
                    data-testid={`bulk-edit-row-create-${pending.localKey}-title-input`}
                    value={pending.title}
                    readOnly={!canEdit}
                    onChange={(e) => editPendingCreateField(pending, 'title', e.target.value)}
                    className="w-full rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.hours_lectures}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'hours_lectures', asInt(e.target.value))
                    }
                    className="w-20 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.hours_practice}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'hours_practice', asInt(e.target.value))
                    }
                    className="w-20 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.hours_lab}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'hours_lab', asInt(e.target.value))
                    }
                    className="w-20 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.hours_self}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'hours_self', asInt(e.target.value))
                    }
                    className="w-20 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <select
                    value={pending.control_form}
                    disabled={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'control_form', e.target.value as ControlForm)
                    }
                    className="rounded border bg-background p-1"
                  >
                    {CONTROL_FORMS.map((cf) => (
                      <option key={cf} value={cf}>
                        {t(`disciplineItems.controlForm.${cf}`)}
                      </option>
                    ))}
                  </select>
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.credits}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'credits', asInt(e.target.value))
                    }
                    className="w-16 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.semester}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'semester', asInt(e.target.value))
                    }
                    className="w-16 rounded border bg-background p-1"
                  />
                </td>
                <td className="p-2">
                  <input
                    type="number"
                    min={0}
                    value={pending.order_index}
                    readOnly={!canEdit}
                    onChange={(e) =>
                      editPendingCreateField(pending, 'order_index', asInt(e.target.value))
                    }
                    className="w-16 rounded border bg-background p-1"
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {isEmpty && (
        <p className="text-sm text-muted-foreground">{t('disciplineItems.bulkEdit.empty')}</p>
      )}

      {canEdit && (
        <button
          type="button"
          onClick={handleAddRow}
          className="rounded border border-dashed px-3 py-1 text-sm hover:bg-accent"
        >
          {t('disciplineItems.bulkEdit.addRow')}
        </button>
      )}
    </div>
  )
}
