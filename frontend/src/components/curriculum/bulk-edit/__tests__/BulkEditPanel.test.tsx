import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { BulkEditPanel } from '../BulkEditPanel'
import type { DisciplineItem } from '@/types/disciplineItem'
import type { BulkEditResult } from '@/types/bulkEdit'

const mockMutate = jest.fn()
const mockUseDisciplineItems = jest.fn()
const mockBulkEditDisciplineItems = jest.fn()
const mockFetchDisciplineItem = jest.fn()

jest.mock('@/hooks/useDisciplineItems', () => ({
  useDisciplineItems: (...args: unknown[]) => mockUseDisciplineItems(...args),
  bulkEditDisciplineItems: (...args: unknown[]) => mockBulkEditDisciplineItems(...args),
  fetchDisciplineItem: (...args: unknown[]) => mockFetchDisciplineItem(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

const sampleItem: DisciplineItem = {
  id: 202,
  section_id: 101,
  title: 'Математический анализ',
  hours_lectures: 36,
  hours_practice: 36,
  hours_lab: 0,
  hours_self: 72,
  control_form: 'exam',
  credits: 4,
  semester: 1,
  order_index: 0,
  version: 5,
  created_at: '2026-05-09T08:00:00Z',
  updated_at: '2026-05-09T08:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
  mockUseDisciplineItems.mockReturnValue({
    items: [sampleItem],
    isLoading: false,
    error: undefined,
    mutate: mockMutate,
  })
})

describe('BulkEditPanel / loading', () => {
  it('renders loading placeholder while items are fetching', () => {
    mockUseDisciplineItems.mockReturnValueOnce({
      items: [],
      isLoading: true,
      error: undefined,
      mutate: mockMutate,
    })
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    expect(screen.getByTestId('bulk-edit-panel-loading')).toBeInTheDocument()
  })
})

describe('BulkEditPanel / Submit button affordance', () => {
  it('Submit is disabled when no pending changes (Q5 affordance)', () => {
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    const submitBtn = screen.getByRole('button', {
      name: 'disciplineItems.bulkEdit.submit',
    })
    expect(submitBtn).toBeDisabled()
  })

  it('Submit becomes enabled when user marks an item for delete', () => {
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeEnabled()
  })
})

describe('BulkEditPanel / Submit success', () => {
  it('on 200 success: clears pending state, fires SWR mutate(), shows success toast', async () => {
    const successResult: BulkEditResult = {
      kind: 'success',
      data: { created: [], updated: [], deleted: [202] },
    }
    mockBulkEditDisciplineItems.mockResolvedValueOnce(successResult)

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    // Mark item for delete to enable Submit.
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() => expect(mockToastSuccess).toHaveBeenCalled())
    expect(mockBulkEditDisciplineItems).toHaveBeenCalledWith(101, {
      creates: [],
      updates: [],
      deletes: [202],
    })
    expect(mockMutate).toHaveBeenCalled()
    // Submit becomes disabled again — pending state was cleared.
    expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeDisabled()
  })
})

describe('BulkEditPanel / Submit 409 conflict', () => {
  it('on 409: stores conflicts, refetches each conflict item, shows per-row banner', async () => {
    const conflictResult: BulkEditResult = {
      kind: 'conflict',
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
    }
    mockBulkEditDisciplineItems.mockResolvedValueOnce(conflictResult)
    mockFetchDisciplineItem.mockResolvedValueOnce({
      ...sampleItem,
      title: 'Renamed by another secretary',
      version: 7,
    })

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() => expect(mockFetchDisciplineItem).toHaveBeenCalledWith(202))
    await waitFor(() =>
      expect(screen.getByTestId('bulk-edit-conflict-banner-202')).toBeInTheDocument()
    )
    expect(screen.getByText('Renamed by another secretary')).toBeInTheDocument()
    // mutate NOT called on conflict — server snapshot gets refreshed via
    // the per-conflict fetchDisciplineItem refetch flow per ADR-12.
    expect(mockMutate).not.toHaveBeenCalled()
  })

  it('clicking "Apply server" reverts pending state for that item (REVERT_ITEM)', async () => {
    const conflictResult: BulkEditResult = {
      kind: 'conflict',
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
    }
    mockBulkEditDisciplineItems.mockResolvedValueOnce(conflictResult)
    mockFetchDisciplineItem.mockResolvedValueOnce({ ...sampleItem, version: 7 })

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() =>
      expect(screen.getByTestId('bulk-edit-conflict-banner-202')).toBeInTheDocument()
    )
    fireEvent.click(screen.getByTestId('bulk-edit-conflict-banner-202-apply-server'))

    // After REVERT_ITEM, pendingDeletes no longer carries 202 → Submit
    // disabled again (no pending changes).
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeDisabled()
    )
  })

  // === Race-fix backfill tests (v0.128.7 — closes round-1 reviewer N1) ===
  // v0.128.4 fix-cycle introduced SUBMIT_CONFLICT_REFRESHED action +
  // Cancel button submitting-guard to prevent submit re-click and
  // DISCARD_ALL during in-flight refetch loop. Reducer-level coverage
  // exists; these tests pin the panel-level integration so the wiring
  // doesn't silently break in future refactors.
  it('Submit stays DISABLED during 409 refetch loop (race fix)', async () => {
    const conflictResult: BulkEditResult = {
      kind: 'conflict',
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
    }
    mockBulkEditDisciplineItems.mockResolvedValueOnce(conflictResult)
    // Manual promise — we control когда refetch resolves.
    let resolveRefetch: (item: DisciplineItem) => void = () => {}
    mockFetchDisciplineItem.mockImplementationOnce(
      () =>
        new Promise<DisciplineItem>((resolve) => {
          resolveRefetch = resolve
        })
    )

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    // Conflict banner appears (SUBMIT_CONFLICT dispatched), но refetch
    // still in flight — Submit MUST stay disabled (would re-fire с
    // stale expected_version otherwise — guaranteed 409 cycle).
    await waitFor(() =>
      expect(screen.getByTestId('bulk-edit-conflict-banner-202')).toBeInTheDocument()
    )
    expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeDisabled()

    // Resolve refetch → SUBMIT_CONFLICT_REFRESHED dispatched → submitting=false.
    resolveRefetch({ ...sampleItem, version: 7 })

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeEnabled()
    )
  })

  it('Cancel button is DISABLED during 409 refetch loop (race fix)', async () => {
    const conflictResult: BulkEditResult = {
      kind: 'conflict',
      conflicts: [{ id: 202, expected_version: 5, current_version: 0 }],
    }
    mockBulkEditDisciplineItems.mockResolvedValueOnce(conflictResult)
    let resolveRefetch: (item: DisciplineItem) => void = () => {}
    mockFetchDisciplineItem.mockImplementationOnce(
      () =>
        new Promise<DisciplineItem>((resolve) => {
          resolveRefetch = resolve
        })
    )

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() =>
      expect(screen.getByTestId('bulk-edit-conflict-banner-202')).toBeInTheDocument()
    )
    // Cancel button must be disabled while refetch in flight — DISCARD_ALL
    // mid-refetch would dispatch ghost SET_REFRESHED writes (also guarded
    // в reducer, but UI gate keeps affordance honest).
    const cancelBtn = screen.getByRole('button', {
      name: 'disciplineItems.bulkEdit.cancel',
    })
    expect(cancelBtn).toBeDisabled()

    resolveRefetch({ ...sampleItem, version: 7 })
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.cancel' })).toBeEnabled()
    )
  })
})

