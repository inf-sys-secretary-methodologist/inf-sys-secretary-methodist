import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { SubmitCurriculumDialog } from '../SubmitCurriculumDialog'

const mockSubmitCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  ...jest.requireActual('@/hooks/useCurricula'),
  submitCurriculum: (...args: unknown[]) => mockSubmitCurriculum(...args),
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

describe('SubmitCurriculumDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<SubmitCurriculumDialog curriculumId={11} open={false} onClose={() => {}} />)
    expect(screen.queryByRole('button', { name: 'submitDialog.confirm' })).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<SubmitCurriculumDialog curriculumId={11} open={true} onClose={() => {}} />)
    expect(screen.getByText('submitDialog.title')).toBeInTheDocument()
    expect(screen.getByText('submitDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'submitDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'submitDialog.confirm' })).toBeInTheDocument()
  })

  it('calls submitCurriculum + onSubmitted + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    mockSubmitCurriculum.mockResolvedValueOnce({})

    render(
      <SubmitCurriculumDialog
        curriculumId={11}
        open={true}
        onClose={onClose}
        onSubmitted={onSubmitted}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'submitDialog.confirm' }))

    await waitFor(() => expect(mockSubmitCurriculum).toHaveBeenCalledWith(11))
    await waitFor(() => expect(onSubmitted).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [422, 'submitToast.errors.notDraft'],
    [403, 'submitToast.errors.forbidden'],
    [500, 'submitToast.errors.generic'],
  ])('maps HTTP %i error to toast key and stays open', async (status, expectedKey) => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status },
    })
    mockSubmitCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(
      <SubmitCurriculumDialog
        curriculumId={11}
        open={true}
        onClose={onClose}
        onSubmitted={onSubmitted}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'submitDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onSubmitted).not.toHaveBeenCalled()
  })

  it('does not double-fire submitCurriculum on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockSubmitCurriculum.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<SubmitCurriculumDialog curriculumId={11} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'submitDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockSubmitCurriculum).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
