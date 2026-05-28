import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { DiscardWorkProgramDialog } from '../DiscardWorkProgramDialog'

const mockDiscardWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  discardWorkProgram: (...args: unknown[]) => mockDiscardWorkProgram(...args),
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

describe('DiscardWorkProgramDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<DiscardWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByRole('button', { name: 'discardDialog.confirm' })).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<DiscardWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('discardDialog.title')).toBeInTheDocument()
    expect(screen.getByText('discardDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'discardDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'discardDialog.confirm' })).toBeInTheDocument()
  })

  it('calls discardWorkProgram + onDiscarded + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onDiscarded = jest.fn()
    mockDiscardWorkProgram.mockResolvedValueOnce({})

    render(
      <DiscardWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onDiscarded={onDiscarded}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'discardDialog.confirm' }))

    await waitFor(() => expect(mockDiscardWorkProgram).toHaveBeenCalledWith(7))
    await waitFor(() => expect(onDiscarded).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [{ response: { data: { error: { code: 'INVALID_TRANSITION' } } } }, 'errors.invalidTransition'],
    [{ response: { status: 403 } }, 'errors.forbidden'],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onDiscarded = jest.fn()
    mockDiscardWorkProgram.mockRejectedValueOnce(err)

    render(
      <DiscardWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onDiscarded={onDiscarded}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'discardDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onDiscarded).not.toHaveBeenCalled()
  })

  it('does not double-fire discardWorkProgram on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockDiscardWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<DiscardWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'discardDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockDiscardWorkProgram).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
