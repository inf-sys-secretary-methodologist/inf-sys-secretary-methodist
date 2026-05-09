import {
  bulkEditReducer,
  initialBulkEditState,
  hasPendingChanges,
  buildBulkEditRequest,
  getConflictForItem,
  type BulkEditState,
  type PendingCreate,
} from '../bulkEditReducer'
import type { BulkEditUpdateInput } from '@/types/bulkEdit'
import type { DisciplineItem } from '@/types/disciplineItem'

// ===== Test fixtures =====

const samplePendingCreate: PendingCreate = {
  localKey: 'tmp-1',
  title: 'Новая дисциплина',
  hours_lectures: 36,
  hours_practice: 0,
  hours_lab: 0,
  hours_self: 36,
  control_form: 'zachet',
  credits: 2,
  semester: 1,
  order_index: 0,
}

const sampleUpdate: BulkEditUpdateInput = {
  id: 202,
  title: 'Обновлённая дисциплина',
  hours_lectures: 36,
  hours_practice: 36,
  hours_lab: 0,
  hours_self: 72,
  control_form: 'exam',
  credits: 4,
  semester: 1,
  order_index: 0,
}

const sampleItem: DisciplineItem = {
  id: 202,
  section_id: 101,
  title: 'Сервер видит так',
  hours_lectures: 36,
  hours_practice: 36,
  hours_lab: 0,
  hours_self: 72,
  control_form: 'exam',
  credits: 4,
  semester: 1,
  order_index: 0,
  version: 7,
  created_at: '2026-05-09T08:00:00Z',
  updated_at: '2026-05-09T08:00:00Z',
}

// ===== Reducer behavior =====

describe('bulkEditReducer / ADD_CREATE', () => {
  it('appends a pending create row', () => {
    const next = bulkEditReducer(initialBulkEditState, {
      type: 'ADD_CREATE',
      payload: samplePendingCreate,
    })
    expect(next.pendingCreates).toEqual([samplePendingCreate])
  })

  it('preserves earlier creates when appending another', () => {
    const after1 = bulkEditReducer(initialBulkEditState, {
      type: 'ADD_CREATE',
      payload: samplePendingCreate,
    })
    const second: PendingCreate = { ...samplePendingCreate, localKey: 'tmp-2', title: 'Вторая' }
    const after2 = bulkEditReducer(after1, { type: 'ADD_CREATE', payload: second })
    expect(after2.pendingCreates.map((c) => c.localKey)).toEqual(['tmp-1', 'tmp-2'])
  })
})

describe('bulkEditReducer / EDIT_CREATE', () => {
  it('patches the matched localKey row', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [samplePendingCreate],
    }
    const next = bulkEditReducer(seeded, {
      type: 'EDIT_CREATE',
      payload: { localKey: 'tmp-1', patch: { title: 'Renamed', credits: 5 } },
    })
    expect(next.pendingCreates[0].title).toBe('Renamed')
    expect(next.pendingCreates[0].credits).toBe(5)
    expect(next.pendingCreates[0].hours_lectures).toBe(36)
  })

  it('is a no-op if localKey not found', () => {
    const next = bulkEditReducer(initialBulkEditState, {
      type: 'EDIT_CREATE',
      payload: { localKey: 'missing', patch: { title: 'Никого нет' } },
    })
    expect(next.pendingCreates).toEqual([])
  })
})

describe('bulkEditReducer / REMOVE_CREATE', () => {
  it('discards the matched localKey row', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        samplePendingCreate,
        { ...samplePendingCreate, localKey: 'tmp-2', title: 'Вторая' },
      ],
    }
    const next = bulkEditReducer(seeded, { type: 'REMOVE_CREATE', payload: { localKey: 'tmp-1' } })
    expect(next.pendingCreates.map((c) => c.localKey)).toEqual(['tmp-2'])
  })
})

