import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { RejectWorkProgramDialog } from '../RejectWorkProgramDialog'

const mockRejectWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  rejectWorkProgram: (...args: unknown[]) => mockRejectWorkProgram(...args),
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

describe('RejectWorkProgramDialog', () => {
  it('does not render reason input when open=false', () => {
    render(<RejectWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByLabelText('rejectDialog.reasonLabel')).not.toBeInTheDocument()
  })

  it('renders title + description + reason input + cancel + confirm when open', () => {
    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('rejectDialog.title')).toBeInTheDocument()
    expect(screen.getByText('rejectDialog.description')).toBeInTheDocument()
    expect(screen.getByLabelText('rejectDialog.reasonLabel')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'rejectDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeInTheDocument()
  })

  it('disables Confirm button initially (empty reason)', () => {
    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeDisabled()
  })

  it('disables Confirm when reason is whitespace-only', () => {
    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: '    ' },
    })
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeDisabled()
  })

  it('enables Confirm when reason is valid', () => {
    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'Часы не соответствуют учебному плану' },
    })
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).not.toBeDisabled()
  })

  it('calls rejectWorkProgram with trimmed reason + onRejected + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onRejected = jest.fn()
    mockRejectWorkProgram.mockResolvedValueOnce({})

    render(
      <RejectWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onRejected={onRejected}
      />
    )
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: '  Часы не соответствуют учебному плану  ' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'rejectDialog.confirm' }))

    await waitFor(() =>
      expect(mockRejectWorkProgram).toHaveBeenCalledWith(7, {
        reason: 'Часы не соответствуют учебному плану',
      })
    )
    await waitFor(() => expect(onRejected).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [
      { response: { data: { error: { code: 'REJECT_REASON_REQUIRED' } } } },
      'errors.rejectReasonRequired',
    ],
    [{ response: { data: { error: { code: 'INVALID_TRANSITION' } } } }, 'errors.invalidTransition'],
    [{ response: { data: { error: { code: 'VERSION_CONFLICT' } } } }, 'errors.versionConflict'],
    [{ response: { status: 403 } }, 'errors.forbidden'],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onRejected = jest.fn()
    mockRejectWorkProgram.mockRejectedValueOnce(err)

    render(
      <RejectWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onRejected={onRejected}
      />
    )
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'reason' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'rejectDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onRejected).not.toHaveBeenCalled()
  })

  it('resets the reason on each fresh open (blank slate per programme)', () => {
    const { rerender } = render(
      <RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />
    )
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'stale draft' },
    })
    rerender(<RejectWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    rerender(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByLabelText('rejectDialog.reasonLabel')).toHaveValue('')
  })

  it('does not close while a reject is in flight (Esc dismiss guarded)', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockRejectWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'reason' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'rejectDialog.confirm' }))
    fireEvent.keyDown(document.body, { key: 'Escape' })
    expect(onClose).not.toHaveBeenCalled()

    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })

  it('does not double-fire rejectWorkProgram on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockRejectWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<RejectWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'reason' },
    })
    const btn = screen.getByRole('button', { name: 'rejectDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockRejectWorkProgram).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
