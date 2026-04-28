import { render, screen, within } from '@/test-utils'
import userEvent from '@testing-library/user-event'
import { TimetableGrid } from '../TimetableGrid'
import type { Lesson } from '@/types/schedule'

const makeMockLesson = (overrides: Partial<Lesson> = {}): Lesson => ({
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
  discipline: { id: 10, name: 'Математика', code: 'MA101' },
  lesson_type: { id: 2, name: 'Лекция', short_name: 'ЛК', color: '#3b82f6' },
  teacher: { id: 5, name: 'Иванов И.И.', email: 'ivanov@example.com' },
  classroom: { id: 7, building: 'А', number: '305', capacity: 100, is_available: true },
  group: { id: 3, name: 'ИВТ-21', course: 2 },
  ...overrides,
})

describe('TimetableGrid', () => {
  it('renders the grid with day headers', () => {
    render(<TimetableGrid lessons={[]} canEdit={false} />)
    expect(screen.getByText('days.monday')).toBeInTheDocument()
    expect(screen.getByText('days.tuesday')).toBeInTheDocument()
    expect(screen.getByText('days.wednesday')).toBeInTheDocument()
    expect(screen.getByText('days.thursday')).toBeInTheDocument()
    expect(screen.getByText('days.friday')).toBeInTheDocument()
    expect(screen.getByText('days.saturday')).toBeInTheDocument()
  })

  it('renders time slots', () => {
    render(<TimetableGrid lessons={[]} canEdit={false} />)
    expect(screen.getByText('09:00')).toBeInTheDocument()
    expect(screen.getByText('10:45')).toBeInTheDocument()
    expect(screen.getByText('13:00')).toBeInTheDocument()
    expect(screen.getByText('14:45')).toBeInTheDocument()
    expect(screen.getByText('16:30')).toBeInTheDocument()
  })

  it('places a lesson in the correct cell (Monday 09:00)', () => {
    const lesson = makeMockLesson({ id: 1, day_of_week: 1, time_start: '09:00' })
    render(<TimetableGrid lessons={[lesson]} canEdit={false} />)

    const cell = screen.getByTestId('cell-1-09:00')
    expect(within(cell).getByTestId('lesson-card')).toBeInTheDocument()
    expect(within(cell).getByText('Математика')).toBeInTheDocument()
  })

  it('places a lesson in Wednesday 13:00 cell', () => {
    const lesson = makeMockLesson({ id: 2, day_of_week: 3, time_start: '13:00' })
    render(<TimetableGrid lessons={[lesson]} canEdit={false} />)

    const cell = screen.getByTestId('cell-3-13:00')
    expect(within(cell).getByTestId('lesson-card')).toBeInTheDocument()
  })

  it('does not place lesson in wrong cells', () => {
    const lesson = makeMockLesson({ id: 1, day_of_week: 1, time_start: '09:00' })
    render(<TimetableGrid lessons={[lesson]} canEdit={false} />)

    // Tuesday 09:00 should be empty
    const tuesdayCell = screen.getByTestId('cell-2-09:00')
    expect(within(tuesdayCell).queryByTestId('lesson-card')).not.toBeInTheDocument()

    // Monday 10:45 should be empty
    const laterCell = screen.getByTestId('cell-1-10:45')
    expect(within(laterCell).queryByTestId('lesson-card')).not.toBeInTheDocument()
  })

  it('places multiple lessons in the same cell', () => {
    const lessons = [
      makeMockLesson({ id: 1, day_of_week: 2, time_start: '10:45', discipline: { id: 10, name: 'Физика' } }),
      makeMockLesson({ id: 2, day_of_week: 2, time_start: '10:45', discipline: { id: 11, name: 'Химия' } }),
    ]
    render(<TimetableGrid lessons={lessons} canEdit={false} />)

    const cell = screen.getByTestId('cell-2-10:45')
    const cards = within(cell).getAllByTestId('lesson-card')
    expect(cards).toHaveLength(2)
  })

  it('calls onLessonClick when a lesson card is clicked', async () => {
    const onLessonClick = jest.fn()
    const lesson = makeMockLesson({ id: 42 })
    const user = userEvent.setup()

    render(<TimetableGrid lessons={[lesson]} canEdit={true} onLessonClick={onLessonClick} />)

    await user.click(screen.getByTestId('lesson-card'))
    expect(onLessonClick).toHaveBeenCalledTimes(1)
    expect(onLessonClick).toHaveBeenCalledWith(expect.objectContaining({ id: 42 }))
  })
})
