import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { RejectRevisionDialog } from '../RejectRevisionDialog'

const mockRejectRevision = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  rejectRevision: (...args: unknown[]) => mockRejectRevision(...args),
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

const reasonLabel = 'rejectRevisionDialog.reasonLabel'
const confirmBtn = { name: 'rejectRevisionDialog.confirm' }

describe('RejectRevisionDialog', () => {
  it('does not render the reason input when open=false', () => {
    render(
      <RejectRevisionDialog workProgramId={7} revisionId={3} open={false} onClose={() => {}} />
    )
    expect(screen.queryByLabelText(reasonLabel)).not.toBeInTheDocument()
  })

  it('renders title + description + reason input + cancel + confirm when open', () => {
    render(<RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />)
    expect(screen.getByText('rejectRevisionDialog.title')).toBeInTheDocument()
    expect(screen.getByText('rejectRevisionDialog.description')).toBeInTheDocument()
    expect(screen.getByLabelText(reasonLabel)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'rejectRevisionDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', confirmBtn)).toBeInTheDocument()
  })

  it('disables Confirm initially (empty reason) and when whitespace-only', () => {
    render(<RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />)
    expect(screen.getByRole('button', confirmBtn)).toBeDisabled()
    fireEvent.change(screen.getByLabelText(reasonLabel), { target: { value: '   ' } })
    expect(screen.getByRole('button', confirmBtn)).toBeDisabled()
  })

  it('enables Confirm when the reason is valid', () => {
    render(<RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText(reasonLabel), { target: { value: 'Не обосновано' } })
    expect(screen.getByRole('button', confirmBtn)).not.toBeDisabled()
  })

  it('calls rejectRevision with trimmed reason + onRejected + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onRejected = jest.fn()
    mockRejectRevision.mockResolvedValueOnce({ id: 7 })

    render(
      <RejectRevisionDialog
        workProgramId={7}
        revisionId={3}
        open={true}
        onClose={onClose}
        onRejected={onRejected}
      />
    )
    fireEvent.change(screen.getByLabelText(reasonLabel), {
      target: { value: '  Изменение не обосновано приказом  ' },
    })
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() =>
      expect(mockRejectRevision).toHaveBeenCalledWith(7, 3, {
        reason: 'Изменение не обосновано приказом',
      })
    )
    await waitFor(() => expect(onRejected).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  // Full code→key table lives in useWorkPrograms.test.ts; representative
  // sentinel + status fallback proves routing + stay-open.
  it.each([
    [
      { response: { data: { error: { code: 'REVISION_NOT_PERMITTED' } } } },
      'errors.revisionNotPermitted',
    ],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    mockRejectRevision.mockRejectedValueOnce(err)

    render(<RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText(reasonLabel), { target: { value: 'reason' } })
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
  })

  it('resets the reason on each fresh open (blank slate per revision)', () => {
    const { rerender } = render(
      <RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />
    )
    fireEvent.change(screen.getByLabelText(reasonLabel), { target: { value: 'stale draft' } })
    rerender(
      <RejectRevisionDialog workProgramId={7} revisionId={3} open={false} onClose={() => {}} />
    )
    rerender(
      <RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />
    )
    expect(screen.getByLabelText(reasonLabel)).toHaveValue('')
  })

  it('does not double-fire rejectRevision on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockRejectRevision.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<RejectRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText(reasonLabel), { target: { value: 'reason' } })
    const btn = screen.getByRole('button', confirmBtn)
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockRejectRevision).toHaveBeenCalledTimes(1)
    resolve({ id: 7 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
