import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockImport = jest.fn()
jest.mock('@/lib/api/studentDebts', () => ({
  studentDebtsApi: {
    import: (...args: unknown[]) => mockImport(...args),
  },
}))

import { toast } from 'sonner'
import { ImportDebtsDialog } from '../ImportDebtsDialog'

const file = new File(['xlsx-bytes'], 'debts.xlsx', {
  type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
})

beforeEach(() => {
  mockImport.mockReset()
  ;(toast.success as jest.Mock).mockClear()
  ;(toast.error as jest.Mock).mockClear()
})

describe('ImportDebtsDialog', () => {
  it('renders title + description when open', () => {
    render(<ImportDebtsDialog open onClose={jest.fn()} />)
    expect(screen.getByText('importDialog.title')).toBeInTheDocument()
  })

  it('disables Import until a file is chosen', () => {
    render(<ImportDebtsDialog open onClose={jest.fn()} />)
    expect(screen.getByRole('button', { name: 'importDialog.confirm' })).toBeDisabled()
  })

  it('imports the chosen file and reports the result', async () => {
    mockImport.mockResolvedValueOnce({ created: 3, updated: 1, skipped: 0, errors: [] })
    const onImported = jest.fn()
    const onClose = jest.fn()
    render(<ImportDebtsDialog open onClose={onClose} onImported={onImported} />)

    const input = screen.getByLabelText('importDialog.selectFile') as HTMLInputElement
    fireEvent.change(input, { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'importDialog.confirm' }))

    await waitFor(() => expect(mockImport).toHaveBeenCalledWith(file))
    expect(toast.success).toHaveBeenCalled()
    expect(onImported).toHaveBeenCalled()
  })

  it('shows an error toast when the import fails (forbidden)', async () => {
    mockImport.mockRejectedValueOnce({ response: { status: 403 } })
    render(<ImportDebtsDialog open onClose={jest.fn()} />)
    const input = screen.getByLabelText('importDialog.selectFile') as HTMLInputElement
    fireEvent.change(input, { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'importDialog.confirm' }))
    await waitFor(() => expect(toast.error).toHaveBeenCalled())
  })

  it('lists per-row errors returned by the import', async () => {
    mockImport.mockResolvedValueOnce({
      created: 1,
      updated: 0,
      skipped: 1,
      errors: [{ row: 4, identity: 'Иванов|ИС-21|БД', message: 'invalid semester' }],
    })
    render(<ImportDebtsDialog open onClose={jest.fn()} />)
    const input = screen.getByLabelText('importDialog.selectFile') as HTMLInputElement
    fireEvent.change(input, { target: { files: [file] } })
    fireEvent.click(screen.getByRole('button', { name: 'importDialog.confirm' }))
    await waitFor(() => expect(screen.getByText('importDialog.rowErrorsTitle')).toBeInTheDocument())
  })
})
