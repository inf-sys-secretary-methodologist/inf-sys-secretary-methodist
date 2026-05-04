import { fireEvent, render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import axios from 'axios'

import { ReturnDialog } from '../ReturnDialog'
import { returnSubmission } from '@/hooks/useAssignments'
import type { SubmissionView } from '@/types/assignments'

jest.mock('@/hooks/useAssignments', () => ({
  returnSubmission: jest.fn(),
}))

jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockedReturnSubmission = jest.mocked(returnSubmission)
// eslint-disable-next-line @typescript-eslint/no-require-imports
const { toast } = require('sonner') as { toast: { success: jest.Mock; error: jest.Mock } }

const pendingSubmission: SubmissionView = {
  id: 1,
  assignment_id: 10,
  student_id: 7,
  student_name: 'Иван Петров',
  status: 'pending',
  created_at: '2026-05-01T00:00:00Z',
  updated_at: '2026-05-01T00:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('ReturnDialog', () => {
  it('renders nothing when open=false', () => {
    render(
      <ReturnDialog
        assignmentId={10}
        submission={pendingSubmission}
        open={false}
        onClose={jest.fn()}
      />
    )
    expect(screen.queryByText(/returnDialog\.title/)).not.toBeInTheDocument()
  })

  it('confirm is disabled while reason is empty', async () => {
    const user = userEvent.setup()
    render(
      <ReturnDialog
        assignmentId={10}
        submission={pendingSubmission}
        open
        onClose={jest.fn()}
      />
    )
    const confirm = screen.getByRole('button', { name: /returnDialog\.confirm/ })
    expect(confirm).toBeDisabled()

    // typing whitespace only — still disabled (trim → empty)
    await user.type(screen.getByLabelText(/returnDialog\.reasonLabel/), '   ')
    expect(confirm).toBeDisabled()
  })

  it('confirm is disabled when reason exceeds 4096 chars', () => {
    render(
      <ReturnDialog
        assignmentId={10}
        submission={pendingSubmission}
        open
        onClose={jest.fn()}
      />
    )
    const textarea = screen.getByLabelText(/returnDialog\.reasonLabel/) as HTMLTextAreaElement
    // Bypass user.type for performance: directly set the value via fireEvent.
    const longText = 'a'.repeat(4097)
    fireEvent.change(textarea, { target: { value: longText } })
    expect(screen.getByRole('button', { name: /returnDialog\.confirm/ })).toBeDisabled()
  })

  it('submits reason and fires onReturned + onClose on success', async () => {
    const user = userEvent.setup()
    const onClose = jest.fn()
    const onReturned = jest.fn()
    mockedReturnSubmission.mockResolvedValueOnce({
      assignment_id: 10,
      student_id: 7,
      reason: 'fix it',
    })

    render(
      <ReturnDialog
        assignmentId={10}
        submission={pendingSubmission}
        open
        onClose={onClose}
        onReturned={onReturned}
      />
    )

    await user.type(screen.getByLabelText(/returnDialog\.reasonLabel/), 'fix it')
    await user.click(screen.getByRole('button', { name: /returnDialog\.confirm/ }))

    await waitFor(() => expect(mockedReturnSubmission).toHaveBeenCalledTimes(1))
    expect(mockedReturnSubmission).toHaveBeenCalledWith(10, { student_id: 7, reason: 'fix it' })
    await waitFor(() => expect(toast.success).toHaveBeenCalled())
    expect(onReturned).toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })

  it('cancel calls onClose without invoking the API', async () => {
    const user = userEvent.setup()
    const onClose = jest.fn()
    render(
      <ReturnDialog
        assignmentId={10}
        submission={pendingSubmission}
        open
        onClose={onClose}
      />
    )
    await user.click(screen.getByRole('button', { name: /returnDialog\.cancel/ }))
    expect(onClose).toHaveBeenCalled()
    expect(mockedReturnSubmission).not.toHaveBeenCalled()
  })

  describe('error mapping', () => {
    const cases: Array<{ status: number; key: string }> = [
      { status: 409, key: 'errors.alreadyReturned' },
      { status: 422, key: 'errors.invalidReason' },
      { status: 403, key: 'errors.forbidden' },
    ]

    it.each(cases)('maps HTTP $status to $key toast and stays open', async ({ status, key }) => {
      const axiosErr = Object.assign(new Error('http'), {
        isAxiosError: true,
        response: { status, data: {} },
        toJSON: () => ({}),
      })
      jest.spyOn(axios, 'isAxiosError').mockReturnValue(true)
      mockedReturnSubmission.mockRejectedValueOnce(axiosErr)

      const user = userEvent.setup()
      const onClose = jest.fn()
      render(
        <ReturnDialog
          assignmentId={10}
          submission={pendingSubmission}
          open
          onClose={onClose}
        />
      )

      await user.type(screen.getByLabelText(/returnDialog\.reasonLabel/), 'fix it')
      await user.click(screen.getByRole('button', { name: /returnDialog\.confirm/ }))

      await waitFor(() => expect(toast.error).toHaveBeenCalled())
      expect((toast.error as jest.Mock).mock.calls[0][0]).toContain(key)
      // Error path must NOT close the dialog — let the user retry.
      expect(onClose).not.toHaveBeenCalled()
    })

    it('falls through to generic on non-axios error', async () => {
      jest.spyOn(axios, 'isAxiosError').mockReturnValue(false)
      mockedReturnSubmission.mockRejectedValueOnce(new Error('network'))

      const user = userEvent.setup()
      render(
        <ReturnDialog
          assignmentId={10}
          submission={pendingSubmission}
          open
          onClose={jest.fn()}
        />
      )

      await user.type(screen.getByLabelText(/returnDialog\.reasonLabel/), 'fix it')
      await user.click(screen.getByRole('button', { name: /returnDialog\.confirm/ }))

      await waitFor(() => expect(toast.error).toHaveBeenCalled())
      expect((toast.error as jest.Mock).mock.calls[0][0]).toContain('errors.generic')
    })
  })
})
