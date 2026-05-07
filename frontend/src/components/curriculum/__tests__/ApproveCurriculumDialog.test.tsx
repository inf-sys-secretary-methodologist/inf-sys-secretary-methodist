import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { ApproveCurriculumDialog } from '../ApproveCurriculumDialog'

const mockApproveCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  ...jest.requireActual('@/hooks/useCurricula'),
  approveCurriculum: (...args: unknown[]) => mockApproveCurriculum(...args),
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

describe('ApproveCurriculumDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<ApproveCurriculumDialog curriculumId={11} open={false} onClose={() => {}} />)
    expect(
      screen.queryByRole('button', { name: 'approveDialog.confirm' })
    ).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<ApproveCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    expect(screen.getByText('approveDialog.title')).toBeInTheDocument()
    expect(screen.getByText('approveDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'approveDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'approveDialog.confirm' })).toBeInTheDocument()
  })

  it('calls approveCurriculum + onApproved + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onApproved = jest.fn()
    mockApproveCurriculum.mockResolvedValueOnce({})

    render(
      <ApproveCurriculumDialog
        curriculumId={11}
        open={true}
        onClose={onClose}
        onApproved={onApproved}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'approveDialog.confirm' }))

    await waitFor(() => expect(mockApproveCurriculum).toHaveBeenCalledWith(11))
    await waitFor(() => expect(onApproved).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [422, 'approveToast.errors.notPending'],
    [403, 'approveToast.errors.forbidden'],
    [500, 'approveToast.errors.generic'],
  ])('maps HTTP %i error to toast key and stays open', async (status, expectedKey) => {
    const onClose = jest.fn()
    const onApproved = jest.fn()
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status },
    })
    mockApproveCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(
      <ApproveCurriculumDialog
        curriculumId={11}
        open={true}
        onClose={onClose}
        onApproved={onApproved}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'approveDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onApproved).not.toHaveBeenCalled()
  })

  it('does not double-fire approveCurriculum on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockApproveCurriculum.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<ApproveCurriculumDialog curriculumId={11} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'approveDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockApproveCurriculum).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
