import { fireEvent, render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import axios from 'axios'

import { GradeForm } from '../GradeForm'
import { saveGrade } from '@/hooks/useAssignments'
import type { SubmissionView } from '@/types/assignments'

jest.mock('@/hooks/useAssignments', () => ({
  saveGrade: jest.fn(),
}))

jest.mock('sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockedSaveGrade = jest.mocked(saveGrade)
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

const gradedSubmission: SubmissionView = {
  ...pendingSubmission,
  id: 2,
  status: 'graded',
  grade_value: 85,
  feedback: 'good',
  graded_at: '2026-05-02T10:00:00Z',
}

beforeEach(() => {
  jest.clearAllMocks()
})

describe('GradeForm', () => {
  it('rejects empty value with NOT_A_NUMBER message', async () => {
    const user = userEvent.setup()
    render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

    await user.click(screen.getByRole('button'))
    expect(screen.getByRole('alert').textContent).toMatch(/validation\.NOT_A_NUMBER/)
    expect(mockedSaveGrade).not.toHaveBeenCalled()
  })

  it('rejects negative value with NEGATIVE message', async () => {
    render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

    // Use fireEvent throughout because user.type with pattern="[0-9]*"
    // skips "-" / "." keystrokes in some userEvent versions, which
    // would defeat the test's intent.
    fireEvent.change(screen.getByLabelText(/valueLabel/), { target: { value: '-5' } })
    fireEvent.submit(screen.getByLabelText(/valueLabel/).closest('form')!)
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument())
    expect(screen.getByRole('alert').textContent).toMatch(/validation\.NEGATIVE/)
    expect(mockedSaveGrade).not.toHaveBeenCalled()
  })

  it('rejects value over max with OVER_MAX message', async () => {
    const user = userEvent.setup()
    render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

    await user.type(screen.getByLabelText(/valueLabel/), '150')
    await user.click(screen.getByRole('button'))
    expect(screen.getByRole('alert').textContent).toMatch(/validation\.OVER_MAX/)
    expect(mockedSaveGrade).not.toHaveBeenCalled()
  })

  it('rejects fractional value with NOT_INTEGER message', async () => {
    render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

    fireEvent.change(screen.getByLabelText(/valueLabel/), { target: { value: '85.5' } })
    fireEvent.submit(screen.getByLabelText(/valueLabel/).closest('form')!)
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument())
    expect(screen.getByRole('alert').textContent).toMatch(/validation\.NOT_INTEGER/)
    expect(mockedSaveGrade).not.toHaveBeenCalled()
  })

  it('POSTs valid grade and shows success toast', async () => {
    mockedSaveGrade.mockResolvedValueOnce({ assignment_id: 10, student_id: 7, value: 85 })
    const user = userEvent.setup()
    const onSaved = jest.fn()
    render(
      <GradeForm
        assignmentId={10}
        maxScore={100}
        submission={pendingSubmission}
        onSaved={onSaved}
      />
    )

    await user.type(screen.getByLabelText(/valueLabel/), '85')
    await user.type(screen.getByLabelText(/feedbackLabel/), 'great work')
    await user.click(screen.getByRole('button'))

    await waitFor(() => expect(mockedSaveGrade).toHaveBeenCalled())
    expect(mockedSaveGrade).toHaveBeenCalledWith(10, {
      student_id: 7,
      value: 85,
      feedback: 'great work',
    })
    await waitFor(() => expect(toast.success).toHaveBeenCalled())
    expect(onSaved).toHaveBeenCalled()
  })

  describe('error mapping', () => {
    const cases: Array<{ status: number; key: string }> = [
      { status: 409, key: 'errors.alreadyGraded' },
      { status: 422, key: 'errors.invalidValue' },
      { status: 403, key: 'errors.forbidden' },
    ]

    it.each(cases)('maps HTTP $status to $key toast', async ({ status, key }) => {
      const axiosErr = Object.assign(new Error('http'), {
        isAxiosError: true,
        response: { status, data: {} },
        toJSON: () => ({}),
      })
      // axios.isAxiosError relies on a marker; force-true via mock
      jest.spyOn(axios, 'isAxiosError').mockReturnValue(true)
      mockedSaveGrade.mockRejectedValueOnce(axiosErr)

      const user = userEvent.setup()
      render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

      await user.type(screen.getByLabelText(/valueLabel/), '85')
      await user.click(screen.getByRole('button'))

      await waitFor(() => expect(toast.error).toHaveBeenCalled())
      expect((toast.error as jest.Mock).mock.calls[0][0]).toContain(key)
    })

    it('falls through to generic on non-axios error', async () => {
      jest.spyOn(axios, 'isAxiosError').mockReturnValue(false)
      mockedSaveGrade.mockRejectedValueOnce(new Error('network'))

      const user = userEvent.setup()
      render(<GradeForm assignmentId={10} maxScore={100} submission={pendingSubmission} />)

      await user.type(screen.getByLabelText(/valueLabel/), '85')
      await user.click(screen.getByRole('button'))

      await waitFor(() => expect(toast.error).toHaveBeenCalled())
      expect((toast.error as jest.Mock).mock.calls[0][0]).toContain('errors.generic')
    })
  })

  it('disables inputs and button when submission is already graded', () => {
    render(<GradeForm assignmentId={10} maxScore={100} submission={gradedSubmission} />)

    const valueInput = screen.getByLabelText(/valueLabel/) as HTMLInputElement
    const feedbackInput = screen.getByLabelText(/feedbackLabel/) as HTMLInputElement
    const button = screen.getByRole('button') as HTMLButtonElement

    expect(valueInput).toBeDisabled()
    expect(feedbackInput).toBeDisabled()
    expect(button).toBeDisabled()
  })
})
