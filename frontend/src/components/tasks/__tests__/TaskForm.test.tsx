import { render, screen, waitFor } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { TaskForm } from '../TaskForm'
import type { Task } from '@/types/tasks'

const baseTask: Task = {
  id: 1,
  title: 'Old title',
  description: 'Old desc',
  status: 'new',
  priority: 'normal',
  author_id: 1,
  progress: 0,
  is_overdue: false,
  created_at: '2026-04-25T10:00:00Z',
  updated_at: '2026-04-25T10:00:00Z',
}

describe('TaskForm', () => {
  it('renders title and description inputs and priority select', () => {
    render(<TaskForm onSubmit={jest.fn()} onCancel={jest.fn()} />)
    expect(screen.getByLabelText(/titleLabel|title/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/descriptionLabel|description/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/priorityLabel|priority/i)).toBeInTheDocument()
  })

  it('pre-fills fields when editing existing task', () => {
    render(<TaskForm task={baseTask} onSubmit={jest.fn()} onCancel={jest.fn()} />)
    expect((screen.getByLabelText(/titleLabel|title/i) as HTMLInputElement).value).toBe('Old title')
    expect((screen.getByLabelText(/descriptionLabel|description/i) as HTMLTextAreaElement).value).toBe('Old desc')
  })

  it('calls onSubmit with form data when submitted', async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined)
    const user = userEvent.setup()
    render(<TaskForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.type(screen.getByLabelText(/titleLabel|title/i), 'New task')
    await user.selectOptions(screen.getByLabelText(/priorityLabel|priority/i), 'high')
    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalled()
    })
    const submitted = onSubmit.mock.calls[0][0]
    expect(submitted.title).toBe('New task')
    expect(submitted.priority).toBe('high')
  })

  it('does not submit when title is empty', async () => {
    const onSubmit = jest.fn()
    const user = userEvent.setup()
    render(<TaskForm onSubmit={onSubmit} onCancel={jest.fn()} />)

    await user.click(screen.getByRole('button', { name: /save|сохран/i }))

    expect(onSubmit).not.toHaveBeenCalled()
    expect(screen.getByText(/titleRequired|обязательно|required/i)).toBeInTheDocument()
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const onCancel = jest.fn()
    const user = userEvent.setup()
    render(<TaskForm onSubmit={jest.fn()} onCancel={onCancel} />)

    await user.click(screen.getByRole('button', { name: /cancel|отмен/i }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })
})
