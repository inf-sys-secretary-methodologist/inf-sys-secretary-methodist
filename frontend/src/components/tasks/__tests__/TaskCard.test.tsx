import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { TaskCard } from '../TaskCard'
import { TASK_STATUSES, TASK_PRIORITIES, type Task, type TaskStatus, type TaskPriority } from '@/types/tasks'

const baseTask: Task = {
  id: 1,
  title: 'Подготовить отчёт',
  description: 'Квартальный отчёт по успеваемости',
  status: 'in_progress',
  priority: 'high',
  author_id: 1,
  progress: 40,
  is_overdue: false,
  created_at: '2026-04-25T10:00:00Z',
  updated_at: '2026-04-25T10:00:00Z',
  due_date: '2026-05-01T00:00:00Z',
  assignee: { id: 5, name: 'Иванов И.И.', email: 'ivanov@example.com' },
}

describe('TaskCard', () => {
  it('renders task title and description', () => {
    render(<TaskCard task={baseTask} />)
    expect(screen.getByText('Подготовить отчёт')).toBeInTheDocument()
    expect(screen.getByText('Квартальный отчёт по успеваемости')).toBeInTheDocument()
  })

  it('shows priority and status badges', () => {
    render(<TaskCard task={baseTask} />)
    // Translation keys used as visible text since useTranslations returns the key in tests
    expect(screen.getByText(/priority\.high|high/i)).toBeInTheDocument()
    expect(screen.getByText(/status\.in_progress|in_progress/i)).toBeInTheDocument()
  })

  it('renders assignee name', () => {
    render(<TaskCard task={baseTask} />)
    expect(screen.getByText('Иванов И.И.')).toBeInTheDocument()
  })

  it('shows progress percentage', () => {
    render(<TaskCard task={baseTask} />)
    expect(screen.getByText(/40/)).toBeInTheDocument()
  })

  it('marks overdue task with overdue indicator', () => {
    render(<TaskCard task={{ ...baseTask, is_overdue: true }} />)
    expect(screen.getByTestId('task-overdue-indicator')).toBeInTheDocument()
  })

  it('does not show overdue indicator for non-overdue task', () => {
    render(<TaskCard task={baseTask} />)
    expect(screen.queryByTestId('task-overdue-indicator')).not.toBeInTheDocument()
  })

  it('calls onClick when card is clicked', async () => {
    const onClick = jest.fn()
    const user = userEvent.setup()
    render(<TaskCard task={baseTask} onClick={onClick} />)
    await user.click(screen.getByText('Подготовить отчёт'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  // Table-driven: every TaskStatus enum value renders without error
  // and produces a status badge containing the i18n key.
  it.each(TASK_STATUSES)('renders status badge for status=%s', (status: TaskStatus) => {
    render(<TaskCard task={{ ...baseTask, status }} />)
    expect(screen.getByText(`status.${status}`)).toBeInTheDocument()
  })

  // Table-driven: every TaskPriority enum value renders without error
  // and produces a priority badge.
  it.each(TASK_PRIORITIES)('renders priority badge for priority=%s', (priority: TaskPriority) => {
    render(<TaskCard task={{ ...baseTask, priority }} />)
    expect(screen.getByText(`priority.${priority}`)).toBeInTheDocument()
  })
})
