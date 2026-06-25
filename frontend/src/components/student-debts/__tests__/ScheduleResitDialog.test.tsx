import { render, screen } from '@/test-utils'
import { fireEvent, waitFor } from '@testing-library/react'

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

const mockScheduleResit = jest.fn()
jest.mock('@/hooks/useStudentDebts', () => ({
  scheduleResit: (...args: unknown[]) => mockScheduleResit(...args),
  pickStudentDebtErrorKey: () => 'generic',
}))

import { toast } from 'sonner'
import { ScheduleResitDialog } from '../ScheduleResitDialog'

beforeEach(() => {
  mockScheduleResit.mockReset()
  ;(toast.success as jest.Mock).mockClear()
  ;(toast.error as jest.Mock).mockClear()
})

describe('ScheduleResitDialog', () => {
  it('renders title when open', () => {
    render(<ScheduleResitDialog debtId={1} open onClose={jest.fn()} />)
    expect(screen.getByText('scheduleDialog.title')).toBeInTheDocument()
  })

  it('disables confirm until date and examiner are filled', () => {
    render(<ScheduleResitDialog debtId={1} open onClose={jest.fn()} />)
    expect(screen.getByRole('button', { name: 'scheduleDialog.confirm' })).toBeDisabled()
  })

  it('schedules the resit with an RFC3339 date and examiner', async () => {
    mockScheduleResit.mockResolvedValueOnce({ id: 1 })
    const onScheduled = jest.fn()
    const onClose = jest.fn()
    render(<ScheduleResitDialog debtId={9} open onClose={onClose} onScheduled={onScheduled} />)

    fireEvent.change(screen.getByLabelText('scheduleDialog.labels.scheduledDate'), {
      target: { value: '2026-07-01' },
    })
    fireEvent.change(screen.getByLabelText('scheduleDialog.labels.examiner'), {
      target: { value: 'Петров П.П.' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'scheduleDialog.confirm' }))

    await waitFor(() => expect(mockScheduleResit).toHaveBeenCalledTimes(1))
    const [id, input] = mockScheduleResit.mock.calls[0]
    expect(id).toBe(9)
    expect(input.examiner).toBe('Петров П.П.')
    // local-midnight parse → RFC3339 timestamp (not a bare date)
    expect(input.scheduled_date).toContain('2026-07-01')
    expect(input.scheduled_date).toMatch(/T/)
    expect(onScheduled).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalled()
  })

  it('shows an error toast and stays open on failure', async () => {
    mockScheduleResit.mockRejectedValueOnce({ response: { status: 409 } })
    const onClose = jest.fn()
    render(<ScheduleResitDialog debtId={9} open onClose={onClose} />)
    fireEvent.change(screen.getByLabelText('scheduleDialog.labels.scheduledDate'), {
      target: { value: '2026-07-01' },
    })
    fireEvent.change(screen.getByLabelText('scheduleDialog.labels.examiner'), {
      target: { value: 'Петров' },
    })
    fireEvent.click(screen.getByRole('button', { name: 'scheduleDialog.confirm' }))
    await waitFor(() => expect(toast.error).toHaveBeenCalled())
    expect(onClose).not.toHaveBeenCalled()
  })
})
