import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { CreateRevisionDialog } from '../CreateRevisionDialog'

const mockCreateRevision = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  createRevision: (...args: unknown[]) => mockCreateRevision(...args),
}))

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

const typeLabel = 'createRevisionDialog.changeTypeLabel'
const summaryLabel = 'createRevisionDialog.summaryLabel'
const createBtn = { name: 'createRevisionDialog.create' }

describe('CreateRevisionDialog', () => {
  it('does not render the form when open=false', () => {
    render(<CreateRevisionDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByLabelText(summaryLabel)).not.toBeInTheDocument()
  })

  it('renders title + description + type select + summary + cancel + create when open', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('createRevisionDialog.title')).toBeInTheDocument()
    expect(screen.getByText('createRevisionDialog.description')).toBeInTheDocument()
    expect(screen.getByLabelText(typeLabel)).toBeInTheDocument()
    expect(screen.getByLabelText(summaryLabel)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'createRevisionDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', createBtn)).toBeInTheDocument()
  })

  it('offers all five change-type options plus a placeholder', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    const select = screen.getByLabelText(typeLabel) as HTMLSelectElement
    // 5 enum options + 1 placeholder
    expect(select.options).toHaveLength(6)
    const values = Array.from(select.options).map((o) => o.value)
    expect(values).toEqual(['', 'hours', 'semester', 'literature', 'assessment', 'other'])
  })

  it('disables Create initially (no type + empty summary)', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByRole('button', createBtn)).toBeDisabled()
  })

  it('disables Create when a type is chosen but summary is whitespace-only', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'hours' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: '   ' } })
    expect(screen.getByRole('button', createBtn)).toBeDisabled()
  })

  it('disables Create when summary is filled but no type is chosen', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: 'Часы 36 → 18' } })
    expect(screen.getByRole('button', createBtn)).toBeDisabled()
  })

  it('enables Create when type + summary are valid', () => {
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'hours' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: 'Часы 36 → 18' } })
    expect(screen.getByRole('button', createBtn)).not.toBeDisabled()
  })

  it('calls createRevision with trimmed summary + onCreated + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onCreated = jest.fn()
    mockCreateRevision.mockResolvedValueOnce({ id: 7 })

    render(
      <CreateRevisionDialog workProgramId={7} open={true} onClose={onClose} onCreated={onCreated} />
    )
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'literature' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), {
      target: { value: '  Обновлён список основной литературы  ' },
    })
    fireEvent.click(screen.getByRole('button', createBtn))

    await waitFor(() =>
      expect(mockCreateRevision).toHaveBeenCalledWith(7, {
        change_type: 'literature',
        change_summary: 'Обновлён список основной литературы',
      })
    )
    await waitFor(() => expect(onCreated).toHaveBeenCalledWith({ id: 7 }))
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  // The full code→key table is covered in useWorkPrograms.test.ts; here we
  // only prove the dialog routes through pickWorkProgramErrorKey and stays
  // open on failure (a representative sentinel + status fallback).
  it.each([
    [
      { response: { data: { error: { code: 'REVISION_NOT_PERMITTED' } } } },
      'errors.revisionNotPermitted',
    ],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onCreated = jest.fn()
    mockCreateRevision.mockRejectedValueOnce(err)

    render(
      <CreateRevisionDialog workProgramId={7} open={true} onClose={onClose} onCreated={onCreated} />
    )
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'hours' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: 'summary' } })
    fireEvent.click(screen.getByRole('button', createBtn))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onCreated).not.toHaveBeenCalled()
  })

  it('resets the form on each fresh open (blank slate per programme)', () => {
    const { rerender } = render(
      <CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />
    )
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'hours' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: 'stale draft' } })
    rerender(<CreateRevisionDialog workProgramId={7} open={false} onClose={() => {}} />)
    rerender(<CreateRevisionDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByLabelText(summaryLabel)).toHaveValue('')
    expect(screen.getByLabelText(typeLabel)).toHaveValue('')
  })

  it('does not double-fire createRevision on rapid double-click', async () => {
    let resolve: (v: unknown) => void = () => {}
    mockCreateRevision.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )
    const onClose = jest.fn()
    render(<CreateRevisionDialog workProgramId={7} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText(typeLabel), { target: { value: 'hours' } })
    fireEvent.change(screen.getByLabelText(summaryLabel), { target: { value: 'summary' } })
    const btn = screen.getByRole('button', createBtn)
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockCreateRevision).toHaveBeenCalledTimes(1)
    resolve({ id: 7 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
