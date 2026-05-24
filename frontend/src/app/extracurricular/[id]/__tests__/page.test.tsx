import { render, screen, fireEvent, waitFor } from '@testing-library/react'

const mockReplace = jest.fn()
const mockPush = jest.fn()
const mockBack = jest.fn()
const mockParams = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: mockPush, back: mockBack }),
  useParams: () => mockParams(),
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

const mockUseEvent = jest.fn()
const mockRegister = jest.fn()
const mockUnregister = jest.fn()
const mockDelete = jest.fn()
jest.mock('@/hooks/useExtracurricularEvents', () => ({
  useExtracurricularEvent: (id: number | null) => mockUseEvent(id),
  registerForExtracurricularEvent: (id: number) => mockRegister(id),
  unregisterFromExtracurricularEvent: (id: number) => mockUnregister(id),
  deleteExtracurricularEvent: (id: number) => mockDelete(id),
  pickExtracurricularErrorKey: () => 'generic',
}))

jest.mock('sonner', () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}))

import ExtracurricularEventDetailPage from '../page'

const studentAuth = {
  user: { id: 9, role: 'student' as const },
  isAuthenticated: true,
  isLoading: false,
}

const baseEvent = {
  id: 7,
  title: 'Spring Concert',
  description: 'Annual evening of music',
  category: 'cultural' as const,
  target_audience: 'all' as const,
  status: 'published' as const,
  location: 'Main hall',
  start_at: '2026-06-15T18:00:00Z',
  end_at: '2026-06-15T21:00:00Z',
  max_capacity: 200,
  organizer_id: 5,
  participants: [
    { user_id: 9, registered_at: '2026-05-20T10:00:00Z' },
    { user_id: 12, registered_at: '2026-05-21T10:00:00Z' },
  ],
  participant_count: 2,
  version: 3,
  created_at: '2026-05-20T10:00:00Z',
  updated_at: '2026-05-25T12:00:00Z',
}

beforeEach(() => {
  mockReplace.mockClear()
  mockPush.mockClear()
  mockBack.mockClear()
  mockParams.mockReturnValue({ id: '7' })
  mockUseAuthCheck.mockReturnValue(studentAuth)
  mockUserStore.mockReturnValue(studentAuth.user)
  mockUseEvent.mockReturnValue({
    event: baseEvent,
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('ExtracurricularEventDetailPage', () => {
  it('renders event title', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText('Spring Concert')).toBeInTheDocument()
  })

  it('renders event description', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText('Annual evening of music')).toBeInTheDocument()
  })

  it('renders location', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText(/Main hall/i)).toBeInTheDocument()
  })

  it('renders participant count / capacity', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText(/2 \/ 200/)).toBeInTheDocument()
  })

  it('shows loading spinner during fetch', () => {
    mockUseEvent.mockReturnValue({
      event: undefined,
      isLoading: true,
      error: undefined,
      mutate: jest.fn(),
    })
    const { container } = render(<ExtracurricularEventDetailPage />)
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })

  it('shows error state on fetch failure', () => {
    mockUseEvent.mockReturnValue({
      event: undefined,
      isLoading: false,
      error: new Error('boom'),
      mutate: jest.fn(),
    })
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it('shows Unregister button when current user is in participants', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText('unregister')).toBeInTheDocument()
  })

  it('shows Register button when current user not in participants', () => {
    mockUserStore.mockReturnValue({ id: 42, role: 'student' as const })
    mockUseAuthCheck.mockReturnValue({
      user: { id: 42, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<ExtracurricularEventDetailPage />)
    expect(screen.getByText('register')).toBeInTheDocument()
  })

  it('calls register mutation when Register clicked', async () => {
    mockUserStore.mockReturnValue({ id: 42, role: 'student' as const })
    mockUseAuthCheck.mockReturnValue({
      user: { id: 42, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    mockRegister.mockResolvedValue(undefined)
    render(<ExtracurricularEventDetailPage />)
    fireEvent.click(screen.getByText('register'))
    await waitFor(() => expect(mockRegister).toHaveBeenCalledWith(7))
  })

  it('lists participants for non-student roles (organizer view)', () => {
    mockUserStore.mockReturnValue({ id: 5, role: 'academic_secretary' as const })
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: 'academic_secretary' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<ExtracurricularEventDetailPage />)
    // Participant rows render some marker для each user_id.
    expect(screen.getByText(/user_id.*9/)).toBeInTheDocument()
    expect(screen.getByText(/user_id.*12/)).toBeInTheDocument()
  })

  it('hides participants list for student viewers', () => {
    render(<ExtracurricularEventDetailPage />)
    expect(screen.queryByText(/user_id.*9/)).not.toBeInTheDocument()
  })
})
