import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { GenerateWorkProgramDialog } from '../GenerateWorkProgramDialog'

const mockGenerateWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  ...jest.requireActual('@/hooks/useWorkPrograms'),
  generateWorkProgram: (...args: unknown[]) => mockGenerateWorkProgram(...args),
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

describe('GenerateWorkProgramDialog', () => {
  it('does not render confirm button when open=false', () => {
    render(<GenerateWorkProgramDialog workProgramId={7} open={false} onClose={() => {}} />)
    expect(screen.queryByRole('button', { name: 'generateDialog.confirm' })).not.toBeInTheDocument()
  })

  it('renders title + description + cancel + confirm when open', () => {
    render(<GenerateWorkProgramDialog workProgramId={7} open={true} onClose={() => {}} />)
    expect(screen.getByText('generateDialog.title')).toBeInTheDocument()
    expect(screen.getByText('generateDialog.description')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'generateDialog.cancel' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'generateDialog.confirm' })).toBeInTheDocument()
  })

  it('calls generateWorkProgram + onGenerated + onClose + toast.success on confirm', async () => {
    const onClose = jest.fn()
    const onGenerated = jest.fn()
    mockGenerateWorkProgram.mockResolvedValueOnce({})

    render(
      <GenerateWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onGenerated={onGenerated}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'generateDialog.confirm' }))

    await waitFor(() => expect(mockGenerateWorkProgram).toHaveBeenCalledWith(7))
    await waitFor(() => expect(onGenerated).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [{ response: { data: { error: { code: 'RATE_LIMITED' } } } }, 'errors.rateLimited'],
    [{ response: { data: { error: { code: 'DRAFT_NOT_EMPTY' } } } }, 'errors.draftNotEmpty'],
    [{ response: { status: 403 } }, 'errors.forbidden'],
    [{ response: { status: 404 } }, 'errors.notFound'],
    [{ response: { status: 500 } }, 'errors.generic'],
  ])('maps backend error %# to an errors.* toast key and stays open', async (err, expectedKey) => {
    const onClose = jest.fn()
    const onGenerated = jest.fn()
    mockGenerateWorkProgram.mockRejectedValueOnce(err)

    render(
      <GenerateWorkProgramDialog
        workProgramId={7}
        open={true}
        onClose={onClose}
        onGenerated={onGenerated}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'generateDialog.confirm' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onGenerated).not.toHaveBeenCalled()
  })

  it('does not close while generation is in flight (Esc dismiss guarded)', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockGenerateWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<GenerateWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    fireEvent.click(screen.getByRole('button', { name: 'generateDialog.confirm' }))
    fireEvent.keyDown(document.body, { key: 'Escape' })
    expect(onClose).not.toHaveBeenCalled()

    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })

  it('does not double-fire generateWorkProgram on rapid double-click', async () => {
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockGenerateWorkProgram.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<GenerateWorkProgramDialog workProgramId={7} open={true} onClose={onClose} />)
    const btn = screen.getByRole('button', { name: 'generateDialog.confirm' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockGenerateWorkProgram).toHaveBeenCalledTimes(1)
    resolve({})
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