describe('bulkEditReducer / EDIT_ITEM', () => {
  it('upserts pendingUpdates by id (insert when new)', () => {
    const next = bulkEditReducer(initialBulkEditState, {
      type: 'EDIT_ITEM',
      payload: sampleUpdate,
    })
    expect(next.pendingUpdates).toEqual([sampleUpdate])
  })

  it('upserts pendingUpdates by id (replace when existing)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
    }
    const updated: BulkEditUpdateInput = { ...sampleUpdate, title: 'Снова обновили' }
    const next = bulkEditReducer(seeded, { type: 'EDIT_ITEM', payload: updated })
    expect(next.pendingUpdates).toHaveLength(1)
    expect(next.pendingUpdates[0].title).toBe('Снова обновили')
  })
})

describe('bulkEditReducer / REVERT_ITEM', () => {
  it('drops pendingUpdate by id', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
    }
    const next = bulkEditReducer(seeded, { type: 'REVERT_ITEM', payload: { id: 202 } })
    expect(next.pendingUpdates).toEqual([])
  })

  it('also removes a pending delete mark on the same id (full revert semantics)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
      pendingDeletes: [202],
    }
    const next = bulkEditReducer(seeded, { type: 'REVERT_ITEM', payload: { id: 202 } })
    expect(next.pendingUpdates).toEqual([])
    expect(next.pendingDeletes).toEqual([])
  })
})

describe('bulkEditReducer / TOGGLE_DELETE', () => {
  it('marks item for delete (initial)', () => {
    const next = bulkEditReducer(initialBulkEditState, {
      type: 'TOGGLE_DELETE',
      payload: { id: 203 },
    })
    expect(next.pendingDeletes).toEqual([203])
  })

  it('unmarks if already pending delete (toggle)', () => {
    const seeded: BulkEditState = { ...initialBulkEditState, pendingDeletes: [203] }
    const next = bulkEditReducer(seeded, { type: 'TOGGLE_DELETE', payload: { id: 203 } })
    expect(next.pendingDeletes).toEqual([])
  })

  it('marking for delete also drops pendingUpdate on same id', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
    }
    const next = bulkEditReducer(seeded, { type: 'TOGGLE_DELETE', payload: { id: 202 } })
    expect(next.pendingDeletes).toEqual([202])
    expect(next.pendingUpdates).toEqual([])
  })
})

describe('bulkEditReducer / SUBMIT_START', () => {
  it('sets submitting=true and clears lastErrorKey', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      lastErrorKey: 'errorGeneric',
    }
    const next = bulkEditReducer(seeded, { type: 'SUBMIT_START' })
    expect(next.submitting).toBe(true)
    expect(next.lastErrorKey).toBeNull()
  })
})

describe('bulkEditReducer / SUBMIT_SUCCESS', () => {
  it('clears all pending state + conflicts + submitting flag', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [samplePendingCreate],
      pendingUpdates: [sampleUpdate],
      pendingDeletes: [203],
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
      refreshedConflictItems: { 202: sampleItem },
      submitting: true,
      lastErrorKey: 'errorGeneric',
    }
    const next = bulkEditReducer(seeded, { type: 'SUBMIT_SUCCESS' })
    expect(next).toEqual(initialBulkEditState)
  })
})

describe('bulkEditReducer / SUBMIT_CONFLICT', () => {
  it('stores conflicts and lifts submitting flag (preserves pending state for retry)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
      submitting: true,
    }
    const conflicts = [{ id: 202, expected_version: 5, current_version: 0 }]
    const next = bulkEditReducer(seeded, { type: 'SUBMIT_CONFLICT', payload: { conflicts } })
    expect(next.conflicts).toEqual(conflicts)
    expect(next.submitting).toBe(false)
    expect(next.pendingUpdates).toEqual([sampleUpdate])
  })
})

describe('bulkEditReducer / SUBMIT_ERROR', () => {
  it('stores errorKey and lifts submitting flag (preserves pending state for retry)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
      submitting: true,
    }
    const next = bulkEditReducer(seeded, {
      type: 'SUBMIT_ERROR',
      payload: { errorKey: 'errorNotEditable' },
    })
    expect(next.lastErrorKey).toBe('errorNotEditable')
    expect(next.submitting).toBe(false)
    expect(next.pendingUpdates).toEqual([sampleUpdate])
  })
})

