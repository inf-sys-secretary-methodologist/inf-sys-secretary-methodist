import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { SubmitWorkProgramDialog } from '../SubmitWorkProgramDialog'

const mockSubmitWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  submitWorkProgram: (...args: unknown[]) => mockSubmitWorkProgram(...args),
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

describe('SubmitWorkProgramDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<SubmitWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByRole('button', { name: 'submitDialog.confirm' })).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<SubmitWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('submitDialog.title')).toBeInTheDocument()
    expect(screen.getByText('submitDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'submitDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'submitDialog.confirm' })).toBeInTheDocument()
  })

  it('calls submitWorkProgram + onSubmitted + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    mockSubmitWorkProgram.mockResolvedValueOnce({})

    render(
      <SubmitWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onSubmitted={onSubmitted}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'submitDialog.confirm' }))

    await waitFor(() => expect(mockSubmitWorkProgram).toHaveBeenCalledWith(7))
    await waitFor(() => expect(onSubmitted).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [{ response: { data: { error: { code: 'INVALID_TRANSITION' } } } }, 'errors.invalidTransition'],
    [{ response: { data: { error: { code: 'VERSION_CONFLICT' } } } }, 'errors.versionConflict'],
    [{ response: { status: 403 } }, 'errors.forbidden'],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onSubmitted = jest.fn()
    mockSubmitWorkProgram.mockRejectedValueOnce(err)

    render(
      <SubmitWorkProgramDialog
        workProgramId={7}
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

  it('does not double-fire submitWorkProgram on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockSubmitWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<SubmitWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'submitDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockSubmitWorkProgram).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
