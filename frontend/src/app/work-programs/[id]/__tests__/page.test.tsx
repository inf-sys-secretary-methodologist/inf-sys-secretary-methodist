import { render, screen, fireEvent } from '@/test-utils'

const mockReplace = jest.fn()
const mockUseParams = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => mockUseParams(),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseWorkProgram = jest.fn()
jest.mock('@/hooks/useWorkPrograms', () => ({
  useWorkPrograms: jest.fn(),
  useWorkProgram: (id: number | null, opts?: { enabled?: boolean }) => mockUseWorkProgram(id, opts),
}))

import WorkProgramDetailPage from '../page'
import type { WorkProgram } from '@/types/workProgram'

const teacherAuth = {
  user: { id: 5, role: 'teacher' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<WorkProgram> = {}): WorkProgram => ({
  id: 7,
  discipline_id: 10,
  specialty_code: '09.03.01',
  applicable_from_year: 2026,
  title: 'Базы данных',
  annotation: 'Аннотация курса БД',
  status: 'approved',
  author_id: 5,
  approver_id: 3,
  approved_at: '2026-05-20T10:00:00Z',
  reject_reason: '',
  version: 2,
  created_at: '2026-05-01T08:00:00Z',
  updated_at: '2026-05-20T10:00:00Z',
  goals: [{ id: 1, text: 'Освоить реляционную модель', order_index: 0 }],
  competences: [{ id: 1, code: 'ПК-1', type: 'pk', description: 'Проектирует БД' }],
  topics: [
    {
      id: 1,
      kind: 'lecture',
      title: 'Нормализация',
      hours: 4,
      week_number: 3,
      learning_outcomes: 'Знает НФ',
      order_index: 0,
    },
  ],
  assessments: [
    {
      id: 1,
      type: 'final',
      description: 'Экзамен',
      max_score: 40,
      example_questions: ['Что такое 3НФ?'],
    },
  ],
  references: [
    { id: 1, kind: 'main', citation: 'Дейт К. Введение в системы БД', year: 2020, order_index: 0 },
  ],
  revisions: [],
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockUseParams.mockReturnValue({ id: '7' })
  mockUseAuthCheck.mockReturnValue(teacherAuth)
  mockUseWorkProgram.mockReturnValue({
    workProgram: sample(),
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })
})

describe('WorkProgramDetailPage', () => {
  it('renders the title and status pill', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Базы данных')).toBeInTheDocument()
    expect(screen.getByText('card.status.approved')).toBeInTheDocument()
  })

  it('does NOT redirect students (273-ФЗ open access)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 9, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramDetailPage />)
    expect(mockReplace).not.toHaveBeenCalled()
  })

  it('fetches for students (enabled=true)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 9, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramDetailPage />)
    const lastCall = mockUseWorkProgram.mock.calls.at(-1)
    expect(lastCall?.[0]).toBe(7)
    expect(lastCall?.[1]).toEqual({ enabled: true })
  })

  it('does NOT fetch while auth is loading', () => {
    mockUseAuthCheck.mockReturnValue({ user: null, isAuthenticated: false, isLoading: true })
    render(<WorkProgramDetailPage />)
    const lastCall = mockUseWorkProgram.mock.calls.at(-1)
    expect(lastCall?.[1]).toEqual({ enabled: false })
  })

  it('renders the annotation', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Аннотация курса БД')).toBeInTheDocument()
  })

  it('renders a goal', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Освоить реляционную модель')).toBeInTheDocument()
  })

  it('renders a competence with its code and type label', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('ПК-1')).toBeInTheDocument()
    expect(screen.getByText('Проектирует БД')).toBeInTheDocument()
    expect(screen.getByText('detail.competenceType.pk')).toBeInTheDocument()
  })

  it('renders a topic with its kind label and title', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Нормализация')).toBeInTheDocument()
    expect(screen.getByText('detail.topicKind.lecture')).toBeInTheDocument()
  })

  it('renders an assessment with its type label', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Экзамен')).toBeInTheDocument()
    expect(screen.getByText('detail.assessmentType.final')).toBeInTheDocument()
  })

  it('renders a reference with its kind label', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Дейт К. Введение в системы БД')).toBeInTheDocument()
    expect(screen.getByText('detail.referenceType.main')).toBeInTheDocument()
  })

  it('renders a revision with change-type + collapsed status labels (not the raw wire key)', () => {
    mockUseWorkProgram.mockReturnValue({
      workProgram: sample({
        revisions: [
          {
            id: 1,
            revision_number: 1,
            change_type: 'hours',
            change_summary: 'Часы скорректированы',
            status: 'pending_approval',
            author_id: 5,
            created_at: '2026-05-21T10:00:00Z',
            updated_at: '2026-05-21T10:00:00Z',
          },
        ],
      }),
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('Часы скорректированы')).toBeInTheDocument()
    expect(screen.getByText('detail.revisionChangeType.hours')).toBeInTheDocument()
    // pending_approval must collapse to the short key, not render the wire value.
    expect(screen.getByText('detail.revisionStatus.pending')).toBeInTheDocument()
    expect(screen.queryByText('detail.revisionStatus.pending_approval')).not.toBeInTheDocument()
  })

  it('shows section empty text when a collection is empty', () => {
    mockUseWorkProgram.mockReturnValue({
      workProgram: sample({
        goals: [],
        competences: [],
        topics: [],
        assessments: [],
        references: [],
      }),
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramDetailPage />)
    // At least one section empty label rendered (t-mock verbatim key).
    expect(screen.getAllByText('detail.sections.empty').length).toBeGreaterThan(0)
  })

  it('shows loading state', () => {
    mockUseWorkProgram.mockReturnValue({
      workProgram: undefined,
      isLoading: true,
      error: undefined,
    })
    render(<WorkProgramDetailPage />)
    expect(screen.queryByText('Базы данных')).not.toBeInTheDocument()
  })

  it('shows error / not-found block when the hook errors', () => {
    mockUseWorkProgram.mockReturnValue({
      workProgram: undefined,
      isLoading: false,
      error: new Error('boom'),
    })
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('detail.loadFailed')).toBeInTheDocument()
  })

  it('shows not-found block for an invalid id', () => {
    mockUseParams.mockReturnValue({ id: 'abc' })
    mockUseWorkProgram.mockReturnValue({
      workProgram: undefined,
      isLoading: false,
      error: undefined,
    })
    render(<WorkProgramDetailPage />)
    expect(screen.getByText('detail.notFound')).toBeInTheDocument()
  })
})

