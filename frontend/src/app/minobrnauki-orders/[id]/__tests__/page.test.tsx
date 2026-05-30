import { render, screen, fireEvent } from '@/test-utils'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => ({ id: '7' }),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseMinobrnaukiOrder = jest.fn()
const mockGenerate = jest.fn()
jest.mock('@/hooks/useMinobrnaukiOrders', () => ({
  useMinobrnaukiOrders: jest.fn(),
  useMinobrnaukiOrder: (id: number | null, opts?: { enabled?: boolean }) =>
    mockUseMinobrnaukiOrder(id, opts),
  generateOrderRevisions: (...args: unknown[]) => mockGenerate(...args),
  pickMinobrnaukiOrderErrorKey: () => 'generic',
}))

import MinobrnaukiOrderDetailPage from '../page'
import type { MinobrnaukiOrder } from '@/types/minobrnaukiOrder'

const methodistAuth = {
  user: { id: 5, role: 'methodist' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sampleOrder = (overrides: Partial<MinobrnaukiOrder> = {}): MinobrnaukiOrder => ({
  id: 7,
  order_number: '№ 1234',
  title: 'О внесении изменений в ФГОС ВО',
  published_at: '2026-03-01',
  change_scope: 'major',
  summary: 'Краткое содержание приказа',
  uploaded_by: 5,
  created_at: '2026-03-02T10:00:00Z',
  affected_work_program_ids: [10, 11],
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockUseAuthCheck.mockReturnValue(methodistAuth)
  mockUseMinobrnaukiOrder.mockReturnValue({
    order: sampleOrder(),
    isLoading: false,
    error: undefined,
  })
})

describe('MinobrnaukiOrderDetailPage', () => {
  it('redirects student → /forbidden', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrderDetailPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it('does not fetch for a student (enabled=false)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrderDetailPage />)
    const lastCall = mockUseMinobrnaukiOrder.mock.calls.at(-1)
    expect(lastCall?.[1]?.enabled).toBe(false)
  })

  it('renders the order metadata + summary', () => {
    render(<MinobrnaukiOrderDetailPage />)
    expect(screen.getByText('О внесении изменений в ФГОС ВО')).toBeInTheDocument()
    expect(screen.getByText('№ 1234')).toBeInTheDocument()
    expect(screen.getByText('Краткое содержание приказа')).toBeInTheDocument()
  })

  it('renders each affected РПД as a link to its detail page', () => {
    render(<MinobrnaukiOrderDetailPage />)
    // next-intl is mocked to identity (interpolation params ignored), so the
    // links share an accessible name — assert on href instead.
    const hrefs = screen.getAllByRole('link').map((a) => a.getAttribute('href'))
    expect(hrefs).toContain('/work-programs/10')
    expect(hrefs).toContain('/work-programs/11')
  })

  it('shows the affected-empty state when no РПД are affected', () => {
    mockUseMinobrnaukiOrder.mockReturnValue({
      order: sampleOrder({ affected_work_program_ids: [] }),
      isLoading: false,
      error: undefined,
    })
    render(<MinobrnaukiOrderDetailPage />)
    expect(screen.getByText('detail.affectedEmpty')).toBeInTheDocument()
  })

  it('shows the not-found state on error', () => {
    mockUseMinobrnaukiOrder.mockReturnValue({
      order: undefined,
      isLoading: false,
      error: new Error('404'),
    })
    render(<MinobrnaukiOrderDetailPage />)
    expect(screen.getByText('detail.notFound')).toBeInTheDocument()
  })

  it('shows the "Сгенерировать правки" button for a methodist', () => {
    render(<MinobrnaukiOrderDetailPage />)
    expect(screen.getByRole('button', { name: 'generateButton' })).toBeInTheDocument()
  })

  it('hides the generate button for a teacher (can view, cannot record)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 8, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrderDetailPage />)
    expect(screen.queryByRole('button', { name: 'generateButton' })).not.toBeInTheDocument()
  })

  it('opens the generate dialog when the button is clicked', () => {
    render(<MinobrnaukiOrderDetailPage />)
    fireEvent.click(screen.getByRole('button', { name: 'generateButton' }))
    expect(screen.getByText('generateDialog.title')).toBeInTheDocument()
  })
})
