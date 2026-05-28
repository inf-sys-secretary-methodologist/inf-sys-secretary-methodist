import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { ApproveWorkProgramDialog } from '../ApproveWorkProgramDialog'

const mockApproveWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  approveWorkProgram: (...args: unknown[]) => mockApproveWorkProgram(...args),
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

describe('ApproveWorkProgramDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<ApproveWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByRole('button', { name: 'approveDialog.confirm' })).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<ApproveWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('approveDialog.title')).toBeInTheDocument()
    expect(screen.getByText('approveDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'approveDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'approveDialog.confirm' })).toBeInTheDocument()
  })

  it('calls approveWorkProgram + onApproved + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onApproved = jest.fn()
    mockApproveWorkProgram.mockResolvedValueOnce({})

    render(
      <ApproveWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onApproved={onApproved}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'approveDialog.confirm' }))

    await waitFor(() => expect(mockApproveWorkProgram).toHaveBeenCalledWith(7))
    await waitFor(() => expect(onApproved).toHaveBeenCalled())
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
    const onApproved = jest.fn()
    mockApproveWorkProgram.mockRejectedValueOnce(err)

    render(
      <ApproveWorkProgramDialog
        workProgramId={7}
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

  it('does not close while an approve is in flight (Esc dismiss guarded)', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockApproveWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<ApproveWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    fireEvent.click(screen.getByRole('button', { name: 'approveDialog.confirm' }))
    fireEvent.keyDown(document.body, { key: 'Escape' })
    expect(onClose).not.toHaveBeenCalled()

    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })

  it('does not double-fire approveWorkProgram on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockApproveWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<ApproveWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'approveDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockApproveWorkProgram).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