describe('BulkEditPanel / Submit error mapping', () => {
  it('on 422 NOT_EDITABLE: shows toast keyed disciplineItems.bulkEdit.errorNotEditable', async () => {
    const axiosErr = Object.assign(new Error('rejected'), {
      isAxiosError: true,
      response: { status: 422, data: { error: { code: 'NOT_EDITABLE', message: '' } } },
      config: {},
    })
    mockBulkEditDisciplineItems.mockRejectedValueOnce(axiosErr)

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    const calls = mockToastError.mock.calls.flat()
    expect(calls.some((c: unknown) => String(c).includes('errorNotEditable'))).toBe(true)
    // Pending state preserved for retry.
    expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeEnabled()
  })

  it('on 500: shows generic toast key', async () => {
    const axiosErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status: 500, data: { error: { code: 'INTERNAL_ERROR', message: '' } } },
      config: {},
    })
    mockBulkEditDisciplineItems.mockRejectedValueOnce(axiosErr)

    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    const calls = mockToastError.mock.calls.flat()
    expect(calls.some((c: unknown) => String(c).includes('errorGeneric'))).toBe(true)
  })
})

describe('BulkEditPanel / Cancel discard-all', () => {
  it('Cancel button hidden when no pending changes', () => {
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    expect(
      screen.queryByRole('button', { name: 'disciplineItems.bulkEdit.cancel' })
    ).not.toBeInTheDocument()
  })

  it('Cancel button visible when there are pending changes; opens confirm dialog', () => {
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    const cancelBtn = screen.getByRole('button', {
      name: 'disciplineItems.bulkEdit.cancel',
    })
    fireEvent.click(cancelBtn)
    expect(screen.getByText('disciplineItems.bulkEdit.cancelDialog.title')).toBeInTheDocument()
  })

  it('confirming cancel dispatches DISCARD_ALL — Submit becomes disabled again', async () => {
    render(<BulkEditPanel sectionID={101} curriculumStatus="draft" />)
    fireEvent.click(screen.getByTestId('bulk-edit-row-202-delete-toggle'))
    fireEvent.click(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.cancel' }))
    fireEvent.click(
      screen.getByRole('button', {
        name: 'disciplineItems.bulkEdit.cancelDialog.confirm',
      })
    )
    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'disciplineItems.bulkEdit.submit' })).toBeDisabled()
    )
  })
})

describe('BulkEditPanel / canEdit gating (frozen state)', () => {
  it.each<['pending_approval' | 'approved' | 'archived']>([
    ['pending_approval'],
    ['approved'],
    ['archived'],
  ])('hides Submit/Cancel when curriculum status=%s', (status) => {
    render(<BulkEditPanel sectionID={101} curriculumStatus={status} />)
    expect(
      screen.queryByRole('button', { name: 'disciplineItems.bulkEdit.submit' })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'disciplineItems.bulkEdit.cancel' })
    ).not.toBeInTheDocument()
  })
})
