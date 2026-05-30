import { render, screen } from '@/test-utils'

const mockReplace = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: mockReplace, push: jest.fn() }),
  useParams: () => ({}),
}))

const mockUseAuthCheck = jest.fn()
jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: () => mockUseAuthCheck(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseMinobrnaukiOrders = jest.fn()
jest.mock('@/hooks/useMinobrnaukiOrders', () => ({
  useMinobrnaukiOrders: (filter?: Record<string, unknown>, opts?: { enabled?: boolean }) =>
    mockUseMinobrnaukiOrders(filter, opts),
  useMinobrnaukiOrder: jest.fn(),
  recordMinobrnaukiOrder: jest.fn(),
  pickMinobrnaukiOrderErrorKey: jest.fn(() => 'generic'),
}))

import MinobrnaukiOrdersPage from '../page'
import type { MinobrnaukiOrderSummary } from '@/types/minobrnaukiOrder'

const methodistAuth = {
  user: { id: 5, role: 'methodist' as const },
  isAuthenticated: true,
  isLoading: false,
}

const sample = (overrides: Partial<MinobrnaukiOrderSummary> = {}): MinobrnaukiOrderSummary => ({
  id: 1,
  order_number: '№ 1234',
  title: 'О внесении изменений в ФГОС ВО',
  published_at: '2026-03-01',
  change_scope: 'major',
  uploaded_by: 5,
  created_at: '2026-03-02T10:00:00Z',
  ...overrides,
})

beforeEach(() => {
  mockReplace.mockClear()
  mockUseAuthCheck.mockReturnValue(methodistAuth)
  mockUseMinobrnaukiOrders.mockReturnValue({
    items: [],
    total: 0,
    isLoading: false,
    error: undefined,
  })
})

describe('MinobrnaukiOrdersPage', () => {
  it('redirects student → /forbidden (orders are non-student per ADR-11)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrdersPage />)
    expect(mockReplace).toHaveBeenCalledWith('/forbidden')
  })

  it.each(['system_admin', 'methodist', 'academic_secretary', 'teacher'])(
    'does NOT redirect %s and fetches the orders',
    (role) => {
      mockUseAuthCheck.mockReturnValue({
        user: { id: 5, role: role as 'methodist' },
        isAuthenticated: true,
        isLoading: false,
      })
      render(<MinobrnaukiOrdersPage />)
      expect(mockReplace).not.toHaveBeenCalled()
      // fetch enabled for staff roles
      const lastCall = mockUseMinobrnaukiOrders.mock.calls.at(-1)
      expect(lastCall?.[1]?.enabled).toBe(true)
    }
  )

  it('does not fetch (enabled=false) for a student before redirect', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 7, role: 'student' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrdersPage />)
    const lastCall = mockUseMinobrnaukiOrders.mock.calls.at(-1)
    expect(lastCall?.[1]?.enabled).toBe(false)
  })

  it('renders the order rows (order number + title)', () => {
    mockUseMinobrnaukiOrders.mockReturnValue({
      items: [sample(), sample({ id: 2, order_number: '№ 99', title: 'Приказ о ФОС' })],
      total: 2,
      isLoading: false,
      error: undefined,
    })
    render(<MinobrnaukiOrdersPage />)
    expect(screen.getByText('О внесении изменений в ФГОС ВО')).toBeInTheDocument()
    expect(screen.getByText('Приказ о ФОС')).toBeInTheDocument()
    expect(screen.getByText('№ 1234')).toBeInTheDocument()
  })

  it('shows the empty state when there are no orders', () => {
    render(<MinobrnaukiOrdersPage />)
    // next-intl is mocked to identity (key → key), so assert on the key.
    expect(screen.getByText('empty.title')).toBeInTheDocument()
    expect(screen.getByText('empty.description')).toBeInTheDocument()
  })

  it('shows the error state when the list fails to load', () => {
    mockUseMinobrnaukiOrders.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: new Error('network'),
    })
    render(<MinobrnaukiOrdersPage />)
    expect(screen.getByText('loadFailed')).toBeInTheDocument()
  })

  it.each(['system_admin', 'methodist', 'academic_secretary'])(
    'shows the Record button for %s (record role)',
    (role) => {
      mockUseAuthCheck.mockReturnValue({
        user: { id: 5, role: role as 'methodist' },
        isAuthenticated: true,
        isLoading: false,
      })
      render(<MinobrnaukiOrdersPage />)
      expect(screen.getByText('recordButton')).toBeInTheDocument()
    }
  )

  it('hides the Record button for a teacher (view-only)', () => {
    mockUseAuthCheck.mockReturnValue({
      user: { id: 8, role: 'teacher' as const },
      isAuthenticated: true,
      isLoading: false,
    })
    render(<MinobrnaukiOrdersPage />)
    expect(screen.queryByText('recordButton')).not.toBeInTheDocument()
  })
})
