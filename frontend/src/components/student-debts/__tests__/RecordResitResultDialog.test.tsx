import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockRecord = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  recordResitResult: (...args: unknown[]) => mockRecord(...args),
  pickStudentDebtErrorKey: () => 'generic',
}))

import { toast } from 'sonner'
import { RecordResitResultDialog } from '../RecordResitResultDialog'

beforeEach(() => {
  mockRecord.mockReset()
  ;(toast.success as jest.Mock).mockClear()
  ;(toast.error as jest.Mock).mockClear()
})

describe('RecordResitResultDialog', () => {
  it('renders title when open', () => {
    render(<RecordResitResultDialog debtId={1} attemptNo={1} open onClose={jest.fn()} />)
    expect(screen.getByText('recordDialog.title')).toBeInTheDocument()
  })

  it('records a passed result with a grade', async () => {
    mockRecord.mockResolvedValueOnce({ id: 1 })
    const onRecorded = jest.fn()
    render(
      <RecordResitResultDialog
        debtId={9}
        attemptNo={2}
        open
        onClose={jest.fn()}
        onRecorded={onRecorded}
      />
    )

    fireEvent.change(screen.getByLabelText('recordDialog.labels.result'), {
      target: { value: 'passed' },
    })
    fireEvent.change(screen.getByLabelText('recordDialog.labels.grade'), {
      target: { value: '5' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.confirm' }))

    await waitFor(() => expect(mockRecord).toHaveBeenCalledTimes(1))
    const [id, attemptNo, input] = mockRecord.mock.calls[0]
    expect(id).toBe(9)
    expect(attemptNo).toBe(2)
    expect(input.result).toBe('passed')
    expect(input.grade).toBe(5)
    expect(onRecorded).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalled()
  })

  it('records a no_show with a null grade (empty field)', async () => {
    mockRecord.mockResolvedValueOnce({ id: 1 })
    render(<RecordResitResultDialog debtId={9} attemptNo={1} open onClose={jest.fn()} />)
    fireEvent.change(screen.getByLabelText('recordDialog.labels.result'), {
      target: { value: 'no_show' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.confirm' }))
    await waitFor(() => expect(mockRecord).toHaveBeenCalledTimes(1))
    const [, , input] = mockRecord.mock.calls[0]
    expect(input.result).toBe('no_show')
    expect(input.grade).toBeNull()
  })

  it('shows an error toast on failure', async () => {
    mockRecord.mockRejectedValueOnce({ response: { status: 409 } })
    render(<RecordResitResultDialog debtId={9} attemptNo={1} open onClose={jest.fn()} />)
    fireEvent.change(screen.getByLabelText('recordDialog.labels.result'), {
      target: { value: 'failed' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'recordDialog.confirm' }))
    await waitFor(() => expect(toast.error).toHaveBeenCalled())
  })
})
