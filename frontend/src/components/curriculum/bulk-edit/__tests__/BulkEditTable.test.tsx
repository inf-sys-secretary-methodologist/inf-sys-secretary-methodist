import { render, screen, fireEvent } from '@/test-utils'
import { useReducer } from 'react'
import { BulkEditTable } from '../BulkEditTable'
import { bulkEditReducer, initialBulkEditState, type BulkEditState } from '../bulkEditReducer'
import type { DisciplineItem } from '@/types/disciplineItem'

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

const secondItem: DisciplineItem = {
  ...sampleItem,
  id: 303,
  title: 'Линейная алгебра',
  control_form: 'zachet',
  credits: 3,
}

// Test helper — renders BulkEditTable wrapped in a host that owns
// useReducer state, mirroring how the parent (Pair 7 BulkEditPanel)
// will mount it.
interface HostProps {
  items: DisciplineItem[]
  canEdit?: boolean
  initialState?: BulkEditState
}
function Host({ items, canEdit = true, initialState }: HostProps) {
  const [state, dispatch] = useReducer(bulkEditReducer, initialState ?? initialBulkEditState)
  return (
    <BulkEditTable
      sectionID={101}
      items={items}
      state={state}
      dispatch={dispatch}
      canEdit={canEdit}
    />
  )
}

describe('BulkEditTable / header', () => {
  it('renders column headers via i18n keys', () => {
    render(<Host items={[]} />)
    expect(screen.getByText('disciplineItems.bulkEdit.columns.title')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.bulkEdit.columns.hoursLectures')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.bulkEdit.columns.controlForm')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.bulkEdit.columns.credits')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.bulkEdit.columns.semester')).toBeInTheDocument()
  })
})

describe('BulkEditTable / empty state', () => {
  it('renders empty placeholder when no items and no pending creates', () => {
    render(<Host items={[]} />)
    expect(screen.getByText('disciplineItems.bulkEdit.empty')).toBeInTheDocument()
  })
})

describe('BulkEditTable / item rows', () => {
  it('renders one row per server item', () => {
    render(<Host items={[sampleItem, secondItem]} />)
    expect(screen.getByDisplayValue('Математический анализ')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Линейная алгебра')).toBeInTheDocument()
  })

  it('shows pending update value over server snapshot for the title field', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingUpdates: [
        {
          id: 202,
          title: 'Renamed by user',
          hours_lectures: 36,
          hours_practice: 36,
          hours_lab: 0,
          hours_self: 72,
          control_form: 'exam',
          credits: 4,
          semester: 1,
          order_index: 0,
        },
      ],
    }
    render(<Host items={[sampleItem]} initialState={seeded} />)
    expect(screen.getByDisplayValue('Renamed by user')).toBeInTheDocument()
    expect(screen.queryByDisplayValue('Математический анализ')).not.toBeInTheDocument()
  })

  it('marks rows in pendingDeletes visually (data-pending-delete attribute)', () => {
    const seeded: BulkEditState = { ...initialBulkEditState, pendingDeletes: [202] }
    render(<Host items={[sampleItem]} initialState={seeded} />)
    const row = screen.getByTestId('bulk-edit-row-202')
    expect(row).toHaveAttribute('data-pending-delete', 'true')
  })
})

describe('BulkEditTable / pending creates', () => {
  it('renders a row for each pendingCreate appended after server items', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        {
          localKey: 'tmp-1',
          title: 'Новая дисциплина',
          hours_lectures: 18,
          hours_practice: 18,
          hours_lab: 0,
          hours_self: 36,
          control_form: 'zachet',
          credits: 2,
          semester: 1,
          order_index: 5,
        },
      ],
    }
    render(<Host items={[sampleItem]} initialState={seeded} />)
    expect(screen.getByDisplayValue('Новая дисциплина')).toBeInTheDocument()
    expect(screen.getByTestId('bulk-edit-row-create-tmp-1')).toBeInTheDocument()
  })
})

describe('BulkEditTable / cell editing', () => {
  it('typing into an existing-item title input dispatches EDIT_ITEM (upserts pendingUpdate)', () => {
    render(<Host items={[sampleItem]} />)
    const input = screen.getByDisplayValue('Математический анализ')
    fireEvent.change(input, { target: { value: 'Edited' } })
    // Re-query post-render — pendingUpdate now drives the displayed value.
    expect(screen.getByDisplayValue('Edited')).toBeInTheDocument()
  })

  it('typing into a pendingCreate title input dispatches EDIT_CREATE', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        {
          localKey: 'tmp-1',
          title: '',
          hours_lectures: 0,
          hours_practice: 0,
          hours_lab: 0,
          hours_self: 0,
          control_form: 'zachet',
          credits: 1,
          semester: 1,
          order_index: 0,
        },
      ],
    }
    render(<Host items={[]} initialState={seeded} />)
    const input = screen.getByTestId('bulk-edit-row-create-tmp-1-title-input')
    fireEvent.change(input, { target: { value: 'Создаваемая' } })
    expect(screen.getByDisplayValue('Создаваемая')).toBeInTheDocument()
  })

  it('toggling delete checkbox on existing row marks pendingDelete', () => {
    render(<Host items={[sampleItem]} />)
    const checkbox = screen.getByTestId('bulk-edit-row-202-delete-toggle')
    fireEvent.click(checkbox)
    const row = screen.getByTestId('bulk-edit-row-202')
    expect(row).toHaveAttribute('data-pending-delete', 'true')
  })

  it('removing a pendingCreate row dispatches REMOVE_CREATE (row disappears)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        {
          localKey: 'tmp-1',
          title: 'Удалить меня',
          hours_lectures: 0,
          hours_practice: 0,
          hours_lab: 0,
          hours_self: 0,
          control_form: 'zachet',
          credits: 1,
          semester: 1,
          order_index: 0,
        },
      ],
    }
    render(<Host items={[]} initialState={seeded} />)
    const removeBtn = screen.getByTestId('bulk-edit-row-create-tmp-1-remove')
    fireEvent.click(removeBtn)
    expect(screen.queryByTestId('bulk-edit-row-create-tmp-1')).not.toBeInTheDocument()
  })
})

