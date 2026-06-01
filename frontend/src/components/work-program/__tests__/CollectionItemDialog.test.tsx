import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { CollectionItemDialog, type CollectionField } from '../CollectionItemDialog'

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

beforeEach(() => {
  jest.clearAllMocks()
})

// A goal-like single-field schema + a competence-like multi-field schema
// exercise text/textarea/select rendering and required-field validation.
const GOAL_FIELDS: CollectionField[] = [
  { name: 'text', labelKey: 'f.text', type: 'textarea', required: true },
]
const COMPETENCE_FIELDS: CollectionField[] = [
  { name: 'code', labelKey: 'f.code', type: 'text', required: true },
  {
    name: 'type',
    labelKey: 'f.type',
    type: 'select',
    required: true,
    options: [
      { value: 'pk', labelKey: 'opt.pk' },
      { value: 'ok', labelKey: 'opt.ok' },
    ],
  },
  { name: 'description', labelKey: 'f.desc', type: 'textarea', required: true },
]

const wp = { id: 7 } as never

describe('CollectionItemDialog', () => {
  it('does not render when open=false', () => {
    render(
      <CollectionItemDialog
        open={false}
        mode="add"
        titleKey="collectionDialog.goals.addTitle"
        fields={GOAL_FIELDS}
        initialValues={{}}
        onSubmit={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    expect(screen.queryByText('collectionDialog.goals.addTitle')).not.toBeInTheDocument()
  })

  it('renders the title + every field when open', () => {
    render(
      <CollectionItemDialog
        open={true}
        mode="add"
        titleKey="collectionDialog.competences.addTitle"
        fields={COMPETENCE_FIELDS}
        initialValues={{}}
        onSubmit={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    expect(screen.getByText('collectionDialog.competences.addTitle')).toBeInTheDocument()
    // 2 textareas (code is an <input>, desc textarea) + 1 select + 1 input
    expect(screen.getByRole('combobox')).toBeInTheDocument()
    // select offers the two options + placeholder
    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.options).toHaveLength(3)
    expect(Array.from(select.options).map((o) => o.value)).toEqual(['', 'pk', 'ok'])
  })

  it('disables Save until all required fields are non-empty', () => {
    render(
      <CollectionItemDialog
        open={true}
        mode="add"
        titleKey="collectionDialog.goals.addTitle"
        fields={GOAL_FIELDS}
        initialValues={{}}
        onSubmit={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    const saveBtn = screen.getByRole('button', { name: 'collectionDialog.save' })
    expect(saveBtn).toBeDisabled()
    fireEvent.change(screen.getByRole('textbox'), { target: { value: '  ' } })
    expect(saveBtn).toBeDisabled() // whitespace-only fails the trim check
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Новая цель' } })
    expect(saveBtn).toBeEnabled()
  })

  it('calls onSubmit with the form values, then onDone + success toast', async () => {
    const onSubmit = jest.fn().mockResolvedValue(wp)
    const onDone = jest.fn()
    const onClose = jest.fn()
    render(
      <CollectionItemDialog
        open={true}
        mode="add"
        titleKey="collectionDialog.goals.addTitle"
        fields={GOAL_FIELDS}
        initialValues={{}}
        onSubmit={onSubmit}
        onDone={onDone}
        onClose={onClose}
      />
    )
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Цель' } })
    fireEvent.click(screen.getByRole('button', { name: 'collectionDialog.save' }))
    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith({ text: 'Цель' }))
    await waitFor(() => expect(onDone).toHaveBeenCalledWith(wp))
    expect(mockToastSuccess).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })

  it('prefills initialValues in edit mode', () => {
    render(
      <CollectionItemDialog
        open={true}
        mode="edit"
        titleKey="collectionDialog.goals.editTitle"
        fields={GOAL_FIELDS}
        initialValues={{ text: 'Существующая цель' }}
        onSubmit={jest.fn()}
        onDone={jest.fn()}
        onClose={jest.fn()}
      />
    )
    expect(screen.getByRole('textbox')).toHaveValue('Существующая цель')
  })

  it('keeps the dialog open and shows an error toast when onSubmit rejects', async () => {
    const onSubmit = jest.fn().mockRejectedValue(new Error('boom'))
    const onDone = jest.fn()
    const onClose = jest.fn()
    render(
      <CollectionItemDialog
        open={true}
        mode="add"
        titleKey="collectionDialog.goals.addTitle"
        fields={GOAL_FIELDS}
        initialValues={{}}
        onSubmit={onSubmit}
        onDone={onDone}
        onClose={onClose}
      />
    )
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Цель' } })
    fireEvent.click(screen.getByRole('button', { name: 'collectionDialog.save' }))
    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(onDone).not.toHaveBeenCalled()
    expect(onClose).not.toHaveBeenCalled()
  })
})
