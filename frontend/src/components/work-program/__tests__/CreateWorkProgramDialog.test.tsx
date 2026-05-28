import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { useState } from 'react'
import { CreateWorkProgramDialog } from '../CreateWorkProgramDialog'

const mockCreateWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  createWorkProgram: (...args: unknown[]) => mockCreateWorkProgram(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

const noop = () => {}

// Fills every required field with a valid value so a single invalid field
// under test is the sole reason Create stays disabled.
function fillValidRequired() {
  fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
    target: { value: 'Базы данных' },
  })
  fireEvent.change(screen.getByLabelText('createDialog.labels.disciplineId'), {
    target: { value: '42' },
  })
  fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
    target: { value: '09.03.01' },
  })
  fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
    target: { value: '2027' },
  })
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('CreateWorkProgramDialog', () => {
  it('does not render fields when open=false', () => {
    render(<CreateWorkProgramDialog open={false} onClose={noop} />)
    expect(screen.queryByLabelText('createDialog.labels.title')).not.toBeInTheDocument()
  })

  it('starts with empty inputs (no pre-fill)', () => {
    render(<CreateWorkProgramDialog open={true} onClose={noop} />)
    expect(screen.getByLabelText('createDialog.labels.title')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.disciplineId')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.specialty')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.year')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.annotation')).toHaveValue('')
  })

  it('keeps Create disabled until all required fields are filled (annotation optional)', () => {
    render(<CreateWorkProgramDialog open={true} onClose={noop} />)
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
    fillValidRequired()
    expect(screen.getByRole('button', { name: 'createDialog.create' })).not.toBeDisabled()
  })

  it.each([1999, 2101, 0])('disables Create when year is out of range (%i)', (year) => {
    render(<CreateWorkProgramDialog open={true} onClose={noop} />)
    fillValidRequired()
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: String(year) },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
  })

  it('disables Create when discipline id is not a positive integer', () => {
    render(<CreateWorkProgramDialog open={true} onClose={noop} />)
    fillValidRequired()
    fireEvent.change(screen.getByLabelText('createDialog.labels.disciplineId'), {
      target: { value: '0' },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
  })

  it('disables Create when annotation exceeds 8192 chars', () => {
    render(<CreateWorkProgramDialog open={true} onClose={noop} />)
    fillValidRequired()
    fireEvent.change(screen.getByLabelText('createDialog.labels.annotation'), {
      target: { value: 'x'.repeat(8193) },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
  })

  it('calls createWorkProgram with the wire payload and fires onCreated + close on success', async () => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    mockCreateWorkProgram.mockResolvedValueOnce({ id: 12, status: 'draft' })

    render(<CreateWorkProgramDialog open={true} onClose={onClose} onCreated={onCreated} />)

    fillValidRequired()
    fireEvent.change(screen.getByLabelText('createDialog.labels.annotation'), {
      target: { value: 'Курс по СУБД' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'createDialog.create' }))

    await waitFor(() => {
      expect(mockCreateWorkProgram).toHaveBeenCalledWith({
        title: 'Базы данных',
        discipline_id: 42,
        specialty_code: '09.03.01',
        applicable_from_year: 2027,
        annotation: 'Курс по СУБД',
      })
    })
    await waitFor(() => expect(onCreated).toHaveBeenCalledWith({ id: 12, status: 'draft' }))
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [{ response: { data: { error: { code: 'IDENTITY_EXISTS' } } } }, 'errors.identityExists'],
    [
      { response: { data: { error: { code: 'INVALID_WORK_PROGRAM' } } } },
      'errors.invalidWorkProgram',
    ],
    [{ response: { status: 403 } }, 'errors.forbidden'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error to toast key and keeps dialog open', async (err, expectedKey) => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    mockCreateWorkProgram.mockRejectedValueOnce(err)

    render(<CreateWorkProgramDialog open={true} onClose={onClose} onCreated={onCreated} />)
    fillValidRequired()
    fireEvent.click(screen.getByRole('button', { name: 'createDialog.create' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onCreated).not.toHaveBeenCalled()
  })

  it('clears form state when dialog is closed and reopened', () => {
    function Harness() {
      const [open, setOpen] = useState(true)
      return (
        <>
          <CreateWorkProgramDialog open={open} onClose={() => setOpen(false)} />
          <button type="button" onClick={() => setOpen(false)}>
            close-host
          </button>
          <button type="button" onClick={() => setOpen(true)}>
            reopen-host
          </button>
        </>
      )
    }

    render(<Harness />)
    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'Draft scratch' },
    })
    expect(screen.getByDisplayValue('Draft scratch')).toBeInTheDocument()

    fireEvent.click(screen.getByText('close-host'))
    fireEvent.click(screen.getByText('reopen-host'))

    expect(screen.getByLabelText('createDialog.labels.title')).toHaveValue('')
    expect(screen.queryByDisplayValue('Draft scratch')).not.toBeInTheDocument()
  })

  it('does not double-fire createWorkProgram on rapid double-click', async () => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockCreateWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<CreateWorkProgramDialog open={true} onClose={onClose} onCreated={onCreated} />)
    fillValidRequired()

    const btn = screen.getByRole('button', { name: 'createDialog.create' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockCreateWorkProgram).toHaveBeenCalledTimes(1)
    resolve({ id: 12 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