describe('BulkEditTable / Add row', () => {
  it('clicking Add row dispatches ADD_CREATE with a fresh localKey', () => {
    render(<Host items={[]} />)
    const addBtn = screen.getByText('disciplineItems.bulkEdit.addRow')
    fireEvent.click(addBtn)
    // After click, exactly one pending create row appears. Filter by
    // <tr> tag because the row's testid prefix is shared by its children
    // (remove button + title input both have testids starting with
    // bulk-edit-row-create-{key}-…).
    const pendingRows = screen
      .getAllByRole('row')
      .filter((r) => r.getAttribute('data-testid')?.startsWith('bulk-edit-row-create-'))
    expect(pendingRows).toHaveLength(1)
  })
})

describe('BulkEditTable / control form select', () => {
  it('renders 4 options matching the typed enum (i18n keys verbatim)', () => {
    render(<Host items={[sampleItem]} />)
    const select = screen.getByTestId('bulk-edit-row-202-control-form-select')
    const options = select.querySelectorAll('option')
    const values = Array.from(options).map((o) => (o as HTMLOptionElement).value)
    expect(values).toEqual(['zachet', 'exam', 'course_project', 'differential_zachet'])
  })

  it('renders i18n key labels per ADR-15 namespace', () => {
    render(<Host items={[sampleItem]} />)
    expect(screen.getByText('disciplineItems.controlForm.zachet')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.controlForm.exam')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.controlForm.course_project')).toBeInTheDocument()
    expect(screen.getByText('disciplineItems.controlForm.differential_zachet')).toBeInTheDocument()
  })
})

describe('BulkEditTable / canEdit gating (frozen state)', () => {
  it('hides Add row button when canEdit=false', () => {
    render(<Host items={[sampleItem]} canEdit={false} />)
    expect(screen.queryByText('disciplineItems.bulkEdit.addRow')).not.toBeInTheDocument()
  })

  it('hides delete toggle controls when canEdit=false', () => {
    render(<Host items={[sampleItem]} canEdit={false} />)
    expect(screen.queryByTestId('bulk-edit-row-202-delete-toggle')).not.toBeInTheDocument()
  })

  it('renders inputs as readOnly when canEdit=false', () => {
    render(<Host items={[sampleItem]} canEdit={false} />)
    const titleInput = screen.getByDisplayValue('Математический анализ')
    expect(titleInput).toHaveAttribute('readonly')
  })
})

describe('BulkEditTable / hardening — empty state hides <table>', () => {
  it('does not render <table> when items=[] and no pending creates', () => {
    const { container } = render(<Host items={[]} />)
    expect(container.querySelector('table')).toBeNull()
    // empty placeholder still visible
    expect(screen.getByText('disciplineItems.bulkEdit.empty')).toBeInTheDocument()
  })

  it('renders <table> when items have entries', () => {
    const { container } = render(<Host items={[sampleItem]} />)
    expect(container.querySelector('table')).not.toBeNull()
  })

  it('renders <table> when only pending creates exist (zero server items but creating new)', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        {
          localKey: 'tmp-empty-but-creating',
          title: '',
          hours_lectures: 0,
          hours_practice: 0,
          hours_lab: 0,
          hours_self: 0,
          control_form: 'zachet',
          credits: 1,
          semester: 1,
          order_index: 0,
        },
      ],
    }
    const { container } = render(<Host items={[]} initialState={seeded} />)
    expect(container.querySelector('table')).not.toBeNull()
  })
})

describe('BulkEditTable / hardening — sectionID data-testid', () => {
  it('wraps table block в div с data-testid="bulk-edit-table-{sectionID}" — query stability under multi-section page', () => {
    render(<Host items={[sampleItem]} />)
    expect(screen.getByTestId('bulk-edit-table-101')).toBeInTheDocument()
  })
})

describe('BulkEditTable / hardening — number input min={0}', () => {
  it('every number input has min="0" — guard against negative entries even before backend rejects 422', () => {
    const seeded: BulkEditState = {
      ...initialBulkEditState,
      pendingCreates: [
        {
          localKey: 'tmp-min',
          title: '',
          hours_lectures: 0,
          hours_practice: 0,
          hours_lab: 0,
          hours_self: 0,
          control_form: 'zachet',
          credits: 1,
          semester: 1,
          order_index: 0,
        },
      ],
    }
    const { container } = render(<Host items={[sampleItem]} initialState={seeded} />)
    const numericInputs = container.querySelectorAll('input[type="number"]')
    // sanity — 7 numeric cells per row (4 hours + credits + semester + order_index;
    // title is text, control_form is select); 1 server item + 1 pending create = 14.
    expect(numericInputs.length).toBe(14)
    numericInputs.forEach((input) => {
      expect(input).toHaveAttribute('min', '0')
    })
  })
})
