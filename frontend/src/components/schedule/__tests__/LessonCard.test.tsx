import { render, screen } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { LessonCard } from '../LessonCard'
import type { Lesson } from '@/types/schedule'

const baseLesson: Lesson = {
  id: 1,
  semester_id: 1,
  discipline_id: 10,
  lesson_type_id: 2,
  teacher_id: 5,
  group_id: 3,
  classroom_id: 7,
  day_of_week: 1,
  time_start: '09:00',
  time_end: '10:30',
  week_type: 'all',
  date_start: '2026-02-01',
  date_end: '2026-06-30',
  is_cancelled: false,
  created_at: '2026-01-15T10:00:00Z',
  updated_at: '2026-01-15T10:00:00Z',
  discipline: { id: 10, name: 'Математический анализ', code: 'MA101' },
  lesson_type: { id: 2, name: 'Лекция', short_name: 'ЛК', color: '#3b82f6' },
  teacher: { id: 5, name: 'Иванов И.И.', email: 'ivanov@example.com' },
  classroom: { id: 7, building: 'А', number: '305', capacity: 100, is_available: true },
  group: { id: 3, name: 'ИВТ-21', course: 2 },
}

describe('LessonCard', () => {
  it('renders discipline name', () => {
    render(<LessonCard lesson={baseLesson} />)
    expect(screen.getByTestId('lesson-discipline')).toHaveTextContent('Математический анализ')
  })

  it('renders lesson type badge', () => {
    render(<LessonCard lesson={baseLesson} />)
    expect(screen.getByTestId('lesson-type-badge')).toHaveTextContent('ЛК')
  })

  it('renders teacher name', () => {
    render(<LessonCard lesson={baseLesson} />)
    expect(screen.getByTestId('lesson-teacher')).toHaveTextContent('Иванов И.И.')
  })

  it('renders classroom building and number', () => {
    render(<LessonCard lesson={baseLesson} />)
    expect(screen.getByTestId('lesson-classroom')).toHaveTextContent('А-305')
  })

  it('calls onClick when clicked', async () => {
    const onClick = jest.fn()
    const user = userEvent.setup()
    render(<LessonCard lesson={baseLesson} onClick={onClick} />)
    await user.click(screen.getByTestId('lesson-card'))
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('shows cancelled state', () => {
    render(<LessonCard lesson={{ ...baseLesson, is_cancelled: true }} />)
    expect(screen.getByText('lesson.cancelled')).toBeInTheDocument()
  })

  it('renders fallback discipline id when discipline is undefined', () => {
    render(<LessonCard lesson={{ ...baseLesson, discipline: undefined }} />)
    expect(screen.getByTestId('lesson-discipline')).toHaveTextContent('#10')
  })

  it('does not render teacher when teacher is undefined', () => {
    render(<LessonCard lesson={{ ...baseLesson, teacher: undefined }} />)
    expect(screen.queryByTestId('lesson-teacher')).not.toBeInTheDocument()
  })

  it('does not render classroom when classroom is undefined', () => {
    render(<LessonCard lesson={{ ...baseLesson, classroom: undefined }} />)
    expect(screen.queryByTestId('lesson-classroom')).not.toBeInTheDocument()
  })
})
