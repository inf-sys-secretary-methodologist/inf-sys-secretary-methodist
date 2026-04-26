import { render, screen } from '@/test-utils'

jest.mock('@/hooks/useAuth', () => ({
  useAuthCheck: jest.fn(),
}))

jest.mock('@/components/layout', () => ({
  AppLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

import AdminAppearancePage from '../page'

describe('AdminAppearancePage', () => {
  it('renders page title', () => {
    render(<AdminAppearancePage />)
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('appearance.title')
  })

  it('renders brand section', () => {
    render(<AdminAppearancePage />)
    expect(screen.getByText('appearance.brandTitle')).toBeInTheDocument()
  })
})
