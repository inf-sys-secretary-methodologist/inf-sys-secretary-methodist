import { render, screen } from '@testing-library/react'

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush, replace: jest.fn() }),
  useParams: () => ({}),
}))

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => ({
    user: { id: 1, role: 'student' as const },
    isAuthenticated: true,
    isLoading: false,
  }),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseEvents = jest.fn()
jest.mock('@/hooks/useExtracurricularEvents', () => ({
  useExtracurricularEvents: (filter?: unknown) => mockUseEvents(filter),
}))

import ExtracurricularCalendarPage from '../page'

beforeEach(() => {
  mockPush.mockClear()
  mockUseEvents.mockReturnValue({
    events: [],
    total: 0,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('ExtracurricularCalendarPage', () => {
  it('renders the page title', () => {
    render(<ExtracurricularCalendarPage />)
    expect(screen.getByText('calendar')).toBeInTheDocument()
  })

  it('shows empty state when no events', () => {
    render(<ExtracurricularCalendarPage />)
    expect(screen.getByText('empty')).toBeInTheDocument()
  })

  it('renders events grouped under their start-month heading', () => {
    mockUseEvents.mockReturnValue({
      events: [
        {
          id: 1,
          title: 'June event',
          category: 'cultural',
          target_audience: 'all',
          status: 'published',
          location: 'Hall',
          start_at: '2026-06-15T18:00:00Z',
          end_at: '2026-06-15T21:00:00Z',
          max_capacity: 100,
          organizer_id: 5,
          participant_count: 10,
          version: 1,
          created_at: '2026-05-01T00:00:00Z',
          updated_at: '2026-05-01T00:00:00Z',
        },
        {
          id: 2,
          title: 'July event',
          category: 'sports',
          target_audience: 'students',
          status: 'published',
          location: 'Stadium',
          start_at: '2026-07-10T09:00:00Z',
          end_at: '2026-07-10T12:00:00Z',
          max_capacity: 50,
          organizer_id: 5,
          participant_count: 5,
          version: 1,
          created_at: '2026-05-01T00:00:00Z',
          updated_at: '2026-05-01T00:00:00Z',
        },
      ],
      total: 2,
      isLoading: false,
      error: undefined,
      mutate: jest.fn(),
    })
    render(<ExtracurricularCalendarPage />)
    expect(screen.getByText('June event')).toBeInTheDocument()
    expect(screen.getByText('July event')).toBeInTheDocument()
  })

  it('shows loading spinner when isLoading', () => {
    mockUseEvents.mockReturnValue({
      events: [],
      total: 0,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    const { container } = render(<ExtracurricularCalendarPage />)
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows error state on hook failure', () => {
    mockUseEvents.mockReturnValue({
      events: [],
      total: 0,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<ExtracurricularCalendarPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it('filters request to status=published (calendar view)', () => {
    render(<ExtracurricularCalendarPage />)
    const filter = mockUseEvents.mock.calls[0]?.[0]
    expect(filter).toEqual(expect.objectContaining({ status: 'published' }))
  })
})
