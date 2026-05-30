import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { SubmitRevisionDialog } from '../SubmitRevisionDialog'

const mockSubmitRevision = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  submitRevision: (...args: unknown[]) => mockSubmitRevision(...args),
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

const confirmBtn = { name: 'submitRevisionDialog.confirm' }

describe('SubmitRevisionDialog', () => {
  it('does not render when open=false', () => {
    render(
      <SubmitRevisionDialog workProgramId={7} revisionId={3} open={false} onClose={() => {}} />
    )
    expect(screen.queryByText('submitRevisionDialog.title')).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<SubmitRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={() => {}} />)
    expect(screen.getByText('submitRevisionDialog.title')).toBeInTheDocument()
    expect(screen.getByText('submitRevisionDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'submitRevisionDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', confirmBtn)).toBeInTheDocument()
  })

  it('calls submitRevision(workProgramId, revisionId) + onSubmitted + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    mockSubmitRevision.mockResolvedValueOnce({ id: 7 })

    render(
      <SubmitRevisionDialog
        workProgramId={7}
        revisionId={3}
        open={true}
        onClose={onClose}
        onSubmitted={onSubmitted}
      />
    )
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() => expect(mockSubmitRevision).toHaveBeenCalledWith(7, 3))
    await waitFor(() => expect(onSubmitted).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  // Full code→key table lives in useWorkPrograms.test.ts; here a
  // representative sentinel + status fallback proves routing + stay-open.
  it.each([
    [
      { response: { data: { error: { code: 'REVISION_NOT_PERMITTED' } } } },
      'errors.revisionNotPermitted',
    ],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    mockSubmitRevision.mockRejectedValueOnce(err)

    render(
      <SubmitRevisionDialog
        workProgramId={7}
        revisionId={3}
        open={true}
        onClose={onClose}
        onSubmitted={onSubmitted}
      />
    )
    fireEvent.click(screen.getByRole('button', confirmBtn))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onSubmitted).not.toHaveBeenCalled()
  })

  it('does not double-fire submitRevision on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockSubmitRevision.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<SubmitRevisionDialog workProgramId={7} revisionId={3} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', confirmBtn)
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockSubmitRevision).toHaveBeenCalledTimes(1)
    resolve({ id: 7 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
