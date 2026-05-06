import { render, screen, fireEvent, waitFor } from '@/test-utils'
import { EditCurriculumDialog } from '../EditCurriculumDialog'
import type { Curriculum } from '@/types/curriculum'

const mockUpdateCurriculum = jest.fn()
jest.mock('@/hooks/useCurricula', () => ({
  ...jest.requireActual('@/hooks/useCurricula'),
  updateCurriculum: (...args: unknown[]) => mockUpdateCurriculum(...args),
}))

const mockToastSuccess = jest.fn()
const mockToastError = jest.fn()
jest.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

const sample: Curriculum = {
  id: 11,
  title: 'ИВТ-2026 / 4 года',
  code: '09.03.04-2026',
  specialty: 'Информатика и вычислительная техника',
  year: 2026,
  description: 'Учебный план направления подготовки',
  status: 'draft',
  created_by: 5,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-01T08:00:00Z',
}

const noop = () => {}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('EditCurriculumDialog', () => {
  it('does not render fields when open=false', () => {
    render(<EditCurriculumDialog curriculum={sample} open={false} onClose={noop} />)
    expect(screen.queryByLabelText('editDialog.labels.title')).not.toBeInTheDocument()
  })

  it('pre-fills inputs with curriculum values when open', () => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    expect(screen.getByDisplayValue('ИВТ-2026 / 4 года')).toBeInTheDocument()
    expect(screen.getByDisplayValue('09.03.04-2026')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Информатика и вычислительная техника')).toBeInTheDocument()
    expect(screen.getByDisplayValue('2026')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Учебный план направления подготовки')).toBeInTheDocument()
  })

  it('disables Save button when title is empty after trim', () => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('editDialog.labels.title'), { target: { value: '   ' } })
    expect(screen.getByRole('button', { name: 'editDialog.save' })).toBeDisabled()
  })

  it('disables Save button when code is empty', () => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('editDialog.labels.code'), { target: { value: '' } })
    expect(screen.getByRole('button', { name: 'editDialog.save' })).toBeDisabled()
  })

  it('disables Save button when specialty is empty', () => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('editDialog.labels.specialty'), {
      target: { value: '' },
    })
    expect(screen.getByRole('button', { name: 'editDialog.save' })).toBeDisabled()
  })

  it.each([1999, 2101, 0, -2026])('disables Save button when year is out of range (%i)', (year) => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    fireEvent.change(screen.getByLabelText('editDialog.labels.year'), {
      target: { value: String(year) },
    })
    expect(screen.getByRole('button', { name: 'editDialog.save' })).toBeDisabled()
  })

  it('disables Save button when description exceeds 4096 chars', () => {
    render(<EditCurriculumDialog curriculum={sample} open={true} onClose={noop} />)
    const long = 'x'.repeat(4097)
    fireEvent.change(screen.getByLabelText('editDialog.labels.description'), {
      target: { value: long },
    })
    expect(screen.getByRole('button', { name: 'editDialog.save' })).toBeDisabled()
  })

  it('calls updateCurriculum with edited body and fires onSaved + close on success', async () => {
    const onSaved = jest.fn()
    const onClose = jest.fn()
    mockUpdateCurriculum.mockResolvedValueOnce({ ...sample, title: 'Updated title' })
    render(
      <EditCurriculumDialog curriculum={sample} open={true} onClose={onClose} onSaved={onSaved} />
    )

    fireEvent.change(screen.getByLabelText('editDialog.labels.title'), {
      target: { value: 'Updated title' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'editDialog.save' }))

    await waitFor(() => {
      expect(mockUpdateCurriculum).toHaveBeenCalledWith(11, {
        title: 'Updated title',
        code: sample.code,
        specialty: sample.specialty,
        year: sample.year,
        description: sample.description,
      })
    })
    await waitFor(() => expect(onSaved).toHaveBeenCalled())
    expect(onClose).toHaveBeenCalled()
    expect(mockToastSuccess).toHaveBeenCalled()
  })

  it.each([
    [409, 'editDialog.errors.codeExists'],
    [422, 'editDialog.errors.notEditable'],
    [403, 'editDialog.errors.forbidden'],
    [500, 'editDialog.errors.generic'],
  ])('maps HTTP %i error to toast key', async (status, expectedKey) => {
    const onSaved = jest.fn()
    const onClose = jest.fn()
    const axiosLikeErr = Object.assign(new Error('boom'), {
      isAxiosError: true,
      response: { status, data: { error: { code: status === 422 ? 'NOT_EDITABLE' : 'X' } } },
    })
    mockUpdateCurriculum.mockRejectedValueOnce(axiosLikeErr)

    render(
      <EditCurriculumDialog curriculum={sample} open={true} onClose={onClose} onSaved={onSaved} />
    )
    fireEvent.click(screen.getByRole('button', { name: 'editDialog.save' }))

    await waitFor(() => expect(mockToastError).toHaveBeenCalled())
    expect(mockToastError.mock.calls[0][0]).toBe(expectedKey)
    // Dialog must stay open on error.
    expect(onClose).not.toHaveBeenCalled()
    expect(onSaved).not.toHaveBeenCalled()
  })
})
