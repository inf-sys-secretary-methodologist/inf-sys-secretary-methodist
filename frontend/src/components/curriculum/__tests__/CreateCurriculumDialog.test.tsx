import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { CreateCurriculumDialog } from '../CreateCurriculumDialog'

const mockCreateCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  ...jest.requireActual('@/hooks/useCurricula'),
  createCurriculum: (...args: unknown[]) => mockCreateCurriculum(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

const noop = () => {}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('CreateCurriculumDialog', () => {
  it('does not render fields when open=false', () => {
    render(<CreateCurriculumDialog open={false} onClose={noop} />)
    expect(screen.queryByLabelText('createDialog.labels.title')).not.toBeInTheDocument()
  })

  it('starts with empty inputs (no pre-fill — Create vs Edit)', () => {
    render(<CreateCurriculumDialog open={true} onClose={noop} />)
    expect(screen.getByLabelText('createDialog.labels.title')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.code')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.specialty')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.year')).toHaveValue('')
    expect(screen.getByLabelText('createDialog.labels.description')).toHaveValue('')
  })

  it('keeps Create button disabled until all required fields are filled', () => {
    render(<CreateCurriculumDialog open={true} onClose={noop} />)
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()

    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'ИВТ-2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: '09.03.04-2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'Информатика' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).not.toBeDisabled()
  })

  it.each([1999, 2101, 0])('disables Create when year is out of range (%i)', (year) => {
    render(<CreateCurriculumDialog open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'X' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: 'C' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'S' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: String(year) },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
  })

  it('disables Create when description exceeds 4096 chars', () => {
    render(<CreateCurriculumDialog open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'X' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: 'C' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'S' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.description'), {
      target: { value: 'x'.repeat(4097) },
    })
    expect(screen.getByRole('button', { name: 'createDialog.create' })).toBeDisabled()
  })

  it('calls createCurriculum and fires onCreated + close on success', async () => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    mockCreateCurriculum.mockResolvedValueOnce({ id: 12, status: 'draft' })

    render(<CreateCurriculumDialog open={true} onClose={onClose} onCreated={onCreated} />)

    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'ИВТ-2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: '09.03.04-2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'Информатика' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.description'), {
      target: { value: 'New plan' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'createDialog.create' }))

    await waitFor(() => {
      expect(mockCreateCurriculum).toHaveBeenCalledWith({
        title: 'ИВТ-2027',
        code: '09.03.04-2027',
        specialty: 'Информатика',
        year: 2027,
        description: 'New plan',
      })
    })
    await waitFor(() => expect(onCreated).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [409, 'createDialog.errors.codeExists'],
    [422, 'createDialog.errors.invalidInput'],
    [403, 'createDialog.errors.forbidden'],
    [500, 'createDialog.errors.generic'],
  ])('maps HTTP %i error to toast key and keeps dialog open', async (status, expectedKey) => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status, data: { error: { code: 'X' } } },
    })
    mockCreateCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(<CreateCurriculumDialog open={true} onClose={onClose} onCreated={onCreated} />)

    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'X' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: 'C' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'S' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'createDialog.create' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    expect(onClose).not.toHaveBeenCalled()
    expect(onCreated).not.toHaveBeenCalled()
  })

  it('does not double-fire createCurriculum on rapid double-click', async () => {
    const onCreated = jest.fn()
    const onClose = jest.fn()
    let resolve: (v: unknown) => void = () => {}
    mockCreateCurriculum.mockImplementation(
      () =>
        new Promise((r) => {
          resolve = r
        })
    )

    render(<CreateCurriculumDialog open={true} onClose={onClose} onCreated={onCreated} />)
    fireEvent.change(screen.getByLabelText('createDialog.labels.title'), {
      target: { value: 'X' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.code'), {
      target: { value: 'C' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.specialty'), {
      target: { value: 'S' },
    })
    fireEvent.change(screen.getByLabelText('createDialog.labels.year'), {
      target: { value: '2027' },
    })

    const btn = screen.getByRole('button', { name: 'createDialog.create' })
    fireEvent.click(btn)
    fireEvent.click(btn)

    expect(mockCreateCurriculum).toHaveBeenCalledTimes(1)
    resolve({ id: 12 })
    await waitFor(() => expect(onClose).toHaveBeenCalled())
  })
})
