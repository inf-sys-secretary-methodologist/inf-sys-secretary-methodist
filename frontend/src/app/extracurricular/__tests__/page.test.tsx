import { render, screen } from '@testing-library/react'

const mockReplace = jest.fn()
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush }),
  useParams: () => ({}),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

const mockUserStore = jest.fn()
jest.mock('@/stores/authStore', () => ({
  useAuthStore: (selector: (s: { user: { id: number; role: string } | null }) => unknown) =>
    selector({ user: mockUserStore() }),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseEvents = jest.fn()
jest.mock('@/hooks/useExtracurricularEvents', () => ({
  useExtracurricularEvents: (filter?: unknown) => mockUseEvents(filter),
  useExtracurricularEvent: jest.fn(),
  createExtracurricularEvent: jest.fn(),
  updateExtracurricularEvent: jest.fn(),
  deleteExtracurricularEvent: jest.fn(),
  registerForExtracurricularEvent: jest.fn(),
  unregisterFromExtracurricularEvent: jest.fn(),
  pickExtracurricularErrorKey: () => 'generic',
}))

import ExtracurricularEventsPage from '../page'

const academicSecretaryAuth = {
  user: { id: 5, role: 'academic_secretary' as const },
  isAuthenticated: true,
  isLoading: false,
}

const studentAuth = {
  user: { id: 9, role: 'student' as const },
  isAuthenticated: true,
  isLoading: false,
}

beforeEach(() => {
  mockReplace.mockClear()
  mockPush.mockClear()
  mockUseAuthCheck.mockReturnValue(academicSecretaryAuth)
  mockUserStore.mockReturnValue(academicSecretaryAuth.user)
  mockUseEvents.mockReturnValue({
    events: [],
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('ExtracurricularEventsPage', () => {
  it('renders the page title', () => {
    render(<ExtracurricularEventsPage />)
    expect(screen.getByText('title')).toBeInTheDocument()
  })

  it('renders create button for non-student roles', () => {
    render(<ExtracurricularEventsPage />)
    expect(screen.getByText('create')).toBeInTheDocument()
  })

  it('hides create button for student role', () => {
    mockUseAuthCheck.mockReturnValue(studentAuth)
    mockUserStore.mockReturnValue(studentAuth.user)
    render(<ExtracurricularEventsPage />)
    expect(screen.queryByText('create')).not.toBeInTheDocument()
  })

  it('shows empty state when no events', () => {
    render(<ExtracurricularEventsPage />)
    expect(screen.getByText('empty')).toBeInTheDocument()
  })

  it('renders event cards when events present', () => {
    mockUseEvents.mockReturnValue({
      events: [
        {
          id: 1,
          title: 'Hackathon 2026',
          category: 'academic',
          target_audience: 'students',
          status: 'published',
          location: 'Lab 3',
          start_at: '2026-06-15T10:00:00Z',
          end_at: '2026-06-15T18:00:00Z',
          max_capacity: 30,
          organizer_id: 5,
          participant_count: 12,
          version: 1,
          created_at: '2026-05-20T10:00:00Z',
          updated_at: '2026-05-20T10:00:00Z',
        },
      ],
      total: 1,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<ExtracurricularEventsPage />)
    expect(screen.getByText('Hackathon 2026')).toBeInTheDocument()
  })

  it('shows loading spinner when isLoading', () => {
    mockUseEvents.mockReturnValue({
      events: [],
      total: 0,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    const { container } = render(<ExtracurricularEventsPage />)
    // Lucide Loader2 sets svg with animate-spin class.
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows error state on load failure', () => {
    mockUseEvents.mockReturnValue({
      events: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<ExtracurricularEventsPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })
})