describe('WorkProgramDetailPage — draft author actions (8d-1)', () => {
  const draftAuthor = {
    user: { id: 5, role: 'teacher' as const },
    isAuthenticated: true,
    isLoading: false,
  }

  const draftWp = (status: WorkProgram['status'] = 'draft') => ({
    workProgram: sample({ status }),
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })

  beforeEach(() => {
    mockUseParams.mockReturnValue({ id: '7' })
    mockUseAuthCheck.mockReturnValue(draftAuthor)
    mockUseWorkProgram.mockReturnValue(draftWp('draft'))
  })

  it('shows submit + discard actions for an author on a draft', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByRole('button', { name: 'detail.actions.submit' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'detail.actions.discard' })).toBeInTheDocument()
  })

  it('opens the submit dialog when the submit action is clicked', () => {
    render(<WorkProgramDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.submit' }))
    expect(screen.getByText('submitDialog.title')).toBeInTheDocument()
  })

  it('opens the discard dialog when the discard action is clicked', () => {
    render(<WorkProgramDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.discard' }))
    expect(screen.getByText('discardDialog.title')).toBeInTheDocument()
  })

  it('hides draft actions for a student (cannot author РПД)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 9, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramDetailPage />)
    expect(screen.queryByRole('button', { name: 'detail.actions.submit' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'detail.actions.discard' })).not.toBeInTheDocument()
  })

  it('hides draft actions when the programme is not a draft (approved)', () => {
    mockUseWorkProgram.mockReturnValue(draftWp('approved'))
    render(<WorkProgramDetailPage />)
    expect(screen.queryByRole('button', { name: 'detail.actions.submit' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'detail.actions.discard' })).not.toBeInTheDocument()
  })
})

describe('WorkProgramDetailPage — approver actions (8d-2)', () => {
  const methodist = {
    user: { id: 3, role: 'methodist' as const },
    isAuthenticated: true,
    isLoading: false,
  }

  const wpInState = (status: WorkProgram['status']) => ({
    workProgram: sample({ status }),
    isLoading: false,
    error: undefined,
    mutate: jest.fn(),
  })

  beforeEach(() => {
    mockUseParams.mockReturnValue({ id: '7' })
    mockUseAuthCheck.mockReturnValue(methodist)
    mockUseWorkProgram.mockReturnValue(wpInState('pending_approval'))
  })

  it('shows approve + reject actions for an approver on a pending programme', () => {
    render(<WorkProgramDetailPage />)
    expect(screen.getByRole('button', { name: 'detail.actions.approve' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'detail.actions.reject' })).toBeInTheDocument()
  })

  it('opens the approve dialog when the approve action is clicked', () => {
    render(<WorkProgramDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.approve' }))
    expect(screen.getByText('approveDialog.title')).toBeInTheDocument()
  })

  it('opens the reject dialog when the reject action is clicked', () => {
    render(<WorkProgramDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'detail.actions.reject' }))
    expect(screen.getByText('rejectDialog.title')).toBeInTheDocument()
  })

  it('hides approver actions for a non-approver (teacher author)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 5, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<WorkProgramDetailPage />)
    expect(screen.queryByRole('button', { name: 'detail.actions.approve' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'detail.actions.reject' })).not.toBeInTheDocument()
  })

  it('hides approver actions when the programme is not pending (draft)', () => {
    mockUseWorkProgram.mockReturnValue(wpInState('draft'))
    render(<WorkProgramDetailPage />)
    expect(screen.queryByRole('button', { name: 'detail.actions.approve' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'detail.actions.reject' })).not.toBeInTheDocument()
  })
})
