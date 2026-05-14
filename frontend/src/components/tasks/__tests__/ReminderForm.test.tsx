import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { ReminderForm } from '../ReminderForm'

describe('ReminderForm', () => {
  it('renders reminder_type select with 4 options + minutes_before input + save/cancel', () => {
    render(<ReminderForm onSubmit={jest.fn()} onCancel={jest.fn()} />)

    expect(screen.getByLabelText(/reminderTypeLabel/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/minutesBeforeLabel/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /taskReminders\.save|save/i })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /taskReminders\.cancel|cancel/i })
    ).toBeInTheDocument()

    const select = screen.getByLabelText(/reminderTypeLabel/i) as HTMLSelectElement
    const optionValues = Array.from(select.options).map((o) => o.value)
    expect(optionValues).toEqual(expect.arrayContaining(['email', 'push', 'in_app', 'telegram']))
    expect(optionValues).toHaveLength(4)
  })

  it('defaults reminder_type to "telegram" and minutes_before to 60', () => {
    render(<ReminderForm onSubmit={jest.fn()} onCancel={jest.fn()} />)
    expect((screen.getByLabelText(/reminderTypeLabel/i) as HTMLSelectElement).value).toBe(
      'telegram'
    )
    expect((screen.getByLabelText(/minutesBeforeLabel/i) as HTMLInputElement).value).toBe('60')
  })

  it('calls onSubmit with selected reminder_type and parsed minutes_before', async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()
    render(<ReminderForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.selectOptions(screen.getByLabelText(/reminderTypeLabel/i), 'email')
    const minutesInput = screen.getByLabelText(/minutesBeforeLabel/i)
    await user.clear(minutesInput)
    await user.type(minutesInput, '120')

    await user.click(screen.getByRole('button', { name: /taskReminders\.save|save/i }))

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1))
    expect(onSubmit).toHaveBeenCalledWith({ reminder_type: 'email', minutes_before: 120 })
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const onCancel = jest.fn()
    const user = userEvent.setup()
    render(<ReminderForm onSubmit={jest.fn()} onCancel={onCancel} />)

    await user.click(screen.getByRole('button', { name: /taskReminders\.cancel|cancel/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not submit when minutes_before is 0 or negative', async () => {
    const onSubmit = jest.fn()
    const user = userEvent.setup()
    render(<ReminderForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    const minutesInput = screen.getByLabelText(/minutesBeforeLabel/i)
    await user.clear(minutesInput)
    await user.type(minutesInput, '0')

    await user.click(screen.getByRole('button', { name: /taskReminders\.save|save/i }))

    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('disables submit button while submitting prop is true', () => {
    render(<ReminderForm onSubmit={jest.fn()} onCancel={jest.fn()} submitting />)
    expect(screen.getByRole('button', { name: /taskReminders\.save|save/i })).toBeDisabled()
  })
})