describe('bulkEditReducer / SET_REFRESHED_CONFLICT_ITEM', () => {
  it('stores the refetched item under its id', () => {
    const next = bulkEditReducer(initialBulkEditState, {
      type: 'SET_REFRESHED_CONFLICT_ITEM',
      payload: sampleItem,
    })
    expect(next.refreshedConflictItems[202]).toEqual(sampleItem)
  })
})

describe('bulkEditReducer / CLEAR_CONFLICTS', () => {
  it('drops conflicts + refreshed snapshots, preserves pending state', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [sampleUpdate],
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
      refreshedConflictItems: { 202: sampleItem },
    }
    const next = bulkEditReducer(seeded, { type: 'CLEAR_CONFLICTS' })
    expect(next.conflicts).toEqual([])
    expect(next.refreshedConflictItems).toEqual({})
    expect(next.pendingUpdates).toEqual([sampleUpdate])
  })
})

describe('bulkEditReducer / DISCARD_ALL', () => {
  it('returns initial state (cancel-all confirm dialog action)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [samplePendingCreate],
      pendingUpdates: [sampleUpdate],
      pendingDeletes: [203],
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
      refreshedConflictItems: { 202: sampleItem },
      lastErrorKey: 'errorGeneric',
    }
    const next = bulkEditReducer(seeded, { type: 'DISCARD_ALL' })
    expect(next).toEqual(initialBulkEditState)
  })
})

// ===== Selectors =====

describe('hasPendingChanges', () => {
  it.each([
    ['no changes (initial)', initialBulkEditState, false],
    [
      'pending create only',
      { ...initialBulkEditState, pendingCreates: [samplePendingCreate] },
      true,
    ],
    ['pending update only', { ...initialBulkEditState, pendingUpdates: [sampleUpdate] }, true],
    ['pending delete only', { ...initialBulkEditState, pendingDeletes: [203] }, true],
    [
      'conflicts present but no pending — not "changes" (the user has nothing to submit)',
      {
        ...initialBulkEditState,
        conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
      },
      false,
    ],
  ])('%s → %p', (_name, state, expected) => {
    expect(hasPendingChanges(state as BulkEditState)).toBe(expected)
  })
})

describe('buildBulkEditRequest', () => {
  it('strips localKey from creates, passes updates + deletes verbatim', () => {
    const state: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [samplePendingCreate],
      pendingUpdates: [sampleUpdate],
      pendingDeletes: [203, 204],
    }
    const body = buildBulkEditRequest(state)
    expect(body).toEqual({
      creates: [
        {
          title: 'Новая дисциплина',
          hours_lectures: 36,
          hours_practice: 0,
          hours_lab: 0,
          hours_self: 36,
          control_form: 'zachet',
          credits: 2,
          semester: 1,
          order_index: 0,
        },
      ],
      updates: [sampleUpdate],
      deletes: [203, 204],
    })
    // Ensure localKey was removed (would be 422 INVALID_INPUT on backend).
    expect(body.creates[0]).not.toHaveProperty('localKey')
  })

  it('returns empty arrays when no pending changes', () => {
    expect(buildBulkEditRequest(initialBulkEditState)).toEqual({
      creates: [],
      updates: [],
      deletes: [],
    })
  })
})

describe('getConflictForItem', () => {
  it('returns the matched conflict entry', () => {
    const state: BulkEditState = {
      ...initialBulkEditState,
      conflicts: [
        { id: 202, expected_version: 5, current_version: 0 },
        { id: 204, expected_version: 3, current_version: 0 },
      ],
    }
    expect(getConflictForItem(state, 202)).toEqual({
      id: 202,
      expected_version: 5,
      current_version: 0,
    })
  })

  it('returns undefined when item id not in conflicts', () => {
    const state: BulkEditState = {
      ...initialBulkEditState,
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
    }
    expect(getConflictForItem(state, 999)).toBeUndefined()
  })
})
