import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'

import { ResubmitDialog } from '../ResubmitDialog'
import { resubmitSubmission } from '@/hooks/useMyAssignments'

jest.mock('@/hooks/useMyAssignments', () => ({
  resubmitSubmission: jest.fn(),
  useMyAssignments: jest.fn(),
  useMyAssignment: jest.fn(),
}))

jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockedResubmit = jest.mocked(resubmitSubmission)
// eslint-disable-next-line @typescript-eslint/no-require-imports
const { toast } = require('sonner') as { toast: { success: jest.Mock; error: jest.Mock } }

beforeEach(() => {
  jest.clearAllMocks()
})

describe('ResubmitDialog', () => {
  it('renders nothing when open=false', () => {
    render(<ResubmitDialog assignmentId={10} open={false} onClose={jest.fn()} />)
    expect(screen.queryByText(/resubmitDialog\.title/)).not.toBeInTheDocument()
  })

  it('renders title, description and confirm button when open', () => {
    render(<ResubmitDialog assignmentId={10} open onClose={jest.fn()} />)
    expect(screen.getByText(/resubmitDialog\.title/)).toBeInTheDocument()
    expect(screen.getByText(/resubmitDialog\.description/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /resubmitDialog\.confirm/ })).toBeInTheDocument()
  })

  it('cancel closes the dialog without invoking the API', async () => {
    const user = userEvent.setup()
    const onClose = jest.fn()
    render(<ResubmitDialog assignmentId={10} open onClose={onClose} />)

    await user.click(screen.getByRole('button', { name: /resubmitDialog\.cancel/ }))
    expect(onClose).toHaveBeenCalled()
    expect(mockedResubmit).not.toHaveBeenCalled()
  })

  it('confirm invokes resubmitSubmission, fires onResubmitted, closes, and toasts success', async () => {
    const user = userEvent.setup()
    const onClose = jest.fn()
    const onResubmitted = jest.fn()
    mockedResubmit.mockResolvedValueOnce({ assignment_id: 10, student_id: 7 })

    render(
      <ResubmitDialog
        assignmentId={10}
        open
        onClose={onClose}
        onResubmitted={onResubmitted}
      />
    )

    await user.click(screen.getByRole('button', { name: /resubmitDialog\.confirm/ }))

    await waitFor(() => expect(mockedResubmit).toHaveBeenCalledWith(10))
    expect(onResubmitted).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalled()
  })

  it('maps 409 NOT_RETURNED error to a specific toast key', async () => {
    const user = userEvent.setup()
    const err = Object.assign(new Error('conflict'), {
      isAxiosError: true,
      response: { status: 409, data: {} },
    })
    mockedResubmit.mockRejectedValueOnce(err)

    render(<ResubmitDialog assignmentId={10} open onClose={jest.fn()} />)
    await user.click(screen.getByRole('button', { name: /resubmitDialog\.confirm/ }))

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith(
        expect.stringContaining('notReturned')
      )
    )
  })

  it('maps 403 forbidden error to a specific toast key', async () => {
    const user = userEvent.setup()
    const err = Object.assign(new Error('forbidden'), {
      isAxiosError: true,
      response: { status: 403, data: {} },
    })
    mockedResubmit.mockRejectedValueOnce(err)

    render(<ResubmitDialog assignmentId={10} open onClose={jest.fn()} />)
    await user.click(screen.getByRole('button', { name: /resubmitDialog\.confirm/ }))

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith(
        expect.stringContaining('forbidden')
      )
    )
  })

  it('falls through to a generic toast for non-mapped errors', async () => {
    const user = userEvent.setup()
    mockedResubmit.mockRejectedValueOnce(new Error('boom'))

    render(<ResubmitDialog assignmentId={10} open onClose={jest.fn()} />)
    await user.click(screen.getByRole('button', { name: /resubmitDialog\.confirm/ }))

    await waitFor(() =>
      expect(toast.error).toHaveBeenCalledWith(expect.stringContaining('generic'))
    )
  })

  it('keeps the dialog open on error so the student can retry', async () => {
    const user = userEvent.setup()
    const onClose = jest.fn()
    mockedResubmit.mockRejectedValueOnce(new Error('boom'))

    render(<ResubmitDialog assignmentId={10} open onClose={onClose} />)
    await user.click(screen.getByRole('button', { name: /resubmitDialog\.confirm/ }))

    await waitFor(() => expect(toast.error).toHaveBeenCalled())
    expect(onClose).not.toHaveBeenCalled()
  })
})
