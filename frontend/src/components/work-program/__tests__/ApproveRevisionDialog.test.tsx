import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { ApproveRevisionDialog } from '../ApproveRevisionDialog'

const mockApproveRevision = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  approveRevision: (...args: unknown[]) => mockApproveRevision(...args),
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

const confirmBtn = { name: 'approveRevisionDialog.confirm' }

describe('ApproveRevisionDialog', () => {
  it('does not render when open=false', () => {
    render(
      <ApproveRevisionDialog workProgramId={7} revisionId={3} open={false} onClose={() => {}} />
    )
    expect(screen.queryByText('approveRevisionDialog.title')).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(
      <ApproveRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />
    )
    expect(screen.getByText('approveRevisionDialog.title')).toBeInTheDocument()
    expect(screen.getByText('approveRevisionDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'approveRevisionDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', confirmBtn)).toBeInTheDocument()
  })

  it('calls approveRevision(workProgramId, revisionId) + onApproved + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onApproved = jest.fn()
    mockApproveRevision.mockResolvedValueOnce({ id: 7 })

    render(
      <ApproveRevisionDialog
        workProgramId={7}
        revisionId={3}
        open={true}
        onClose={onClose}
        onApproved={onApproved}
      />
    )
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() => expect(mockApproveRevision).toHaveBeenCalledWith(7, 3))
    await waitFor(() => expect(onApproved).toHaveBeenCalled())
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
    const onApproved = jest.fn()
    mockApproveRevision.mockRejectedValueOnce(err)

    render(
      <ApproveRevisionDialog
        workProgramId={7}
        revisionId={3}
        open={true}
        onClose={onClose}
        onApproved={onApproved}
      />
    )
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onApproved).not.toHaveBeenCalled()
  })

  it('does not double-fire approveRevision on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockApproveRevision.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<ApproveRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', confirmBtn)
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockApproveRevision).toHaveBeenCalledTimes(1)
    resolve({ id: 7 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
