import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockImport1C = jest.fn()
jest.mock('@/lib/api/studentDebts', () => ({
  studentDebtsApi: {
    import1C: (...args: unknown[]) => mockImport1C(...args),
  },
}))

import { toast } from 'sonner'
import { Import1CDialog } from '../Import1CDialog'

beforeEach(() => {
  mockImport1C.mockReset()
  ;(toast.success as jest.Mock).mockClear()
  ;(toast.error as jest.Mock).mockClear()
})

describe('Import1CDialog', () => {
  it('renders title + description when open', () => {
    render(<Import1CDialog open onClose={jest.fn()} />)
    expect(screen.getByText('import1CDialog.title')).toBeInTheDocument()
  })

  it('pulls from 1С on confirm and reports the result', async () => {
    mockImport1C.mockResolvedValueOnce({ created: 3, updated: 1, skipped: 0, errors: [] })
    const onImported = jest.fn()
    render(<Import1CDialog open onClose={jest.fn()} onImported={onImported} />)

    fireEvent.click(screen.getByRole('button', { name: 'import1CDialog.confirm' }))

    await waitFor(() => expect(mockImport1C).toHaveBeenCalledTimes(1))
    expect(toast.success).toHaveBeenCalled()
    expect(onImported).toHaveBeenCalled()
  })

  it('shows an error toast when the 1С import fails (forbidden)', async () => {
    mockImport1C.mockRejectedValueOnce({ response: { status: 403 } })
    render(<Import1CDialog open onClose={jest.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'import1CDialog.confirm' }))
    await waitFor(() => expect(toast.error).toHaveBeenCalled())
  })

  it('lists per-row errors returned by the 1С import', async () => {
    mockImport1C.mockResolvedValueOnce({
      created: 1,
      updated: 0,
      skipped: 1,
      errors: [{ row: 2, identity: 'Волков|ПИ-22|Программирование', message: 'invalid semester' }],
    })
    render(<Import1CDialog open onClose={jest.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'import1CDialog.confirm' }))
    await waitFor(() =>
      expect(screen.getByText('import1CDialog.rowErrorsTitle')).toBeInTheDocument()
    )
  })
})
