import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

jest.mock('@/hooks/useSchedule', () => ({
  useScheduleTimetable: () => ({ lessons: [], isLoading: false, error: null, mutate: jest.fn() }),
  useClassrooms: () => ({ classrooms: [], isLoading: false, error: null }),
  useStudentGroups: () => ({ groups: [], isLoading: false, error: null }),
  useSemesters: () => ({ semesters: [], isLoading: false, error: null }),
  useLessonTypes: () => ({ lessonTypes: [], isLoading: false, error: null }),
}))

jest.mock('@/stores/authStore', () => ({
  useAuthStore: (selector: (s: { user: null }) => unknown) => selector({ user: null }),
}))

import SchedulePage from '../page'

describe('SchedulePage', () => {
  it('renders schedule title', () => {
    render(<SchedulePage />)
    expect(screen.getByText('title')).toBeInTheDocument()
  })

  it('shows empty state when no lessons', () => {
    render(<SchedulePage />)
    expect(screen.getByText('empty')).toBeInTheDocument()
  })

  it('renders week type tabs', () => {
    render(<SchedulePage />)
    expect(screen.getByText('filters.allWeeks')).toBeInTheDocument()
    expect(screen.getByText('filters.oddWeeks')).toBeInTheDocument()
    expect(screen.getByText('filters.evenWeeks')).toBeInTheDocument()
  })
})
