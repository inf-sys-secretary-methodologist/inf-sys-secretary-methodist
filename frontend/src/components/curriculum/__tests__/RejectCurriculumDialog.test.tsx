import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { RejectCurriculumDialog } from '../RejectCurriculumDialog'

const mockRejectCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  ...jest.requireActual('@/hooks/useCurricula'),
  rejectCurriculum: (...args: unknown[]) => mockRejectCurriculum(...args),
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

describe('RejectCurriculumDialog', () => {
  it('does not render reason input when open=false', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={false} onClose={() => {}} />)
    expect(screen.queryByLabelText('rejectDialog.reasonLabel')).not.toBeInTheDocument()
  })

  it('renders title + description + reason input + cancel + confirm when open', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    expect(screen.getByText('rejectDialog.title')).toBeInTheDocument()
    expect(screen.getByText('rejectDialog.description')).toBeInTheDocument()
    expect(screen.getByLabelText('rejectDialog.reasonLabel')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'rejectDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeInTheDocument()
  })

  it('disables Confirm button initially (empty reason)', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeDisabled()
  })

  it('disables Confirm when reason is whitespace-only', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: '    ' },
    })
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeDisabled()
  })

  it('disables Confirm when reason exceeds 4096 chars', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'x'.repeat(4097) },
    })
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).toBeDisabled()
  })

  it('enables Confirm when reason is valid', () => {
    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'Не соответствует ФГОС' },
    })
    expect(screen.getByRole('button', { name: 'rejectDialog.confirm' })).not.toBeDisabled()
  })

  it('calls rejectCurriculum with trimmed reason + onRejected + onClose + toast.success', async () => {
    const onClose = jest.fn()
    const onRejected = jest.fn()
    mockRejectCurriculum.mockResolvedValueOnce({})

    render(
      <RejectCurriculumDialog
        curriculumId={11}
        open={true}
        onClose={onClose}
        onRejected={onRejected}
      />
    )
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: '  Не соответствует ФГОС  ' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'rejectDialog.confirm' }))

    await waitFor(() =>
      expect(mockRejectCurriculum).toHaveBeenCalledWith(11, {
        reason: 'Не соответствует ФГОС',
      })
    )
    await waitFor(() => expect(onRejected).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [422, 'rejectToast.errors.notPending'],
    [400, 'rejectToast.errors.invalidReason'],
    [403, 'rejectToast.errors.forbidden'],
    [500, 'rejectToast.errors.generic'],
  ])('maps HTTP %i error to toast key and stays open', async (status, expectedKey) => {
    const onClose = jest.fn()
    const onRejected = jest.fn()
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status },
    })
    mockRejectCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(
      <RejectCurriculumDialog
        curriculumId={11}
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

  it('does not double-fire rejectCurriculum on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockRejectCurriculum.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<RejectCurriculumDialog curriculumId={11} open={true} onClose={onClose} />)
    fireEvent.change(screen.getByLabelText('rejectDialog.reasonLabel'), {
      target: { value: 'reason' },
    })
    const btn = screen.getByRole('button', { name: 'rejectDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockRejectCurriculum).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
