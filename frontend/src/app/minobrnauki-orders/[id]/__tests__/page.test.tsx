import { render, screen } from '@/test-utils'

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
jest.mock('@/hooks/useMinobrnaukiOrders', () => ({
  useMinobrnaukiOrders: jest.fn(),
  useMinobrnaukiOrder: (id: number | null, opts?: { enabled?: boolean }) =>
    mockUseMinobrnaukiOrder(id, opts),
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
    const link10 = screen.getByRole('link', { name: /10/ })
    const link11 = screen.getByRole('link', { name: /11/ })
    expect(link10).toHaveAttribute('href', '/work-programs/10')
    expect(link11).toHaveAttribute('href', '/work-programs/11')
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
})
